package services

import "github.com/labstack/echo/v4"

type MigratorService interface {
	SourceAuthenticated(c echo.Context) bool
	TargetAuthenticated(c echo.Context) bool
	Run() (string, error)
	Output(token string) ([]string, bool, error)
}

func NewMigratorService(s GitHubService, t GitHubService) MigratorService {
	return &MigratorServiceImpl{
		sourceGitHubService: s,
		targetGitHubService: t,
	}
}

type MigratorServiceImpl struct {
	sourceGitHubService GitHubService
	targetGitHubService GitHubService
}

func (ms *MigratorServiceImpl) SourceAuthenticated(c echo.Context) bool {
	_, err := ms.sourceGitHubService.AccessToken(c)
	return err == nil
}

func (ms *MigratorServiceImpl) TargetAuthenticated(c echo.Context) bool {
	_, err := ms.targetGitHubService.AccessToken(c)
	return err == nil
}

func (*MigratorServiceImpl) Run() (string, error) {
	return "", nil
}

func (*MigratorServiceImpl) Output(_ string) ([]string, bool, error) {
	return []string{}, true, nil
}
