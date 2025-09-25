package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v74/github"
	githubClient "github.com/google/go-github/v74/github"
	"github.com/labstack/echo/v4"
)

type ClientType int

const (
	Source ClientType = iota
	Target
)

func (e ClientType) String() string {
	switch e {
	case Source:
		return "source"
	case Target:
		return "destination"
	default:
		return "unknown"
	}
}

type GitHubService interface {
	ClearSession(c echo.Context)
	Token(c echo.Context, t ClientType) (string, error)
	Orgs(c echo.Context, t ClientType) ([]string, error)
	Repos(c echo.Context, t ClientType, org string) ([]string, error)
	Scopes(c echo.Context, t ClientType) ([]string, error)
}

func NewGitHubService(tokenService TokenService) *GitHubAPIService {
	return &GitHubAPIService{
		tokenService: tokenService,
	}
}

type GitHubAPIService struct {
	tokenService TokenService
}

func (gs *GitHubAPIService) ClearSession(c echo.Context) {
	gs.tokenService.ClearTokens(c)
}

func (gs *GitHubAPIService) Token(c echo.Context, t ClientType) (string, error) {
	token, err := gs.tokenService.Token(c, t)
	if err != nil {
		return "", err
	}
	return token.PersonalAccess, nil
}

func (gs *GitHubAPIService) Orgs(c echo.Context, t ClientType) ([]string, error) {
	ctx := context.Background()
	client, err := gs.Client(c, t)
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

func (gs *GitHubAPIService) Repos(c echo.Context, t ClientType, org string) ([]string, error) {
	ctx := context.Background()
	client, err := gs.Client(c, t)
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

func (gs *GitHubAPIService) Scopes(c echo.Context, t ClientType) ([]string, error) {
	ctx := context.Background()
	client, err := gs.Client(c, t)
	if err != nil {
		return []string{}, fmt.Errorf("error getting scopes: %w", err)
	}
	_, resp, err := client.RateLimit.Get(ctx)
	if err != nil {
		return []string{}, fmt.Errorf("error getting scopes: %w", err)
	}
	scopesStr := resp.Header.Get("x-oauth-scopes")
	scopes := strings.Split(scopesStr, ", ")
	return scopes, nil
}

func (gs *GitHubAPIService) Client(c echo.Context, t ClientType) (*githubClient.Client, error) {
	token, err := gs.tokenService.Token(c, t)
	if err != nil {
		return nil, err
	}
	return github.NewClient(nil).WithAuthToken(token.PersonalAccess), nil
}
