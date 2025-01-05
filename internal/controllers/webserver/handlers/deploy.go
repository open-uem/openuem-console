package handlers

import (
	"encoding/json"
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
	return RenderView(c, deploy_views.DeployIndex("| Deploy", deploy_views.Deploy(h.SessionManager, true, "")))
}

func (h *Handler) DeployUninstall(c echo.Context) error {
	return RenderView(c, deploy_views.DeployIndex("| Deploy", deploy_views.Deploy(h.SessionManager, false, "")))
}

func (h *Handler) SearchPackagesAction(c echo.Context, install bool) error {
	var err error

	search := c.FormValue("filterByAppName")
	if search == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "install.search_empty_error"), true))
	}

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	// Default sort
	if p.SortBy == "" {
		p.SortBy = "name"
		p.SortOrder = "asc"
	}

	packages, err := models.SearchPackages(search, p)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	p.NItems, err = models.CountPackages(search)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, deploy_views.SearchPacketResult(install, packages, c, p))
}

func (h *Handler) SelectPackageDeployment(c echo.Context) error {
	var err error

	packageId := c.FormValue("filterByPackageId")
	packageName := c.FormValue("filterByPackageName")
	installParam := c.FormValue("filterByInstallationType")

	if packageId == "" || packageName == "" || installParam == "" {
		return RenderError(c, partials.ErrorMessage("required params not found", true))
	}

	f := filters.AgentFilter{}

	nSelectedItems := c.FormValue("filterBySelectedItems")
	f.SelectedItems, err = strconv.Atoi(nSelectedItems)
	if err != nil {
		f.SelectedItems = 0
	}

	tmpAllAgents := []string{}
	allAgents, err := h.Model.GetAdmittedAgents(f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	for _, a := range allAgents {
		tmpAllAgents = append(tmpAllAgents, "\""+a.ID+"\"")
	}
	f.SelectedAllAgents = "[" + strings.Join(tmpAllAgents, ",") + "]"

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	p.SortBy = "hostname"
	p.NItems, err = h.Model.CountAllAgents(filters.AgentFilter{}, true)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	agents, err := h.Model.GetAgentsByPage(p, filters.AgentFilter{}, true)
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

	return RenderView(c, deploy_views.DeployIndex("", deploy_views.SelectPackageDeployment(c, p, f, h.SessionManager, packageId, packageName, agents, install, refreshTime)))
}

func (h *Handler) DeployPackageToSelectedAgents(c echo.Context) error {
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
			Repository:  "winget",
			Action:      "install",
		}

		actionBytes, err := json.Marshal(action)
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

			if err := h.Model.SaveDeployInfo(&action); err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), true))
			}
		} else {
			if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
			}

			if err := h.NATSConnection.Publish("agent.uninstallpackage."+agent, actionBytes); err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), true))
			}

			if err := h.Model.SaveDeployInfo(&action); err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), true))
			}
		}
	}

	if install {
		return RenderView(c, deploy_views.DeployIndex("| Deploy", deploy_views.Deploy(h.SessionManager, true, i18n.T(c.Request().Context(), "install.requested"))))
	} else {
		return RenderView(c, deploy_views.DeployIndex("| Deploy", deploy_views.Deploy(h.SessionManager, false, i18n.T(c.Request().Context(), "uninstall.requested"))))
	}
}
