package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	openuem_nats "github.com/open-uem/nats"
	models "github.com/open-uem/openuem-console/internal/models/winget"
	"github.com/open-uem/openuem-console/internal/views/deploy_views"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (h *Handler) DeployInstall(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	return RenderView(c, deploy_views.DeployIndex("| Deploy", deploy_views.Deploy(c, true, "", commonInfo), commonInfo))
}

func (h *Handler) DeployUninstall(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	return RenderView(c, deploy_views.DeployIndex("| Deploy", deploy_views.Deploy(c, false, "", commonInfo), commonInfo))
}

func (h *Handler) SearchPackagesAction(c echo.Context, install bool) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	search := c.FormValue("filterByAppName")
	if search == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "install.search_empty_error"), true))
	}

	allSources := []string{}

	useWinget, err := h.Model.GetDefaultUseWinget(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if useWinget {
		allSources = append(allSources, "winget")
	}

	useFlatpak, err := h.Model.GetDefaultUseFlatpak(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if useFlatpak {
		allSources = append(allSources, "flatpak")
	}

	filteredSources := []string{}
	for index := range allSources {
		value := c.FormValue(fmt.Sprintf("filterBySource%d", index))
		if value != "" {
			filteredSources = append(filteredSources, value)
		}
	}

	if len(filteredSources) == 0 {
		if useWinget {
			filteredSources = append(allSources, "winget")
		}
		if useFlatpak {
			filteredSources = append(allSources, "flatpak")
		}
	}

	f := filters.DeployPackageFilter{}
	f.Sources = filteredSources

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

	// Default sort
	if p.SortBy == "" {
		p.SortBy = "name"
		p.SortOrder = "asc"
	}

	packages, err := models.SearchPackages(search, p, h.CommonFolder, f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	p.NItems, err = models.CountPackages(search, h.CommonFolder, f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, deploy_views.SearchPacketResult(install, packages, c, p, f, allSources, commonInfo))
}

func (h *Handler) SelectPackageDeployment(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	packageId := c.FormValue("filterByPackageId")
	packageName := c.FormValue("filterByPackageName")
	installParam := c.FormValue("filterByInstallationType")
	source := c.FormValue("filterBySource")

	if packageId == "" || packageName == "" || installParam == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "packages.required_params_missing"), false))
	}

	f := filters.AgentFilter{}

	nSelectedItems := c.FormValue("filterBySelectedItems")
	f.SelectedItems, err = strconv.Atoi(nSelectedItems)
	if err != nil {
		f.SelectedItems = 0
	}

	switch source {
	case "winget":
		f.AgentOSVersions = []string{"windows"}
	case "flatpak":
		f.AgentOSVersions = []string{"ubuntu", "debian", "opensuse-leap", "linuxmint", "fedora", "manjaro", "arch", "almalinux"}
	}

	tmpAllAgents := []string{}
	allAgents, err := h.Model.GetAllAgents(f, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	for _, a := range allAgents {
		tmpAllAgents = append(tmpAllAgents, "\""+a.ID+"\"")
	}
	f.SelectedAllAgents = "[" + strings.Join(tmpAllAgents, ",") + "]"

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

	p.SortBy = "hostname"
	p.NItems, err = h.Model.CountAllAgents(filters.AgentFilter{}, true, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	agents, err := h.Model.GetAgentsByPage(p, f, true, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	install, err := strconv.ParseBool(installParam)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	return RenderView(c, deploy_views.DeployIndex("", deploy_views.SelectPackageDeployment(c, p, f, packageId, packageName, agents, install, refreshTime, commonInfo), commonInfo))
}

func (h *Handler) DeployPackageToSelectedAgents(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	checkedItems := c.FormValue("selectedAgents")
	packageId := c.FormValue("filterByPackageId")
	packageName := c.FormValue("filterByPackageName")
	installParam := c.FormValue("filterByInstallationType")

	agents := strings.Split(checkedItems, ",")
	if len(agents) == 0 {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_selected_agents_to_deploy"), true))
	}

	install, err := strconv.ParseBool(installParam)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	for _, agent := range agents {
		action := openuem_nats.DeployAction{
			AgentId:     agent,
			PackageId:   packageId,
			PackageName: packageName,
			// Repository:  "winget",
		}

		if install {
			action.Action = "install"
		} else {
			action.Action = "uninstall"
		}

		actionBytes, err := json.Marshal(action)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), true))
		}

		deploymentFailed, err := h.Model.DeploymentFailed(agent, packageId, commonInfo)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), true))
		}

		if install {
			if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
			}

			if err := h.NATSConnection.Publish("agent.installpackage."+agent, actionBytes); err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), true))
			}

			if err := h.Model.SaveDeployInfo(&action, deploymentFailed, commonInfo); err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), true))
			}
		} else {
			if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
			}

			if err := h.NATSConnection.Publish("agent.uninstallpackage."+agent, actionBytes); err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), true))
			}

			if err := h.Model.SaveDeployInfo(&action, deploymentFailed, commonInfo); err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), true))
			}
		}
	}

	if install {
		return RenderView(c, deploy_views.DeployIndex("| Deploy", deploy_views.Deploy(c, true, i18n.T(c.Request().Context(), "install.requested"), commonInfo), commonInfo))
	} else {
		return RenderView(c, deploy_views.DeployIndex("| Deploy", deploy_views.Deploy(c, false, i18n.T(c.Request().Context(), "uninstall.requested"), commonInfo), commonInfo))
	}
}
