package common

import (
	"log"

	"github.com/doncicuto/openuem-console/internal/controllers/authserver"
	"github.com/doncicuto/openuem-console/internal/controllers/sessions"
	"github.com/doncicuto/openuem-console/internal/controllers/webserver"
	"github.com/doncicuto/openuem-console/internal/models"
	"github.com/doncicuto/openuem_utils"
	"github.com/go-co-op/gocron/v2"
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
	SFTPPrivateKeyPath    string
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
	WinGetDBFolder        string
	OrgName               string
	OrgProvince           string
	OrgLocality           string
	OrgAddress            string
	Country               string
	ReverseProxyAuthPort  string
	ReverseProxyServer    string
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
		log.Fatalf("[FATAL]: could not create task scheduler, reason: %s", err.Error())
		return
	}
	w.TaskScheduler.Start()
	log.Println("[INFO]: task scheduler has been started")

	// Start a job to try to connect with the database
	if err := w.StartDBConnectJob(); err != nil {
		log.Fatalf("[FATAL]: could not start DB connect job, reason: %s", err.Error())
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
