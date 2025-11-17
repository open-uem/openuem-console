package handlers

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"image/png"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/alexedwards/argon2id"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent"
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

	csrfToken, ok := c.Get("csrf").(string)
	if !ok || csrfToken == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, i18n.T(c.Request().Context(), "authentication.could_not_get_settings"))
	}

	return RenderLogin(c, login_views.LoginIndex(login_views.Login(settings), csrfToken))
}

func (h *Handler) LoginUserPassword(c echo.Context) error {
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

	return RenderView(c, login_views.LoginUserPassword(settings))
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
		// error should go to auth log
		log.Printf("[ERROR]: could not get user account for username %s, reason: %v", username, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.wrong_username_or_password"), true))
	}

	if user.Hash == "" {
		// error should go to auth log
		log.Println("[ERROR]: hash is empty, maybe there was an issue with migration!")
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.wrong_username_or_password"), true))
	}

	// Check if passwords match
	match, err := argon2id.ComparePasswordAndHash(password, user.Hash)
	if err != nil {
		// error should go to auth log
		log.Printf("[ERROR]: could not compare password and hash for user %s, reason: %v", username, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.wrong_username_or_password"), true))
	}

	if !match {
		// error should go to auth log
		log.Printf("[ERROR]: user %s entered a wrong password", username)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.wrong_username_or_password"), true))
	}

	// Passwords match, create a new session
	if err := h.NewSession(c, user); err != nil {
		log.Printf("[ERROR]: could not create session, reason: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "could not create session")
	}

	// Check if user is forced to change password
	if user.Register == "users.force_change_password" {
		return RenderLoginPartial(c, login_views.ChangePassword(username))
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

	user, err := h.Model.GetUserById(username)
	if err != nil {
		// error should go to auth log
		log.Printf("[ERROR]: could not get user account for username %s, reason: %v", username, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.wrong_username_or_password"), true))
	}

	if err := h.Model.ChangePassword(username, password); err != nil {
		log.Printf("[ERROR]: could not save the new password %s, reason: %v", username, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.could_not_save_new_password"), true))
	}

	// TODO - must redirect after password change to enter 2FA?

	// Access granted
	return h.AccessGranted(c, user)
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
		// error should go to auth log
		log.Printf("[ERROR]: could not generate totp key, reason: %v", err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.could_not_generate_totp_key"), true))
	}

	// Convert TOTP key into a PNG
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		// error should go to auth log
		log.Printf("[ERROR]: could not generate QR, reason: %v", err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.could_not_generate_qr"), true))
	}
	if err := png.Encode(&buf, img); err != nil {
		// error should go to auth log
		log.Printf("[ERROR]: could not encode image as PNG, reason: %v", err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.could_not_generate_qr"), true))
	}

	qrCode := base64.StdEncoding.EncodeToString(buf.Bytes())

	if err := h.Model.SaveTOTPSecretKey(username, key.Secret()); err != nil {
		// error should go to auth log
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
		// error should go to auth log
		log.Printf("[ERROR]: could not get user account for username %s, reason: %v", username, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_wrong_setup"), true))
	}

	valid := totp.Validate(passcode, user.TotpSecret)
	if !valid {
		// error should go to auth log
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

	if err := h.SessionManager.Manager.RenewToken(c.Request().Context()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	h.SessionManager.Manager.Put(c.Request().Context(), "uid", user.ID)
	h.SessionManager.Manager.Put(c.Request().Context(), "username", user.Name)
	h.SessionManager.Manager.Put(c.Request().Context(), "user-agent", c.Request().UserAgent())
	h.SessionManager.Manager.Put(c.Request().Context(), "ip-address", c.Request().RemoteAddr)
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
		// error should go to auth log
		log.Printf("[ERROR]: could not get user account for username %s, reason: %v", username, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_wrong_setup"), true))
	}

	valid := totp.Validate(passcode, user.TotpSecret)
	if !valid {
		// error should go to auth log
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
		// error should go to auth log
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
	return RenderView(c, login_views.LostPassword())
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
	h.SessionManager.Manager.Put(c.Request().Context(), "ip-address", c.Request().RemoteAddr)
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
