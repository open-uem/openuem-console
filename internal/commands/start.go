package commands

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/go-co-op/gocron/v2"
	"github.com/open-uem/openuem-console/internal/common"
	"github.com/open-uem/utils"
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
	cwd, err := utils.GetWd()
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

	// Create server releases directory
	worker.ServerReleasesFolder = filepath.Join(cwd, "tmp", "server-releases")
	if strings.HasSuffix(cwd, "tmp") {
		worker.ServerReleasesFolder = filepath.Join(cwd, "server-releases")
	}
	if err := worker.CreateServerReleasesDir(); err != nil {
		log.Fatalf("[FATAL]: could not create server releases temp dir: %v", err)
	}

	// Create winget directory
	worker.WinGetDBFolder = filepath.Join(cwd, "tmp", "winget")
	if strings.HasSuffix(cwd, "tmp") {
		worker.ServerReleasesFolder = filepath.Join(cwd, "winget")
	}
	if err := worker.CreateWingetDBDir(); err != nil {
		log.Fatalf("[FATAL]: could not create winget temp dir: %v", err)
	}

	// Save pid to PIDFILE
	if err := os.WriteFile("PIDFILE", []byte(strconv.Itoa(os.Getpid())), 0666); err != nil {
		return err
	}

	// Start Task Scheduler
	worker.TaskScheduler, err = gocron.NewScheduler()
	if err != nil {
		log.Fatalf("[FATAL]: could not create task scheduler, reason: %s", err.Error())
	}
	worker.TaskScheduler.Start()
	log.Println("[INFO]: task scheduler has been started")

	// Start worker
	worker.StartWorker()

	// Keep the connection alive
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-done

	worker.StopWorker()

	return nil
}
