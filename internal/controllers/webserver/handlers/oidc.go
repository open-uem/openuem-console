package handlers

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent"
	"github.com/open-uem/nats"
	"golang.org/x/oauth2"
)

func (h *Handler) OIDCLogIn(c echo.Context) error {

	provider, err := oidc.NewProvider(context.Background(), "https://openuem-console-rms331.us1.zitadel.cloud") // TODO - hardcoded must come from config
	if err != nil {
		log.Printf("[ERROR]: we could not instantiate OIDC provider, reason: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not instantiate OIDC provider")
	}

	oauth2Config := oauth2.Config{
		ClientID:    "329227437038756021", // TODO - hardcoded must come from config
		RedirectURL: h.GetRedirectURI(c),
		Endpoint:    provider.Endpoint(),
	}

	authProvider := "zitadel"                                 // TODO - hardcoded must come from config
	cookieEncryptionKey := "LnQaKMKzSxL5MEY3fXSFDyYK5Jmi7rzi" // TODO - hardcoded must come from config

	switch authProvider {
	case "zitadel":
		oauth2Config.Scopes = []string{oidc.ScopeOpenID, "profile", "email", "phone", "urn:zitadel:iam:org:project:id:zitadel:aud"}
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
	if err := WriteOIDCCookie(c, "state", state, cookieEncryptionKey); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not generate OIDC state cookie")
	}

	if err := WriteOIDCCookie(c, "verifier", verifier, cookieEncryptionKey); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not generate OIDC verifier cookie")
	}

	u := oauth2Config.AuthCodeURL(state, codeChallenge, codeChallengeMethod)

	// TODO - debug
	// log.Println("[INFO]: the OIDC auth code url is: ", u)

	return c.Redirect(http.StatusFound, u)
}

func (h *Handler) OIDCCallback(c echo.Context) error {

	var oidcUser *ent.User

	// Get code from request
	code := c.QueryParam("code")
	if code == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not get OIDC code from request")
	}

	// Get state from request
	state := c.QueryParam("state")
	if state == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not get OIDC state from request")
	}

	cookieEncryptionKey := "LnQaKMKzSxL5MEY3fXSFDyYK5Jmi7rzi" // TODO - hardcoded must come from config

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

	authProvider := "zitadel" // TODO - hardcoded must come from config

	switch authProvider {
	case "zitadel":
		oidcUser, err = h.ZitadelOIDCLogIn(c, code, verifierFromCookie)
		if err != nil {
			return err
		}
	}

	// Manage session
	return h.ManageOIDCSession(c, oidcUser)
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

func WriteOIDCCookie(c echo.Context, name string, value string, secretKey string) error {
	expiry := time.Now().Add(10 * time.Minute)

	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   "andromeda.openuem.eu",
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
		token, expiry, err := h.SessionManager.Manager.Commit(c.Request().Context())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		h.SessionManager.Manager.WriteSessionCookie(c.Request().Context(), c.Response().Writer, token, expiry)

		_, err = h.Model.Client.Sessions.UpdateOneID(token).SetOwnerID(user.ID).Save(context.Background())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		// if it's the first time let's confirm login and remove the cert password
		if err := h.Model.ConfirmLogIn(user.ID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	return nil
}

func (h *Handler) GetRedirectURI(c echo.Context) string {
	url := fmt.Sprintf("https://%s:%s/oidc/callback", h.ServerName, h.ConsolePort)
	if h.ReverseProxyAuthPort != "" {
		url = fmt.Sprintf("https://%s/oidc/callback", strings.TrimSuffix(c.Request().Referer(), "/"))
	}

	return url
}

func (h *Handler) ManageOIDCSession(c echo.Context, u *ent.User) error {
	// Check if user exists
	userExists, err := h.Model.UserExists(u.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "cannot check if user exists in database")
	}

	// If user doesn't exist create user in database that must await for review and redirect to message to wait for validation
	if !userExists {
		if err := h.Model.AddOIDCUser(u.ID, u.Name, u.Email, u.Phone, u.EmailVerified); err != nil {
			log.Printf("[ERROR]: we could not create the OIDC user, reason: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "cannot create new OIDC user")
		}
	}

	// If user exists, check if account is in a valid state
	account, err := h.Model.GetUserById(u.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "cannot get user from database")
	}

	if account.Register == nats.REGISTER_APPROVED || account.Register == nats.REGISTER_COMPLETE {
		if err := h.CreateSession(c, account); err != nil {
			log.Printf("[ERROR]: could not create session, reason: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "could not create session")
		}

		myTenant, err := h.Model.GetDefaultTenant()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		mySite, err := h.Model.GetDefaultSite(myTenant)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		if h.ReverseProxyAuthPort != "" {
			url := strings.TrimSuffix(c.Request().Referer(), "/")
			url += fmt.Sprintf("/tenant/%d/site/%d/dashboard", myTenant.ID, mySite.ID)
			return c.Redirect(http.StatusFound, url)
		} else {
			return c.Redirect(http.StatusFound, fmt.Sprintf("https://%s:%s/tenant/%d/site/%d/dashboard", h.ServerName, h.ConsolePort, myTenant.ID, mySite.ID))
		}
	}

	return echo.NewHTTPError(http.StatusForbidden, "An admin must approve your account")
}
