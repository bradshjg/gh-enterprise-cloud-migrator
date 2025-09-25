package main

import (
	"os"
	"time"

	"github.com/bradshjg/gh-enterprise-server-to-enterprise-cloud-migrator/handlers"
	migratorMiddleware "github.com/bradshjg/gh-enterprise-server-to-enterprise-cloud-migrator/middleware"
	"github.com/bradshjg/gh-enterprise-server-to-enterprise-cloud-migrator/services"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	e.Debug = os.Getenv("DEBUG") == "true"
	e.HideBanner = true
	e.HidePort = true
	e.DisableHTTP2 = true
	e.Server.WriteTimeout = 10 * time.Second
	e.Server.ReadTimeout = 10 * time.Second

	e.Use(migratorMiddleware.LoggingMiddleware())
	e.Use(migratorMiddleware.RequestLoggingMiddleware())
	sessionStore := migratorMiddleware.SessionStore()
	e.Use(session.Middleware(sessionStore))

	e.HTTPErrorHandler = handlers.HTTPErrorHandler

	e.Static("/static", "assets")

	ts := services.NewTokenService(sessionStore)
	gs := services.NewGitHubService(ts)
	ms := services.NewMigratorService(gs)

	th := handlers.NewTokenHandler(ts)
	gh := handlers.NewGitHubHandler(gs)
	mh := handlers.NewMigratorHandler(ms)

	e.GET("/", mh.IndexHandler)
	e.POST("/run", mh.StartRunHandler)
	e.GET("/run", mh.RunHandler)
	e.GET("/output", mh.OutputHandler)
	e.POST("/token", th.TokenHandler)
	e.GET("/orgs", gh.OrgsHandler)
	e.GET("/repos", gh.ReposHandler)
	e.GET("/*", handlers.RouteNotFoundHandler)

	e.Logger.Fatal(e.Start(":8080"))
}
