package handlers

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/doncicuto/openuem-console/internal/models"
	"github.com/doncicuto/openuem-console/internal/views/admin_views"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/go-playground/validator/v10"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/wneessen/go-mail"
)

func (h *Handler) SMTPSettings(c echo.Context) error {
	var err error

	if c.Request().Method == "POST" {

		settings, err := validateSMTPSettings(c)
		if err != nil {
			return renderError(c, partials.ErrorMessage(err.Error(), false))
		}
		if err := h.Model.UpdateSMTPSettings(settings); err != nil {
			return renderError(c, partials.ErrorMessage(err.Error(), false))
		}

		// Notification Worker must reload its smtp settings
		if err := h.NATSConnection.Publish("notification.reload_settings", nil); err != nil {
			return renderError(c, partials.ErrorMessage(err.Error(), false))
		}

		return renderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "smtp.saved")))
	}

	settings, err := h.Model.GetSMTPSettings()
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return renderView(c, admin_views.SMTPSettingsIndex(" | SMTP Settings", admin_views.SMTPSettings(c, settings)))
}

func (h *Handler) TestSMTPSettings(c echo.Context) error {
	var err error

	settings, err := validateSMTPSettings(c)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := sendEmailTest(settings, settings.MailFrom); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}
	return renderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "smtp.test_success", settings.MailFrom)))
}

func validateSMTPSettings(c echo.Context) (*models.SMTPSettings, error) {
	var err error

	validate := validator.New()
	settings := models.SMTPSettings{}

	settingsId := c.FormValue("settingsId")
	settings.Server = c.FormValue("server")
	port := c.FormValue("port")
	settings.User = c.FormValue("user")
	settings.Password = c.FormValue("password")
	settings.Auth = c.FormValue("auth")
	settings.MailFrom = c.FormValue("mail-from")

	if settingsId == "" {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "smtp.id_cannot_be_empty"))
	}

	settings.ID, err = strconv.Atoi(settingsId)
	if err != nil {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "smtp.id_invalid"))
	}

	if settings.Server == "" {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "smtp.server_cannot_be_empty"))
	}

	if errs := validate.Var(settings.Server, "hostname"); errs != nil {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "smtp.server_invalid"))
	}

	if port == "" {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "smtp.port_cannot_be_empty"))
	}

	settings.Port, err = strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "smtp.port_invalid"))
	}

	if settings.Port < 0 || settings.Port > 65535 {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "smtp.port_invalid"))
	}

	if !slices.Contains(admin_views.AuthTypes, settings.Auth) {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "smtp.auth_invalid"))
	}

	if settings.MailFrom == "" {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "smtp.mailfrom_cannot_be_empty"))
	}

	if errs := validate.Var(settings.MailFrom, "email"); errs != nil {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "smtp.mailfrom_invalid"))
	}

	return &settings, nil
}

func sendEmailTest(settings *models.SMTPSettings, to string) error {
	var err error
	var c *mail.Client
	if settings.Auth == "NOAUTH" || (settings.User == "" && settings.Password == "") {
		c, err = mail.NewClient(settings.Server, mail.WithPort(settings.Port))
	} else {
		c, err = mail.NewClient(settings.Server, mail.WithPort(settings.Port), mail.WithSMTPAuth(mail.SMTPAuthType(settings.Auth)),
			mail.WithUsername(settings.User), mail.WithPassword(settings.Password))
	}
	if err != nil {
		return err
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