package common

import (
	"log"
	"net/http"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/open-uem/openuem-console/internal/controllers/authserver"
	"github.com/open-uem/openuem-console/internal/controllers/sessions"
	"github.com/open-uem/openuem-console/internal/controllers/webserver"
	"github.com/open-uem/openuem-console/internal/models"
	"github.com/open-uem/utils"
	"golang.org/x/mod/semver"
)

// StartDBConnectJob initializes the connection with the database and so some initial tasks and migrations
func (w *Worker) StartDBConnectJob() error {
	var err error

	w.Model, err = models.New(w.DBUrl, "pgx", w.Domain)
	if err == nil {
		log.Println("[INFO]: connection established with database")

		if err := w.Model.CreateInitialSettings(); err != nil {
			log.Println("[WARN]: could not create initial settings")
		}

		// Create default orgs and sites #feat-119
		if err := w.Model.CreateDefaultTenantAndSite(); err != nil {
			log.Println("[WARN]: could not create initial settings")
		}

		// Associate agents without site to default site #feat-119
		if err := w.Model.AssociateAgentsToDefaultTenantAndSite(); err != nil {
			log.Println("[WARN]: could not associate agents to default tenant and site")
		}

		// Associate tags without tenant to default tenant #feat-119
		if err := w.Model.AssociateTagsToDefaultTenant(); err != nil {
			log.Println("[WARN]: could not associate tags to default tenant")
		}

		// Associate metadata without tenant to default tenant #feat-119
		if err := w.Model.AssociateMetadataToDefaultTenant(); err != nil {
			log.Println("[WARN]: could not associate metadata to default tenant")
		}

		// Associate profiles without tenant to default tenant #feat-119 and if version < 0.13.0
		if semver.Compare("v"+w.Version, "v0.13.0") < 0 {
			if err := w.Model.AssociateProfilesToDefaultTenantAndSite(); err != nil {
				log.Println("[WARN]: could not associate profiles to default tenant and site")
			}
		}

		// Associate domain to default site #feat-119
		if err := w.Model.AssociateDomainToDefaultSite(w.Domain); err != nil {
			log.Println("[WARN]: could not associate domain to default site")
		}

		// Nickname uses the hostname as the default value
		if err := w.Model.SetDefaultNickname(); err != nil {
			log.Println("[WARN]: could not default nickname to default site")
		}

		// Create argon2 default password for openuem admin if not exist
		if err := w.Model.CreateDefaultAdminPassword(w.ResetOpenUEMUser); err != nil {
			log.Println("[WARN]: could not create default openuem password")
		}

		// Encrypt sensitive fields if they're set in clear and we have a master key
		if err := w.EncryptSensitiveFields(); err != nil {
			log.Printf("[WARN]: could not encrypt sensitive fields, reason: %v", err)
		}

		w.StartConsoleService()

		// Start a job to check latest OpenUEM releases
		channel, err := w.Model.GetDefaultUpdateChannel()
		if err != nil {
			log.Println("[ERROR]: could not get updates channel settings")
			channel = "stable"
		}

		// Start a job to download server releases version
		if err := w.StartServerReleasesDownloadJob(); err != nil {
			log.Printf("[ERROR]: could not start server releases download job, reason: %s", err.Error())
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
				w.Model, err = models.New(w.DBUrl, "pgx", w.Domain)
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

				// Create default orgs and sites #feat-119
				if err := w.Model.CreateDefaultTenantAndSite(); err != nil {
					log.Println("[WARN]: could not create default tenant and site")
				}

				// Associate agents without site to default site #feat-119
				if err := w.Model.AssociateAgentsToDefaultTenantAndSite(); err != nil {
					log.Println("[WARN]: could not associate agents to default tenant and site")
				}

				// Associate tags without tenant to default tenant #feat-119
				if err := w.Model.AssociateTagsToDefaultTenant(); err != nil {
					log.Println("[WARN]: could not associate tags to default tenant")
				}

				// Associate metadata without tenant to default tenant #feat-119
				if err := w.Model.AssociateMetadataToDefaultTenant(); err != nil {
					log.Println("[WARN]: could not associate metadata to default tenant")
				}

				// Associate profiles without tenant to default tenant #feat-119 and if version < 0.13.0
				if semver.Compare("v"+w.Version, "v0.13.0") < 0 {
					if err := w.Model.AssociateProfilesToDefaultTenantAndSite(); err != nil {
						log.Println("[WARN]: could not associate profiles to default tenant and site")
					}
				}

				// Associate domain to default site #feat-119
				if err := w.Model.AssociateDomainToDefaultSite(w.Domain); err != nil {
					log.Println("[WARN]: could not associate domain to default site")
				}

				// Nickname uses the hostname as the default value
				if err := w.Model.SetDefaultNickname(); err != nil {
					log.Println("[WARN]: could not default nickname to default site")
				}

				// Create argon2 default password for openuem admin if not exist
				if err := w.Model.CreateDefaultAdminPassword(w.ResetOpenUEMUser); err != nil {
					log.Println("[WARN]: could not create default openuem password")
				}

				// Encrypt sensitive fields if they're set in clear and we have a master key
				if err := w.EncryptSensitiveFields(); err != nil {
					log.Printf("[WARN]: could not encrypt sensitive fields, reason: %v", err)
				}

				w.StartConsoleService()

				// Start a job to check latest OpenUEM releases
				channel, err := w.Model.GetDefaultUpdateChannel()
				if err != nil {
					log.Println("[ERROR]: could not get updates channel settings")
					channel = "stable"
				}

				// Start a job to download server releases version
				if err := w.StartServerReleasesDownloadJob(); err != nil {
					log.Printf("[ERROR]: could not start server releases download job, reason: %s", err.Error())
					return
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
	log.Printf("[INFO]: new DB connect job has been scheduled every %d seconds", 30)
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
	w.WebServer = webserver.New(w.Model, w.NATSServers, w.SessionManager, w.TaskScheduler, w.JWTKey, w.ConsoleCertPath, w.ConsolePrivateKeyPath, w.SFTPPrivateKeyPath, w.CACertPath, serverName, consolePort, authPort, w.DownloadDir, w.Domain, w.OrgName, w.OrgProvince, w.OrgLocality, w.OrgAddress, w.Country, w.ReverseProxyAuthPort, w.ReverseProxyServer, w.ServerReleasesFolder, w.CommonSoftwareDBFolder, w.Version, w.EncryptionMasterKey, w.ReenableCertAuth, w.ReenablePasswdAuth, w.ResetOpenUEMUser, w.AuthLogger)
	go func() {
		if err := w.WebServer.Serve(":"+consolePort, w.ConsoleCertPath, w.ConsolePrivateKeyPath); err != http.ErrServerClosed {
			log.Printf("[ERROR]: the server has stopped, reason: %v", err.Error())
		}
	}()
	log.Println("[INFO]: console is running")

	// HTTPS auth server
	w.AuthServer = authserver.New(w.Model, w.SessionManager, w.CACertPath, serverName, consolePort, authPort, w.ReverseProxyAuthPort)
	go func() {
		if err := w.AuthServer.Serve(":"+authPort, w.ConsoleCertPath, w.ConsolePrivateKeyPath); err != http.ErrServerClosed {
			log.Printf("[ERROR]: the server has stopped, reason: %v", err.Error())
		}
	}()
	log.Println("[INFO]: auth server is running")
}

func (w *Worker) EncryptSensitiveFields() error {
	// 1. Check if we have the ENCRYPTION_MASTER_KEY env variable
	if w.EncryptionMasterKey == "" {
		return nil
	}

	// 2. Get SMTP password, check if encrypted and encrypt if needed
	credentials, err := w.Model.GetSMTPPasswords()
	if err != nil {
		return err
	}

	for _, c := range credentials {
		if c.SMTPPassword != "" {
			isEncrypted, err := utils.IsSensitiveFieldEncrypted(c.SMTPPassword, w.EncryptionMasterKey)
			if err != nil {
				return err
			}

			if !isEncrypted {
				encryptedPassword, err := utils.EncryptSensitiveField(c.SMTPPassword, w.EncryptionMasterKey)
				if err != nil {
					log.Printf("[ERROR]: could not encrypt SMTP password, reason: %v", err)
					continue
				}

				if err := w.Model.UpdateSMTPPassword(c.ID, encryptedPassword); err != nil {
					log.Printf("[ERROR]: could not save encrypted SMTP password, reason: %v", err)
					continue
				}
			}
		}
	}

	// 3. Get NetBird access tokens, check if encrypted and encrypt if needed
	tokens, err := w.Model.GetNetbirdAccessTokens()
	if err != nil {
		return err
	}

	for _, t := range tokens {
		if t.AccessToken != "" {
			isEncrypted, err := utils.IsSensitiveFieldEncrypted(t.AccessToken, w.EncryptionMasterKey)
			if err != nil {
				return err
			}

			if !isEncrypted {
				encryptedToken, err := utils.EncryptSensitiveField(t.AccessToken, w.EncryptionMasterKey)
				if err != nil {
					log.Printf("[ERROR]: could not encrypt NetBird access token, reason: %v", err)
					continue
				}

				if err := w.Model.UpdateNetbirdAccessToken(t.ID, encryptedToken); err != nil {
					log.Printf("[ERROR]: could not save encrypted NetBird access token, reason: %v", err)
					continue
				}
			}
		}
	}

	return nil
}
