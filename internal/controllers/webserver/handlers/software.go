package handlers

import (
	"log"

	"github.com/doncicuto/openuem-console/internal/models"
	"github.com/doncicuto/openuem-console/internal/views/filters"
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
	filterByName := c.FormValue("filterByAppName")
	filterByPublisher := c.FormValue("filterByAppPublisher")

	f := filters.ApplicationsFilter{}
	if filterByName != "" {
		f.AppName = filterByName
	}
	if filterByPublisher != "" {
		f.Vendor = filterByPublisher
	}

	apps, err = h.Model.GetAppsByPage(p, f)
	if err != nil {
		return RenderView(c, software_views.SoftwareIndex(" | Software", partials.Error(err.Error(), "Software", "/software")))
	}

	p.NItems, err = h.Model.CountAllApps(f)
	if err != nil {
		return RenderView(c, software_views.SoftwareIndex(" | Software", partials.Error(err.Error(), "Software", "/software")))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	return RenderView(c, software_views.SoftwareIndex(" | Software", software_views.Software(c, p, apps, f, refreshTime)))
}
