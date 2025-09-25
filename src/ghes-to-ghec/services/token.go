package services

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

const (
	sessionName = "ghes-to-ghec"
)

var ErrCookieNotValid = errors.New("cookie not valid")
var ErrTokenNotFound = errors.New("key not found")

type Token struct {
	PersonalAccess string
	Type           ClientType
}

type TokenService interface {
	ClearTokens(c echo.Context)
	StoreToken(c echo.Context, t Token) error
	Token(c echo.Context, t ClientType) (Token, error)
}

func NewTokenService(sessionStore *sessions.CookieStore) TokenService {
	return &TokenServiceImpl{
		sessionStore: sessionStore,
		sessionName:  sessionName,
	}
}

type TokenServiceImpl struct {
	sessionStore *sessions.CookieStore
	sessionName  string
}

func (os *TokenServiceImpl) ClearTokens(c echo.Context) {
	session, _ := os.sessionStore.Get(c.Request(), sessionName)
	session.Options.MaxAge = -1

	os.sessionStore.Save(c.Request(), c.Response(), session)
}

func (os *TokenServiceImpl) StoreToken(c echo.Context, t Token) error {
	session, err := os.sessionStore.Get(c.Request(), os.sessionName)
	if err != nil {
		return err
	}
	tokenJson, err := json.Marshal(t)
	if err != nil {
		return err
	}
	session.Values[os.tokenKey(t.Type)] = tokenJson
	return session.Save(c.Request(), c.Response())
}

func (os *TokenServiceImpl) Token(c echo.Context, e ClientType) (Token, error) {
	session, err := os.sessionStore.Get(c.Request(), os.sessionName)
	if err != nil {
		return Token{}, ErrCookieNotValid
	}
	tokenJSON, ok := session.Values[os.tokenKey(e)].(string)
	if !ok {
		return Token{}, ErrTokenNotFound
	}
	token := new(Token)
	err = json.Unmarshal([]byte(tokenJSON), token)
	if err != nil {
		return Token{}, err
	}
	return *token, nil
}

func (os *TokenServiceImpl) tokenKey(t ClientType) string {
	return fmt.Sprintf("%s-token", t)
}
