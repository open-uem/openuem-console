package common

import (
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/open-uem/utils"
	"gopkg.in/ini.v1"
)

func (w *Worker) GenerateConsoleConfig() error {
	var err error

	w.DBUrl, err = utils.CreatePostgresDatabaseURL()
	if err != nil {
		log.Printf("[ERROR]: %v", err)
		return err
	}

	// Open ini file
	configFile := utils.GetConfigFile()
	cfg, err := ini.Load(configFile)
	if err != nil {
		return err
	}

	key, err := cfg.Section("Certificates").GetKey("CACert")
	if err != nil {
		return err
	}

	w.CACertPath = key.String()
	_, err = utils.ReadPEMCertificate(w.CACertPath)
	if err != nil {
		log.Printf("[ERROR]: could not read CA certificate in %s", w.CACertPath)
		return err
	}

	key, err = cfg.Section("Certificates").GetKey("ConsoleCert")
	if err != nil {
		return err
	}

	w.ConsoleCertPath = key.String()
	_, err = utils.ReadPEMCertificate(w.ConsoleCertPath)
	if err != nil {
		log.Println("[ERROR]: could not read Console certificate")
		return err
	}

	key, err = cfg.Section("Certificates").GetKey("ConsoleKey")
	if err != nil {
		return err
	}

	w.ConsolePrivateKeyPath = key.String()
	_, err = utils.ReadPEMPrivateKey(w.ConsolePrivateKeyPath)
	if err != nil {
		log.Println("[ERROR]: could not read Console private key")
		return err
	}

	key, err = cfg.Section("Certificates").GetKey("SFTPKey")
	if err != nil {
		return err
	}

	w.SFTPPrivateKeyPath = key.String()
	_, err = utils.ReadPEMPrivateKey(w.SFTPPrivateKeyPath)
	if err != nil {
		log.Println("[ERROR]: could not read SFTP private key")
		return err
	}

	w.JWTKey, err = utils.GetJWTKey()
	if err != nil {
		return err
	}

	key, err = cfg.Section("Console").GetKey("hostname")
	if err != nil {
		return err
	}
	w.ServerName = key.String()

	key, err = cfg.Section("Console").GetKey("port")
	if err != nil {
		return err
	}
	w.ConsolePort = key.String()

	key, err = cfg.Section("Console").GetKey("authport")
	if err != nil {
		return err
	}
	w.AuthPort = key.String()

	key, err = cfg.Section("Console").GetKey("domain")
	if err != nil {
		return err
	}
	w.Domain = key.String()

	key, err = cfg.Section("NATS").GetKey("NATSServers")
	if err != nil {
		return err
	}
	w.NATSServers = key.String()

	key, err = cfg.Section("Certificates").GetKey("OrgName")
	if err != nil {
		return err
	}
	w.OrgName = key.String()

	key, err = cfg.Section("Certificates").GetKey("OrgProvince")
	if err != nil {
		return err
	}
	w.OrgProvince = key.String()

	key, err = cfg.Section("Certificates").GetKey("OrgLocality")
	if err != nil {
		return err
	}
	w.OrgLocality = key.String()

	key, err = cfg.Section("Certificates").GetKey("OrgAddress")
	if err != nil {
		return err
	}
	w.OrgAddress = key.String()

	key, err = cfg.Section("Certificates").GetKey("OrgCountry")
	if err != nil {
		return err
	}
	w.Country = key.String()

	key, err = cfg.Section("Console").GetKey("reverseproxyauthport")
	if err != nil {
		return err
	}
	w.ReverseProxyAuthPort = key.String()

	key, err = cfg.Section("Console").GetKey("reverseproxyserver")
	if err != nil {
		return err
	}
	w.ReverseProxyServer = key.String()

	key, err = cfg.Section("Server").GetKey("Version")
	if err != nil {
		return err
	}
	w.Version = key.String()

	return nil
}

func (w *Worker) StartGenerateConsoleConfigJob() error {
	var err error

	// Create task for getting the worker config
	w.ConfigJob, err = w.TaskScheduler.NewJob(
		gocron.DurationJob(
			time.Duration(time.Duration(1*time.Minute)),
		),
		gocron.NewTask(
			func() {
				err = w.GenerateConsoleConfig()
				if err != nil {
					log.Printf("[ERROR]: could not generate config for console, reason: %v", err)
					return
				}

				log.Println("[INFO]: console's config has been successfully generated")
				if err := w.TaskScheduler.RemoveJob(w.ConfigJob.ID()); err != nil {
					return
				}
			},
		),
	)
	if err != nil {
		log.Fatalf("[FATAL]: could not start the generate console config job: %v", err)
		return err
	}
	log.Printf("[INFO]: new generate console config job has been scheduled every %d minute", 1)
	return nil
}
