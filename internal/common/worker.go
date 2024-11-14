package common

import (
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/danieljoos/wincred"
	"github.com/doncicuto/openuem-console/internal/controllers/authserver"
	"github.com/doncicuto/openuem-console/internal/controllers/sessions"
	"github.com/doncicuto/openuem-console/internal/controllers/webserver"
	"github.com/doncicuto/openuem-console/internal/models"
	"github.com/doncicuto/openuem_utils"
	"github.com/go-co-op/gocron/v2"
	"golang.org/x/sys/windows/registry"
)

type Worker struct {
	Model                 *models.Model
	Logger                *openuem_utils.OpenUEMLogger
	DBConnectJob          gocron.Job
	TaskScheduler         gocron.Scheduler
	DBUrl                 string
	CACertPath            string
	ConsoleCertPath       string
	ConsolePrivateKeyPath string
	JWTKey                string
	SessionManager        *sessions.SessionManager
	WebServer             *webserver.WebServer
	AuthServer            *authserver.AuthServer
	DownloadDir           string
	ConsolePort           string
	AuthPort              string
	ServerName            string
	Domain                string
	NATSServers           string
}

func NewWorker(logName string) *Worker {
	worker := Worker{}
	if logName != "" {
		worker.Logger = openuem_utils.NewLogger(logName)
	}
	return &worker
}

func (w *Worker) StartWorker() {
	var err error

	// Start Task Scheduler
	w.TaskScheduler, err = gocron.NewScheduler()
	if err != nil {
		log.Printf("[ERROR]: could not create task scheduler, reason: %s", err.Error())
		return
	}
	w.TaskScheduler.Start()
	log.Println("[INFO]: task scheduler has been started")

	// Start a job to try to connect with the database
	if err := w.StartDBConnectJob(); err != nil {
		log.Printf("[ERROR]: could not start DB connect job, reason: %s", err.Error())
		return
	}

	// Start a job to clean tmp download directory
	if err := w.StartDownloadCleanJob(); err != nil {
		log.Printf("[ERROR]: could not start Dowload dir clean job, reason: %s", err.Error())
		return
	}

	// Start a job to download Microsoft Winget database
	if err := w.StartWinGetDBDownloadJob(); err != nil {
		log.Printf("[ERROR]: could not start index.db download job, reason: %s", err.Error())
		return
	}

	// Get port information
	consolePort := "1323"
	if w.ConsolePort != "" {
		consolePort = w.ConsolePort
	}

	authPort := "1324"
	if w.AuthPort != "" {
		authPort = w.AuthPort
	}

	// Get server name
	serverName := "localhost"
	if w.ServerName != "" {
		serverName = w.ServerName
	}

	// Session handler
	sessionLifetimeInMinutes, err := w.Model.GetDefaultSessionLifetime()
	if err != nil {
		log.Printf("[ERROR]: could not get session lifetime from database, reason: %v", err.Error())
		sessionLifetimeInMinutes = 1440
	}

	w.SessionManager = sessions.New(w.DBUrl, sessionLifetimeInMinutes)

	// HTTPS web server
	w.WebServer = webserver.New(w.Model, w.NATSServers, w.SessionManager, w.TaskScheduler, w.JWTKey, w.ConsoleCertPath, w.ConsolePrivateKeyPath, w.CACertPath, serverName, consolePort, authPort, w.DownloadDir, w.Domain)
	go func() {
		if err := w.WebServer.Serve(":"+consolePort, w.ConsoleCertPath, w.ConsolePrivateKeyPath); err != http.ErrServerClosed {
			log.Printf("[ERROR]: the server has stopped, reason: %v", err.Error())
		}
	}()
	log.Println("[INFO]: console is running")

	// HTTPS auth server
	w.AuthServer = authserver.New(w.Model, w.SessionManager, w.CACertPath, serverName, consolePort, authPort)
	go func() {
		if err := w.AuthServer.Serve(":"+authPort, w.ConsoleCertPath, w.ConsolePrivateKeyPath); err != http.ErrServerClosed {
			log.Printf("[ERROR]: the server has stopped, reason: %v", err.Error())
		}
	}()
	log.Println("[INFO]: auth server is running")
}

func (w *Worker) StopWorker() {
	w.Model.Close()
	if err := w.TaskScheduler.Shutdown(); err != nil {
		log.Printf("[ERROR]: could not stop the task scheduler, reason: %s", err.Error())
	}

	if w.SessionManager != nil {
		w.SessionManager.Close()
	}

	if w.WebServer != nil {
		if err := w.WebServer.Close(); err != nil {
			log.Println("[ERROR]: Error closing the web server")
		}
	}

	if w.AuthServer != nil {
		if err := w.AuthServer.Close(); err != nil {
			log.Println("[ERROR]: Error closing the auth server")
		}
	}

	if w.Logger != nil {
		w.Logger.Close()
	}
}

func (w *Worker) GenerateConsoleConfig() error {
	var err error

	cwd, err := GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return err
	}

	w.DBUrl, err = openuem_utils.CreatePostgresDatabaseURL()
	if err != nil {
		log.Printf("[ERROR]: %v", err)
		return err
	}

	w.CACertPath = filepath.Join(cwd, "certificates/ca/ca.cer")
	_, err = openuem_utils.ReadPEMCertificate(w.CACertPath)
	if err != nil {
		log.Printf("[ERROR]: could not read CA certificate in %s", w.CACertPath)
		return err
	}

	w.ConsoleCertPath = filepath.Join(cwd, "certificates/console/console.cer")
	_, err = openuem_utils.ReadPEMCertificate(w.ConsoleCertPath)
	if err != nil {
		log.Println("[ERROR]: could not read OCSP certificate")
		return err
	}

	w.ConsolePrivateKeyPath = filepath.Join(cwd, "certificates/console/console.key")
	_, err = openuem_utils.ReadPEMPrivateKey(w.ConsolePrivateKeyPath)
	if err != nil {
		log.Println("[ERROR]: could not read OCSP private key")
		return err
	}

	encodedKey, err := wincred.GetGenericCredential("OpenUEM JWT Key")
	if err != nil {
		return fmt.Errorf("could not read JWTKey from Windows Credential Manager")
	}

	w.JWTKey = openuem_utils.UTF16BytesToString(encodedKey.CredentialBlob, binary.LittleEndian)

	k, err := openuem_utils.OpenRegistryForQuery(registry.LOCAL_MACHINE, `SOFTWARE\OpenUEM\Server`)
	if err != nil {
		log.Println("[ERROR]: could not open registry")
		return err
	}
	defer k.Close()

	w.ServerName, err = openuem_utils.GetValueFromRegistry(k, "ConsoleServer")
	if err != nil {
		log.Println("[ERROR]: could not read console server name from registry")
		return err
	}

	w.ConsolePort, err = openuem_utils.GetValueFromRegistry(k, "ConsolePort")
	if err != nil {
		log.Println("[ERROR]: could not read console port from registry")
		return err
	}

	w.AuthPort, err = openuem_utils.GetValueFromRegistry(k, "AuthPort")
	if err != nil {
		log.Println("[ERROR]: could not read auth port from registry")
		return err
	}

	w.Domain, err = openuem_utils.GetValueFromRegistry(k, "Domain")
	if err != nil {
		log.Println("[ERROR]: could not read domain from registry")
		return err
	}

	w.NATSServers, err = openuem_utils.GetValueFromRegistry(k, "NATSServers")
	if err != nil {
		log.Println("[ERROR]: could not read NATS servers from registry")
		return err
	}

	return nil
}

func (w *Worker) CreateDowloadTempDir() error {
	cwd, err := GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return err
	}

	w.DownloadDir = filepath.Join(cwd, "tmp", "download")

	if strings.HasSuffix(cwd, "tmp") {
		w.DownloadDir = filepath.Join(cwd, "download")
	}

	if _, err := os.Stat(w.DownloadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(w.DownloadDir, 0666); err != nil {
			log.Printf("[ERROR]: could not create temp download directory, reason: %v", err)
			return err
		}
	}

	return nil
}
