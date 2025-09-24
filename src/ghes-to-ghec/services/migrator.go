package services

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/labstack/echo/v4"
)

var (
	outputMap sync.Map
)

type MigratorService interface {
	SourceAuthenticated(c echo.Context) bool
	TargetAuthenticated(c echo.Context) bool
	Run(m Migration) (string, error)
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

type Migration struct {
	Context    echo.Context
	SourceOrg  string
	SourceRepo string
	TargetOrg  string
}

// Run executes a series of commands as documented by the ghes to ghec docs and returns an opaque string token for output polling.
// See https://docs.github.com/en/migrations/using-github-enterprise-importer/migrating-between-github-products/migrating-repositories-from-github-enterprise-server-to-github-enterprise-cloud
// In summary, it runs:
// `gh gei generate-script --github-source-org SOURCE_ORG --github-target-org TARGET_ORG --output FILE`
// and then
// `./FILE` to migrate all repositories (if no target repository specified)
// `gh gei migrate-repo --github-source-org SOURCE_ORG --source-repo SOURCE_REPO --github-target-org TARGET_ORG` to migrate a single repo
func (ms *MigratorServiceImpl) Run(m Migration) (string, error) {
	streamName, err := generateStreamName()
	if err != nil {
		return "", err
	}
	err = ms.run(m, streamName)
	if err != nil {
		return "", err
	}
	return streamName, nil
}

func (ms *MigratorServiceImpl) run(m Migration, sn string) error {
	// gh-gei flings files all around, create a temp directory to hold them
	dir, err := os.MkdirTemp("", "*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	sourceToken, err := ms.sourceGitHubService.AccessToken(m.Context)
	if err != nil {
		return err
	}
	targetToken, err := ms.targetGitHubService.AccessToken(m.Context)
	if err != nil {
		return err
	}

	runEnv := []string{
		fmt.Sprintf("GH_TOKEN=%s", targetToken),
		fmt.Sprintf("GH_SOURCE_PAT=%s", sourceToken),
		fmt.Sprintf("GH_PAT=%s", targetToken),
	}
	defaultArgs := []string{
		"--github-source-org", m.SourceOrg,
		"--github-target-org", m.TargetOrg,
	}
	ghesApiUrl := os.Getenv("GITHUB_SOURCE_API_URL")
	if ghesApiUrl != "" {
		defaultArgs = append(defaultArgs, "--ghes-api-url", ghesApiUrl)
	}
	migrateScript := "migrate"
	ghCLICmd := "gh"
	// run `gh gei generate-script --github-source-org SOURCE_ORG --github-target-org TARGET_ORG --output FILE`
	genScriptCmdArgs := []string{
		"gei",
		"generate-script",
		"--output", migrateScript,
	}
	genScriptCmdArgs = append(genScriptCmdArgs, defaultArgs...)
	genScriptCmd := exec.Command(ghCLICmd, genScriptCmdArgs...)
	genScriptCmd.Dir = dir
	genScriptCmd.Env = runEnv

	_, err = genScriptCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error generating migration script: %w", err)
	}
	if err = os.Chmod(filepath.Join(dir, migrateScript), 0755); err != nil {
		return err
	}

	var runMigrationCmd *exec.Cmd
	var runMigrationArgs []string

	if m.SourceRepo == "" {
		// run migration script
		runMigrationCmd = exec.Command(fmt.Sprintf("./%s", migrateScript))
	} else {
		// run single repo migration
		runMigrationArgs = []string{
			"gei",
			"migrate-repo",
			"--source-repo", m.SourceRepo,
		}
		runMigrationArgs = append(runMigrationArgs, defaultArgs...)
		runMigrationCmd = exec.Command(ghCLICmd, runMigrationArgs...)
	}

	stdoutPipe, err := runMigrationCmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderrPipe, err := runMigrationCmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := runMigrationCmd.Start(); err != nil {
		return err
	}

	ch := make(chan string, 10)
	outputMap.Store(sn, ch)

	go func() {
		defer close(ch)
		var wg sync.WaitGroup

		wg.Add(1)
		go ms.collectOutput(ch, stdoutPipe)

		wg.Add(1)
		go ms.collectOutput(ch, stderrPipe)

		if err := runMigrationCmd.Wait(); err != nil {
			log.Printf("command finished with error: %v", err)
		}

		wg.Wait()
	}()
	return nil
}

func (*MigratorServiceImpl) collectOutput(ch chan string, readPipe io.ReadCloser) {
	scanner := bufio.NewScanner(readPipe)
	for scanner.Scan() {
		ch <- scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		log.Printf("error reading pipe: %v", err)
	}
}

// Accepts an opaque string token and returns available output as slice of strings and whether output is done as a bool
func (*MigratorServiceImpl) Output(_ string) ([]string, bool, error) {
	return []string{}, true, nil
}

// generateStreamName generates a cryptographically secure random string for output streams.
func generateStreamName() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}
	name := base64.URLEncoding.EncodeToString(b)
	return name, nil
}
