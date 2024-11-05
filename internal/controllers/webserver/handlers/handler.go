package handlers

import (
	"log"

	"github.com/doncicuto/openuem-console/internal/controllers/sessions"
	"github.com/doncicuto/openuem-console/internal/models"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Handler struct {
	Model          *models.Model
	NATSConnection *nats.Conn
	JetStream      jetstream.JetStream
	SessionManager *sessions.SessionManager
	JWTKey         string
	CertPath       string
	KeyPath        string
	CACertPath     string
	DownloadDir    string
	ServerName     string
	AuthPort       string
	ConsolePort    string
	Domain         string
	NATSTimeout    int
}

func NewHandler(model *models.Model, nc *nats.Conn, s *sessions.SessionManager, jwtKey, certPath, keyPath, caCertPath, server, authPort, tmpDownloadDir, domain string) *Handler {

	// Get NATS request timeout seconds
	timeout, err := model.GetNATSTimeout()
	if err != nil {
		timeout = 20
		log.Println("[ERROR]: could not get NATS request timeout from database")
	}

	return &Handler{
		Model:          model,
		NATSConnection: nc,
		SessionManager: s,
		JWTKey:         jwtKey,
		CertPath:       certPath,
		KeyPath:        keyPath,
		CACertPath:     caCertPath,
		DownloadDir:    tmpDownloadDir,
		ServerName:     server,
		AuthPort:       authPort,
		Domain:         domain,
		NATSTimeout:    timeout,
	}
}
