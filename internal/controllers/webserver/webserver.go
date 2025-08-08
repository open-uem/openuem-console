package webserver

import (
	"log"
	"net/http"

	"github.com/go-co-op/gocron/v2"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/controllers/router"
	"github.com/open-uem/openuem-console/internal/controllers/sessions"
	"github.com/open-uem/openuem-console/internal/controllers/webserver/handlers"
	"github.com/open-uem/openuem-console/internal/models"
)

type WebServer struct {
	Router         *echo.Echo
	Handler        *handlers.Handler
	Server         *http.Server
	SessionManager *sessions.SessionManager
}

func New(m *models.Model, natsServers string, s *sessions.SessionManager, ts gocron.Scheduler, jwtKey, certPath, keyPath, sftpKeyPath, caCertPath, server, consolePort, authPort, tmpDownloadDir, domain, orgName, orgProvince, orgLocality, orgAddress, country, reverseProxyAuthPort, reverseProxyServer, serverReleasesFolder, wingetFolder, flatpakFolder, brewFolder, commonFolder, version string, reEnableCertAuth bool) *WebServer {
	var err error
	w := WebServer{}

	// Get max upload size setting
	maxUploadSize, err := m.GetMaxUploadSize()
	if err != nil {
		maxUploadSize = "512M"
		log.Println("[ERROR]: could not get max upload size from database")
	}

	// Router
	w.Router = router.New(s, server, consolePort, maxUploadSize)

	// Create Handler and register its router
	w.Handler = handlers.NewHandler(m, natsServers, s, ts, jwtKey, certPath, keyPath, sftpKeyPath, caCertPath, server, consolePort, authPort, tmpDownloadDir, domain, orgName, orgProvince, orgLocality, orgAddress, country, reverseProxyAuthPort, reverseProxyServer, serverReleasesFolder, wingetFolder, flatpakFolder, brewFolder, commonFolder, version, reEnableCertAuth)
	w.Handler.Register(w.Router)

	// Add the session manager
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
