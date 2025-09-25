package handlers

import (
	"net/http"

	"github.com/bradshjg/gh-enterprise-server-to-enterprise-cloud-migrator/services"
	"github.com/bradshjg/gh-enterprise-server-to-enterprise-cloud-migrator/views"
	"github.com/labstack/echo/v4"
)

func NewGitHubHandler(githubService services.GitHubService) *GitHubHandler {
	return &GitHubHandler{
		githubService: githubService,
	}
}

type GitHubHandler struct {
	githubService services.GitHubService
}

type OrgsQuery struct {
	ClientType services.ClientType `query:"client"`
}

func (gh *GitHubHandler) OrgsHandler(c echo.Context) error {
	orgsQuery := new(OrgsQuery)
	err := c.Bind(orgsQuery)
	if err != nil {
		return err
	}
	orgs, err := gh.githubService.Orgs(c, orgsQuery.ClientType)
	if err != nil {
		return err
	}
	data := views.OrgFormData{
		Orgs:       orgs,
		ClientType: orgsQuery.ClientType,
	}
	return renderView(c, views.OrgsForm(data))
}

type Org struct {
	Name string `query:"source-org"`
}

func (gh *GitHubHandler) ReposHandler(c echo.Context) error {
	org := new(Org)
	err := c.Bind(org)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: %w", err)
	}
	if org.Name == "" {
		data := views.SourceRepoOptionsData{
			Repos: []string{},
		}
		return renderView(c, views.SourceRepoOptions(data))
	}
	repos, err := gh.githubService.Repos(c, services.Source, org.Name)
	if err != nil {
		return err
	}
	data := views.SourceRepoOptionsData{
		Repos: repos,
	}
	return renderView(c, views.SourceRepoOptions(data))
}
