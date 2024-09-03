package main

import (
	"log"

	"github.com/Marattttt/portfolio/frontend/internal/handlers"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	handlers.SetupRoutes(e)

	e.Server.Addr = ":8080"

	log.Fatal(e.Server.ListenAndServe())
}
