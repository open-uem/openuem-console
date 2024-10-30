package commands

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/doncicuto/openuem-console/internal/common"
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

/*
	 func startConsole(cCtx *cli.Context) error {
		var err error
		command := ConsoleCommand{}

		command.CheckRequisites(cCtx)

		// TODO: NATS connection close in stop?
		command.NATSConnection, err = openuem_nats.ConnectWithNATS(command.NATSServers, command.CertPath, command.CertKey, command.CACertPath)
		if err != nil {
			log.Println("[ERROR]: Error connecting to NATS", err.Error())
		}
		// Session handler
		sessionsHandler := sessions.New(command.DBUrl)
		defer sessionsHandler.Close()

		// HTTPS web server
		w := webserver.New(command.Model, command.NATSConnection, sessionsHandler, command.JWTKey, command.CertPath, command.CertKey, command.CACertPath)
		go w.Serve(":1323", command.CertPath, command.CertKey)
		defer func() {
			if err := w.Close(); err != nil {
				log.Println("[ERROR]: Error closing the web server")
			} else {
				log.Println("[ERROR]: Closing Web server")
			}
		}()

		// HTTPS auth server
		a := authserver.New(command.Model, sessionsHandler, command.CACertPath)
		go a.Serve(":1324", command.CertPath, command.CertKey)
		defer func() {
			if err := w.Close(); err != nil {
				log.Println("[ERROR]: Error closing the auth server")
			} else {
				log.Println("[ERROR]: Closing Auth server")
			}
		}()

		// Keep the connection alive
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
		log.Printf("[INFO]: OpenUEM console is ready\n\n")
		<-done

		log.Printf("[INFO]: OpenUEM console has stopped listening\n\n")
		return nil
	}
*/
func startConsole(cCtx *cli.Context) error {
	worker := common.NewWorker("")

	if err := worker.GenerateConsoleConfigFromCLI(cCtx); err != nil {
		log.Printf("[ERROR]: could not generate config for OpenUEM Console: %v", err)
	}

	// Save pid to PIDFILE
	if err := os.WriteFile("PIDFILE", []byte(strconv.Itoa(os.Getpid())), 0666); err != nil {
		return err
	}

	worker.StartWorker()

	// Keep the connection alive
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	log.Printf("[INFO]: OpenUEM Console is ready and listening on %s\n", cCtx.String("address"))
	<-done

	worker.StopWorker()

	log.Printf("[INFO]: OpenUEM Console has stopped listening\n")
	return nil
}
