package handlers

import (
	"net/http"
	"slices"

	"github.com/coreos/go-oidc/v3/oidc"
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

type AuthentikRole struct {
	Name string `json:"name"`
}
type AuthentikRolesResponse struct {
	Roles []AuthentikRole `json:"results"`
}

func (h *Handler) AuthentikOIDCLogIn(c echo.Context, code string, verifier string, settings *ent.Authentication, provider *oidc.Provider) (*ent.User, error) {
	// Request token
	accessToken, err := h.ExchangeCodeForAccessToken(c, code, verifier, provider.Endpoint().TokenURL, settings.OIDCClientID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "could not exchange OIDC code for token")
	}

	// Get user account info from remote endpoint
	u, err := h.GetUserInfo(accessToken, provider.UserInfoEndpoint())
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "could not get user info from OIDC endpoint")
	}

	if settings.OIDCRole != "" {
		if !slices.Contains(u.Groups, settings.OIDCRole) {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "user has no permission to log in to OpenUEM")
		}
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
