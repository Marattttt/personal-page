package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Marattttt/personal-page/authorizer/internal/db"
	"github.com/Marattttt/personal-page/authorizer/pkg/config"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	ctx := context.TODO()

	checkFail(godotenv.Load(), "Loading .env file")

	conf, err := config.ConfigFromEnv(ctx)
	checkFail(err, "Parsing env vars to create config")

	dbConnn, err := sqlx.Connect("postgres", conf.PostgresURL)
	checkFail(err, "Connecting to db")

	checkFail(db.Migrate(dbConnn, conf.MigrationsSource), "Migrating db")

	e := echo.New()
	e.Server.Addr = fmt.Sprintf(":%d", conf.Port)
	AddRoutes(e, conf, dbConnn)

	checkFail(e.Server.ListenAndServe(), "Unexpected server shutdown")
}

func checkFail(err error, msg string) {
	if err != nil {
		slog.Error(msg, slog.String("err", err.Error()))
	}
}
