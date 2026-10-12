package middleware

import (
	"net/http"
	"strings"

	"github.com/invopop/ctxi18n"
	"github.com/labstack/echo/v4"
)

const LocaleCookieName = "openuem_lang"

// NormalizeLocale standardizes language codes to the ones supported by OpenUEM locales
func NormalizeLocale(lang string) string {
	lang = strings.TrimSpace(strings.ToLower(lang))
	if strings.HasPrefix(lang, "zh") || lang == "cn" {
		return "cn"
	}
	switch lang {
	case "en", "es", "ca", "de", "fr", "no", "pt", "cn":
		return lang
	default:
		return ""
	}
}

// MapAcceptLanguage converts browser language tags like zh-CN to OpenUEM's cn locale
func MapAcceptLanguage(accept string) string {
	if accept == "" {
		return ""
	}
	tags := strings.Split(accept, ",")
	for i, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		base := trimmed
		if idx := strings.Index(trimmed, ";"); idx != -1 {
			base = trimmed[:idx]
		}
		baseLower := strings.ToLower(strings.TrimSpace(base))
		if strings.HasPrefix(baseLower, "zh") || baseLower == "cn" {
			if idx := strings.Index(trimmed, ";"); idx != -1 {
				tags[i] = "cn" + trimmed[idx:]
			} else {
				tags[i] = "cn"
			}
		}
	}
	return strings.Join(tags, ",")
}

func GetLocale(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var selectedLang string

		// 1. Check query parameter
		if qLang := c.QueryParam("lang"); qLang != "" {
			if normalized := NormalizeLocale(qLang); normalized != "" {
				selectedLang = normalized
				// Persist via cookie
				c.SetCookie(&http.Cookie{
					Name:     LocaleCookieName,
					Value:    selectedLang,
					Path:     "/",
					MaxAge:   365 * 24 * 3600,
					SameSite: http.SameSiteLaxMode,
				})
			}
		}

		// 2. Check cookie if not set by query param
		if selectedLang == "" {
			if cookie, err := c.Cookie(LocaleCookieName); err == nil && cookie.Value != "" {
				selectedLang = NormalizeLocale(cookie.Value)
			}
			if selectedLang == "" {
				if cookie, err := c.Cookie("lang"); err == nil && cookie.Value != "" {
					selectedLang = NormalizeLocale(cookie.Value)
				}
			}
		}

		// 3. If a language is chosen via cookie or query param, apply it
		if selectedLang != "" {
			ctx, err := ctxi18n.WithLocale(c.Request().Context(), selectedLang)
			if err == nil {
				c.SetRequest(c.Request().WithContext(ctx))
				return next(c)
			}
		}

		// 4. Fallback to Accept-Language header
		accept := MapAcceptLanguage(c.Request().Header.Get("Accept-Language"))
		ctx, err := ctxi18n.WithLocale(c.Request().Context(), accept)
		if err != nil {
			// Fallback to default
			ctx, err = ctxi18n.WithLocale(c.Request().Context(), "en")
			if err != nil {
				return err
			}
		}

		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}
