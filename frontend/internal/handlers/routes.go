package handlers

import "github.com/labstack/echo/v4"

func SetupRoutes(r echo.Router) {
	r.Add("GET", "/", HandleIndex())
}
