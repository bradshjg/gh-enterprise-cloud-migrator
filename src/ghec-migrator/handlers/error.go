package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func HTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	content := fmt.Sprintf("HTTP %d: %s", code, err)
	c.String(code, content)
}
