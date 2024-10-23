package commands

import (
	"log"

	"github.com/doncicuto/openuem-console/internal/models"
	"github.com/doncicuto/openuem_utils"
	"github.com/go-playground/validator"
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

	log.Println("... reading console's certificate")
	_, err = openuem_utils.ReadPEMCertificate(cCtx.String("cert"))
	if err != nil {
		return err
	}
	command.CertPath = cCtx.String("cert")

	log.Println("... reading console's private key")
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

	validate := validator.New()
	err = validate.Var(cCtx.String("nats-host"), "hostname")
	if err != nil {
		return err
	}
	command.NATSHost = cCtx.String("nats-host")

	err = validate.Var(cCtx.String("nats-port"), "numeric")
	if err != nil {
		return err
	}
	command.NATSPort = cCtx.String("nats-port")

	command.JWTKey = cCtx.String("jwt-key")
	return nil
}
