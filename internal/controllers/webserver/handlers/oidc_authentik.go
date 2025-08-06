package handlers

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent"
)

type AuthentikRealmAccess struct {
	Roles []string `json:"roles"`
}

type AuthentikClaims struct {
	Name          string               `json:"name"`
	Username      string               `json:"preferred_username"`
	GivenName     string               `json:"given_name"`
	FamilyName    string               `json:"family_name"`
	Email         string               `json:"email"`
	EmailVerified bool                 `json:"email_verified"`
	RealmAccess   AuthentikRealmAccess `json:"realm_access"`
	jwt.RegisteredClaims
}

func (h *Handler) AuthentikOIDCLogIn(c echo.Context, code string, verifier string, settings *ent.Authentication) (*ent.User, error) {
	// Request token
	endpoint := fmt.Sprintf("%s/token/", settings.OIDCServer)
	client := settings.OIDCClientID
	accessToken, err := h.ExchangeCodeForAccessToken(c, code, verifier, endpoint, client)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "could not exchange OIDC code for token")
	}

	// Get user account info from remote endpoint
	endpoint = fmt.Sprintf("%s/userinfo/", settings.OIDCServer)
	u, err := h.GetUserInfo(accessToken, endpoint)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "could not get user info from OIDC endpoint")
	}

	if settings.OIDCRole != "" && !slices.Contains(u.Groups, settings.OIDCRole) {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "user has no permission to log in to OpenUEM")
	}

	// Get user information
	myUser := ent.User{
		ID:            u.PreferredUsername,
		Name:          u.GivenName,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
	}

	return &myUser, nil
}
