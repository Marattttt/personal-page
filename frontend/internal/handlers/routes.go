package handlers

import (
	"context"

	"github.com/Marattttt/portfolio/frontend/internal/runners"
	"github.com/Marattttt/portfolio/frontend/static"
	"github.com/labstack/echo/v4"
)

type GoRunner interface {
	Run(context.Context, string) (*runners.RunResult, error)
}

type JsRunner interface {
	Run(context.Context, string) (*runners.RunResult, error)
}

func SetupRoutes(e *echo.Echo, gorunner GoRunner, jsrunner JsRunner) {
	e.Add("GET", "/", HandleIndex())
	e.Add("POST", "/run", HandleRun(gorunner, jsrunner))

	e.StaticFS("/static", static.Get())
}
