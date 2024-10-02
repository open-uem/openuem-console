package webserver

import (
	"log"
	"net/http"

	"github.com/doncicuto/openuem-console/internal/controllers/router"
	"github.com/doncicuto/openuem-console/internal/controllers/sessions"
	"github.com/doncicuto/openuem-console/internal/controllers/webserver/handlers"
	"github.com/doncicuto/openuem-console/internal/models"
	"github.com/doncicuto/openuem_nats"
	"github.com/labstack/echo/v4"
)

type WebServer struct {
	Router         *echo.Echo
	Handler        *handlers.Handler
	Server         *http.Server
	SessionManager *sessions.SessionManager
}

func New(m *models.Model, ms *openuem_nats.MessageServer, s *sessions.SessionManager, jwtKey, certPath, keyPath, caCertPath string) *WebServer {
	w := WebServer{}
	// Router
	w.Router = router.New(s)

	// Create Handlers and register its router
	w.Handler = handlers.NewHandler(m, ms, s, jwtKey, certPath, keyPath, caCertPath)
	w.Handler.Register(w.Router)

	w.SessionManager = s

	return &w
}

func (w *WebServer) Serve(address, certFile, certKey string) {
	w.Server = &http.Server{
		Addr:    address,
		Handler: w.Router,
	}
	if err := w.Server.ListenAndServeTLS(certFile, certKey); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func (w *WebServer) Close() error {
	return w.Server.Close()
}
