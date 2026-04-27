package handlers

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	smtpsettings "github.com/open-uem/ent/settings"
	"github.com/open-uem/openuem-console/internal/models"
	"github.com/open-uem/openuem-console/internal/views/admin_views"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/utils"
	"github.com/wneessen/go-mail"
)

func (h *Handler) SMTPSettings(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	if c.Request().Method == "POST" {

		settings, err := validateSMTPSettings(c)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}

		// encrypt the SMTP Password if we have the encryption master key and the password is not already encrypted
		if h.EncryptionMasterKey != "" && settings.Password != "" {
			isSMTPPasswordEncrypted, err := utils.IsSensitiveFieldEncrypted(settings.Password, h.EncryptionMasterKey)
			if err != nil {
				return err
			}

			if !isSMTPPasswordEncrypted {
				settings.Password, err = utils.EncryptSensitiveField(settings.Password, h.EncryptionMasterKey)
				if err != nil {
					return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "smtp.cannot_be_encrypted"), true))
				}
			}
		}

		if err := h.Model.UpdateSMTPSettings(settings); err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}

		// Notification Worker must reload its smtp settings
		if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
		}

		if err := h.NATSConnection.Publish("notification.reload_settings", nil); err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}

		return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "smtp.saved")))
	}

	settings, err := h.Model.GetSMTPSettings(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// decrypt access token
	if h.EncryptionMasterKey != "" && settings.SMTPPassword != "" {
		isSecretEncrypted, err := utils.IsSensitiveFieldEncrypted(settings.SMTPPassword, h.EncryptionMasterKey)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "smtp.cannot_be_decrypted", err.Error()), true))
		}

		if isSecretEncrypted {
			settings.SMTPPassword, err = utils.DecryptSensitiveField(settings.SMTPPassword, h.EncryptionMasterKey)
			if err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "smtp.cannot_be_decrypted", err.Error()), true))
			}
		}
	}

	agentsExists, err := h.Model.AgentsExists(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return RenderView(c, admin_views.SMTPSettingsIndex(" | SMTP Settings", admin_views.SMTPSettings(c, settings, agentsExists, serversExists, commonInfo, h.GetAdminTenantName(commonInfo)), commonInfo))
}

func (h *Handler) TestSMTPSettings(c echo.Context) error {
	var err error

	settings, err := validateSMTPSettings(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := h.SendEmailTest(settings, settings.MailFrom); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "smtp.test_success", settings.MailFrom)))
}

func validateSMTPSettings(c echo.Context) (*models.SMTPSettings, error) {
	var err error

	validate := validator.New()
	settings := models.SMTPSettings{}

	settingsId := c.FormValue("settingsId")
	settings.Server = strings.TrimSpace(c.FormValue("server"))
	port := c.FormValue("port")
	settings.User = c.FormValue("user")
	settings.Password = c.FormValue("password")
	settings.Auth = c.FormValue("auth")
	settings.MailFrom = c.FormValue("mail-from")
	settings.EncryptionType = c.FormValue("encryption")

	if settingsId == "" {
		return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "smtp.id_cannot_be_empty"))
	}

	settings.ID, err = strconv.Atoi(settingsId)
	if err != nil {
		return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "smtp.id_invalid"))
	}

	if settings.Server == "" {
		return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "smtp.server_cannot_be_empty"))
	}

	if errs := validate.Var(settings.Server, "hostname"); errs != nil {
		return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "smtp.server_invalid"))
	}

	if port == "" {
		return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "smtp.port_cannot_be_empty"))
	}

	settings.Port, err = strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "smtp.port_invalid"))
	}

	if settings.Port < 0 || settings.Port > 65535 {
		return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "smtp.port_invalid"))
	}

	if !slices.Contains(admin_views.AuthTypes, settings.Auth) {
		return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "smtp.auth_invalid"))
	}

	if settings.MailFrom == "" {
		return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "smtp.mailfrom_cannot_be_empty"))
	}

	if !slices.Contains([]string{string(smtpsettings.SMTPEncryptionTypeNone), string(smtpsettings.SMTPEncryptionTypeSmtps), string(smtpsettings.SMTPEncryptionTypeStarttls)}, settings.EncryptionType) {
		return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "smtp.encription_invalid"))
	}

	if errs := validate.Var(settings.MailFrom, "email"); errs != nil {
		return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "smtp.mailfrom_invalid"))
	}

	return &settings, nil
}

func (h *Handler) SendEmailTest(settings *models.SMTPSettings, to string) error {
	var err error
	var c *mail.Client
	if settings.Auth == "NOAUTH" || (settings.User == "" && settings.Password == "") {
		c, err = mail.NewClient(settings.Server, mail.WithPort(settings.Port))
	} else {

		// We need the SMTP password in clear
		smtpPassword := settings.Password

		// if not empty check if we have the key to decrypt it
		if settings.Password != "" {
			if h.EncryptionMasterKey != "" {
				isSMTPPasswordEncrypted, err := utils.IsSensitiveFieldEncrypted(settings.Password, h.EncryptionMasterKey)
				if err != nil {
					return err
				}

				if isSMTPPasswordEncrypted {
					smtpPassword, err = utils.DecryptSensitiveField(settings.Password, h.EncryptionMasterKey)
					if err != nil {
						return err
					}
				}
			}
		}
		c, err = mail.NewClient(settings.Server, mail.WithPort(settings.Port), mail.WithSMTPAuth(mail.SMTPAuthType(settings.Auth)),
			mail.WithUsername(settings.User), mail.WithPassword(smtpPassword))
	}
	if err != nil {
		return err
	}

	if settings.EncryptionType == string(smtpsettings.SMTPEncryptionTypeSmtps) {
		c.SetSSL(true)
	}

	if settings.EncryptionType == string(smtpsettings.SMTPEncryptionTypeStarttls) {
		c.SetTLSPortPolicy(mail.TLSMandatory)
	}

	m := mail.NewMsg()
	if err := m.From(settings.MailFrom); err != nil {
		return err
	}
	if err := m.To(to); err != nil {
		return err
	}
	m.Subject("This is a test email from OpenUEM")

	return c.DialAndSend(m)
}
