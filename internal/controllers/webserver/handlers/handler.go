package handlers

import (
	"github.com/doncicuto/openuem-console/internal/controllers/sessions"
	"github.com/doncicuto/openuem-console/internal/models"
	"github.com/doncicuto/openuem_nats"
)

type Handler struct {
	Model          *models.Model
	MessageServer  *openuem_nats.MessageServer
	SessionManager *sessions.SessionManager
	JWTKey         string
	CertPath       string
	KeyPath        string
	CACertPath     string
}

func NewHandler(model *models.Model, ms *openuem_nats.MessageServer, s *sessions.SessionManager, jwtKey, certPath, keyPath, caCertPath string) *Handler {
	return &Handler{
		Model:          model,
		MessageServer:  ms,
		SessionManager: s,
		JWTKey:         jwtKey,
		CertPath:       certPath,
		KeyPath:        keyPath,
		CACertPath:     caCertPath,
	}
}
