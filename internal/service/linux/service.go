//go:build linux

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-co-op/gocron/v2"
	"github.com/open-uem/openuem-console/internal/common"
)

func main() {
	var err error

	w := common.NewWorker("openuem-console-service")

	// Start Task Scheduler
	w.TaskScheduler, err = gocron.NewScheduler()
	if err != nil {
		log.Fatalf("[FATAL]: could not create task scheduler, reason: %s", err.Error())
		return
	}
	w.TaskScheduler.Start()
	log.Println("[INFO]: task scheduler has been started")

	if err := w.GenerateConsoleConfig(); err != nil {
		log.Printf("[ERROR]: could not generate config for OpenUEM console: %v", err)
		if err := w.StartGenerateConsoleConfigJob(); err != nil {
			log.Fatalf("[FATAL]: could not start job to generate config for OpenUEM console: %v", err)
		}
	}

	// Create temp directory for downloads
	w.DownloadDir = "/tmp/downloads"
	if err := w.CreateDowloadTempDir(); err != nil {
		log.Fatalf("[FATAL]: could not create download temp dir: %v", err)
	}

	// Create winget directory for index.db
	w.WinGetDBFolder = "/tmp/winget"
	if err := w.CreateWingetDBDir(); err != nil {
		log.Fatalf("[FATAL]: could not create winget temp dir: %v", err)
	}

	// Create flatpak directory for flatpak.db
	w.FlatpakDBFolder = "/tmp/flatpak"
	if err := w.CreateFlatpakDBDir(); err != nil {
		log.Fatalf("[FATAL]: could not create flatpak temp dir: %v", err)
	}

	// Create common software directory for common.db
	w.CommonSoftwareDBFolder = "/tmp/commondb"
	if err := w.CreateCommonSoftwareDBDir(); err != nil {
		log.Fatalf("[FATAL]: could not create commondb temp dir: %v", err)
	}

	// Create server releases directory
	w.ServerReleasesFolder = "/tmp/server-releases"
	if err := w.CreateServerReleasesDir(); err != nil {
		log.Fatalf("[FATAL]: could not create server releases temp dir: %v", err)
	}

	// Start the worker
	w.StartWorker()

	// Keep the connection alive
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-done

	w.StopWorker()
	log.Printf("[INFO]: the Console and Auth servers have stopped\n")
}
