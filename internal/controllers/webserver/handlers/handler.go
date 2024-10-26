package handlers

import (
	"github.com/doncicuto/openuem-console/internal/controllers/sessions"
	"github.com/doncicuto/openuem-console/internal/models"
	"github.com/nats-io/nats.go"
)

type Handler struct {
	Model          *models.Model
	NATSConnection *nats.Conn
	SessionManager *sessions.SessionManager
	JWTKey         string
	CertPath       string
	KeyPath        string
	CACertPath     string
}

func NewHandler(model *models.Model, nc *nats.Conn, s *sessions.SessionManager, jwtKey, certPath, keyPath, caCertPath string) *Handler {
	return &Handler{
		Model:          model,
		NATSConnection: nc,
		SessionManager: s,
		JWTKey:         jwtKey,
		CertPath:       certPath,
		KeyPath:        keyPath,
		CACertPath:     caCertPath,
	}
}
