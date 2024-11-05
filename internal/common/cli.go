package common

import (
	"github.com/doncicuto/openuem_utils"
	"github.com/urfave/cli/v2"
)

func (w *Worker) GenerateConsoleConfigFromCLI(cCtx *cli.Context) error {
	var err error

	w.DBUrl = cCtx.String("dburl")

	w.CACertPath = cCtx.String("cacert")
	_, err = openuem_utils.ReadPEMCertificate(w.CACertPath)
	if err != nil {
		return err
	}

	w.ConsoleCertPath = cCtx.String("cert")
	_, err = openuem_utils.ReadPEMCertificate(w.ConsoleCertPath)
	if err != nil {
		return err
	}

	w.ConsolePrivateKeyPath = cCtx.String("key")
	_, err = openuem_utils.ReadPEMPrivateKey(w.ConsolePrivateKeyPath)
	if err != nil {
		return err
	}

	w.NATSServers = cCtx.String("nats-servers")

	w.JWTKey = cCtx.String("jwt-key")

	w.ConsolePort = cCtx.String("console-port")
	w.AuthPort = cCtx.String("auth-port")
	w.ServerName = cCtx.String("server-name")
	w.Domain = cCtx.String("domain")

	return nil
}
