package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/Marattttt/portfolio/gorunner/pkg/runtime"
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
	ctx := context.TODO()

	checkFatal(godotenv.Load(), "Loading environment with godotenv")

	conf, err := CreateConfig(ctx)
	checkFatal(err, "Could not create config")

	checkFatal(conf.Apply(), "Could not apply config")

	conn, err := amqp091.Dial(conf.MQ.URL())
	checkFatal(err, "Dialling "+conf.MQ.URL())

	sendmsg := make(chan Resp)

	go func() {
		consume(ctx, conf, conn, sendmsg)
		close(sendmsg)
	}()

	go func() {
		produce(ctx, conf, conn, sendmsg)
		slog.Info("Stopped message production")
	}()

	<-ctx.Done()
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

		run := createRuntime(conf.Runtime, &runtimeLock)
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

func createRuntime(conf RuntimeConfig, runtimeLock sync.Locker) Runtime {
	// Run as same user
	if conf.RunAs == nil {
		env := runtime.SameUserEnv{}

		return runtime.NewRuntime(runtimeLock, conf.Dir, env)
	}

	// App cannot continuie working if no runtime can be constructed
	err := fmt.Errorf("Could not construct runtime from config %+V", conf)
	panic(err)
}

func produce(ctx context.Context, conf *Config, conn *amqp091.Connection, sendCh chan Resp) {
	ch, err := conn.Channel()
	// Cannot continue operationg on an error of such level
	checkFatal(err, "Obtaining a channel from MQ")

	q, err := ch.QueueDeclare(conf.MQ.RespondQ, true, false, false, false, nil)
	checkFatal(err, "Declaring a response queue with MQ")

	for r := range sendCh {
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
