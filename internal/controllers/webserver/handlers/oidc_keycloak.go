package handlers

import (
	"log"
	"net/http"
	"slices"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent"
)

type KeycloakRealmAccess struct {
	Roles []string `json:"roles"`
}

type KeycloakClaims struct {
	Name          string              `json:"name"`
	Username      string              `json:"preferred_username"`
	GivenName     string              `json:"given_name"`
	FamilyName    string              `json:"family_name"`
	Email         string              `json:"email"`
	EmailVerified bool                `json:"email_verified"`
	RealmAccess   KeycloakRealmAccess `json:"realm_access"`
	jwt.RegisteredClaims
}

func (h *Handler) KeycloakOIDCLogIn(c echo.Context, code string, verifier string, settings *ent.Authentication, provider *oidc.Provider) (*ent.User, error) {
	// Request token
	accessToken, err := h.ExchangeCodeForAccessToken(c, code, verifier, provider.Endpoint().TokenURL, settings.OIDCClientID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "could not exchange OIDC code for token")
	}

	// Use public key from Realms Settings -> Keys -> RS256 (get public key)
	publicKey, err := parseRSAPublicKey(settings.OIDCKeycloakPublicKey)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "could not parse Keycloak RSA public key")
	}

	// Parse JWT token
	t, err := jwt.ParseWithClaims(accessToken, &KeycloakClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, echo.NewHTTPError(http.StatusInternalServerError, "unexpected signing method")
		}

		// return the public key that is used to validate the token.
		return publicKey, nil
	})

	if err != nil {
		log.Printf("[ERROR]: could not parse access token, reason: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "could not parse access token")
	}

	// Check if JWT is valid
	if !t.Valid {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "the token is not valid")
	}

	// Get claims from JWT
	claims, ok := t.Claims.(*KeycloakClaims)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "could not get Keycloak claims")
	}

	roles := claims.RealmAccess.Roles
	if settings.OIDCRole != "" && !slices.Contains(roles, settings.OIDCRole) {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "user has no permission to log in to OpenUEM")
	}

	// Get user information
	myUser := ent.User{
		ID:            claims.Username,
		Name:          claims.GivenName + " " + claims.FamilyName,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
	}

	return &myUser, nil
}
