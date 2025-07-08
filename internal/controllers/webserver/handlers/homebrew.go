package handlers

import (
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	models "github.com/open-uem/openuem-console/internal/models/winget"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (h *Handler) SearchHomeBrewFormulaePackages(c echo.Context) error {
	var err error

	search := c.FormValue("brew-search")
	if search == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "install.search_empty_error"), true))
	}

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

	// Default sort
	if p.SortBy == "" {
		p.SortBy = "name"
		p.SortOrder = "asc"
	}

	packages, err := models.SearchAllHomeBrewFormulaePackages(search, h.BrewFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, partials.SearchHomeBrewResult(packages))
}

func (h *Handler) SearchHomeBrewCasksPackages(c echo.Context) error {
	var err error

	search := c.FormValue("brew-search")
	if search == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "install.search_empty_error"), true))
	}

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

	// Default sort
	if p.SortBy == "" {
		p.SortBy = "name"
		p.SortOrder = "asc"
	}

	packages, err := models.SearchAllHomeBrewCasksPackages(search, h.BrewFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, partials.SearchHomeBrewResult(packages))
}
