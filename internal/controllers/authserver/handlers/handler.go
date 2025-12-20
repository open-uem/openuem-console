package handlers

import (
	"crypto/x509"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/controllers/sessions"
	"github.com/open-uem/openuem-console/internal/models"
)

type Handler struct {
	Model                *models.Model
	SessionManager       *sessions.SessionManager
	CACert               *x509.Certificate
	ServerName           string
	ConsolePort          string
	ReverseProxyAuthPort string
}

func NewHandler(model *models.Model, sm *sessions.SessionManager, cert *x509.Certificate, server, consolePort, reverseProxyAuthPort string) *Handler {
	return &Handler{
		Model:                model,
		SessionManager:       sm,
		CACert:               cert,
		ServerName:           server,
		ConsolePort:          consolePort,
		ReverseProxyAuthPort: reverseProxyAuthPort,
	}
}

func RenderLoginPartial(c echo.Context, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	c.Response().Header().Set(echo.HeaderXContentTypeOptions, "nosniff")
	c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
	c.Response().Header().Set("HX-Retarget", "#login")
	c.Response().Header().Set("HX-Reswap", "outerHTML")
	return cmp.Render(c.Request().Context(), c.Response().Writer)
}
