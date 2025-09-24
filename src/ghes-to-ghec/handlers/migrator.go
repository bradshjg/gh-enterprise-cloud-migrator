package handlers

import (
	"fmt"
	"net/http"
	"net/url"

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

type Migration struct {
	SourceOrg  string `form:"source-org"`
	SourceRepo string `form:"source-repo"`
	TargetOrg  string `form:"target-org"`
}

func (mh *MigratorHandler) StartRunHandler(c echo.Context) error {
	migration := new(Migration)
	c.Bind(migration)
	migrationData := services.Migration{
		Context:    c,
		SourceOrg:  migration.SourceOrg,
		SourceRepo: migration.SourceRepo,
		TargetOrg:  migration.TargetOrg,
	}
	token, err := mh.migratorService.Run(migrationData)
	if err != nil {
		return fmt.Errorf("error handling run: %w", err)
	}
	queryParams := url.Values{}
	queryParams.Set("token", token)
	targetURL := fmt.Sprintf("/run?%s", queryParams.Encode())
	return c.Redirect(http.StatusFound, targetURL)
}

type Output struct {
	Token string `query:"token"`
}

const StopPollingStatus = 286

func (mh *MigratorHandler) RunHandler(c echo.Context) error {
	var output Output
	err := c.Bind(&output)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: %w", err)
	}
	data := views.RunData{
		Token: output.Token,
	}
	return renderView(c, views.Run(data))
}

func (mh *MigratorHandler) OutputHandler(c echo.Context) error {
	var output Output
	err := c.Bind(&output)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: %w", err)
	}
	lines, done, err := mh.migratorService.Output(output.Token)
	if err != nil {
		return fmt.Errorf("error getting output: %w", err)
	}
	if done {
		c.Response().Writer.WriteHeader(StopPollingStatus) // HTMX handles the semantics here
	}
	outputData := views.OutputData{
		Lines: lines,
	}
	return renderView(c, views.Output(outputData))
}
