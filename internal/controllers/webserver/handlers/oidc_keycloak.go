package handlers

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"slices"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent"
)

type KeycloakOAuth2TokenResponse struct {
	AccessToken      string `json:"access_token,omitempty"`
	ExpiresIn        int    `json:"expires_in,omitempty"`
	IDToken          string `json:"id_token,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

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

func (h *Handler) KeycloakExchangeCodeForAccessToken(c echo.Context, code string, verifier string) (string, error) {
	var z KeycloakOAuth2TokenResponse

	v := url.Values{}

	url := "http://localhost:8080/realms/openuem/protocol/openid-connect/token" // TODO - remove hardcoded url
	v.Set("grant_type", "authorization_code")
	v.Set("code", code)
	v.Set("redirect_uri", h.GetRedirectURI(c)) // TODO - remove hardcoded url
	v.Set("client_id", "openuem")              // TODO - remove hardcoded client ID
	v.Set("code_verifier", verifier)

	resp, err := http.PostForm(url, v)
	if err != nil {
		log.Printf("[ERROR]: could not send request to token endpoint, reason: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error while reading the response bytes:", err)
	}

	// Debug
	//log.Println(string([]byte(body)))

	if err := json.Unmarshal(body, &z); err != nil {
		log.Printf("[ERROR]: could not decode response from token endpoint, reason: %v", err)
		return "", err
	}

	if z.Error != "" {
		return "", errors.New(z.Error + " " + z.ErrorDescription)
	}

	return z.AccessToken, nil
}

func (h *Handler) KeycloakOIDCLogIn(c echo.Context, code string, verifier string) (*ent.User, error) {
	// Request token
	accessToken, err := h.KeycloakExchangeCodeForAccessToken(c, code, verifier)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "could not exchange OIDC code for token")
	}

	// TODO - hardcoded the public key from Realms Settings -> Keys -> RS256 (get public key)
	base64EncodedPublicKey := `MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAsER9yw5s1zA3X9zb3SgThZr32sI4AEzp+LVlIdGYLfY6L1zFNBcGThcfiE7pHNVnDtwPOBVAdXQksx8xV/fIgC4W1wFXFLQmKGJ3civRsYLjPntJmnppb7TyIwXBMNnWvikn9gCMB1guUqgP1WCdaU7ur0J1oSlLpjQBQEfHBs56mjCQxnUvYl+hShECjxOhdvX6g6d/R2tmfTP/ix/zyY3XT9wvwf34ZQ1cVCzXlX0Y8qRaQ+SoPcjsH3TT4fRQp3Us+WR1qqzV6BdUbeikcFdSeH/63vANP2Fj1EiePC25xAGulZIMSE1scQWkp7Dz4HjapahYDCZ6kBgkcPGi5QIDAQAB`
	publicKey, err := parseKeycloakRSAPublicKey(base64EncodedPublicKey)
	if err != nil {
		log.Println(err)
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
	if !slices.Contains(roles, "openuem") { // TODO - hardcoded role
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

// Reference: https://stackoverflow.com/questions/76077022/how-to-validate-json-token-generated-by-keycloak-using-golang
func parseKeycloakRSAPublicKey(base64Encoded string) (*rsa.PublicKey, error) {
	buf, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		return nil, err
	}
	parsedKey, err := x509.ParsePKIXPublicKey(buf)
	if err != nil {
		return nil, err
	}
	publicKey, ok := parsedKey.(*rsa.PublicKey)
	if ok {
		return publicKey, nil
	}
	return nil, fmt.Errorf("unexpected key type %T", publicKey)
}
