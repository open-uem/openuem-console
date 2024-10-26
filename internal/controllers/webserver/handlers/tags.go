package handlers

import (
	"strconv"

	"github.com/doncicuto/openuem-console/internal/views/admin_views"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/labstack/echo/v4"
)

func (h *Handler) TagManager(c echo.Context) error {
	var err error

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	p.NItems, err = h.Model.CountAllTags()
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if c.Request().Method == "POST" {
		tagId := c.FormValue("tagId")
		tag := c.FormValue("tag")
		description := c.FormValue("description")
		color := c.FormValue("color")

		if tag != "" && color != "" {
			if tagId == "" {
				if err := h.Model.NewTag(tag, description, color); err != nil {
					return renderError(c, partials.ErrorMessage(err.Error(), false))
				}
			} else {
				id, err := strconv.Atoi(tagId)
				if err != nil {
					return renderError(c, partials.ErrorMessage(err.Error(), false))
				}
				if err := h.Model.UpdateTag(id, tag, description, color); err != nil {
					return renderError(c, partials.ErrorMessage(err.Error(), false))
				}
			}

		}
	}

	if c.Request().Method == "DELETE" {
		tagId := c.FormValue("tagId")
		if tagId == "" {
			return renderError(c, partials.ErrorMessage("tag cannot be empty", false))
		}

		id, err := strconv.Atoi(tagId)
		if err != nil {
			return renderError(c, partials.ErrorMessage(err.Error(), false))
		}

		if err := h.Model.DeleteTag(id); err != nil {
			return renderError(c, partials.ErrorMessage(err.Error(), false))
		}
	}

	tags, err := h.Model.GetTagsByPage(p)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return renderView(c, admin_views.TagsIndex(" | Tags", admin_views.Tags(c, p, tags)))
}