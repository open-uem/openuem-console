package partials

import (
	"slices"
	"strconv"

	"github.com/a-h/templ"
	"github.com/doncicuto/openuem-console/internal/views/utils"
	"github.com/labstack/echo/v4"
)

type PaginationAndSort struct {
	CurrentPage int
	PageSize    int
	NItems      int
	SortBy      string
	SortOrder   string
}

func NewPaginationAndSort() PaginationAndSort {
	// TODO set a MIN value for Page size
	return PaginationAndSort{
		CurrentPage: 1,
		PageSize:    5,
	}
}

type page struct {
	Number   string
	Href     templ.SafeURL
	Active   bool
	Disabled bool
}

func (p *PaginationAndSort) GetPaginationAndSortParams(c echo.Context) {
	var err error

	if p.CurrentPage, err = strconv.Atoi(c.QueryParam("page")); err != nil {
		p.CurrentPage = 1
	}

	if p.PageSize, err = strconv.Atoi(c.QueryParam("pageSize")); err != nil {
		p.PageSize = 5
	}
	// TODO set a MAX value for pageSize
	p.PageSize = min(p.PageSize, 25)
	p.SortBy = c.QueryParam("sortBy")
	p.SortOrder = c.QueryParam("sortOrder")
	if !slices.Contains([]string{"asc", "desc"}, p.SortOrder) {
		p.SortOrder = "desc"
	}
}

func GetPaginationEntries(c echo.Context, currentPage, pageSize, nApps int, sortBy, sortOrder string) []page {
	pages := []page{}
	nPages := nApps / pageSize
	if nApps%pageSize != 0 {
		nPages++
	}

	// First page
	pages = append(pages, page{
		Number:   "1",
		Href:     utils.GetURLWithParams(c, 1, pageSize, sortBy, sortOrder),
		Active:   currentPage == 1,
		Disabled: false,
	})

	if currentPage < 8 || nPages <= 10 {
		// TODO let's see what we can do with that 8
		maxPages := 8
		if nPages <= 10 {
			maxPages = nPages - 1
		}

		for i := 2; i <= maxPages; i++ {
			newPage := page{
				Number:   strconv.Itoa(i),
				Href:     utils.GetURLWithParams(c, i, pageSize, sortBy, sortOrder),
				Active:   currentPage == i,
				Disabled: false,
			}
			pages = append(pages, newPage)
		}
		if nPages > 10 {
			pages = append(pages, page{
				Number:   "...",
				Href:     templ.URL("#"),
				Active:   false,
				Disabled: true,
			})
		}
	} else if currentPage >= 8 && nPages-currentPage >= 7 {
		pages = append(pages, page{
			Number:   "...",
			Href:     templ.URL("#"),
			Active:   false,
			Disabled: true,
		})

		for i := currentPage - 3; i <= currentPage+2; i++ {
			newPage := page{
				Number:   strconv.Itoa(i),
				Href:     utils.GetURLWithParams(c, i, pageSize, sortBy, sortOrder),
				Active:   currentPage == i,
				Disabled: false,
			}
			pages = append(pages, newPage)
		}

		pages = append(pages, page{
			Number:   "...",
			Href:     templ.URL("#"),
			Active:   false,
			Disabled: true,
		})
	} else if nPages-currentPage < 7 {
		pages = append(pages, page{
			Number:   "...",
			Href:     templ.URL("#"),
			Active:   false,
			Disabled: true,
		})
		for i := nPages - 7; i <= nPages-1; i++ {
			newPage := page{
				Number:   strconv.Itoa(i),
				Href:     utils.GetURLWithParams(c, i, pageSize, sortBy, sortOrder),
				Active:   currentPage == i,
				Disabled: false,
			}
			pages = append(pages, newPage)
		}
	}

	// Last page
	pages = append(pages, page{
		Number:   strconv.Itoa(nPages),
		Href:     utils.GetURLWithParams(c, nPages, pageSize, sortBy, sortOrder),
		Active:   currentPage == nPages,
		Disabled: false,
	})

	return pages
}
