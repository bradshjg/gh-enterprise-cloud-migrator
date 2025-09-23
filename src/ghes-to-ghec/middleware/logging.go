package middleware

import (
	"context"
	"log/slog"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

func logValuesFunc(c echo.Context, v middleware.RequestLoggerValues) error {
	commonAttrs := []slog.Attr{
		slog.String("method", v.Method),
		slog.String("path", v.URIPath),
		slog.Int("status", v.Status),
		slog.Int("latency", int(v.Latency.Milliseconds())),
	}
	if v.Error == nil {
		logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
			commonAttrs...,
		)
	} else {
		errorAttrs := append(commonAttrs, slog.String("err", v.Error.Error()))
		logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
			errorAttrs...,
		)
	}
	return nil
}

func RequestLoggingMiddleware() echo.MiddlewareFunc {
	config := middleware.RequestLoggerConfig{
		LogStatus:     true,
		LogURIPath:    true,
		LogError:      true,
		LogLatency:    true,
		LogMethod:     true,
		HandleError:   true,
		LogValuesFunc: logValuesFunc,
	}
	return middleware.RequestLoggerWithConfig(config)
}

type SLoggerContext struct {
	echo.Context
}

func (c *SLoggerContext) SLogger() *slog.Logger {
	return logger
}

func LoggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sc := &SLoggerContext{c}
			return next(sc)
		}
	}
}
