package webserver

import (
	"log"
	"net/http"

	"github.com/doncicuto/openuem-console/internal/controllers/router"
	"github.com/doncicuto/openuem-console/internal/controllers/sessions"
	"github.com/doncicuto/openuem-console/internal/controllers/webserver/handlers"
	"github.com/doncicuto/openuem-console/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type WebServer struct {
	Router         *echo.Echo
	Handler        *handlers.Handler
	Server         *http.Server
	SessionManager *sessions.SessionManager
}

func New(m *models.Model, nc *nats.Conn, s *sessions.SessionManager, jwtKey, certPath, keyPath, caCertPath string) *WebServer {
	var err error
	w := WebServer{}
	// Router
	w.Router = router.New(s)

	// Create Handlers and register its router
	w.Handler = handlers.NewHandler(m, nc, s, jwtKey, certPath, keyPath, caCertPath)

	w.Handler.JetStream, err = jetstream.New(w.Handler.NATSConnection)
	if err != nil {
		log.Fatalf("[FATAL]: could not instantiate JetStream, reason: %v", err)
	}
	w.Handler.Register(w.Router)

	w.SessionManager = s

	return &w
}

func (w *WebServer) Serve(address, certFile, certKey string) error {
	w.Server = &http.Server{
		Addr:    address,
		Handler: w.Router,
	}
	return w.Server.ListenAndServeTLS(certFile, certKey)
}

func (w *WebServer) Close() error {
	return w.Server.Close()
}
