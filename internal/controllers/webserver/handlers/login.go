package handlers

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views/login_views"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/pquerna/otp/totp"
)

func (h *Handler) Login(c echo.Context) error {
	// if accidentally we disable the use of certificates this allows us to reenable it again
	if h.ReenableCertAuth {
		if err := h.Model.ReEnableCertificatesAuth(); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, i18n.T(c.Request().Context(), "authentication.could_not_reenable_certs", err.Error()))
		}
	}

	settings, err := h.Model.GetAuthenticationSettings()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, i18n.T(c.Request().Context(), "authentication.could_not_get_settings"))
	}

	// Destroy session if any
	if err := h.SessionManager.Manager.Destroy(c.Request().Context()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok || csrfToken == "" {
		return echo.NewHTTPError(http.StatusForbidden, i18n.T(c.Request().Context(), "authentication.csrf_token_not_found"))
	}

	return RenderLogin(c, login_views.LoginIndex(login_views.Login(settings), csrfToken))
}

func (h *Handler) LoginPasswordAuth(c echo.Context) error {
	username := c.FormValue("username")
	if username == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.username_empty"), true))
	}

	password := c.FormValue("password")
	if password == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.password_empty"), true))
	}

	user, err := h.Model.GetUserById(username)
	if err != nil {
		log.Printf("[ERROR]: could not get user account for username %s, reason: %v", username, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.wrong_username_or_password"), true))
	}

	if user.Hash == "" {
		log.Println("[ERROR]: hash is empty, maybe there was an issue with migration!")
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.wrong_username_or_password"), true))
	}

	// Check if passwords match
	match, err := argon2id.ComparePasswordAndHash(password, user.Hash)
	if err != nil {
		log.Printf("[ERROR]: could not compare password and hash for user %s, reason: %v", username, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.wrong_username_or_password"), true))
	}

	if !match {
		h.AuthLogger.Printf("user %s entered a wrong password", username)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.wrong_username_or_password"), true))
	}

	// Check if user is forced to change password
	if user.Register == openuem_nats.REGISTER_FORCE_PASSWORD_CHANGE {
		csrfToken, ok := c.Get("csrf").(string)
		if !ok || csrfToken == "" {
			return echo.NewHTTPError(http.StatusForbidden, i18n.T(c.Request().Context(), "authentication.csrf_token_not_found"))
		}

		// Create a session as we'll require the user to change the password
		if err := h.CreateForgotPasswordSession(c, user); err != nil {
			log.Printf("[ERROR]: could not create a forgot password session for user %s, reason: %v", user.ID, err)
		}

		return RenderLogin(c, login_views.LoginIndex(login_views.ChangePassword(), csrfToken))
	}

	// Passwords match, create a new session
	if err := h.NewSession(c, user); err != nil {
		log.Printf("[ERROR]: could not create a new session after passwords match, reason: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "could not create session")
	}

	if user.Use2fa {
		if user.TotpSecretConfirmed {
			return RenderLoginPartial(c, login_views.Use2FA(username))
		} else {
			return h.Register2FA(c)
		}
	}

	return h.AccessGranted(c, user)
}

func (h *Handler) LoginPasswordChange(c echo.Context) error {
	username := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
	if username == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.username_empty"), true))
	}

	password := c.FormValue("password")
	if password == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.password_empty"), true))
	}

	confirmPassword := c.FormValue("confirm-password")
	if confirmPassword == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.password_empty"), true))
	}

	if password != confirmPassword {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.passwords_dont_match"), true))
	}

	if err := ValidatePasswordComplexity(password); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.password_complexity_invalid"), true))
	}

	if err := h.Model.ChangePassword(username, password); err != nil {
		log.Printf("[ERROR]: could not save the new password %s, reason: %v", username, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.could_not_save_new_password"), true))
	}

	// Invalidate code to set new password
	if err := h.Model.RemoveForgotCode(username); err != nil {
		log.Printf("[ERROR]: could not remove forgot code, reason: %v", err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.could_not_remove_forgot_code"), true))
	}

	// Password has been changed
	h.AuthLogger.Printf("user %s has changed the password", username)

	// Redirect to login
	return h.Login(c)
}

func (h *Handler) Register2FA(c echo.Context) error {
	username := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
	if username == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.username_empty"), true))
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "OpenUEM",
		AccountName: username,
	})
	if err != nil {
		log.Printf("[ERROR]: could not generate totp key, reason: %v", err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.could_not_generate_totp_key"), true))
	}

	// Convert TOTP key into a PNG
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		log.Printf("[ERROR]: could not generate QR, reason: %v", err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.could_not_generate_qr"), true))
	}
	if err := png.Encode(&buf, img); err != nil {
		log.Printf("[ERROR]: could not encode image as PNG, reason: %v", err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.could_not_generate_qr"), true))
	}

	qrCode := base64.StdEncoding.EncodeToString(buf.Bytes())

	if err := h.Model.SaveTOTPSecretKey(username, key.Secret()); err != nil {
		log.Printf("[ERROR]: could not save TOTP secret key, reason: %v", err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_could_not_save_secret"), true))
	}

	return RenderLoginPartial(c, login_views.Register2FA(username, qrCode, key.Secret()))
}

func (h *Handler) LoginTOTPConfirm(c echo.Context) error {
	username := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
	if username == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.username_empty"), true))
	}

	passcode := c.FormValue("confirm-code")
	if passcode == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_empty_code"), true))
	}

	user, err := h.Model.GetUserById(username)
	if err != nil {
		log.Printf("[ERROR]: could not get user account for username %s, reason: %v", username, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_wrong_setup"), true))
	}

	valid := totp.Validate(passcode, user.TotpSecret)
	if !valid {
		log.Println("[ERROR]: the TOTP code is not valid")
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_wrong_setup"), true))
	}

	// Generate codes
	codes := []string{}
	for range 10 {
		code, err := generateRecoveryCode()
		if err != nil {
			log.Printf("[ERROR]: could not generate recovery codes, reason: %v", err)
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_wrong_setup"), true))
		}
		codes = append(codes, code)
	}

	// Save recovery codse
	if err := h.Model.SaveRecoveryCodes(username, codes); err != nil {
		log.Printf("[ERROR]: could not save recovery codes, reason: %v", err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_wrong_setup"), true))
	}

	// 2FA has been enabled
	h.AuthLogger.Printf("user %s has enabled 2FA", username)

	if err := h.SessionManager.Manager.RenewToken(c.Request().Context()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	h.SessionManager.Manager.Put(c.Request().Context(), "uid", user.ID)
	h.SessionManager.Manager.Put(c.Request().Context(), "username", user.Name)
	h.SessionManager.Manager.Put(c.Request().Context(), "user-agent", c.Request().UserAgent())
	h.SessionManager.Manager.Put(c.Request().Context(), "ip-address", c.Request().RemoteAddr)
	h.SessionManager.Manager.Put(c.Request().Context(), "usepasswd", user.Passwd)
	h.SessionManager.Manager.Put(c.Request().Context(), "email", user.Email)
	h.SessionManager.Manager.Put(c.Request().Context(), "twofa", true)
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

	// TODO - Get user's default tenant and site
	myTenant, err := h.Model.GetDefaultTenant()
	if err != nil {
		log.Printf("[ERROR]: could not get default tenant, reason: %v", err)
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	mySite, err := h.Model.GetDefaultSite(myTenant)
	if err != nil {
		log.Printf("[ERROR]: could not get default site, reason: %v", err)
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	url := ""
	if h.ReverseProxyAuthPort != "" {
		url := strings.TrimSuffix(c.Request().Referer(), "/")
		url += fmt.Sprintf("/tenant/%d/site/%d/dashboard", myTenant.ID, mySite.ID)
	} else {
		url = fmt.Sprintf("https://%s:%s/tenant/%d/site/%d/dashboard", h.ServerName, h.ConsolePort, myTenant.ID, mySite.ID)
	}

	return RenderLoginPartial(c, login_views.ShowRecoveryCodes(strings.Join(codes, "\n"), url))
}

func (h *Handler) LoginTOTPValidate(c echo.Context) error {
	username := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
	if username == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.username_empty"), true))
	}

	passcode := c.FormValue("confirm-code")
	if passcode == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_empty_code"), true))
	}

	user, err := h.Model.GetUserById(username)
	if err != nil {
		log.Printf("[ERROR]: could not get user account for username %s, reason: %v", username, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_wrong_setup"), true))
	}

	valid := totp.Validate(passcode, user.TotpSecret)
	if !valid {
		log.Println("[ERROR]: the TOTP code is not valid")
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_wrong_setup"), true))
	}

	// Access granted
	return h.AccessGranted(c, user)
}

func (h *Handler) LoginTOTPBackupRequest(c echo.Context) error {
	return RenderLoginPartial(c, login_views.EnterRecoveryCode())
}

func (h *Handler) LoginTOTPBackupCheck(c echo.Context) error {
	username := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
	if username == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.username_empty"), true))
	}

	user, err := h.Model.GetUserById(username)
	if err != nil {
		log.Printf("[ERROR]: could not get user account for username %s, reason: %v", username, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_wrong_setup"), true))
	}

	code := c.FormValue("recovery-code")
	if code == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_empty_code"), true))
	}

	isValid := h.Model.ConsumeRecoveryCode(username, code)
	if !isValid {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_wrong_recovery_code"), true))
	}

	// Access granted
	return h.AccessGranted(c, user)
}

func (h *Handler) LoginForgotPass(c echo.Context) error {
	csrfToken, ok := c.Get("csrf").(string)
	if !ok || csrfToken == "" {
		return echo.NewHTTPError(http.StatusForbidden, i18n.T(c.Request().Context(), "authentication.csrf_token_not_found"))
	}

	return RenderLogin(c, login_views.LoginIndex(login_views.LostPassword(), csrfToken))
}

func (h *Handler) NewSession(c echo.Context, user *ent.User) error {
	sessionUID := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
	if sessionUID != user.ID {
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
		h.SessionManager.Manager.Put(c.Request().Context(), "twofa", false)
		token, expiry, err := h.SessionManager.Manager.Commit(c.Request().Context())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		h.SessionManager.Manager.WriteSessionCookie(c.Request().Context(), c.Response().Writer, token, expiry)
	}

	return nil
}

func (h *Handler) AccessGranted(c echo.Context, user *ent.User) error {
	err := h.SessionManager.Manager.RenewToken(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	h.SessionManager.Manager.Put(c.Request().Context(), "uid", user.ID)
	h.SessionManager.Manager.Put(c.Request().Context(), "username", user.Name)
	h.SessionManager.Manager.Put(c.Request().Context(), "user-agent", c.Request().UserAgent())
	h.SessionManager.Manager.Put(c.Request().Context(), "usepasswd", user.Passwd)
	h.SessionManager.Manager.Put(c.Request().Context(), "email", user.Email)
	h.SessionManager.Manager.Put(c.Request().Context(), "ip-address", c.Request().RemoteAddr)
	if user.Use2fa {
		h.SessionManager.Manager.Put(c.Request().Context(), "twofa", true)
	}
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

	// TODO - Get user's default tenant and site
	myTenant, err := h.Model.GetDefaultTenant()
	if err != nil {
		log.Printf("[ERROR]: could not get default tenant, reason: %v", err)
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	mySite, err := h.Model.GetDefaultSite(myTenant)
	if err != nil {
		log.Printf("[ERROR]: could not get default site, reason: %v", err)
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if h.AuthLogger != nil {
		if user.Passwd {
			if user.Use2fa {
				h.AuthLogger.Printf("user %s has logged in with a password and using 2FA", user.ID)
			} else {
				h.AuthLogger.Printf("user %s has logged in with a password", user.ID)
			}
		} else {
			if user.Use2fa {
				h.AuthLogger.Printf("user %s has logged in with a certificate and using 2FA", user.ID)
			}
		}
	}

	if h.ReverseProxyAuthPort != "" {
		url := strings.TrimSuffix(c.Request().Referer(), "/")
		url += fmt.Sprintf("/tenant/%d/site/%d/dashboard", myTenant.ID, mySite.ID)
		return c.Redirect(http.StatusFound, url)
	} else {
		return c.Redirect(http.StatusFound, fmt.Sprintf("https://%s:%s/tenant/%d/site/%d/dashboard", h.ServerName, h.ConsolePort, myTenant.ID, mySite.ID))
	}
}

func generateRecoveryCode() (string, error) {
	var charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var randomCode string

	length := 16
	randomBytes := make([]byte, length)

	for i := range length {
		_, err := io.ReadFull(rand.Reader, randomBytes)
		if err != nil {
			return "", fmt.Errorf("failed to generate recovery code: %v", err)
		}
		randomIndex := int(randomBytes[i] % byte(len(charset)))
		randomCode += string(charset[randomIndex])
	}

	return fmt.Sprintf("%s-%s-%s-%s", randomCode[0:4], randomCode[4:8], randomCode[8:12], randomCode[12:16]), nil
}

func (h *Handler) ForgotPasswordEmail(c echo.Context) error {

	email := c.FormValue("email")
	if email == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.email_empty"), true))
	}

	username := h.Model.GetUserIDByEmail(email)
	if username != "" {
		code, err := generateForgotCode()
		if err != nil {
			return err
		}

		hash, err := argon2id.CreateHash(code, argon2id.DefaultParams)
		if err != nil {
			return err
		}

		if err := h.Model.SaveForgotCode(username, hash); err != nil {
			return err
		}

		notification := openuem_nats.Notification{
			To:               email,
			Subject:          "Request to set a new password",
			MessageTitle:     "OpenUEM | Your code to create a new password",
			MessageText:      fmt.Sprintf("Here’s your confirmation code: %s. You can copy it into the open browser window or click the link below to confirm this request", code),
			MessageGreeting:  "You or someone else has indicated that you have forgotten your login password",
			MessageAction:    "Generate a new password",
			MessageActionURL: c.Request().Header.Get("Origin") + fmt.Sprintf("/login/forgotverify?code=%s", code),
		}

		data, err := json.Marshal(notification)
		if err != nil {
			return err
		}

		if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
			return fmt.Errorf("%s", i18n.T(c.Request().Context(), "nats.not_connected"))
		}

		if err := h.NATSConnection.Publish("notification.confirm_email", data); err != nil {
			return err
		}

		user, err := h.Model.GetUserById(username)
		if err != nil {
			log.Printf("[ERROR]: could not get user account for username %s, reason: %v", username, err)
			return err
		}

		// Create a session as we'll require the username to change the password
		if err := h.CreateForgotPasswordSession(c, user); err != nil {
			return err
		}
	}

	return RenderLoginPartial(c, login_views.LostPasswordCode(email))
}

func (h *Handler) VerifyForgotPasswordCode(c echo.Context) error {
	confirmCode := ""
	if c.Request().Method == "GET" {
		confirmCode = c.QueryParam("code")
	}

	if c.Request().Method == "POST" {
		confirmCode = c.FormValue("confirm-code")
	}

	username := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
	if username == "" {
		log.Println("[ERROR]: could not find a valid username in the session")
		if c.Request().Method == "GET" {
			return echo.NewHTTPError(http.StatusUnauthorized, i18n.T(c.Request().Context(), "login.forgot_verify_error"))
		} else {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.forgot_verify_error"), true))
		}
	}

	confirmCode = strings.ToUpper(confirmCode)
	if confirmCode == "" {
		if c.Request().Method == "GET" {
			return echo.NewHTTPError(http.StatusUnauthorized, i18n.T(c.Request().Context(), "login.forgot_code_empty"))
		} else {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.forgot_code_empty"), true))
		}
	}

	valid := h.Model.IsForgotCodeValid(username, confirmCode)
	if !valid {
		log.Printf("[ERROR]: %s is not a valid code", confirmCode)
		if c.Request().Method == "GET" {
			return echo.NewHTTPError(http.StatusUnauthorized, i18n.T(c.Request().Context(), "login.forgot_verify_error"))
		} else {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.forgot_verify_error"), true))
		}
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok || csrfToken == "" {
		return echo.NewHTTPError(http.StatusForbidden, i18n.T(c.Request().Context(), "authentication.csrf_token_not_found"))
	}

	return RenderLogin(c, login_views.LoginIndex(login_views.ChangePassword(), csrfToken))
}

func (h *Handler) CreateForgotPasswordSession(c echo.Context, user *ent.User) error {
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
		h.SessionManager.Manager.Put(c.Request().Context(), "forgot", true)
		token, expiry, err := h.SessionManager.Manager.Commit(c.Request().Context())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		h.SessionManager.Manager.WriteSessionCookie(c.Request().Context(), c.Response().Writer, token, expiry)

		_, err = h.Model.Client.Sessions.UpdateOneID(token).SetOwnerID(user.ID).Save(context.Background())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	return nil
}

func (h *Handler) UpdateForgotPasswordSession(c echo.Context, user *ent.User) error {
	if err := h.SessionManager.Manager.RenewToken(c.Request().Context()); err != nil {
		return err
	}

	h.SessionManager.Manager.Remove(c.Request().Context(), "forgot")
	token, expiry, err := h.SessionManager.Manager.Commit(c.Request().Context())
	if err != nil {
		return err
	}
	h.SessionManager.Manager.WriteSessionCookie(c.Request().Context(), c.Response().Writer, token, expiry)
	return nil
}

func (h *Handler) LoginNewUser(c echo.Context) error {
	// 1. Parse token
	tokenString := c.QueryParam("token")

	if tokenString == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, i18n.T(c.Request().Context(), "login.token_invalid"))
	}

	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.JWTKey), nil
	})

	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "could not parse claims")
	}

	if claims, ok := token.Claims.(*MyCustomClaims); ok {
		// Is the token expired?
		if time.Now().After(claims.ExpiresAt.Time) {
			return echo.NewHTTPError(http.StatusForbidden, "token has expired, please contact your administrator to request a new email to set your initial password")
		}

		// Get user from database
		user, err := h.Model.GetUserById(claims.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		// Check if token exists in database for this user
		if user.NewUserToken != tokenString {
			return echo.NewHTTPError(http.StatusForbidden, "token is not valid, please contact your administrator to request a new email to set your initial password")
		}

		// Delete token
		if err := h.Model.DeleteNewAccountToken(user.ID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "could not delete token")
		}

		// Create a session as we'll require the user to change the password
		csrfToken, ok := c.Get("csrf").(string)
		if !ok || csrfToken == "" {
			return echo.NewHTTPError(http.StatusForbidden, i18n.T(c.Request().Context(), "authentication.csrf_token_not_found"))
		}

		if err := h.CreateForgotPasswordSession(c, user); err != nil {
			return err
		}
		return RenderLogin(c, login_views.LoginIndex(login_views.ChangePassword(), csrfToken))

	} else {
		return echo.NewHTTPError(http.StatusBadRequest, "unknown claims type, cannot proceed")
	}
}

func generateForgotCode() (string, error) {
	var charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var randomCode string

	length := 6
	randomBytes := make([]byte, length)

	for i := range length {
		_, err := io.ReadFull(rand.Reader, randomBytes)
		if err != nil {
			return "", fmt.Errorf("failed to generate forgot password code: %v", err)
		}
		randomIndex := int(randomBytes[i] % byte(len(charset)))
		randomCode += string(charset[randomIndex])
	}

	return fmt.Sprintf("%s", randomCode[0:6]), nil
}
