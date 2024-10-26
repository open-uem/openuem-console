package commands

import (
	"log"

	"github.com/doncicuto/openuem-console/internal/models"
	"github.com/doncicuto/openuem_utils"
	"github.com/urfave/cli/v2"
)

func (command *ConsoleCommand) CheckRequisites(cCtx *cli.Context) error {
	var err error

	log.Println("... reading CA certificate", cCtx.String("cacert"))
	command.CACert, err = openuem_utils.ReadPEMCertificate(cCtx.String("cacert"))
	if err != nil {
		return err
	}
	command.CACertPath = cCtx.String("cacert")

	log.Println("... reading console's certificate", cCtx.String("cert"))
	_, err = openuem_utils.ReadPEMCertificate(cCtx.String("cert"))
	if err != nil {
		return err
	}
	command.CertPath = cCtx.String("cert")

	log.Println("... reading console's private key", cCtx.String("key"))
	_, err = openuem_utils.ReadPEMPrivateKey(cCtx.String("key"))
	if err != nil {
		return err
	}
	command.CertKey = cCtx.String("key")

	log.Println("... connecting to database")
	command.DBUrl = cCtx.String("dburl")
	command.Model, err = models.New(command.DBUrl)
	if err != nil {
		log.Fatalf("❌ could not connect to database, reason: %s", err.Error())
	}

	command.NATSServers = cCtx.String("nats-servers")

	command.JWTKey = cCtx.String("jwt-key")
	return nil
}
