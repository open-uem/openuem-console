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

func startConsole(cCtx *cli.Context) error {
	worker := common.NewWorker("")

	if err := worker.GenerateConsoleConfigFromCLI(cCtx); err != nil {
		log.Printf("[ERROR]: could not generate config for OpenUEM Console: %v", err)
	}

	// Create temp directory for downloads
	if err := worker.CreateDowloadTempDir(); err != nil {
		log.Printf("[ERROR]: could not create download temp dir: %v", err)
	}

	// Save pid to PIDFILE
	if err := os.WriteFile("PIDFILE", []byte(strconv.Itoa(os.Getpid())), 0666); err != nil {
		return err
	}

	worker.StartWorker()

	// Keep the connection alive
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-done

	worker.StopWorker()

	return nil
}
