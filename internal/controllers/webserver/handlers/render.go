package handlers

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func renderView(c echo.Context, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)

	return cmp.Render(c.Request().Context(), c.Response().Writer)
}

func renderError(c echo.Context, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	c.Response().Header().Set("HX-Retarget", "#error")
	c.Response().Header().Set("HX-Reswap", "outerHTML")

	return cmp.Render(c.Request().Context(), c.Response().Writer)
}

func renderConfirm(c echo.Context, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	c.Response().Header().Set("HX-Retarget", "#confirm")
	c.Response().Header().Set("HX-Reswap", "outerHTML")

	return cmp.Render(c.Request().Context(), c.Response().Writer)
}

func renderSuccess(c echo.Context, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	c.Response().Header().Set("HX-Retarget", "#success")
	c.Response().Header().Set("HX-Reswap", "outerHTML")

	return cmp.Render(c.Request().Context(), c.Response().Writer)
}
