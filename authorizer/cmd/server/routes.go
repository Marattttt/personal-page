package main

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Marattttt/personal-page/authorizer/internal/auth"
	"github.com/Marattttt/personal-page/authorizer/internal/db"
	"github.com/Marattttt/personal-page/authorizer/pkg/config"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

func AddRoutes(e *echo.Echo, conf *config.Config, dbconn *sqlx.DB) {
	e.GET("/login/:id", handleLogin(dbconn, conf))
}

var reqId int

func handleLogin(dbConn *sqlx.DB, conf *config.Config) func(c echo.Context) error {
	return func(c echo.Context) error {
		logger := reqLogger()

		uid := intParam(c, "id")
		if uid == nil {
			return c.String(http.StatusBadRequest, "id path param is missing")
		}

		repo := db.NewUserRepo(dbConn, logger)

		user, err := repo.Get(c.Request().Context(), *uid)
		if err != nil {
			return c.String(http.StatusNotFound, "user no found")
		}

		access, refresh, err := auth.GeneratePair(*user, &conf.AuthConfig)
		if err != nil {
			logger.Error("Could not generate error", slog.String("err", err.Error()))
			return c.String(http.StatusInternalServerError, "could not generate token")
		}

		return c.JSON(
			http.StatusOK,
			map[string]string{
				"access":  *access,
				"refresh": *refresh,
			})

	}
}

type ValidateRequest struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

// func handleValidate(conf *config.Config) func(c echo.Context) error {
// 	return func(c echo.Context) error {
// 		logger := reqLogger()
//
// 		var req ValidateRequest
// 		if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
// 			logger.Error("Could not decode req body", slog.Any("err", err.Error()))
// 			return c.String(http.StatusBadRequest, "Could not decode body")
// 		}
// 		defer c.Request().Body.Close()
//
// 	}
// }

func intParam(c echo.Context, name string) *int {
	param := c.Param(name)
	i, err := strconv.Atoi(param)

	if err != nil {
		return nil
	}

	return &i
}

func reqLogger() *slog.Logger {
	logger := slog.Default().With("id", reqId)
	reqId++
	return logger
}
