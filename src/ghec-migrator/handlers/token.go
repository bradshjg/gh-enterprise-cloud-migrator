package handlers

import (
	"net/http"

	"github.com/bradshjg/ghec-migrator/services"
	"github.com/labstack/echo/v4"
)

func NewTokenHandler(tokenService services.TokenService) *TokenHandler {
	return &TokenHandler{
		tokenService: tokenService,
	}
}

type TokenHandler struct {
	tokenService services.TokenService
}

type TokenPayload struct {
	Token      string              `form:"token"`
	ClientType services.ClientType `form:"client"`
}

func (th *TokenHandler) TokenHandler(c echo.Context) error {
	tp := new(TokenPayload)
	err := c.Bind(tp)
	if err != nil {
		return err
	}
	token := services.Token{
		PersonalAccess: tp.Token,
		Type:           tp.ClientType,
	}
	err = th.tokenService.StoreToken(c, token)
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/")
}

func (th *TokenHandler) ResetTokensHandler(c echo.Context) error {
	th.tokenService.ClearSession(c)
	return c.Redirect(http.StatusFound, "/")
}
