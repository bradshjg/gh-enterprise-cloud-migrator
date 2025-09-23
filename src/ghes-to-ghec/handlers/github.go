package handlers

import (
	"fmt"
	"net/http"

	"github.com/bradshjg/gh-enterprise-server-to-enterprise-cloud-migrator/services"
	"github.com/bradshjg/gh-enterprise-server-to-enterprise-cloud-migrator/views"
	"github.com/labstack/echo/v4"
)

func NewGitHubHandler(oauthService services.OAuthService, githubService services.GitHubService, isSource bool) *GitHubHandler {
	return &GitHubHandler{
		oauthService:  oauthService,
		githubService: githubService,
		isSource:      isSource,
	}
}

type GitHubHandler struct {
	oauthService  services.OAuthService
	githubService services.GitHubService
	isSource      bool
}

func (gh *GitHubHandler) OrgsHandler(c echo.Context) error {
	orgs, err := gh.githubService.Orgs(c)
	if err != nil {
		return err
	}
	data := views.OrgFormData{
		Orgs:   orgs,
		Source: gh.isSource,
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
	repos, err := gh.githubService.Repos(c, org.Name)
	if err != nil {
		return err
	}
	data := views.SourceRepoOptionsData{
		Repos: repos,
	}
	return renderView(c, views.SourceRepoOptions(data))
}

func (gh *GitHubHandler) OAuthHandler(c echo.Context) error {
	redirectURL, err := gh.oauthService.RedirectURL(c)
	if err != nil {
		return fmt.Errorf("error generating redirect url: %w", err)
	}
	return c.Redirect(http.StatusFound, redirectURL)
}

func (gh *GitHubHandler) OAuthCallbackHandler(c echo.Context) error {
	err := gh.oauthService.StoreToken(c)
	if err != nil {
		return fmt.Errorf("error storing token in oauth callback: %w", err)
	}
	return c.Redirect(http.StatusFound, "/")
}
