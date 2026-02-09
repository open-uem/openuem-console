package handlers

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent"
	"github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/auth"
	"github.com/open-uem/openuem-console/internal/models"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"golang.org/x/oauth2"
)

type OAuth2TokenResponse struct {
	AccessToken      string `json:"access_token,omitempty"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	ExpiresIn        int    `json:"expires_in,omitempty"`
	IDToken          string `json:"id_token,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

type UserInfoResponse struct {
	Subject           string   `json:"sub,omitempty"`
	Name              string   `json:"name,omitempty"`
	GivenName         string   `json:"given_name,omitempty"`
	FamilyName        string   `json:"family_name,omitempty"`
	UpdatedAt         int      `json:"updated_at,omitempty"`
	PreferredUsername string   `json:"preferred_username,omitempty"`
	Email             string   `json:"email,omitempty"`
	EmailVerified     bool     `json:"email_verified,omitempty"`
	Phone             string   `json:"phone_number,omitempty"`
	Error             string   `json:"error,omitempty"`
	ErrorDescription  string   `json:"error_description,omitempty"`
	Groups []string `json:"groups"`
	// OIDC provider organization info (provider-specific claim names)
	// Zitadel uses: urn:zitadel:iam:user:resourceowner:id/name
	// Other providers may use: org_id, organization, tenant_id, etc.
	OIDCOrgID   string `json:"urn:zitadel:iam:user:resourceowner:id,omitempty"`
	OIDCOrgName string `json:"urn:zitadel:iam:user:resourceowner:name,omitempty"`
}

type OIDCRolesResponse struct {
	Roles   []string `json:"result"`
	Message string   `json:"message"`
}

func (h *Handler) OIDCLogIn(c echo.Context) error {

	settings, err := h.Model.GetAuthenticationSettings()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, i18n.T(c.Request().Context(), "authentication.could_not_get_settings"))
	}

	provider, err := oidc.NewProvider(context.Background(), settings.OIDCIssuerURL)
	if err != nil {
		log.Printf("[ERROR]: we could not instantiate OIDC provider, reason: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not instantiate OIDC provider")
	}

	oauth2Config := oauth2.Config{
		ClientID:    settings.OIDCClientID,
		RedirectURL: h.GetRedirectURI(c),
		Endpoint:    provider.Endpoint(),
	}

	authProvider := settings.OIDCProvider
	cookieEncryptionKey := settings.OIDCCookieEncriptionKey

	oauth2Config.Scopes = []string{"openid", "profile", "email"}
	switch authProvider {
	case auth.AUTHELIA:
		oauth2Config.Scopes = append(oauth2Config.Scopes, "groups")
	case auth.ZITADEL:
		oauth2Config.Scopes = append(oauth2Config.Scopes,
			"phone",
			"urn:zitadel:iam:org:project:id:zitadel:aud",  // Get project roles
			"urn:zitadel:iam:user:resourceowner",          // Get org ID and name
		)
	}

	state, err := randomBytestoHex(32)
	if err != nil {
		log.Printf("[ERROR]: we could not generate random OIDC state, reason: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not generate random OIDC state")
	}

	verifier := oauth2.GenerateVerifier()
	codeChallenge := oauth2.S256ChallengeOption(verifier)
	codeChallengeMethod := oauth2.SetAuthURLParam("code_challenge_method", "S256")

	// Create encrypted cookies
	if err := h.WriteOIDCCookie(c, "state", state, cookieEncryptionKey); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not generate OIDC state cookie")
	}

	if err := h.WriteOIDCCookie(c, "verifier", verifier, cookieEncryptionKey); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not generate OIDC verifier cookie")
	}

	u := oauth2Config.AuthCodeURL(state, codeChallenge, codeChallengeMethod)

	// TODO - debug
	// log.Println("[INFO]: the OIDC auth code url is: ", u)

	return c.Redirect(http.StatusFound, u)
}

func (h *Handler) OIDCCallback(c echo.Context) error {

	settings, err := h.Model.GetAuthenticationSettings()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, i18n.T(c.Request().Context(), "authentication.could_not_get_settings"))
	}

	provider, err := oidc.NewProvider(context.Background(), settings.OIDCIssuerURL)
	if err != nil {
		log.Printf("[ERROR]: we could not instantiate OIDC provider, reason: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not instantiate OIDC provider")
	}

	errorDescription := c.QueryParam("error_description")

	// Get code from request
	code := c.QueryParam("code")
	if code == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, errorDescription)
	}

	// Get state from request
	state := c.QueryParam("state")
	if state == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, errorDescription)
	}

	cookieEncryptionKey := settings.OIDCCookieEncriptionKey

	// Get state from cookie
	stateFromCookie, err := ReadOIDCCookie(c, "state", cookieEncryptionKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not read OIDC state from cookie")
	}

	// Check if states match
	if stateFromCookie != state {
		return echo.NewHTTPError(http.StatusInternalServerError, "OIDC state doesn't match")
	}

	// Get verifier from cookie
	verifierFromCookie, err := ReadOIDCCookie(c, "verifier", cookieEncryptionKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not read OIDC verifier from cookie")
	}

	// TODO Verify code if possible, I've verifier and I've the code how I can check if the code is valid? Is this needed?

	// Get access token in exchange of code
	oAuth2TokenResponse, err := h.ExchangeCodeForAccessToken(c, code, verifierFromCookie, provider.Endpoint().TokenURL, settings.OIDCClientID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not exchange OIDC code for token")
	}

	authProvider := settings.OIDCProvider

	// Get user account info from remote endpoint
	u, err := GetUserInfo(oAuth2TokenResponse.AccessToken, provider.UserInfoEndpoint())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not get user info from OIDC endpoint")
	}

	// Validate email is present (required for user ID)
	if u.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "OIDC provider did not return an email address")
	}

	// Get user information - use email as unique identifier
	oidcUser := ent.User{
		ID:            u.Email,
		Name:          u.Name,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		Phone:         u.Phone,
		RefreshToken:  oAuth2TokenResponse.RefreshToken,
		AccessToken:   oAuth2TokenResponse.AccessToken,
		TokenType:     oAuth2TokenResponse.TokenType,
		TokenExpiry:   oAuth2TokenResponse.ExpiresIn,
		IDToken:       oAuth2TokenResponse.IDToken,
	}

	// Collect roles for tenant assignment
	var oidcRoles []string

	// Check if user is member of specified group or role
	if authProvider == auth.ZITADEL {
		// Get roles info from remote endpoint
		data, err := h.GetOIDCUserRoles(oAuth2TokenResponse.AccessToken, settings)
		if err != nil {
			log.Printf("[WARN]: could not get roles from permissions endpoint: %v", err)
		} else {
			oidcRoles = data.Roles
		}

		// Check if any of the configured roles are present
		if !h.userHasAllowedOIDCRole(oidcRoles, settings) {
			return echo.NewHTTPError(http.StatusUnauthorized, "user has no permission to log in to OpenUEM")
		}
	} else {
		// Check if any of the configured groups are present
		if !h.userHasAllowedOIDCRole(u.Groups, settings) {
			return echo.NewHTTPError(http.StatusUnauthorized, "user has no permission to log in to OpenUEM")
		}
	}

	// Try to get org ID from userinfo first, fallback to ID token
	orgID := u.OIDCOrgID
	if orgID == "" && oAuth2TokenResponse.IDToken != "" {
		// Parse ID token to extract org ID (Zitadel may not include it in userinfo)
		if claims, err := parseIDTokenClaims(oAuth2TokenResponse.IDToken); err == nil {
			if oid, ok := claims["urn:zitadel:iam:user:resourceowner:id"].(string); ok {
				orgID = oid
			}
		}
	}

	// Build OIDC info for tenant assignment
	oidcInfo := OIDCTenantInfo{
		OrgID:  orgID,
		Roles:  oidcRoles,
		Groups: u.Groups,
	}

	// Manage session and assign tenant
	return h.ManageOIDCSession(c, &oidcUser, oidcInfo)
}

// Reference: https://chrisguitarguy.com/2022/12/07/oauth-pkce-with-go/
func randomBytestoHex(count int) (string, error) {
	buf := make([]byte, count)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}

func (h *Handler) WriteOIDCCookie(c echo.Context, name string, value string, secretKey string) error {
	expiry := time.Now().Add(10 * time.Minute)

	domain := h.ServerName
	if h.ReverseProxyServer != "" {
		domain = h.ReverseProxyServer
	}

	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   domain,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiry,
		MaxAge:   int(time.Until(expiry).Seconds() + 1),
	}

	// Reference: https://www.alexedwards.net/blog/working-with-cookies-in-go#encrypted-cookies

	// Create a new AES cipher block from the secret key.
	block, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		log.Printf("[ERROR]: we could not create AES cipher block, reason: %v", err)
		return err
	}

	// Wrap the cipher block in Galois Counter Mode.
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("[ERROR]: we could not wrap AES cipher block, reason: %v", err)
		return err
	}

	// Create a unique nonce containing 12 random bytes.
	nonce := make([]byte, aesGCM.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		log.Printf("[ERROR]: we could not create nonce, reason: %v", err)
		return err
	}

	// Prepare the plaintext input for encryption
	plaintext := fmt.Sprintf("%s:%s", cookie.Name, cookie.Value)

	// Encrypt the data using aesGCM.Seal()
	encryptedValue := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

	// Set the cookie value to the encryptedValue.
	cookie.Value = base64.StdEncoding.EncodeToString(encryptedValue)

	c.SetCookie(cookie)

	return nil
}

func ReadOIDCCookie(c echo.Context, name string, secretKey string) (string, error) {
	// Reference: https://www.alexedwards.net/blog/working-with-cookies-in-go#encrypted-cookies

	// Read the encrypted value from the cookie as normal.
	cookie, err := c.Request().Cookie(name)
	if err != nil {
		log.Printf("[ERROR]: we could not read the cookie, reason: %v", err)
		return "", err
	}

	// Create a new AES cipher block from the secret key.
	block, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		log.Printf("[ERROR]: we could not create the cipher block, reason: %v", err)
		return "", err
	}

	// Wrap the cipher block in Galois Counter Mode.
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("[ERROR]: we could not wrap the cipher block, reason: %v", err)
		return "", err
	}

	// Get the nonce size.
	nonceSize := aesGCM.NonceSize()

	// Convert from base64
	enc, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		log.Printf("[ERROR]: could not base64 decode the value, reason: %v", err)
		return "", err
	}

	// To avoid a potential 'index out of range' panic in the next step, we
	// check that the length of the encrypted value is at least the nonce
	// size.
	if len(enc) < nonceSize {
		log.Printf("[ERROR]: invalid value in cookie, reason: %v", err)
		return "", errors.New("invalid value")
	}

	// Split apart the nonce from the actual encrypted data.
	nonce := enc[:nonceSize]
	ciphertext := enc[nonceSize:]

	// Use aesGCM.Open() to decrypt and authenticate the data. If this fails,
	// return a ErrInvalidValue error.
	plaintext, err := aesGCM.Open(nil, []byte(nonce), []byte(ciphertext), nil)
	if err != nil {
		log.Printf("[ERROR]: could not decrypt value in cookie, reason: %v", err)
		return "", errors.New("invalid value")
	}

	// The plaintext value is in the format "{cookie name}:{cookie value}". We
	// use strings.Cut() to split it on the first ":" character.
	expectedName, value, ok := strings.Cut(string(plaintext), ":")
	if !ok {
		log.Printf("[ERROR]: could not find the expected value, reason: %v", err)
		return "", errors.New("invalid value")
	}

	// Check that the cookie name is the expected one and hasn't been changed.
	if expectedName != name {
		log.Printf("[ERROR]: unexpected cookie name, reason: %v", err)
		return "", errors.New("invalid value")
	}

	return value, nil
}

func (h *Handler) CreateSession(c echo.Context, user *ent.User) error {
	msg := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
	if msg != user.ID {
		err := h.SessionManager.Manager.RenewToken(c.Request().Context())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		h.SessionManager.Manager.Put(c.Request().Context(), "uid", user.ID)
		h.SessionManager.Manager.Put(c.Request().Context(), "username", user.Name)
		h.SessionManager.Manager.Put(c.Request().Context(), "user-agent", c.Request().UserAgent())
		h.SessionManager.Manager.Put(c.Request().Context(), "ip-address", c.Request().RemoteAddr)
		h.SessionManager.Manager.Put(c.Request().Context(), "usepasswd", user.Passwd)
		h.SessionManager.Manager.Put(c.Request().Context(), "email", user.Email)
		token, expiry, err := h.SessionManager.Manager.Commit(c.Request().Context())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		h.SessionManager.Manager.WriteSessionCookie(c.Request().Context(), c.Response().Writer, token, expiry)

		_, err = h.Model.Client.Sessions.UpdateOneID(token).SetOwnerID(user.ID).Save(context.Background())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		// if it's the first time let's confirm login
		if err := h.Model.ConfirmLogIn(user.ID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	return nil
}

func (h *Handler) GetRedirectURI(c echo.Context) string {
	u := fmt.Sprintf("https://%s:%s/oidc/callback", h.ServerName, h.ConsolePort)
	if h.ReverseProxyServer != "" {
		referer, err := url.Parse(c.Request().Referer())
		if err != nil {
			return u
		}
		u = fmt.Sprintf("https://%s:%s/oidc/callback", referer.Hostname(), referer.Port())
	}

	h.OIDCRedirectURI = u
	return u
}

// OIDCTenantInfo contains information from the OIDC provider for tenant assignment
type OIDCTenantInfo struct {
	OrgID  string   // Organization ID from OIDC provider (e.g. Zitadel resource owner)
	Roles  []string // Project roles from OIDC provider (e.g. "openuem_admin")
	Groups []string // OIDC groups (e.g. Authelia groups)
}

func (h *Handler) ManageOIDCSession(c echo.Context, u *ent.User, oidcInfo OIDCTenantInfo) error {
	settings, err := h.Model.GetAuthenticationSettings()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.could_not_get_settings", err.Error()), true))
	}

	// Check if user exists
	userExists, err := h.Model.UserExists(u.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, i18n.T(c.Request().Context(), "authentication.cannot_check_if_user_exists"))
	}

	// If user doesn't exist create user in database if auto creation is enabled
	if !userExists {
		if settings.OIDCAutoCreateAccount {
			if err := h.Model.AddOIDCUser(u.ID, u.Name, u.Email, u.Phone, u.EmailVerified, settings.OIDCAutoApprove); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, i18n.T(c.Request().Context(), "authentication.cannot_create_oidc_user", err.Error()))
			}
		} else {
			return echo.NewHTTPError(http.StatusForbidden, i18n.T(c.Request().Context(), "authentication.an_admin_must_create_your_account"))
		}
	}

	// Assign tenant based on OIDC org info (every login)
	if err := h.AssignTenantFromOIDC(u.ID, oidcInfo, settings); err != nil {
		log.Printf("[WARN]: could not assign tenant from OIDC for user %s: %v", u.ID, err)
	}

	// If user exists, check if account is in a valid state
	account, err := h.Model.GetUserById(u.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "cannot get user from database")
	}

	// If user has been approved by admin, auto approve is on or user already logged in (register completed)
	if account.Register == nats.REGISTER_APPROVED || settings.OIDCAutoApprove || account.Register == nats.REGISTER_COMPLETE {
		if err := h.CreateSession(c, account); err != nil {
			log.Printf("[ERROR]: could not create session, reason: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "could not create session")
		}

		if err := h.Model.SaveOIDCTokenInfo(u.ID, u.AccessToken, u.RefreshToken, u.IDToken, u.TokenType, u.TokenExpiry); err != nil {
			log.Printf("[ERROR]: could not save refresh token, reason: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "could not save refresh token for user")
		}

		if h.AuthLogger != nil {
			h.AuthLogger.Printf("user %s has logged in with OpenID (%s)", u.ID, settings.OIDCProvider)
		}

		myTenant, err := h.Model.GetDefaultTenant()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		mySite, err := h.Model.GetDefaultSite(myTenant)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		if h.ReverseProxyServer != "" {
			return h.Dashboard(c)
		} else {
			return c.Redirect(http.StatusFound, fmt.Sprintf("https://%s:%s/tenant/%d/site/%d/dashboard", h.ServerName, h.ConsolePort, myTenant.ID, mySite.ID))
		}
	}

	return echo.NewHTTPError(http.StatusForbidden, "An admin must approve your account")
}

func (h *Handler) ExchangeCodeForAccessToken(c echo.Context, code string, verifier string, endpoint string, clientID string) (*OAuth2TokenResponse, error) {
	var z OAuth2TokenResponse

	v := url.Values{}

	url := endpoint
	v.Set("grant_type", "authorization_code")
	v.Set("code", code)
	v.Set("redirect_uri", h.OIDCRedirectURI)
	v.Set("client_id", clientID)
	v.Set("code_verifier", verifier)

	resp, err := http.PostForm(url, v)
	if err != nil {
		log.Printf("[ERROR]: could not send request to token endpoint, reason: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR]: Error while reading the response bytes, reason: %v", err)
	}

	// Debug
	// log.Println(string([]byte(body)))

	if err := json.Unmarshal(body, &z); err != nil {
		log.Printf("[ERROR]: could not decode response from token endpoint, reason: %v", err)
		return nil, err
	}

	if z.Error != "" {
		log.Printf("[ERROR]: found an error in the response from token endpoint, reason: %v", z.Error+" "+z.ErrorDescription)
		return nil, errors.New(z.Error + " " + z.ErrorDescription)
	}

	return &z, nil
}

func GetUserInfo(accessToken string, endpoint string) (*UserInfoResponse, error) {
	user := UserInfoResponse{}

	// create request
	req, err := http.NewRequest("GET", endpoint, nil)
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

// parseIDTokenClaims parses a JWT ID token and returns the claims without signature verification
// (signature was already verified by the OIDC provider exchange)
func parseIDTokenClaims(idToken string) (map[string]interface{}, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid JWT format")
	}

	// Decode the payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("could not decode JWT payload: %w", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("could not parse JWT claims: %w", err)
	}

	return claims, nil
}

func (h *Handler) GetOIDCUserRoles(accessToken string, settings *ent.Authentication) (*OIDCRolesResponse, error) {
	u := fmt.Sprintf("%s/auth/v1/permissions/me/_search", settings.OIDCIssuerURL)
	roles := OIDCRolesResponse{}

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

// AssignTenantFromOIDC assigns a user to a tenant based on OIDC provider information.
// Strategy 1: Uses the org ID to find the matching tenant (e.g. Zitadel resource owner ID).
// Strategy 2: Uses groups in format "openuem:<org>:<role>" (e.g. Authelia).
func (h *Handler) AssignTenantFromOIDC(userID string, info OIDCTenantInfo, settings *ent.Authentication) error {
	// Strategy 1: Org ID mapping (preferred)
	if info.OrgID != "" {
		return h.assignTenantByOrgID(userID, info.OrgID, info.Roles, settings)
	}

	// Strategy 2: Generic OIDC groups (fallback for Authelia etc.)
	if len(info.Groups) > 0 {
		return h.assignTenantByGroups(userID, info.Groups, settings)
	}

	// No org ID or groups found - user may be manually assigned to tenants
	return nil
}

// assignTenantByOrgID maps OIDC org ID to a tenant and assigns the user with the correct role
func (h *Handler) assignTenantByOrgID(userID, orgID string, roles []string, settings *ent.Authentication) error {
	t, err := h.Model.GetTenantByOIDCOrgID(orgID)
	if err != nil {
		log.Printf("[WARN]: no tenant found for OIDC org ID '%s', skipping assignment", orgID)
		return nil
	}

	// Determine role from OIDC roles using configured role names from settings
	role := h.resolveOIDCRoleFromSettings(roles, settings)

	// Check if user is already assigned
	hasAccess, _ := h.Model.UserHasAccessToTenant(userID, t.ID)
	if hasAccess {
		// Update role if changed
		currentRole, err := h.Model.GetUserRoleInTenant(userID, t.ID)
		if err == nil && currentRole != role {
			if err := h.Model.UpdateUserTenantRole(userID, t.ID, role); err != nil {
				log.Printf("[ERROR]: could not update role for user %s in tenant %d: %v", userID, t.ID, err)
			} else {
				log.Printf("[INFO]: updated user %s role to %s in tenant '%s' via OIDC org mapping", userID, role, t.Description)
			}
		}
	} else {
		// New assignment
		if err := h.Model.AssignUserToTenant(userID, t.ID, role, true); err != nil {
			log.Printf("[ERROR]: could not assign user %s to tenant '%s': %v", userID, t.Description, err)
			return err
		}
		log.Printf("[INFO]: assigned user %s as %s to tenant '%s' via OIDC org ID '%s'", userID, role, t.Description, orgID)
	}
	return nil
}

// userHasAllowedOIDCRole checks if the user has any of the configured OIDC roles/groups
// Returns true if no roles are configured (allow all), or if user has at least one matching role
func (h *Handler) userHasAllowedOIDCRole(userRoles []string, settings *ent.Authentication) bool {
	// If no roles are configured, allow all users
	if settings.OIDCRoleAdmin == "" && settings.OIDCRoleOperator == "" && settings.OIDCRoleUser == "" {
		return true
	}

	// Check if user has any of the configured roles
	for _, role := range userRoles {
		if settings.OIDCRoleAdmin != "" && role == settings.OIDCRoleAdmin {
			return true
		}
		if settings.OIDCRoleOperator != "" && role == settings.OIDCRoleOperator {
			return true
		}
		if settings.OIDCRoleUser != "" && role == settings.OIDCRoleUser {
			return true
		}
	}

	return false
}

// resolveOIDCRoleFromSettings maps user's OIDC roles to OpenUEM role based on authentication settings
// Returns the highest privilege role the user has (admin > operator > user)
func (h *Handler) resolveOIDCRoleFromSettings(userRoles []string, settings *ent.Authentication) models.UserTenantRole {
	hasAdmin := false
	hasOperator := false
	hasUser := false

	for _, role := range userRoles {
		if settings.OIDCRoleAdmin != "" && role == settings.OIDCRoleAdmin {
			hasAdmin = true
		}
		if settings.OIDCRoleOperator != "" && role == settings.OIDCRoleOperator {
			hasOperator = true
		}
		if settings.OIDCRoleUser != "" && role == settings.OIDCRoleUser {
			hasUser = true
		}
	}

	// Return highest privilege role
	if hasAdmin {
		return models.UserTenantRoleAdmin
	}
	if hasOperator {
		return models.UserTenantRoleOperator
	}
	if hasUser {
		return models.UserTenantRoleUser
	}

	// Default to user role
	return models.UserTenantRoleUser
}

// assignTenantByGroups handles generic OIDC group-based tenant assignment (Authelia etc.)
// Expected group format: openuem:<organization>:<role>
func (h *Handler) assignTenantByGroups(userID string, groups []string, settings *ent.Authentication) error {
	for _, group := range groups {
		parts := strings.Split(group, ":")
		if len(parts) != 3 || parts[0] != "openuem" {
			continue
		}

		orgName := parts[1]
		roleName := parts[2]

		// Resolve role using configured settings or fallback to standard names
		var role models.UserTenantRole
		var roleFound bool
		if settings.OIDCRoleAdmin != "" && roleName == settings.OIDCRoleAdmin {
			role = models.UserTenantRoleAdmin
			roleFound = true
		} else if settings.OIDCRoleOperator != "" && roleName == settings.OIDCRoleOperator {
			role = models.UserTenantRoleOperator
			roleFound = true
		} else if settings.OIDCRoleUser != "" && roleName == settings.OIDCRoleUser {
			role = models.UserTenantRoleUser
			roleFound = true
		}

		// Fallback to standard role names if no settings match
		if !roleFound {
			switch roleName {
			case "admin":
				role = models.UserTenantRoleAdmin
			case "operator":
				role = models.UserTenantRoleOperator
			case "user":
				role = models.UserTenantRoleUser
			default:
				continue
			}
		}

		t, err := h.Model.GetTenantByName(orgName)
		if err != nil {
			log.Printf("[WARN]: organization '%s' from OIDC group not found, skipping", orgName)
			continue
		}

		hasAccess, _ := h.Model.UserHasAccessToTenant(userID, t.ID)
		if hasAccess {
			currentRole, err := h.Model.GetUserRoleInTenant(userID, t.ID)
			if err == nil && currentRole != role {
				if err := h.Model.UpdateUserTenantRole(userID, t.ID, role); err != nil {
					log.Printf("[ERROR]: could not update role for user %s in tenant %d: %v", userID, t.ID, err)
				}
			}
		} else {
			if err := h.Model.AssignUserToTenant(userID, t.ID, role, true); err != nil {
				log.Printf("[ERROR]: could not assign user %s to org '%s': %v", userID, orgName, err)
				continue
			}
			log.Printf("[INFO]: assigned user %s as %s to '%s' from OIDC group", userID, role, orgName)
		}
	}
	return nil
}
