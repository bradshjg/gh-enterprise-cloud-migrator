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

	sos := services.NewOauthService(sessionStore, services.SourceOAuthEndpoint)
	tos := services.NewOauthService(sessionStore, services.TargetOAuthEndpoint)
	sgs := services.NewGitHubService(sos)
	tgs := services.NewGitHubService(tos)
	ms := services.NewMigratorService(sgs, tgs)

	ghs := handlers.NewGitHubHandler(sos, sgs, true)
	ght := handlers.NewGitHubHandler(tos, tgs, false)
	mh := handlers.NewMigratorHandler(ms)

	e.GET("/", mh.IndexHandler)
	e.POST("/run", mh.StartRunHandler)
	e.GET("/run", mh.RunHandler)
	e.GET("/output", mh.OutputHandler)
	e.GET("/source-orgs", ghs.OrgsHandler)
	e.GET("/source-repos", ghs.ReposHandler)
	e.GET("/target-orgs", ght.OrgsHandler)
	e.GET("/github-source/login", ghs.OAuthHandler)
	e.GET("/github-source/callback", ghs.OAuthCallbackHandler)
	e.GET("/github-target/login", ght.OAuthHandler)
	e.GET("/github-target/callback", ght.OAuthCallbackHandler)
	e.GET("/*", handlers.RouteNotFoundHandler)

	e.Logger.Fatal(e.Start(":8080"))
}
