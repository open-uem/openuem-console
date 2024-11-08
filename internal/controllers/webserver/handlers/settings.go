package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/doncicuto/openuem-console/internal/models"
	"github.com/doncicuto/openuem-console/internal/views/admin_views"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/go-playground/validator/v10"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
)

func (h *Handler) GeneralSettings(c echo.Context) error {
	var err error

	if c.Request().Method == "POST" {

		settings, err := validateGeneralSettings(c)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), true))
		}

		// TODO - This setting may not be effective until the console service is restarted
		if settings.MaxUploadSize != "" {
			if err := h.Model.UpdateMaxUploadSizeSetting(settings.ID, settings.MaxUploadSize); err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), true))
			}
			return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "settings.reload")))
		}

		if settings.NATSTimeout != 0 {
			if err := h.Model.UpdateNATSTimeoutSetting(settings.ID, settings.NATSTimeout); err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), true))
			}
		}

		if settings.Country != "" {
			if err := h.Model.UpdateCountrySetting(settings.ID, settings.Country); err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), true))
			}
		}

		if settings.UserCertYears != 0 {
			if err := h.Model.UpdateUserCertDurationSetting(settings.ID, settings.UserCertYears); err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), true))
			}
		}

		if settings.Refresh != 0 {
			if err := h.Model.UpdateRefreshTimeSetting(settings.ID, settings.Refresh); err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), true))
			}
		}

		if settings.SessionLifetime != 0 {
			if err := h.Model.UpdateSessionLifetime(settings.ID, settings.SessionLifetime); err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), true))
			}
			return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "settings.reload")))
		}

		return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "settings.saved")))
	}

	settings, err := h.Model.GetGeneralSettings()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, admin_views.GeneralSettingsIndex(" | General Settings", admin_views.GeneralSettings(c, settings)))
}

func validateGeneralSettings(c echo.Context) (*models.GeneralSettings, error) {
	var err error

	validate := validator.New()
	settings := models.GeneralSettings{}

	settingsId := c.FormValue("settingsId")
	country := c.FormValue("country")
	natsTimeout := c.FormValue("nats-timeout")
	maxUploadSize := c.FormValue("max-upload-size")
	certYear := c.FormValue("cert-years")
	refresh := c.FormValue("refresh")
	sessionLifetime := c.FormValue("session-lifetime")

	if settingsId == "" {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "settings.id_cannot_be_empty"))
	}

	settings.ID, err = strconv.Atoi(settingsId)
	if err != nil {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "settings.id_invalid"))
	}

	if country != "" {
		if errs := validate.Var(country, "iso3166_1_alpha2"); errs != nil {
			return nil, fmt.Errorf(i18n.T(c.Request().Context(), "settings.country_invalid"))
		}
		settings.Country = country
	}

	if certYear != "" {
		settings.UserCertYears, err = strconv.Atoi(certYear)
		if err != nil {
			return nil, fmt.Errorf(i18n.T(c.Request().Context(), "settings.cert_years_invalid"))
		}

		if settings.UserCertYears <= 0 {
			return nil, fmt.Errorf(i18n.T(c.Request().Context(), "settings.cert_years_invalid"))
		}
	}

	if natsTimeout != "" {
		settings.NATSTimeout, err = strconv.Atoi(natsTimeout)
		if err != nil {
			return nil, fmt.Errorf(i18n.T(c.Request().Context(), "settings.nats_timeout_invalid"))
		}

		if settings.NATSTimeout <= 0 {
			return nil, fmt.Errorf(i18n.T(c.Request().Context(), "settings.nats_timeout_invalid"))
		}
	}

	if maxUploadSize != "" {
		if !strings.HasSuffix(maxUploadSize, "M") && !strings.HasSuffix(maxUploadSize, "K") && !strings.HasSuffix(maxUploadSize, "G") {
			return nil, fmt.Errorf(i18n.T(c.Request().Context(), "settings.max_upload_size_invalid"))
		}
		settings.MaxUploadSize = maxUploadSize
	}

	if refresh != "" {
		settings.Refresh, err = strconv.Atoi(refresh)
		if err != nil {
			return nil, fmt.Errorf(i18n.T(c.Request().Context(), "settings.refresh_invalid"))
		}

		if settings.Refresh <= 0 {
			return nil, fmt.Errorf(i18n.T(c.Request().Context(), "settings.refresh_invalid"))
		}
	}

	if sessionLifetime != "" {
		settings.SessionLifetime, err = strconv.Atoi(sessionLifetime)
		if err != nil {
			return nil, fmt.Errorf(i18n.T(c.Request().Context(), "settings.refresh_invalid"))
		}

		if settings.SessionLifetime <= 0 {
			return nil, fmt.Errorf(i18n.T(c.Request().Context(), "settings.refresh_invalid"))
		}
	}

	return &settings, nil
}