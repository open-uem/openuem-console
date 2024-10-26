package commands

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/doncicuto/openuem-console/internal/controllers/authserver"
	"github.com/doncicuto/openuem-console/internal/controllers/sessions"
	"github.com/doncicuto/openuem-console/internal/controllers/webserver"
	"github.com/doncicuto/openuem_nats"

	"github.com/urfave/cli/v2"
)

func StartConsole() *cli.Command {
	return &cli.Command{
		Name:   "start",
		Usage:  "Start the OpenUEM console",
		Action: startConsole,
		Flags:  StartConsoleFlags(),
	}
}

func startConsole(cCtx *cli.Context) error {
	var err error
	command := ConsoleCommand{}

	command.CheckRequisites(cCtx)

	// TODO: NATS connection close in stop?
	command.NATSConnection, err = openuem_nats.ConnectWithNATS(command.NATSServers, command.CertPath, command.CertKey, command.CACertPath)
	if err != nil {
		log.Println("Error connecting to NATS", err.Error())
	}
	// Session handler
	sessionsHandler := sessions.New(command.DBUrl)
	defer sessionsHandler.Close()

	// HTTPS web server
	w := webserver.New(command.Model, command.NATSConnection, sessionsHandler, command.JWTKey, command.CertPath, command.CertKey, command.CACertPath)
	go w.Serve(":1323", command.CertPath, command.CertKey)
	defer func() {
		if err := w.Close(); err != nil {
			log.Println("Error closing the web server")
		} else {
			log.Println("Closing Web server")
		}
	}()

	// HTTPS auth server
	a := authserver.New(command.Model, sessionsHandler, command.CACertPath)
	go a.Serve(":1324", command.CertPath, command.CertKey)
	defer func() {
		if err := w.Close(); err != nil {
			log.Println("Error closing the auth server")
		} else {
			log.Println("Closing Auth server")
		}
	}()

	// Keep the connection alive
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("✅ Done! OpenUEM console is ready\n\n")
	<-done

	log.Printf("👋 Done! OpenUEM console has stopped listening\n\n")
	return nil
}
