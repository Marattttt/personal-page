package handlers

import (
	"fmt"
	"io"
	"net/url"

	"github.com/Marattttt/portfolio/frontend/internal/handlers/templates"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

type runRequest struct {
	Code string `schema:"code,required"`
	Lang string `schema:"lang,required"`
}

func (r *runRequest) fillFromUrlEncoded(urlEncoded string) error {
	values, err := url.ParseQuery(urlEncoded)
	if err != nil {
		return fmt.Errorf("parsing url-encoded: %w", err)
	}

	code := values.Get("code")
	if len(code) == 0 {
		return fmt.Errorf("property code is required")
	}
	r.Code = code

	lang := values.Get("lang")
	if len(lang) == 0 {
		return fmt.Errorf("property lang is required")
	}

	r.Lang = lang

	return nil
}

func HandleRun(gorunner GoRunner) func(c echo.Context) error {
	return func(c echo.Context) error {
		urlEncoded, err := io.ReadAll(c.Request().Body)
		defer c.Request().Body.Close()
		if err != nil {
			return fmt.Errorf("reading request body: %w", err)
		}

		var req runRequest
		if err := req.fillFromUrlEncoded(string(urlEncoded)); err != nil {
			return fmt.Errorf("parsing url encoded request: %w", err)
		}

		if req.Lang != "golang" {
			return fmt.Errorf("lang %s is not allowed", req.Lang)
		}

		resp, err := gorunner.Run(c.Request().Context(), req.Code)
		if err != nil {
			return fmt.Errorf("running code: %w", err)
		}

		writeView(
			c,
			templates.RunResult(
				string(resp.Sstdout),
				string(resp.Sstderr),
				resp.ExitCode,
				resp.ExecutionTime),
		)

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
