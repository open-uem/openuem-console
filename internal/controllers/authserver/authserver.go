package authserver

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/doncicuto/openuem-console/internal/controllers/authserver/handlers"
	"github.com/doncicuto/openuem-console/internal/controllers/router"
	"github.com/doncicuto/openuem-console/internal/controllers/sessions"
	"github.com/doncicuto/openuem-console/internal/models"
	"github.com/labstack/echo/v4"
)

type AuthServer struct {
	Router         *echo.Echo
	Handler        *handlers.Handler
	Server         *http.Server
	SessionManager *sessions.SessionManager
	CACert         *x509.Certificate
}

func New(m *models.Model, s *sessions.SessionManager, caCert string) *AuthServer {
	var err error
	a := AuthServer{}
	// Router
	a.Router = router.New(s)

	// Session Manager
	a.SessionManager = s

	a.CACert, err = readPEMCertificate(caCert)
	if err != nil {
		log.Fatal(err)
	}

	// Create Handlers and register its router
	a.Handler = handlers.NewHandler(m, s, a.CACert)
	a.Handler.Register(a.Router)

	return &a
}

func (a *AuthServer) Serve(address, certFile, certKey string) {
	cp := x509.NewCertPool()
	cp.AddCert(a.CACert)

	a.Server = &http.Server{
		Addr:    address,
		Handler: a.Router,
		TLSConfig: &tls.Config{
			ClientAuth: tls.RequestClientCert,
			ClientCAs:  cp,
		},
	}
	if err := a.Server.ListenAndServeTLS(certFile, certKey); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func (a *AuthServer) Close() error {
	return a.Server.Close()
}

// TODO use utils from cert-manager
func readPEMCertificate(path string) (*x509.Certificate, error) {
	certBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	certBlock, _ := pem.Decode(certBytes)
	if certBlock.Type != "CERTIFICATE" || certBlock.Bytes == nil {
		return nil, fmt.Errorf("file does not content a certificate")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}
