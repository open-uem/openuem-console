//go:build linux

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
	cfg, err := ini.Load("/etc/openuem-server/openuem.ini")
	if err != nil {
		return err
	}

	key, err := cfg.Section("Server").GetKey("ca_cert_path")
	if err != nil {
		return err
	}

	w.CACertPath = key.String()
	_, err = openuem_utils.ReadPEMCertificate(w.CACertPath)
	if err != nil {
		log.Printf("[ERROR]: could not read CA certificate in %s", w.CACertPath)
		return err
	}

	key, err = cfg.Section("Server").GetKey("console_cert_path")
	if err != nil {
		return err
	}

	w.ConsoleCertPath = key.String()
	_, err = openuem_utils.ReadPEMCertificate(w.ConsoleCertPath)
	if err != nil {
		log.Println("[ERROR]: could not read OCSP certificate")
		return err
	}

	key, err = cfg.Section("Server").GetKey("console_key_path")
	if err != nil {
		return err
	}

	w.ConsolePrivateKeyPath = key.String()
	_, err = openuem_utils.ReadPEMPrivateKey(w.ConsolePrivateKeyPath)
	if err != nil {
		log.Println("[ERROR]: could not read OCSP private key")
		return err
	}

	key, err = cfg.Section("Server").GetKey("console_jwt_key")
	if err != nil {
		return err
	}
	w.JWTKey = key.String()

	key, err = cfg.Section("Server").GetKey("console_server_name")
	if err != nil {
		return err
	}
	w.ServerName = key.String()

	key, err = cfg.Section("Server").GetKey("console_port")
	if err != nil {
		return err
	}
	w.ConsolePort = key.String()

	key, err = cfg.Section("Server").GetKey("auth_port")
	if err != nil {
		return err
	}
	w.AuthPort = key.String()

	key, err = cfg.Section("Server").GetKey("domain")
	if err != nil {
		return err
	}
	w.Domain = key.String()

	key, err = cfg.Section("Server").GetKey("nats_url")
	if err != nil {
		return err
	}
	w.NATSServers = key.String()

	return nil
}
