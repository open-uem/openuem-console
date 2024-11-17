//go:build windows

package common

import (
	"encoding/binary"
	"fmt"
	"log"
	"path/filepath"

	"github.com/danieljoos/wincred"
	"github.com/doncicuto/openuem_utils"
	"golang.org/x/sys/windows/registry"
)

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
