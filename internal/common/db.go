package common

import (
	"log"
	"net/http"
	"time"

	"github.com/doncicuto/openuem-console/internal/controllers/authserver"
	"github.com/doncicuto/openuem-console/internal/controllers/sessions"
	"github.com/doncicuto/openuem-console/internal/controllers/webserver"
	"github.com/doncicuto/openuem-console/internal/models"
	"github.com/go-co-op/gocron/v2"
)

func (w *Worker) StartDBConnectJob() error {
	var err error

	w.Model, err = models.New(w.DBUrl)
	if err == nil {
		log.Println("[INFO]: connection established with database")

		if err := w.Model.CreateInitialSettings(); err != nil {
			log.Println("[WARN]: could not create initial settings")
		}

		w.StartConsoleService()

		// Start a job to check latest OpenUEM releases
		channel, err := w.Model.GetDefaultUpdateChannel()
		if err != nil {
			log.Println("[ERROR]: could not get updates channel settings")
			channel = "stable"
		}
		if err := w.StartCheckLatestReleasesJob(channel); err != nil {
			log.Printf("[ERROR]: could not start check latest releases job, reason: %s", err.Error())
		}
		return nil
	}
	log.Printf("[ERROR]: could not connect with database %v", err)

	// Create task
	w.DBConnectJob, err = w.TaskScheduler.NewJob(
		gocron.DurationJob(
			time.Duration(time.Duration(30*time.Second)),
		),
		gocron.NewTask(
			func() {
				w.Model, err = models.New(w.DBUrl)
				if err != nil {
					log.Printf("[ERROR]: could not connect with database %v", err)
					return
				}
				log.Println("[INFO]: connection established with database")

				if err := w.TaskScheduler.RemoveJob(w.DBConnectJob.ID()); err != nil {
					return
				}

				if err := w.Model.CreateInitialSettings(); err != nil {
					log.Println("[WARN]: could not create initial settings")
				}

				w.StartConsoleService()

				// Start a job to check latest OpenUEM releases
				channel, err := w.Model.GetDefaultUpdateChannel()
				if err != nil {
					log.Println("[ERROR]: could not get updates channel settings")
					channel = "stable"
				}
				if err := w.StartCheckLatestReleasesJob(channel); err != nil {
					log.Printf("[ERROR]: could not start check latest releases job, reason: %s", err.Error())
					return
				}
			},
		),
	)
	if err != nil {
		log.Fatalf("[FATAL]: could not start the DB connect job: %v", err)
		return err
	}
	log.Printf("[INFO]: new DB connect job has been scheduled every %d minutes", 2)
	return nil
}

func (w *Worker) StartConsoleService() {
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
	w.WebServer = webserver.New(w.Model, w.NATSServers, w.SessionManager, w.TaskScheduler, w.JWTKey, w.ConsoleCertPath, w.ConsolePrivateKeyPath, w.CACertPath, serverName, consolePort, authPort, w.DownloadDir, w.Domain, w.OrgName, w.OrgProvince, w.OrgLocality, w.OrgAddress, w.Country)
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
