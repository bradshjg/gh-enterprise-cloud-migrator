package services

import (
	"encoding/json"
	"errors"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

const (
	sessionName = "ghec-migrator"
)

var ErrTokenNotFound = errors.New("missing token")

type Token struct {
	PersonalAccess string
	Type           ClientType
}

type TokenService interface {
	ClearSession(c echo.Context)
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

func (ts *TokenServiceImpl) ClearSession(c echo.Context) {
	session, _ := ts.sessionStore.Get(c.Request(), sessionName)
	session.Options.MaxAge = -1

	ts.sessionStore.Save(c.Request(), c.Response(), session)
}

func (ts *TokenServiceImpl) StoreToken(c echo.Context, t Token) error {
	session, err := ts.sessionStore.Get(c.Request(), ts.sessionName)
	if err != nil {
		return err
	}
	tokenJson, err := json.Marshal(t)
	if err != nil {
		return err
	}
	session.Values[string(t.Type)] = tokenJson
	return session.Save(c.Request(), c.Response())
}

func (ts *TokenServiceImpl) Token(c echo.Context, e ClientType) (Token, error) {
	session, err := ts.sessionStore.Get(c.Request(), ts.sessionName)
	if err != nil {
		ts.ClearSession(c)
		return Token{}, ErrTokenNotFound
	}
	tokenJSON, ok := session.Values[string(e)].([]byte)
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
