package handlers

import (
	"strconv"

	"github.com/labstack/echo/v4"
	model "github.com/open-uem/openuem-console/internal/models/servers"
	"github.com/open-uem/openuem-console/internal/views/admin_views"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (h *Handler) OrgMetadataManager(c echo.Context) error {
	var err error

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

	if c.Request().Method == "POST" {
		orgMetadataId := c.FormValue("orgMetadataId")
		name := c.FormValue("name")
		description := c.FormValue("description")

		if name != "" {
			if orgMetadataId == "" {
				if err := h.Model.NewOrgMetadata(name, description); err != nil {
					return RenderError(c, partials.ErrorMessage(err.Error(), false))
				}
			} else {
				id, err := strconv.Atoi(orgMetadataId)
				if err != nil {
					return RenderError(c, partials.ErrorMessage(err.Error(), false))
				}
				if err := h.Model.UpdateOrgMetadata(id, name, description); err != nil {
					return RenderError(c, partials.ErrorMessage(err.Error(), false))
				}
			}
		}
	}

	if c.Request().Method == "DELETE" {
		orgMetadataId := c.FormValue("orgMetadataId")
		if orgMetadataId == "" {
			return RenderError(c, partials.ErrorMessage("tag cannot be empty", false))
		}

		id, err := strconv.Atoi(orgMetadataId)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}

		if err := h.Model.DeleteOrgMetadata(id); err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}
	}

	p.NItems, err = h.Model.CountAllOrgMetadata()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	data, err := h.Model.GetOrgMetadataByPage(p)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	agentsExists, err := h.Model.AgentsExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, admin_views.OrgMetadataIndex(" | Tags", admin_views.OrgMetadata(c, p, h.SessionManager, h.Version, latestServerRelease.Version, data, agentsExists, serversExists)))
}
