package common

import (
	"log"

	"github.com/doncicuto/openuem_utils"
	"gopkg.in/ini.v1"
)

func (w *Worker) GenerateConsoleConfig() error {
	var err error

	w.DBUrl, err = openuem_utils.CreatePostgresDatabaseURL()
	if err != nil {
		log.Printf("[ERROR]: %v", err)
		return err
	}

	// Open ini file
	configFile := openuem_utils.GetConfigFile()
	cfg, err := ini.Load(configFile)
	if err != nil {
		return err
	}

	key, err := cfg.Section("Certificates").GetKey("CACert")
	if err != nil {
		return err
	}

	w.CACertPath = key.String()
	_, err = openuem_utils.ReadPEMCertificate(w.CACertPath)
	if err != nil {
		log.Printf("[ERROR]: could not read CA certificate in %s", w.CACertPath)
		return err
	}

	key, err = cfg.Section("Certificates").GetKey("ConsoleCert")
	if err != nil {
		return err
	}

	w.ConsoleCertPath = key.String()
	_, err = openuem_utils.ReadPEMCertificate(w.ConsoleCertPath)
	if err != nil {
		log.Println("[ERROR]: could not read Console certificate")
		return err
	}

	key, err = cfg.Section("Certificates").GetKey("ConsoleKey")
	if err != nil {
		return err
	}

	w.ConsolePrivateKeyPath = key.String()
	_, err = openuem_utils.ReadPEMPrivateKey(w.ConsolePrivateKeyPath)
	if err != nil {
		log.Println("[ERROR]: could not read Console private key")
		return err
	}

	key, err = cfg.Section("Certificates").GetKey("SFTPKey")
	if err != nil {
		return err
	}

	w.SFTPPrivateKeyPath = key.String()
	_, err = openuem_utils.ReadPEMPrivateKey(w.SFTPPrivateKeyPath)
	if err != nil {
		log.Println("[ERROR]: could not read SFTP private key")
		return err
	}

	w.JWTKey, err = openuem_utils.GetJWTKey()
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

	key, err = cfg.Section("NATS").GetKey("servers")
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

	return nil
}
