package common

import (
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/open-uem/openuem-console/internal/controllers/authserver"
	"github.com/open-uem/openuem-console/internal/controllers/sessions"
	"github.com/open-uem/openuem-console/internal/controllers/webserver"
	"github.com/open-uem/openuem-console/internal/models"
	"github.com/open-uem/utils"
)

type Worker struct {
	Model                             *models.Model
	Logger                            *utils.OpenUEMLogger
	DBConnectJob                      gocron.Job
	ConfigJob                         gocron.Job
	TaskScheduler                     gocron.Scheduler
	DBUrl                             string
	CACertPath                        string
	ConsoleCertPath                   string
	ConsolePrivateKeyPath             string
	SFTPPrivateKeyPath                string
	JWTKey                            string
	SessionManager                    *sessions.SessionManager
	WebServer                         *webserver.WebServer
	AuthServer                        *authserver.AuthServer
	DownloadDir                       string
	ConsolePort                       string
	AuthPort                          string
	ServerName                        string
	Domain                            string
	NATSServers                       string
	WinGetDBFolder                    string
	FlatpakDBFolder                   string
	BrewDBFolder                      string
	CommonSoftwareDBFolder            string
	OrgName                           string
	OrgProvince                       string
	OrgLocality                       string
	OrgAddress                        string
	Country                           string
	ReverseProxyAuthPort              string
	ReverseProxyServer                string
	ServerReleasesFolder              string
	DownloadWingetDBJob               gocron.Job
	DownloadWingetJobDuration         time.Duration
	DownloadServerReleasesJob         gocron.Job
	DownloadServerReleasesJobDuration time.Duration
	DownloadLatestReleaseJob          gocron.Job
	DownloadLatestReleaseJobDuration  time.Duration
	DownloadFlatpakDBJob              gocron.Job
	DownloadFlatpakJobDuration        time.Duration
	DownloadBrewDBJob                 gocron.Job
	DownloadBrewJobDuration           time.Duration
	CommonSoftwareDBJob               gocron.Job
	CommonSoftwareJobDuration         time.Duration
	Version                           string
}

func NewWorker(logName string) *Worker {
	worker := Worker{}
	if logName != "" {
		worker.Logger = utils.NewLogger(logName)
	}

	return &worker
}

func (w *Worker) StartWorker() {
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

	// Start a job to download Flatpak database
	if err := w.StartFlatpakDBDownloadJob(); err != nil {
		log.Printf("[ERROR]: could not start flatpak.db download job, reason: %s", err.Error())
		return
	}

	// Start a job to download Brew database
	if err := w.StartBrewDBDownloadJob(); err != nil {
		log.Printf("[ERROR]: could not start brew.db download job, reason: %s", err.Error())
		return
	}

	// Start a job to create sofwate package table from flatpak, brew, and winget databases
	if err := w.StartCommonPackagesDBJob(); err != nil {
		log.Printf("[ERROR]: could not start job to create common packages db, reason: %s", err.Error())
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
