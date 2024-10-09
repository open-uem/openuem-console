package handlers

import (
	"github.com/doncicuto/openuem-console/internal/models"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/doncicuto/openuem-console/internal/views/software_views"
	"github.com/labstack/echo/v4"
)

func (h *Handler) Software(c echo.Context) error {
	var err error
	var apps []models.App

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	// Default sort
	if p.SortBy == "" {
		p.SortBy = "name"
		p.SortOrder = "asc"
	}

	// Get filters
	focus := c.FormValue("focus")
	filterByName := c.FormValue("filterByName")
	filterByPublisher := c.FormValue("filterByPublisher")

	apps, err = h.Model.GetAppsByPage(p, filterByName, filterByPublisher)
	if err != nil {
		return renderView(c, software_views.SoftwareIndex(" | Software", partials.Error(err.Error(), "Software", "/software")))
	}

	p.NItems, err = h.Model.CountAllApps(filterByName, filterByPublisher)
	if err != nil {
		return renderView(c, software_views.SoftwareIndex(" | Software", partials.Error(err.Error(), "Software", "/software")))
	}

	return renderView(c, software_views.SoftwareIndex(" | Software", software_views.Software(c, p, apps, filterByName, filterByPublisher, focus)))
}
