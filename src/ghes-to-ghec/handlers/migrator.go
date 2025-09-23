package handlers

import (
	"net/http"

	"github.com/bradshjg/gh-enterprise-server-to-enterprise-cloud-migrator/services"
	"github.com/bradshjg/gh-enterprise-server-to-enterprise-cloud-migrator/views"
	"github.com/labstack/echo/v4"
)

func NewMigratorHandler(migratorService services.MigratorService) *MigratorHandler {
	return &MigratorHandler{
		migratorService: migratorService,
	}
}

type MigratorHandler struct {
	migratorService services.MigratorService
}

func (fh *MigratorHandler) IndexHandler(c echo.Context) error {
	indexData := views.IndexData{
		SourceAuthenticated: fh.migratorService.SourceAuthenticated(c),
		TargetAuthenticated: fh.migratorService.TargetAuthenticated(c),
	}
	return renderView(c, views.Index(indexData))
}

func (fh *MigratorHandler) RunHandler(c echo.Context) error {
	return c.String(http.StatusOK, "Howdy!")
}

func (fh *MigratorHandler) OutputHandler(c echo.Context) error {
	return c.String(http.StatusOK, "Howdy!")
}
