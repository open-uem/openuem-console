package commands

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/doncicuto/openuem-console/internal/common"
	"github.com/doncicuto/openuem_utils"
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
		log.Fatalf("[FATAL]: could not generate config for OpenUEM Console: %v", err)
	}

	// Get working directory
	cwd, err := openuem_utils.GetWd()
	if err != nil {
		log.Fatal("[FATAL]: could not get working directory")
	}

	// Create temp directory for downloads
	worker.DownloadDir = filepath.Join(cwd, "tmp", "download")
	if strings.HasSuffix(cwd, "tmp") {
		worker.DownloadDir = filepath.Join(cwd, "download")
	}

	if err := worker.CreateDowloadTempDir(); err != nil {
		log.Fatalf("[ERROR]: could not create download temp dir: %v", err)
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
