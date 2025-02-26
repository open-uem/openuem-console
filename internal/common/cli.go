package common

import (
	"github.com/open-uem/utils"
	"github.com/urfave/cli/v2"
)

func (w *Worker) GenerateConsoleConfigFromCLI(cCtx *cli.Context) error {
	var err error

	w.DBUrl = cCtx.String("dburl")

	w.CACertPath = cCtx.String("cacert")
	_, err = utils.ReadPEMCertificate(w.CACertPath)
	if err != nil {
		return err
	}

	w.ConsoleCertPath = cCtx.String("cert")
	_, err = utils.ReadPEMCertificate(w.ConsoleCertPath)
	if err != nil {
		return err
	}

	w.ConsolePrivateKeyPath = cCtx.String("key")
	_, err = utils.ReadPEMPrivateKey(w.ConsolePrivateKeyPath)
	if err != nil {
		return err
	}

	w.SFTPPrivateKeyPath = cCtx.String("sftpkey")
	_, err = utils.ReadPEMPrivateKey(w.SFTPPrivateKeyPath)
	if err != nil {
		return err
	}

	w.NATSServers = cCtx.String("nats-servers")

	w.JWTKey = cCtx.String("jwt-key")

	w.ConsolePort = cCtx.String("console-port")
	w.AuthPort = cCtx.String("auth-port")
	w.ServerName = cCtx.String("server-name")
	w.Domain = cCtx.String("domain")
	w.OrgName = cCtx.String("org-name")
	w.OrgProvince = cCtx.String("org-province")
	w.OrgLocality = cCtx.String("org-locality")
	w.OrgAddress = cCtx.String("org-address")
	w.Country = cCtx.String("country")
	w.ReverseProxyAuthPort = cCtx.String("reverse-proxy-auth-port")
	w.ReverseProxyServer = cCtx.String("reverse-proxy-server")
	w.Version = "0.3.0"

	return nil
}
