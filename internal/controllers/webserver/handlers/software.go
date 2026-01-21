package handlers

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/models"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem-console/internal/views/software_views"
)

func (h *Handler) Software(c echo.Context) error {
	var apps []models.App

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

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

	// Get filters
	f, err := h.GetSoftwareFilters(c)
	if err != nil {
		return RenderView(c, software_views.SoftwareIndex(" | Software", partials.Error(c, err.Error(), "Software", partials.GetNavigationUrl(commonInfo, "/software"), commonInfo), commonInfo))
	}

	apps, err = h.Model.GetAppsByPage(p, *f, commonInfo)
	if err != nil {
		return RenderView(c, software_views.SoftwareIndex(" | Software", partials.Error(c, err.Error(), "Software", partials.GetNavigationUrl(commonInfo, "/software"), commonInfo), commonInfo))
	}

	p.NItems, err = h.Model.CountAllApps(*f, commonInfo)
	if err != nil {
		return RenderView(c, software_views.SoftwareIndex(" | Software", partials.Error(c, err.Error(), "Software", partials.GetNavigationUrl(commonInfo, "/software"), commonInfo), commonInfo))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	return RenderView(c, software_views.SoftwareIndex(" | Software", software_views.Software(c, p, *f, apps, refreshTime, itemsPerPage, commonInfo), commonInfo))
}

func (h *Handler) GetSoftwareFilters(c echo.Context) (*filters.ApplicationsFilter, error) {
	// Get filters
	filterByName := c.FormValue("filterByAppName")
	filterByPublisher := c.FormValue("filterByAppPublisher")
	filterByVersion := c.FormValue("filterByAppVersion")
	filterBySearch := c.FormValue("filterBySearch")

	f := filters.ApplicationsFilter{}
	if filterByName != "" {
		f.AppName = filterByName
	}
	if filterByPublisher != "" {
		f.Vendor = filterByPublisher
	}

	if filterByVersion != "" {
		f.Version = filterByVersion
	}

	if filterBySearch != "" {
		f.Search = filterBySearch
	}

	return &f, nil
}
