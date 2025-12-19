package handlers

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"log"
	"strings"

	"github.com/alexedwards/argon2id"
	validator "github.com/go-passwd/validator"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/views/account_views"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/pquerna/otp/totp"
)

func (h *Handler) MyAccount(c echo.Context) error {
	username := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
	if username == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.username_empty"), true))
	}

	user, err := h.Model.GetUserById(username)
	if err != nil {
		log.Printf("[ERROR]: could not get user account for username %s, reason: %v", username, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_wrong_setup"), true))
	}

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	defaultCountry, err := h.Model.GetDefaultCountry()
	if err != nil {
		return err
	}

	return RenderView(c, account_views.MyAccountIndex("| My Account", account_views.MyAccount(c, user, defaultCountry, commonInfo, ""), commonInfo))
}

func (h *Handler) UpdatePersonalInfo(c echo.Context) error {
	username := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
	if username == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.username_empty"), true))
	}

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	defaultCountry, err := h.Model.GetDefaultCountry()
	if err != nil {
		return err
	}

	if err := h.Model.UpdateUser(username, c.FormValue("name"), c.FormValue("email"), c.FormValue("phone"), c.FormValue("country")); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.could_not_update_personal_info", err.Error()), true))
	}

	h.SessionManager.Manager.Put(c.Request().Context(), "email", c.FormValue("email"))

	user, err := h.Model.GetUserById(username)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.could_not_find_user"), true))
	}

	return RenderView(c, account_views.MyAccountIndex("| My Account", account_views.MyAccount(c, user, defaultCountry, commonInfo, i18n.T(c.Request().Context(), "login.personal_info_updated")), commonInfo))
}

func (h *Handler) MyAccountPassword(c echo.Context) error {
	username := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
	if username == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.username_empty"), true))
	}

	user, err := h.Model.GetUserById(username)
	if err != nil {
		log.Printf("[ERROR]: could not get user account for username %s, reason: %v", username, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_wrong_setup"), true))
	}

	currentPassword := c.FormValue("current-password")
	if currentPassword == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.current_password_empty"), true))
	}

	newPassword := c.FormValue("new-password")
	if currentPassword == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.new_password_empty"), true))
	}

	confirmNewPassword := c.FormValue("confirm-new-password")
	if currentPassword == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.confirm_new_password_empty"), true))
	}

	if newPassword != confirmNewPassword {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.passwords_dont_match"), true))
	}

	if err := ValidatePasswordComplexity(newPassword); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.password_complexity_invalid"), true))
	}

	// Check if current password is valid
	match, err := argon2id.ComparePasswordAndHash(currentPassword, user.Hash)
	if err != nil {
		log.Printf("[ERROR]: could not compare password and hash for user %s, reason: %v", username, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.current_password_not_valid"), true))
	}

	if !match {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.current_password_not_valid"), true))
	}

	// Change password in database
	if err := h.Model.ChangePassword(username, newPassword); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.could_not_save_new_password"), true))
	}

	// Password has been changed, log out
	return h.Logout(c)
}

func (h *Handler) Enable2FA(c echo.Context) error {
	username := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
	if username == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.username_empty"), true))
	}

	user, err := h.Model.GetUserHash(username)
	if err != nil {
		log.Printf("[ERROR]: could not get user account for username %s, reason: %v", username, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_wrong_setup"), true))
	}

	if user.Passwd {
		currentPassword := c.FormValue("current-password")
		if currentPassword == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.current_password_empty"), true))
		}

		// Check if current password is valid
		match, err := argon2id.ComparePasswordAndHash(currentPassword, user.Hash)
		if err != nil {
			log.Printf("[ERROR]: could not compare password and hash for user %s, reason: %v", username, err)
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.current_password_not_valid"), true))
		}
		if !match {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.current_password_not_valid"), true))
		}
	}

	// Generate TOTP key and QR code
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

	return RenderAccountPartial(c, account_views.Enable2FA(username, qrCode, key.Secret()))
}

func (h *Handler) Enabled2FA(c echo.Context) error {
	username := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
	if username == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.username_empty"), true))
	}

	passcode := c.FormValue("confirm-code")
	if passcode == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_empty_code"), true))
	}

	user, err := h.Model.GetUserTOTPSecret(username)
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

	// Save recovery codes
	if err := h.Model.SaveRecoveryCodes(username, codes); err != nil {
		log.Printf("[ERROR]: could not save recovery codes, reason: %v", err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_wrong_setup"), true))
	}

	return RenderAccountPartial(c, account_views.Enabled2FA(strings.Join(codes, "\n")))
}

func (h *Handler) Disable2FA(c echo.Context) error {
	username := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
	if username == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.username_empty"), true))
	}

	user, err := h.Model.GetUserHash(username)
	if err != nil {
		log.Printf("[ERROR]: could not get user account for username %s, reason: %v", username, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_wrong_setup"), true))
	}

	if user.Passwd {
		currentPassword := c.FormValue("current-password")
		if currentPassword == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.current_password_empty"), true))
		}

		// Check if current password is valid
		match, err := argon2id.ComparePasswordAndHash(currentPassword, user.Hash)
		if err != nil {
			log.Printf("[ERROR]: could not compare password and hash for user %s, reason: %v", username, err)
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.current_password_not_valid"), true))
		}
		if !match {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.current_password_not_valid"), true))
		}
	}

	// Remove 2FA from database
	if err := h.Model.Disable2FA(username); err != nil {
		log.Printf("[ERROR]: could not disable 2FA for %s, reason: %v", username, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.could_not_disable_2fa"), true))
	}

	user, err = h.Model.GetUserById(username)
	if err != nil {
		log.Printf("[ERROR]: could not get user account for username %s, reason: %v", username, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "login.totp_wrong_setup"), true))
	}

	// 2FA has been disabled, log out
	return h.Logout(c)
}

func ValidatePasswordComplexity(password string) error {
	passwordValidator := validator.New(
		validator.MinLength(15, nil),
		validator.MaxLength(64, nil),
	)

	if err := passwordValidator.Validate(password); err != nil {
		return err
	}

	return nil
}
