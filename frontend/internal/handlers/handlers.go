package handlers

import (
	"fmt"
	"io"

	"github.com/Marattttt/portfolio/frontend/internal/handlers/templates"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func HandleRun() func(c echo.Context) error {
	return func(c echo.Context) error {
		req, err := io.ReadAll(c.Request().Body)
		defer c.Request().Body.Close()

		fmt.Println("received: ", req)
		if err != nil {
			return err
		}

		r := c.Response()
		r.Header().Add(echo.HeaderContentType, echo.MIMETextHTML)

		body := fmt.Sprintf("<p>%s</p>", string(req))
		r.Write([]byte(body))

		return nil
	}
}

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
