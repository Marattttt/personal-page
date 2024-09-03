package handlers

import (
	"github.com/Marattttt/portfolio/frontend/internal/handlers/templates"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func HandleIndex() func(c echo.Context) error {
	return func(c echo.Context) error {
		tpl := templates.Index()
		writeView(c, tpl)
		return nil
	}
}

func writeView(c echo.Context, tpl templ.Component) {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	tpl.Render(c.Request().Context(), c.Response().Writer)
}
