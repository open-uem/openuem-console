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
		RefreshToken:  oAuth2TokenResponse.RefreshToken,
		AccessToken:   oAuth2TokenResponse.AccessToken,
		TokenType:     oAuth2TokenResponse.TokenType,
		TokenExpiry:   oAuth2TokenResponse.ExpiresIn,
		IDToken:       oAuth2TokenResponse.IDToken,
	}

	return &myUser, nil
}

// func (h *Handler) AuthentikGetUserRoles(userID string, idToken string, settings *ent.Authentication) (*AuthentikRolesResponse, error) {

// 	u, err := url.Parse(settings.OIDCIssuerURL)
// 	if err != nil {
// 		return nil, err
// 	}

// 	api := fmt.Sprintf("%s://%s/api/v3/rbac/permissions/users/?user_id=%s", u.Scheme, u.Host, userID)
// 	log.Println(api)

// 	roles := AuthentikRolesResponse{}

// 	// create request
// 	req, err := http.NewRequest("GET", api, nil)
// 	if err != nil {
// 		log.Printf("[ERROR]: could not prepare HTTP get for RBAC permissions endpoint, reason: %v", err)
// 		return nil, err
// 	}

// 	// add access token
// 	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", idToken))
// 	req.Header.Add("Accept", "application/json")

// 	// send request
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Printf("[ERROR]: could not get HTTP response from permissions endpoint, reason: %v", err)
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Printf("[ERROR]: could not read response bytes, reason: %v", err)
// 		return nil, err
// 	}

// 	// DEBUG
// 	log.Println(string([]byte(body)))

// 	if err := json.Unmarshal(body, &roles); err != nil {
// 		log.Printf("[ERROR]: could not unmarshal response from permissions endpoint, reason: %v", err)
// 		return nil, err
// 	}

// 	// if roles.Message != "" {
// 	// 	log.Printf("[ERROR]: could not get roles from permissions endpoint, reason: %v", err)
// 	// 	return nil, errors.New(roles.Message)
// 	// }

// 	return &roles, nil
// }
