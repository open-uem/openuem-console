package commands

import (
	"crypto/x509"

	"github.com/doncicuto/openuem-console/internal/models"
	"github.com/nats-io/nats.go"
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
}
