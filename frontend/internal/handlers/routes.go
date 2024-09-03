package handlers

import (
	"github.com/Marattttt/portfolio/frontend/static"
	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo) {
	e.Add("GET", "/", HandleIndex())
	e.StaticFS("/static", static.Get())
}
