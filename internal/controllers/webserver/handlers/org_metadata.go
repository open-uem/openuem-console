package handlers

import (
	"strconv"

	"github.com/doncicuto/openuem-console/internal/views/admin_views"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/labstack/echo/v4"
)

func (h *Handler) OrgMetadataManager(c echo.Context) error {
	var err error
	comesFromDialog := false

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

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
		comesFromDialog = true
	}

	p.NItems, err = h.Model.CountAllOrgMetadata()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	data, err := h.Model.GetOrgMetadataByPage(p)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return RenderView(c, admin_views.OrgMetadataIndex(" | Tags", admin_views.OrgMetadata(c, p, h.SessionManager, data, comesFromDialog)))
}
