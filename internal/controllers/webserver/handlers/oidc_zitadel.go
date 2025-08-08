package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent"
)

type ZitadelRolesResponse struct {
	Roles   []string `json:"result"`
	Message string   `json:"message"`
}

func (h *Handler) ZitadelOIDCLogIn(c echo.Context, code string, verifier string, settings *ent.Authentication, provider *oidc.Provider) (*ent.User, error) {
	// Request token
	oAuth2TokenResponse, err := h.ExchangeCodeForAccessToken(c, code, verifier, provider.Endpoint().TokenURL, settings.OIDCClientID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "could not exchange OIDC code for token")
	}

	// Get user account info from remote endpoint
	u, err := h.GetUserInfo(oAuth2TokenResponse.AccessToken, provider.UserInfoEndpoint())
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "could not get user info from OIDC endpoint")
	}

	if settings.OIDCRole != "" {
		// Get roles info from remote endpoint
		data, err := h.ZitadelGetUserRoles(oAuth2TokenResponse.AccessToken, settings)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusInternalServerError, "could not get roles from permissions endpoint")
		}

		if !slices.Contains(data.Roles, settings.OIDCRole) {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "user has no permission to log in to OpenUEM")
		}
	}

	myUser := ent.User{
		ID:            u.PreferredUsername,
		Name:          u.GivenName + " " + u.FamilyName,
		Email:         u.Email,
		Phone:         u.Phone,
		EmailVerified: u.EmailVerified,
		RefreshToken:  oAuth2TokenResponse.RefreshToken,
		AccessToken:   oAuth2TokenResponse.AccessToken,
		TokenType:     oAuth2TokenResponse.TokenType,
		TokenExpiry:   oAuth2TokenResponse.ExpiresIn,
		IDToken:       oAuth2TokenResponse.IDToken,
	}

	return &myUser, nil
}

func (h *Handler) ZitadelGetUserRoles(accessToken string, settings *ent.Authentication) (*ZitadelRolesResponse, error) {
	u := fmt.Sprintf("%s/auth/v1/permissions/me/_search", settings.OIDCIssuerURL)
	roles := ZitadelRolesResponse{}

	// create request
	req, err := http.NewRequest("POST", u, nil)
	if err != nil {
		log.Printf("[ERROR]: could not prepare HTTP get for permissions endpoint, reason: %v", err)
		return nil, err
	}

	// add access token
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Add("Accept", "application/json")

	// send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR]: could not get HTTP response from permissions endpoint, reason: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR]: could not read response bytes, reason: %v", err)
		return nil, err
	}

	// DEBUG
	// log.Println(string([]byte(body)))

	if err := json.Unmarshal(body, &roles); err != nil {
		log.Printf("[ERROR]: could not unmarshal response from permissions endpoint, reason: %v", err)
		return nil, err
	}

	if roles.Message != "" {
		log.Printf("[ERROR]: could not get roles from permissions endpoint, reason: %v", roles.Message)
		return nil, errors.New(roles.Message)
	}

	return &roles, nil
}
