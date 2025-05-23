package commands

import (
	"crypto/x509"

	"github.com/nats-io/nats.go"
	"github.com/open-uem/openuem-console/internal/models"
)

type ConsoleCommand struct {
	NATSConnection *nats.Conn
	Model          *models.Model
	CACert         *x509.Certificate
	DBUrl          string
	CertPath       string
	CertKey        string
	CACertPath     string
	NATSServers    string
	JWTKey         string
	Domain         string
}
