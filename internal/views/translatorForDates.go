package views

import (
	"github.com/gohugoio/locales"
	"github.com/gohugoio/locales/en"
	"github.com/gohugoio/locales/es"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
)

func GetTranslatorForDates(c echo.Context) locales.Translator {
	var l locales.Translator

	i18nCode := i18n.GetLocale(c.Request().Context()).Code().String()

	switch i18nCode {
	case "es":
		l = es.New()
	default:
		l = en.New()
	}
	return l
}
