package handlers

import (
	"log"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (h *Handler) SearchWingetPackages(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	search := c.FormValue("package-search")
	if search == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "install.search_empty_error"), true))
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

	// The task form sends its profile, whose organizations decide which packages of the
	// console's winget repository its tasks can install
	packages, err := h.Model.SearchAllWingetPackages(search, h.hiddenLibraryPackagesForProfile(c.FormValue("task-profile-id"), commonInfo))
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, partials.SearchPacketResult(packages))
}
