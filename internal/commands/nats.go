package commands

import (
	"log"

	"github.com/doncicuto/openuem_nats"
)

func (command *ConsoleCommand) connectToNATS() error {
	log.Println("🔌  connecting to NATS cluster")
	command.MessageServer = openuem_nats.New(command.NATSHost, command.NATSPort, command.CertPath, command.CertKey, command.CACert)
	return command.MessageServer.Connect()
}
