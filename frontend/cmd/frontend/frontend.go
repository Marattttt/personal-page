package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/Marattttt/portfolio/frontend/internal/handlers"
	"github.com/Marattttt/portfolio/frontend/internal/runners"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("Could not load from godotenv", slog.String("err", err.Error()))
	} else {
		slog.Info("Successfully godotenv")
	}

	ctx := context.TODO()

	conf, err := createConfig(ctx)
	checkFatal(err, "Creating app config")

	mqConn, err := connectMQ(conf)
	checkFatal(err, "Connecting to mq")

	gorunner := runners.NewGoRunner(conf.Runners, mqConn)

	e := echo.New()
	handlers.SetupRoutes(e, gorunner)

	e.Server.Addr = ":8080"

	log.Fatal(e.Server.ListenAndServe())
}

func checkFatal(err error, msg string) {
	if err != nil {
		slog.Error(msg, slog.String("err", err.Error()))
		os.Exit(1)
	}
}

func connectMQ(conf *Config) (*amqp091.Connection, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s", conf.Runners.MqUser, conf.Runners.MqPass, conf.Runners.MQAddr)
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}
	return conn, nil

}
