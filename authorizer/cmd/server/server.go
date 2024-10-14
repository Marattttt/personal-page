package main

import (
	"context"
	"log/slog"

	"github.com/Marattttt/personal-page/authorizer/pkg/config"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

func main() {
	ctx := context.TODO()

	conf, err := config.ConfigFromEnv(ctx)
	checkFail(err, "Parsing env vars to create config")

	dbConnn, err := sqlx.Connect("postgres", conf.DBConfig.PostgresURL)
	checkFail(err, "Connecting to db")

	e := echo.New()
	AddRoutes(e, conf, dbConnn)

	checkFail(e.Server.ListenAndServe(), "Unexpected server shutdown")
}

func checkFail(err error, msg string) {
	if err != nil {
		slog.Error(msg, slog.String("err", err.Error()))
	}
}
