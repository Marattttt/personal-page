package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/Marattttt/personal-page-libs/userenv"
	"github.com/Marattttt/personal-page/gorunner/pkg/runtime"
	"github.com/joho/godotenv"
	"github.com/rabbitmq/amqp091-go"
)

type Req struct {
	Code string `json:"code"`
}

type Resp struct {
	Stdout   []byte        `json:"stdout"`
	Stderr   []byte        `json:"stderr"`
	TimeTook time.Duration `json:"timeTook"`
	ExitCode int           `json:"exitCode"`

	CorrelationID string `json:"-"`
}

func main() {
	appctx, appcancel := context.WithCancel(context.TODO())
	defer appcancel()

	if err := godotenv.Load(); err != nil {
		slog.Warn("Could not load from godotenv", slog.String("err", err.Error()))
	} else {
		slog.Info("Successfully godotenv")
	}

	conf, err := CreateConfig(appctx)
	checkFatal(err, "Could not create config")

	checkFatal(conf.Apply(), "Could not apply config")

	conn, err := amqp091.Dial(conf.MQ.URL())
	checkFatal(err, "Dialling "+conf.MQ.URL())

	sendmsg := make(chan Resp)

	go func() {
		consume(appctx, conf, conn, sendmsg)
		// Closing the sendmsg channel signals to finish reading from it and stop the producer
		// goroutine, which leads to all remaining messages being sent to mq before shutdown
		close(sendmsg)
	}()

	go func() {
		produce(appctx, conf, conn, sendmsg)
		slog.Info("Stopped message production")
		appcancel()
	}()

	<-appctx.Done()
	slog.Info("Shutting down application")
}

func checkFatal(err error, msg string) {
	if err != nil {
		slog.Error(msg, slog.String("err", err.Error()))
		os.Exit(1)
	}
}

func consume(ctx context.Context, conf *Config, conn *amqp091.Connection, send chan Resp) {
	ch, err := conn.Channel()
	// Cannot continue operationg on an error of such level
	checkFatal(err, "Obtaining a channel from MQ")

	q, err := ch.QueueDeclare(conf.MQ.RecvQ, true, false, false, false, nil)
	checkFatal(err, "Declaring receive queue")

	d, err := ch.ConsumeWithContext(ctx, q.Name, "", false, false, false, false, nil)
	checkFatal(err, "Creating a consume channel")

	runtimeLock := sync.Mutex{}
	run, err := createRuntime(*conf, &runtimeLock)
	checkFatal(err, "Cretaing runtime")

	for msg := range d {
		var req Req
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			msg.Reject(false)
			slog.Warn("Could not decode broker's message body", slog.String("err", err.Error()))
			continue
		}

		defer func() {
			if cause := recover(); cause != nil {
				if err, ok := cause.(error); ok {
					slog.Error("Recovered from panic in main/consume (error)", slog.String("err", err.Error()))
				} else {
					slog.Error("Recovered from panic in main/consume (not an error)", slog.Any("cause", cause))
				}
			}
		}()

		rex, err := run.Run(ctx, req.Code)
		if err != nil {
			msg.Reject(true)
			slog.Error("Could not execute code from mq", slog.String("err", err.Error()))
			continue
		}

		resp := Resp{
			Stdout:   rex.Stdout,
			Stderr:   rex.Stderr,
			ExitCode: rex.ExitCode,
			TimeTook: rex.TimeTook,

			CorrelationID: msg.CorrelationId,
		}

		send <- resp

		msg.Ack(false)
	}
}

type Runtime interface {
	Run(ctx context.Context, code string) (*runtime.RunResult, error)
}

// Function may panic due to invalid app configuration
func createRuntime(conf Config, runtimeLock sync.Locker) (Runtime, error) {
	// Run as same user
	if conf.Runtime.RunAs == nil {
		slog.Info("Creating same user environment")
		env := userenv.SameUserEnv{}

		if conf.Mode != "debug" {
			return nil, fmt.Errorf("Not specifying user to run the application as is not allowed outside of debug mode")
		}
		return runtime.NewRuntime(runtimeLock, conf.Runtime.Dir, env), nil
	}

	if conf.Runtime.RunAsPass != nil {
		slog.Warn("Password authentication for a user is not supported")
	}

	slog.Info("Creating environment for a different user", slog.String("runAs", *conf.Runtime.RunAs))
	diffUserEnv, err := userenv.NewDiffUserEnv(*conf.Runtime.RunAs, nil)
	if err != nil {
		return nil, err
	}

	return runtime.NewRuntime(runtimeLock, conf.Runtime.Dir, diffUserEnv), nil
}

func produce(ctx context.Context, conf *Config, conn *amqp091.Connection, sendCh chan Resp) {
	ch, err := conn.Channel()
	// Cannot continue operationg on an error of such level
	checkFatal(err, "Obtaining a channel from MQ")

	q, err := ch.QueueDeclare(conf.MQ.RespondQ, true, false, false, false, nil)
	checkFatal(err, "Declaring a response queue with MQ")

	for {
		select {
		case <-ctx.Done():
			slog.Warn("Message production context cancelled")
			return
		case r := <-sendCh:
			marshalled, err := json.Marshal(r)
			if err != nil {
				slog.Error("Could not marshall message", slog.String("err", err.Error()), slog.Any("val", r))
				continue
			}

			slog.Info("Producing message to mq", slog.Int("msgLen", len(marshalled)))

			err = ch.Publish("", q.Name, true, false, amqp091.Publishing{
				CorrelationId: r.CorrelationID,
				ContentType:   "application/json",
				Body:          marshalled,
			})
			if err != nil {
				slog.Error("Could not send a message to mq", slog.String("err", err.Error()))
			}
		}
	}
}
