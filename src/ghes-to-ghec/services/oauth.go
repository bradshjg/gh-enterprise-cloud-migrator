package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	githubClient "github.com/google/go-github/v74/github"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

const (
	sessionName = "ghes-to-ghec"
)

var ErrSessionNotValid = errors.New("session not valid")
var ErrKeyNotFound = errors.New("key not found")

type OAuthService interface {
	ClearSession(c echo.Context)
	RedirectURL(c echo.Context) (string, error)
	StoreToken(c echo.Context) error
	AccessToken(c echo.Context) (string, error)
	Client(c echo.Context) (*githubClient.Client, error)
}

func NewOauthService(sessionStore *sessions.CookieStore, t OAuthEndpointType) OAuthService {
	return &OAuthServiceImpl{
		oauthConfig:  githubOAuthConfig(t),
		sessionStore: sessionStore,
		sessionName:  sessionName,
		tokenKey:     fmt.Sprintf("%s-token", t),
		stateKey:     fmt.Sprintf("%s-state", t),
		verifier:     oauth2.GenerateVerifier(),
	}
}

type OAuthServiceImpl struct {
	oauthConfig  *oauth2.Config
	sessionStore *sessions.CookieStore
	sessionName  string
	tokenKey     string
	stateKey     string
	verifier     string
}

type OAuthCallbackParams struct {
	State string `query:"state"`
	Code  string `query:"code"`
}

func (os *OAuthServiceImpl) ClearSession(c echo.Context) {
	session, _ := os.sessionStore.Get(c.Request(), sessionName)
	session.Options.MaxAge = -1

	os.sessionStore.Save(c.Request(), c.Response(), session)
}

func (os *OAuthServiceImpl) RedirectURL(c echo.Context) (string, error) {
	state, err := generateRandomState()
	if err != nil {
		return "", err
	}

	os.storeState(c, state)

	return os.oauthConfig.AuthCodeURL(state, oauth2.S256ChallengeOption(os.verifier)), nil
}

func (os *OAuthServiceImpl) StoreToken(c echo.Context) error {
	ctx := context.Background()
	var oauthCallbackParams OAuthCallbackParams
	err := c.Bind(&oauthCallbackParams)
	if err != nil {
		return err
	}
	state, err := os.state(c)
	if err != nil {
		return err
	}
	if state != oauthCallbackParams.State {
		return fmt.Errorf("state values doen't match: %v, %v", state, oauthCallbackParams.State)
	}
	token, err := os.oauthConfig.Exchange(ctx, oauthCallbackParams.Code, oauth2.VerifierOption(os.verifier))
	if err != nil {
		return err
	}
	tokenJson, err := json.Marshal(token)
	if err != nil {
		return err
	}
	os.storeToken(c, string(tokenJson))
	return nil
}

func (os *OAuthServiceImpl) AccessToken(c echo.Context) (string, error) {
	token, err := os.token(c)
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

func (os *OAuthServiceImpl) Client(c echo.Context) (*githubClient.Client, error) {
	token, err := os.token(c)
	if err != nil {
		return nil, err
	}
	return githubClient.NewClient(nil).WithAuthToken(token.AccessToken), nil
}

func (os *OAuthServiceImpl) state(c echo.Context) (string, error) {
	v, err := os.get(c, os.stateKey)
	if err != nil {
		return "", err
	}
	return v, nil
}

func (os *OAuthServiceImpl) storeState(c echo.Context, value string) error {
	return os.store(c, os.stateKey, value)
}

func (os *OAuthServiceImpl) token(c echo.Context) (oauth2.Token, error) {
	tokenJSON, err := os.get(c, os.tokenKey)
	if err != nil {
		return oauth2.Token{}, err
	}
	var token oauth2.Token
	err = json.Unmarshal([]byte(tokenJSON), &token)
	if err != nil {
		return oauth2.Token{}, err
	}
	return token, nil
}

func (os *OAuthServiceImpl) storeToken(c echo.Context, value string) {
	os.store(c, os.tokenKey, value)
}

func (os *OAuthServiceImpl) store(c echo.Context, key string, value string) error {
	session, err := os.sessionStore.Get(c.Request(), os.sessionName)
	if err != nil {
		return err
	}
	session.Values[key] = value
	return session.Save(c.Request(), c.Response())
}

func (os *OAuthServiceImpl) get(c echo.Context, key string) (string, error) {
	session, err := os.sessionStore.Get(c.Request(), os.sessionName)
	if err != nil {
		return "", ErrSessionNotValid
	}
	if v, ok := session.Values[key].(string); ok {
		return v, nil
	}
	return "", ErrKeyNotFound
}

type OAuthEndpointType string

const (
	SourceOAuthEndpoint OAuthEndpointType = "SOURCE"
	TargetOAuthEndpoint OAuthEndpointType = "TARGET"
)

func oauthEndpoint(endpointType OAuthEndpointType) oauth2.Endpoint {
	var authURL string
	var tokenURL string

	if endpointType == SourceOAuthEndpoint {
		authURL = os.Getenv("GITHUB_SOURCE_OAUTH_AUTH_URL")
		tokenURL = os.Getenv("GITHUB_SOURCE_OAUTH_TOKEN_URL")
	}

	if authURL == "" {
		authURL = github.Endpoint.AuthURL
	}

	if tokenURL == "" {
		tokenURL = github.Endpoint.TokenURL
	}

	return oauth2.Endpoint{
		AuthURL:  authURL,
		TokenURL: tokenURL,
	}
}

var scopes = []string{"repo", "admin:org", "workflow"}

func githubOAuthConfig(endpointType OAuthEndpointType) *oauth2.Config {
	clientIDEnvVar := fmt.Sprintf("GITHUB_%s_OAUTH_CLIENT_ID", endpointType)
	clientSecretEnvVar := fmt.Sprintf("GITHUB_%s_OAUTH_CLIENT_SECRET", endpointType)
	redirectURLEnvVar := fmt.Sprintf("GITHUB_%s_OAUTH_REDIRECT_URL", endpointType)

	return &oauth2.Config{
		ClientID:     os.Getenv(clientIDEnvVar),
		ClientSecret: os.Getenv(clientSecretEnvVar),
		RedirectURL:  os.Getenv(redirectURLEnvVar),
		Scopes:       scopes,
		Endpoint:     oauthEndpoint(endpointType),
	}

}

// generateRandomState generates a cryptographically secure random string for OAuth state.
func generateRandomState() (string, error) {
	b := make([]byte, 32) // Generate a 32-byte random string
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
