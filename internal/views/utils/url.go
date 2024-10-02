package utils

import (
	"strconv"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func GetURLWithParams(c echo.Context, page, pageSize int, sortBy, sortOrder string) templ.SafeURL {
	if pageSize == 0 {
		c.QueryParams().Del("pageSize")
	} else {
		c.QueryParams().Set("pageSize", strconv.Itoa(pageSize))
	}

	c.QueryParams().Set("page", strconv.Itoa(page))
	c.QueryParams().Set("sortBy", sortBy)
	c.QueryParams().Set("sortOrder", sortOrder)
	c.Request().URL.RawQuery = c.QueryParams().Encode()
	return templ.URL(c.Request().URL.RequestURI())
}
