package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/controllers/router/middleware"
)

func (h *Handler) ChangeLanguage(c echo.Context) error {
	lang := c.QueryParam("lang")
	if lang == "" {
		lang = c.FormValue("lang")
	}

	targetLang := middleware.NormalizeLocale(lang)
	if targetLang == "" {
		targetLang = "en"
	}

	cookie := &http.Cookie{
		Name:     middleware.LocaleCookieName,
		Value:    targetLang,
		Path:     "/",
		MaxAge:   365 * 24 * 3600,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)

	if h.SessionManager != nil {
		h.SessionManager.Manager.Put(c.Request().Context(), middleware.LocaleCookieName, targetLang)
	}

	if c.Request().Header.Get("HX-Request") == "true" {
		c.Response().Header().Set("HX-Refresh", "true")
		return c.NoContent(http.StatusOK)
	}

	referer := c.Request().Referer()
	if referer == "" {
		referer = "/"
	}
	return c.Redirect(http.StatusSeeOther, referer)
}
