package handlers

import (
	"crypto/x509"

	"github.com/doncicuto/openuem-console/internal/controllers/sessions"
	"github.com/doncicuto/openuem-console/internal/models"
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
