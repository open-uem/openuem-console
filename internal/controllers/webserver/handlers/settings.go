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
	"github.com/open-uem/openuem-console/internal/views/admin_views"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

var UpdateChannels = []string{"stable", "devel", "testing"}

func (h *Handler) GeneralSettings(c echo.Context) error {
	var err error
	successMessage := ""

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

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

		if settings.ItemsPerPage != 0 {
			if err := h.Model.UpdateDefaultItemsPerPageSetting(settings.ID, settings.ItemsPerPage); err != nil {
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
			return h.ChangeAgentFrequency(c, settings)
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

		if c.FormValue("use-winget") != "" {
			if err := h.Model.UpdateUseWinget(settings.ID, settings.UseWinget); err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.use_winget_could_not_be_saved"), true))
			}
		}

		if c.FormValue("use-flatpak") != "" {
			if err := h.Model.UpdateUseFlatpak(settings.ID, settings.UseFlatpak); err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.use_flatpak_could_not_be_saved"), true))
			}
		}

		if c.FormValue("use-brew") != "" {
			if err := h.Model.UpdateUseBrew(settings.ID, settings.UseBrew); err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.use_brew_could_not_be_saved"), true))
			}
		}

		if settings.WinGetFrequency != 0 {
			return h.ChangeWingetFrequency(c, settings)
		}

		if c.FormValue("disable-sftp") != "" {
			return h.ChangeSFTPSetting(c, settings)
		}

		if c.FormValue("disable-remote-assistance") != "" {
			return h.ChangeRemoteAssistanceSetting(c, settings)
		}

		if c.FormValue("detect-remote-agents") != "" {
			if err := h.Model.UpdateDetectRemoteAgents(settings.ID, settings.DetectRemoteAgents); err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.detect_remote_agents_could_not_be_saved"), true))
			}
		}

		if c.FormValue("auto-admit-agents") != "" {
			if err := h.Model.UpdateAutoAdmitAgents(settings.ID, settings.AutoAdmitAgents); err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.auto_admit_agents_could_not_be_saved"), true))
			}
		}

		successMessage = i18n.T(c.Request().Context(), "settings.saved")
	}

	settings, err := h.Model.GetGeneralSettings(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	agentsExists, err := h.Model.AgentsExists(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	allTags, err := h.Model.GetAllTags(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return RenderView(c, admin_views.GeneralSettingsIndex(" | General Settings", admin_views.GeneralSettings(c, settings, agentsExists, serversExists, allTags, commonInfo, h.GetAdminTenantName(commonInfo), successMessage), commonInfo))
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
	useWinget := c.FormValue("use-winget")
	useFlatpak := c.FormValue("use-flatpak")
	useBrew := c.FormValue("use-brew")
	disableSFTP := c.FormValue("disable-sftp")
	disableRemoteAssistance := c.FormValue("disable-remote-assistance")
	detectRemoteAgents := c.FormValue("detect-remote-agents")
	autoAdmitAgents := c.FormValue("auto-admit-agents")
	netbird := c.FormValue("netbird")
	itemsPerPage := c.FormValue("items-per-page")

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

	if itemsPerPage != "" {
		settings.ItemsPerPage, err = strconv.Atoi(itemsPerPage)
		if err != nil {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.items_per_page_invalid"))
		}

		if settings.ItemsPerPage <= 0 {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.items_per_page_invalid"))
		}
	}

	if sessionLifetime != "" {
		settings.SessionLifetime, err = strconv.Atoi(sessionLifetime)
		if err != nil {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.session_lifetime_invalid"))
		}

		if settings.SessionLifetime <= 0 {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.session_lifetime_invalid"))
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
		if settings.WinGetFrequency < 30 {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.winget_configure_frequency_invalid"))
		}
	}

	if useWinget != "" {
		settings.UseWinget, err = strconv.ParseBool(useWinget)
		if err != nil {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.use_winget_invalid"))
		}
	}

	if useFlatpak != "" {
		settings.UseFlatpak, err = strconv.ParseBool(useFlatpak)
		if err != nil {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.use_flatpak_invalid"))
		}
	}

	if useBrew != "" {
		settings.UseBrew, err = strconv.ParseBool(useBrew)
		if err != nil {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.use_brew_invalid"))
		}
	}

	if disableSFTP != "" {
		settings.SFTPDisabled, err = strconv.ParseBool(disableSFTP)
		if err != nil {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.disable_sftp_invalid"))
		}
	}

	if disableRemoteAssistance != "" {
		settings.RemoteAssistanceDisabled, err = strconv.ParseBool(disableRemoteAssistance)
		if err != nil {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.disable_remote_assistance_invalid"))
		}
	}

	if detectRemoteAgents != "" {
		settings.DetectRemoteAgents, err = strconv.ParseBool(detectRemoteAgents)
		if err != nil {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.detect_remote_agents_invalid"))
		}
	}

	if autoAdmitAgents != "" {
		settings.AutoAdmitAgents, err = strconv.ParseBool(autoAdmitAgents)
		if err != nil {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.auto_admit_agents_invalid"))
		}
	}

	if autoAdmitAgents != "" {
		settings.AutoAdmitAgents, err = strconv.ParseBool(autoAdmitAgents)
		if err != nil {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.auto_admit_agents_invalid"))
		}
	}

	if netbird != "" {
		settings.NetBird, err = strconv.ParseBool(netbird)
		if err != nil {
			return nil, fmt.Errorf("%s", i18n.T(c.Request().Context(), "settings.netbird_invalid"))
		}
	}

	return &settings, nil
}

func (h *Handler) ChangeAgentFrequency(c echo.Context, settings *models.GeneralSettings) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	// Get current frequency
	currentFrequency, err := h.Model.GetDefaultAgentFrequency(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	// Get winget frequency
	wingetFrequency, err := h.Model.GetDefaultWingetFrequency(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	// Get SFTPDisabled
	sftpDisabled, err := h.Model.GetDefaultSFTPDisabled(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	// Get RemoteAssistanceDisabled
	remoteAssistanceDisabled, err := h.Model.GetDefaultRemoteAssistanceDisabled(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), true))
	}

	config := openuem_nats.Config{}
	config.AgentFrequency = settings.AgentFrequency
	config.WinGetFrequency = wingetFrequency
	config.SFTPDisabled = sftpDisabled
	config.RemoteAssistanceDisabled = remoteAssistanceDisabled
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
		config.SFTPDisabled = sftpDisabled
		config.RemoteAssistanceDisabled = remoteAssistanceDisabled
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

func (h *Handler) ChangeWingetFrequency(c echo.Context, settings *models.GeneralSettings) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	// Get current frequency
	currentFrequency, err := h.Model.GetDefaultAgentFrequency(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	// Get winget frequency
	wingetFrequency, err := h.Model.GetDefaultWingetFrequency(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	// Get SFTP disabled
	sftpDisabled, err := h.Model.GetDefaultSFTPDisabled(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	// Get RemoteAssistanceDisabled
	remoteAssistanceDisabled, err := h.Model.GetDefaultRemoteAssistanceDisabled(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), true))
	}

	config := openuem_nats.Config{}
	config.AgentFrequency = currentFrequency
	config.WinGetFrequency = settings.WinGetFrequency
	config.SFTPDisabled = sftpDisabled
	config.RemoteAssistanceDisabled = remoteAssistanceDisabled
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
		config.SFTPDisabled = sftpDisabled
		config.RemoteAssistanceDisabled = remoteAssistanceDisabled
		data, err := json.Marshal(config)
		if err != nil {
			return err
		}

		if err := h.NATSConnection.Publish("agent.newconfig", data); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.winget_configure_frequency_error"), true))
		}
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.winget_configure_frequency_could_not_be_saved"), true))
	}

	return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "settings.disable_sftp_success")))
}

func (h *Handler) ChangeSFTPSetting(c echo.Context, settings *models.GeneralSettings) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	// Get current frequency
	currentFrequency, err := h.Model.GetDefaultAgentFrequency(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	// Get winget frequency
	wingetFrequency, err := h.Model.GetDefaultWingetFrequency(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	// Get SFTP disabled settings
	sftpDisabled, err := h.Model.GetDefaultSFTPDisabled(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	// Get RemoteAssistanceDisabled
	remoteAssistanceDisabled, err := h.Model.GetDefaultRemoteAssistanceDisabled(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), true))
	}

	config := openuem_nats.Config{}
	config.AgentFrequency = currentFrequency
	config.WinGetFrequency = wingetFrequency
	config.SFTPDisabled = settings.SFTPDisabled
	config.RemoteAssistanceDisabled = remoteAssistanceDisabled
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	if err := h.NATSConnection.Publish("agent.newconfig", data); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.disable_sftp_error"), true))
	}

	if err := h.Model.UpdateSFTPDisabled(settings.ID, settings.SFTPDisabled); err != nil {
		// Rollback communication
		config := openuem_nats.Config{}
		config.AgentFrequency = currentFrequency
		config.WinGetFrequency = wingetFrequency
		config.SFTPDisabled = sftpDisabled
		config.RemoteAssistanceDisabled = remoteAssistanceDisabled
		data, err := json.Marshal(config)
		if err != nil {
			return err
		}

		if err := h.NATSConnection.Publish("agent.newconfig", data); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.disable_sftp_error"), true))
		}
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.disable_sftp_could_not_be_saved"), true))
	}

	// Apply change to all agents
	if err := h.Model.UpdateSFTPServiceToAllAgents(!settings.SFTPDisabled, commonInfo); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.disable_sftp_to_all"), true))
	}

	return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "settings.disable_sftp_success")))
}

func (h *Handler) ChangeRemoteAssistanceSetting(c echo.Context, settings *models.GeneralSettings) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	// Get current frequency
	currentFrequency, err := h.Model.GetDefaultAgentFrequency(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	// Get winget frequency
	wingetFrequency, err := h.Model.GetDefaultWingetFrequency(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	// Get SFTP disabled settings
	sftpDisabled, err := h.Model.GetDefaultSFTPDisabled(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	// Get RemoteAssistanceDisabled
	remoteAssistanceDisabled, err := h.Model.GetDefaultRemoteAssistanceDisabled(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), true))
	}

	config := openuem_nats.Config{}
	config.AgentFrequency = currentFrequency
	config.WinGetFrequency = wingetFrequency
	config.SFTPDisabled = sftpDisabled
	config.RemoteAssistanceDisabled = settings.RemoteAssistanceDisabled
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	if err := h.NATSConnection.Publish("agent.newconfig", data); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.disable_remote_assistance_error"), true))
	}

	if err := h.Model.UpdateRemoteAssistanceDisabled(settings.ID, settings.RemoteAssistanceDisabled); err != nil {
		// Rollback communication
		config := openuem_nats.Config{}
		config.AgentFrequency = currentFrequency
		config.WinGetFrequency = wingetFrequency
		config.SFTPDisabled = sftpDisabled
		config.RemoteAssistanceDisabled = remoteAssistanceDisabled
		data, err := json.Marshal(config)
		if err != nil {
			return err
		}

		if err := h.NATSConnection.Publish("agent.newconfig", data); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.disable_remote_assistance_error"), true))
		}
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.disable_remote_assistance_could_not_be_saved"), true))
	}

	// Apply change to all agents
	if err := h.Model.UpdateRemoteAssistanceToAllAgents(!settings.RemoteAssistanceDisabled, commonInfo); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.disable_remote_assistance_to_all"), true))
	}

	return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "settings.disable_remote_assistance_success")))
}

func (h *Handler) ApplyGlobalSettings(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	if commonInfo.TenantID == "-1" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.tenant_cannot_be_empty"), true))
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.could_not_convert_to_int", commonInfo.TenantID), true))
	}

	if err := h.Model.ApplyGlobalSettings(tenantID); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_apply_global_settings", err.Error()), true))
	}

	settings, err := h.Model.GetGeneralSettings(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	agentsExists, err := h.Model.AgentsExists(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	allTags, err := h.Model.GetAllTags(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return RenderView(c, admin_views.GeneralSettingsIndex(" | General Settings", admin_views.GeneralSettings(c, settings, agentsExists, serversExists, allTags, commonInfo, h.GetAdminTenantName(commonInfo), i18n.T(c.Request().Context(), "settings.saved")), commonInfo))
}
