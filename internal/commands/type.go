package commands

import (
	"crypto/x509"

	"github.com/doncicuto/openuem-console/internal/models"
	"github.com/doncicuto/openuem_nats"
)

type ConsoleCommand struct {
	MessageServer *openuem_nats.MessageServer
	Model         *models.Model
	CACert        *x509.Certificate
	DBUrl         string
	CertPath      string
	CertKey       string
	CACertPath    string
	NATSHost      string
	NATSPort      string
	JWTKey        string
}
