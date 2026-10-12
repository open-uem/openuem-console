package views

import (
	"github.com/gohugoio/locales"
	"github.com/gohugoio/locales/ca"
	"github.com/gohugoio/locales/de"
	"github.com/gohugoio/locales/en"
	"github.com/gohugoio/locales/es"
	"github.com/gohugoio/locales/fr"
	"github.com/gohugoio/locales/no"
	"github.com/gohugoio/locales/pt"
	"github.com/gohugoio/locales/zh"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
)

func GetTranslatorForDates(c echo.Context) locales.Translator {
	var l locales.Translator

	i18nCode := i18n.GetLocale(c.Request().Context()).Code().String()

	switch i18nCode {
	case "es":
		l = es.New()
	case "de":
		l = de.New()
	case "ca":
		l = ca.New()
	case "fr":
		l = fr.New()
	case "no":
		l = no.New()
	case "pt":
		l = pt.New()
	case "cn", "zh":
		l = zh.New()
	default:
		l = en.New()
	}
	return l
}
