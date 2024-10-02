package middleware

import (
	"github.com/invopop/ctxi18n"
	"github.com/labstack/echo/v4"
)

func GetLocale(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		accept := c.Request().Header.Get("Accept-Language")
		ctx, err := ctxi18n.WithLocale(c.Request().Context(), accept)
		if err != nil {
			return err
		}
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}
