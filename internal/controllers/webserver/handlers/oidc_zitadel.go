package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"slices"

	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent"
)

type ZitadelOAuth2TokenResponse struct {
	AccessToken      string `json:"access_token,omitempty"`
	ExpiresIn        int    `json:"expires_in,omitempty"`
	IDToken          string `json:"id_token,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

type ZitadelUserInfoResponse struct {
	Subject           string `json:"sub,omitempty"`
	Name              string `json:"name,omitempty"`
	GivenName         string `json:"given_name,omitempty"`
	FamilyName        string `json:"family_name,omitempty"`
	UpdatedAt         int    `json:"updated_at,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Email             string `json:"email,omitempty"`
	EmailVerified     bool   `json:"email_verified,omitempty"`
	Phone             string `json:"phone_number,omitempty"`
	Error             string `json:"error,omitempty"`
	ErrorDescription  string `json:"error_description,omitempty"`
}

type ZitadelRolesResponse struct {
	Roles   []string `json:"result"`
	Message string   `json:"message"`
}

func (h *Handler) ZitadelExchangeCodeForAccessToken(c echo.Context, code string, verifier string) (string, error) {
	var z ZitadelOAuth2TokenResponse

	v := url.Values{}

	url := "https://openuem-console-rms331.us1.zitadel.cloud/oauth/v2/token" // TODO - remove hardcoded url
	v.Set("grant_type", "authorization_code")
	v.Set("code", code)
	v.Set("redirect_uri", h.GetRedirectURI(c)) // TODO - remove hardcoded url
	v.Set("client_id", "329227437038756021")   // TODO - remove hardcoded client ID
	v.Set("code_verifier", verifier)

	resp, err := http.PostForm(url, v)
	if err != nil {
		log.Printf("[ERROR]: could not send request to token endpoint, reason: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&z); err != nil {
		log.Printf("[ERROR]: could not decode response from token endpoint, reason: %v", err)
		return "", err
	}

	if z.Error != "" {
		return "", errors.New(z.Error + " " + z.ErrorDescription)
	}

	return z.AccessToken, nil
}

func (h *Handler) ZitadelGetUserInfo(accessToken string) (*ZitadelUserInfoResponse, error) {
	u := "https://openuem-console-rms331.us1.zitadel.cloud/oidc/v1/userinfo" // TODO - remove hardcoded url
	user := ZitadelUserInfoResponse{}

	// create request
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Printf("[ERROR]: could not prepare HTTP get for user info endpoint, reason: %v", err)
		return nil, err
	}

	// add access token
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	// send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR]: could not get HTTP response for user info endpoint, reason: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error while reading the response bytes:", err)
	}

	// Debug
	// log.Println(string([]byte(body)))

	if err := json.Unmarshal(body, &user); err != nil {
		log.Printf("[ERROR]: could not decode response from user info endpoint, reason: %v", err)
		return nil, err
	}

	if user.Error != "" {
		log.Printf("[ERROR]: could not get user info from endpoint, reason: %v", err)
		return nil, errors.New(user.Error)
	}

	return &user, nil
}

func (h *Handler) ZitadelGetUserRoles(accessToken string) (*ZitadelRolesResponse, error) {
	u := "https://openuem-console-rms331.us1.zitadel.cloud/auth/v1/permissions/me/_search" // TODO - remove hardcoded url
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
		log.Printf("[ERROR]: could not get roles from permissions endpoint, reason: %v", err)
		return nil, errors.New(roles.Message)
	}

	return &roles, nil
}

func (h *Handler) ZitadelOIDCLogIn(c echo.Context, code string, verifier string) (*ent.User, error) {
	// Request token
	accessToken, err := h.ZitadelExchangeCodeForAccessToken(c, code, verifier)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "could not exchange OIDC code for token")
	}

	// Get user account info from remote endpoint
	u, err := h.ZitadelGetUserInfo(accessToken)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "could not get user info from OIDC endpoint")
	}

	// Get roles info from remote endpoint
	data, err := h.ZitadelGetUserRoles(accessToken)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "could not get roles from permissions endpoint")
	}

	if !slices.Contains(data.Roles, "openuem") { // TODO - hardcoded role
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "user has no permission to log in to OpenUEM")
	}

	myUser := ent.User{
		ID:            u.PreferredUsername,
		Name:          u.GivenName + " " + u.FamilyName,
		Email:         u.Email,
		Phone:         u.Phone,
		EmailVerified: u.EmailVerified,
	}

	return &myUser, nil
}
