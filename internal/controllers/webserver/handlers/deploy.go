package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views/deploy_views"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

// agentPackageSubjects are the NATS subjects used to ask an agent to install, update or uninstall a package
var agentPackageSubjects = map[string]string{
	deploy_views.ActionInstall:   "agent.installpackage.",
	deploy_views.ActionUpdate:    "agent.updatepackage.",
	deploy_views.ActionUninstall: "agent.uninstallpackage.",
}

func (h *Handler) DeployInstall(c echo.Context) error {
	return h.DeployAction(c, deploy_views.ActionInstall)
}

func (h *Handler) DeployUninstall(c echo.Context) error {
	return h.DeployAction(c, deploy_views.ActionUninstall)
}

func (h *Handler) DeployAction(c echo.Context, action string) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	return RenderView(c, deploy_views.DeployIndex("| Deploy", deploy_views.Deploy(c, action, "", commonInfo), commonInfo))
}

func (h *Handler) SearchPackagesAction(c echo.Context, action string) error {
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

	useBrew, err := h.Model.GetDefaultUseBrew(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if useBrew {
		allSources = append(allSources, "brew")
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
		if useBrew {
			filteredSources = append(allSources, "brew")
		}
	}

	f := filters.DeployPackageFilter{}
	f.Sources = filteredSources

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.NewPaginationAndSort(itemsPerPage)
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"), itemsPerPage)

	// Default sort
	if p.SortBy == "" {
		p.SortBy = "name"
		p.SortOrder = "asc"
	}

	packages, err := h.Model.SearchPackages(search, p, f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	p.NItems, err = h.Model.CountPackages(search, f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, deploy_views.SearchPacketResult(action, packages, c, p, f, allSources, itemsPerPage, commonInfo))
}

func (h *Handler) SelectPackageDeployment(c echo.Context) error {
	action, ok := deploy_views.ParseAction(c.FormValue("filterByInstallationType"))
	if !ok {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "packages.required_params_missing"), false))
	}

	return h.SelectAgentsForDeployment(c, action)
}

// SelectAgentsForDeployment shows the agents where a package can be installed, updated or uninstalled.
// When every package must be updated, no package is selected and agents with any OS are shown.
func (h *Handler) SelectAgentsForDeployment(c echo.Context, action string) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	packageId := c.FormValue("filterByPackageId")
	packageName := c.FormValue("filterByPackageName")
	packageBranch := c.FormValue("filterByPackageBranch")
	packageBrewType := c.FormValue("filterByPackageBrewType")
	packageVerified := c.FormValue("filterByPackageVerified")
	source := c.FormValue("filterBySource")

	if action != deploy_views.ActionUpdateAll && (packageId == "" || packageName == "") {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "packages.required_params_missing"), false))
	}

	if packageBrewType != "" && !slices.Contains([]string{"cask", "formula"}, packageBrewType) {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "packages.wrong_brew_type"), false))
	}

	isPackageVerified := false
	if packageVerified != "" {
		isPackageVerified, err = strconv.ParseBool(packageVerified)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "packages.invalid_verified"), false))
		}
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
		f.AgentOSVersions = []string{"ubuntu", "neon", "debian", "opensuse-leap", "linuxmint", "fedora", "manjaro", "arch", "almalinux", "rocky"}
	case "brew":
		f.AgentOSVersions = []string{"macOS"}
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

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.NewPaginationAndSort(itemsPerPage)
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"), itemsPerPage)

	p.SortBy = "nickname"
	p.NItems, err = h.Model.CountAllAgents(f, true, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	agents, err := h.Model.GetAgentsByPage(p, f, true, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	return RenderView(c, deploy_views.DeployIndex("", deploy_views.SelectPackageDeployment(c, p, f, packageId, packageName, source, packageBranch, packageBrewType, isPackageVerified, agents, action, refreshTime, itemsPerPage, commonInfo), commonInfo))
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
	packageBranch := c.FormValue("filterByPackageBranch")
	packageBrewType := c.FormValue("filterByPackageBrewType")
	packageVerified := c.FormValue("filterByPackageVerified")

	if packageBrewType != "" && !slices.Contains([]string{"cask", "formula"}, packageBrewType) {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "packages.wrong_brew_type"), false))
	}

	isPackageVerified := false
	if packageVerified != "" {
		isPackageVerified, err = strconv.ParseBool(packageVerified)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "packages.invalid_verified"), false))
		}
	}

	agents := []string{}
	for _, agent := range strings.Split(checkedItems, ",") {
		if agent != "" {
			agents = append(agents, agent)
		}
	}
	if len(agents) == 0 {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_selected_agents_to_deploy"), true))
	}

	action, ok := deploy_views.ParseAction(c.FormValue("filterByInstallationType"))
	if !ok {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "packages.required_params_missing"), true))
	}

	if action == deploy_views.ActionUpdateAll {
		return h.UpdateAllPackagesOnAgents(c, agents, commonInfo)
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
	}

	for _, agent := range agents {
		deployAction := openuem_nats.DeployAction{
			AgentId:         agent,
			Action:          action,
			PackageId:       packageId,
			PackageName:     packageName,
			PackageBranch:   packageBranch,
			PackageBrewType: packageBrewType,
			PackageVerified: isPackageVerified,
		}

		actionBytes, err := json.Marshal(deployAction)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), true))
		}

		deploymentFailed, err := h.Model.DeploymentFailed(agent, packageId, commonInfo)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), true))
		}

		if err := h.NATSConnection.Publish(agentPackageSubjects[action]+agent, actionBytes); err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), true))
		}

		if err := h.Model.SaveDeployInfo(&deployAction, deploymentFailed, commonInfo); err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), true))
		}
	}

	return RenderView(c, deploy_views.DeployIndex("| Deploy", deploy_views.Deploy(c, action, i18n.T(c.Request().Context(), action+".requested"), commonInfo), commonInfo))
}
