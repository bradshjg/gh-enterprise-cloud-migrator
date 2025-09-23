package services

import (
	"context"
	"fmt"

	"github.com/google/go-github/v74/github"
	"github.com/labstack/echo/v4"
)

type GitHubService interface {
	ClearSession(c echo.Context)
	Orgs(c echo.Context) ([]string, error)
	Repos(c echo.Context, org string) ([]string, error)
	AccessToken(c echo.Context) (string, error)
}

func NewGitHubService(oauthService OAuthService) *GitHubAPIService {
	return &GitHubAPIService{
		oauthService: oauthService,
	}
}

type GitHubAPIService struct {
	oauthService OAuthService
}

func (gs *GitHubAPIService) ClearSession(c echo.Context) {
	gs.oauthService.ClearSession(c)
}

func (gs *GitHubAPIService) Orgs(c echo.Context) ([]string, error) {
	ctx := context.Background()
	client, err := gs.oauthService.Client(c)
	if err != nil {
		return []string{}, fmt.Errorf("error getting client: %w", err)
	}
	opt := &github.ListOptions{
		PerPage: 100,
	}
	var allOrgs []string
	for {
		orgs, resp, err := client.Organizations.List(ctx, "", opt)
		if err != nil {
			return []string{}, fmt.Errorf("error listing orgs: %w", err)
		}
		for _, org := range orgs {
			allOrgs = append(allOrgs, org.GetLogin())
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allOrgs, nil
}

func (gs *GitHubAPIService) Repos(c echo.Context, org string) ([]string, error) {
	ctx := context.Background()
	client, err := gs.oauthService.Client(c)
	if err != nil {
		return []string{}, fmt.Errorf("error getting client: %w", err)
	}
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	var allRepos []string
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			return []string{}, fmt.Errorf("error listing orgs: %w", err)
		}
		for _, repo := range repos {
			allRepos = append(allRepos, repo.GetName())
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allRepos, nil
}

func (gs *GitHubAPIService) AccessToken(c echo.Context) (string, error) {
	token, err := gs.oauthService.AccessToken(c)
	if err != nil {
		return "", err
	}
	return token, nil
}
