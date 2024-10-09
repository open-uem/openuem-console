package handlers

import (
	"encoding/json"
	"strconv"

	models "github.com/doncicuto/openuem-console/internal/models/winget"
	"github.com/doncicuto/openuem-console/internal/views/agents_views"
	"github.com/doncicuto/openuem-console/internal/views/deploy_views"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/doncicuto/openuem_nats"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
)

func (h *Handler) DeployInstall(c echo.Context) error {
	return renderView(c, deploy_views.DeployIndex("| Deploy", deploy_views.Deploy(true)))
}

func (h *Handler) DeployUninstall(c echo.Context) error {
	return renderView(c, deploy_views.DeployIndex("| Deploy", deploy_views.Deploy(false)))
}

func (h *Handler) SearchPackagesAction(c echo.Context, install bool) error {
	var err error
	search := c.FormValue("search")
	if search == "" {
		return renderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "install.search_empty_error"), true))
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
		return renderError(c, partials.ErrorMessage(err.Error(), true))
	}

	p.NItems, err = models.CountPackages(search)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return renderView(c, deploy_views.SearchPacketResult(install, packages, c, p))
}

func (h *Handler) SelectPackageDeployment(c echo.Context) error {
	var err error

	packageId := c.QueryParam("packageId")
	packageName := c.QueryParam("packageName")
	installParam := c.QueryParam("install")

	if packageId == "" || packageName == "" || installParam == "" {
		return renderError(c, partials.ErrorMessage("required params not found", true))
	}

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	p.SortBy = "hostname"
	p.NItems, err = h.Model.CountAllAgents(agents_views.AgentFilter{})
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), true))
	}

	agents, err := h.Model.GetAgentsByPage(p, agents_views.AgentFilter{})
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), true))
	}

	install, err := strconv.ParseBool(installParam)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return renderView(c, deploy_views.DeployIndex("", deploy_views.SelectPackageDeployment(c, p, packageId, packageName, agents, install)))
}

func (h *Handler) DeployPackageToSelectedAgents(c echo.Context) error {
	checkedItems := c.FormValue("checkedItems")
	packageId := c.FormValue("packageId")
	packageName := c.FormValue("packageName")
	installParam := c.FormValue("install")
	var agents []string

	err := json.Unmarshal([]byte(checkedItems), &agents)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), true))
	}

	install, err := strconv.ParseBool(installParam)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), true))
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
			return renderError(c, partials.ErrorMessage(err.Error(), true))
		}

		if install {
			if err := h.MessageServer.Connection.Publish("agent.installpackage."+agent, actionBytes); err != nil {
				return renderError(c, partials.ErrorMessage(err.Error(), true))
			}
		} else {
			if err := h.MessageServer.Connection.Publish("agent.uninstallpackage."+agent, actionBytes); err != nil {
				return renderError(c, partials.ErrorMessage(err.Error(), true))
			}
		}
	}

	if install {
		return renderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "install.requested")))
	} else {
		return renderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "uninstall.requested")))
	}
}
