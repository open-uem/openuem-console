package handlers

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/models"
	model "github.com/open-uem/openuem-console/internal/models/servers"
	"github.com/open-uem/openuem-console/internal/views/admin_views"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

var UpdateChannels = []string{"stable", "devel", "testing"}

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

		if settings.UpdateChannel != "" {
			if err := h.Model.UpdateOpenUEMChannel(settings.ID, settings.UpdateChannel); err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), true))
			}
		}

		if settings.AgentFrequency != 0 {
			// Get current frequency
			currentFrequency, err := h.Model.GetDefaultAgentFrequency()
			if err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), true))
			}

			// Get winget frequency
			wingetFrequency, err := h.Model.GetDefaultWingetFrequency()
			if err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), true))
			}

			if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), true))
			}

			config := openuem_nats.Config{}
			config.AgentFrequency = settings.AgentFrequency
			config.WinGetFrequency = wingetFrequency
			data, err := json.Marshal(config)
			if err != nil {
				return err
			}

			if err := h.NATSConnection.Publish("agent.newconfig", data); err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.agent_frequency_error"), true))
			}

			if err := h.Model.UpdateAgentFrequency(settings.ID, settings.AgentFrequency); err != nil {
				// Rollback communication
				config := openuem_nats.Config{}
				config.AgentFrequency = currentFrequency
				config.WinGetFrequency = wingetFrequency
				data, err := json.Marshal(config)
				if err != nil {
					return err
				}

				if err := h.NATSConnection.Publish("agent.newconfig", data); err != nil {
					return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.agent_frequency_error"), true))
				}
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.agent_frequency_could_not_be_saved"), true))
			}

			return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "settings.agent_frequency_success")))
		}

		if c.FormValue("request-pin") != "" {
			if err := h.Model.UpdateRequestVNCPIN(settings.ID, settings.RequestVNCPIN); err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.request_pin_could_not_be_saved"), true))
			}
		}

		if c.FormValue("admitted-agent-tag") != "" {
			if settings.Tag == -1 {
				if err := h.Model.RemoveAdmittedTag(settings.ID); err != nil {
					return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.add_tag_admitted_could_not_be_cleared"), true))
				}
			} else {
				if err := h.Model.AddAdmittedTag(settings.ID, settings.Tag); err != nil {
					return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.add_tag_admitted_could_not_be_saved"), true))
				}
			}
		}

		if settings.WinGetFrequency != 0 {
			// Get current frequency
			currentFrequency, err := h.Model.GetDefaultAgentFrequency()
			if err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), true))
			}

			// Get winget frequency
			wingetFrequency, err := h.Model.GetDefaultWingetFrequency()
			if err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), true))
			}
			if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), true))
			}

			config := openuem_nats.Config{}
			config.AgentFrequency = currentFrequency
			config.WinGetFrequency = settings.WinGetFrequency
			data, err := json.Marshal(config)
			if err != nil {
				return err
			}

			if err := h.NATSConnection.Publish("agent.newconfig", data); err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.winget_configure_frequency_error"), true))
			}

			if err := h.Model.UpdateWingetFrequency(settings.ID, settings.WinGetFrequency); err != nil {
				// Rollback communication
				config := openuem_nats.Config{}
				config.AgentFrequency = currentFrequency
				config.WinGetFrequency = wingetFrequency
				data, err := json.Marshal(config)
				if err != nil {
					return err
				}

				if err := h.NATSConnection.Publish("agent.newconfig", data); err != nil {
					return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.winget_configure_frequency_error"), true))
				}
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.winget_configure_frequency_could_not_be_saved"), true))
			}

			return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "settings.winget_configure_frequency_success")))
		}

		return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "settings.saved")))
	}

	settings, err := h.Model.GetGeneralSettings()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	agentsExists, err := h.Model.AgentsExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	allTags, err := h.Model.GetAllTags()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, admin_views.GeneralSettingsIndex(" | General Settings", admin_views.GeneralSettings(c, h.SessionManager, h.Version, latestServerRelease.Version, settings, agentsExists, serversExists, allTags)))
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
	updateChannel := c.FormValue("update-channel")
	agentFrequency := c.FormValue("agent-frequency")
	requestPIN := c.FormValue("request-pin")
	admittedTag := c.FormValue("admitted-agent-tag")
	wingetFrequency := c.FormValue("winget-configure-frequency")

	if settingsId == "" {
		return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.id_cannot_be_empty"))
	}

	settings.ID, err = strconv.Atoi(settingsId)
	if err != nil {
		return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.id_invalid"))
	}

	if country != "" {
		if errs := validate.Var(country, "iso3166_1_alpha2"); errs != nil {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.country_invalid"))
		}
		settings.Country = country
	}

	if certYear != "" {
		settings.UserCertYears, err = strconv.Atoi(certYear)
		if err != nil {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.cert_years_invalid"))
		}

		if settings.UserCertYears <= 0 {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.cert_years_invalid"))
		}
	}

	if natsTimeout != "" {
		settings.NATSTimeout, err = strconv.Atoi(natsTimeout)
		if err != nil {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.nats_timeout_invalid"))
		}

		if settings.NATSTimeout <= 0 {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.nats_timeout_invalid"))
		}
	}

	if maxUploadSize != "" {
		if !strings.HasSuffix(maxUploadSize, "M") && !strings.HasSuffix(maxUploadSize, "K") && !strings.HasSuffix(maxUploadSize, "G") {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.max_upload_size_invalid"))
		}
		settings.MaxUploadSize = maxUploadSize
	}

	if refresh != "" {
		settings.Refresh, err = strconv.Atoi(refresh)
		if err != nil {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.refresh_invalid"))
		}

		if settings.Refresh <= 0 {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.refresh_invalid"))
		}
	}

	if sessionLifetime != "" {
		settings.SessionLifetime, err = strconv.Atoi(sessionLifetime)
		if err != nil {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.refresh_invalid"))
		}

		if settings.SessionLifetime <= 0 {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.refresh_invalid"))
		}
	}

	if updateChannel != "" {
		if !slices.Contains(UpdateChannels, updateChannel) {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.upload_channel_invalid"))
		}
		settings.UpdateChannel = updateChannel
	}

	if agentFrequency != "" {
		settings.AgentFrequency, err = strconv.Atoi(agentFrequency)
		if err != nil {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.agent_frequency_invalid"))
		}

		if settings.AgentFrequency <= 0 {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.agent_frequency_invalid"))
		}
	}

	if requestPIN != "" {
		settings.RequestVNCPIN, err = strconv.ParseBool(requestPIN)
		if err != nil {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.request_pin_invalid"))
		}
	}

	if admittedTag != "" {
		settings.Tag, err = strconv.Atoi(admittedTag)
		if err != nil {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.add_tag_invalid"))
		}
	}

	if wingetFrequency != "" {
		settings.WinGetFrequency, err = strconv.Atoi(wingetFrequency)
		if err != nil {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.winget_configure_frequency_invalid"))
		}

		// Min WinGetFrequency is 30
		if settings.WinGetFrequency < 0 {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.winget_configure_frequency_invalid"))
		}
	}

	return &settings, nil
}
