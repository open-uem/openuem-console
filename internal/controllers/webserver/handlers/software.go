package handlers

import (
	"slices"
	"strconv"

	"github.com/doncicuto/openuem-console/internal/models"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/doncicuto/openuem-console/internal/views/software_views"
	"github.com/labstack/echo/v4"
)

func (h *Handler) Software(c echo.Context) error {
	var err error
	var apps []models.App

	currentPage := 1
	if currentPage, err = strconv.Atoi(c.QueryParam("page")); err != nil {
		currentPage = 1
	}

	// TODO NAPPSPERPAGE set a constant
	pageSize := 5
	if pageSize, err = strconv.Atoi(c.QueryParam("pageSize")); err != nil {
		pageSize = 5
	}
	// TODO set a MAX value for pageSize
	pageSize = min(pageSize, 25)

	sortBy := c.QueryParam("sortBy")
	if !slices.Contains([]string{"name", "publisher", "installations"}, sortBy) {
		sortBy = "name"
	}

	sortOrder := c.QueryParam("sortOrder")
	if !slices.Contains([]string{"asc", "desc"}, sortOrder) {
		sortOrder = "asc"
	}

	nApps := 0
	nApps, err = h.Model.CountAllApps()
	if err != nil {
		return renderView(c, software_views.SoftwareIndex(" | Software", partials.Error(err.Error(), "Software", "/software")))
	}

	apps, err = h.Model.GetAppsByPage(currentPage, pageSize, sortBy, sortOrder)
	if err != nil {
		return renderView(c, software_views.SoftwareIndex(" | Software", partials.Error(err.Error(), "Software", "/software")))
	}

	return renderView(c, software_views.SoftwareIndex(" | Software", software_views.Software(apps, c, currentPage, int(pageSize), nApps, sortBy, sortOrder)))
}
