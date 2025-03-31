//go:build windows

package main

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/go-co-op/gocron/v2"
	"github.com/open-uem/openuem-console/internal/common"
	"github.com/open-uem/utils"
	"golang.org/x/sys/windows/svc"
)

func main() {
	var err error

	w := common.NewWorker("openuem-console-service.txt")

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

	// Get working directory
	cwd, err := utils.GetWd()
	if err != nil {
		log.Fatal("[FATAL]: could not get working directory")
	}

	// Create temp directory for downloads
	w.DownloadDir = filepath.Join(cwd, "tmp", "download")
	if strings.HasSuffix(cwd, "tmp") {
		w.DownloadDir = filepath.Join(cwd, "download")
	}

	if err := w.CreateDowloadTempDir(); err != nil {
		log.Fatalf("[FATAL]: could not create download temp dir: %v", err)
	}

	// Create winget directory for index.db
	w.WinGetDBFolder = filepath.Join(cwd, "tmp", "winget")
	if err := w.CreateWingetDBDir(); err != nil {
		log.Fatalf("[FATAL]: could not create winget temp dir: %v", err)
	}

	// Create winget directory for index.db
	w.FlatpakDBFolder = filepath.Join(cwd, "tmp", "flatpak")
	if err := w.CreateFlatpakDBDir(); err != nil {
		log.Fatalf("[FATAL]: could not create flatpak temp dir: %v", err)
	}

	// Create server releases directory
	w.ServerReleasesFolder = filepath.Join(cwd, "tmp", "server-releases")
	if err := w.CreateServerReleasesDir(); err != nil {
		log.Fatalf("[FATAL]: could not create server releases temp dir: %v", err)
	}

	// Configure the windows service
	s := utils.NewOpenUEMWindowsService()
	s.ServiceStart = w.StartWorker
	s.ServiceStop = w.StopWorker

	// Run service
	if err := svc.Run("openuem-console-service", s); err != nil {
		log.Printf("[ERROR]: could not run service: %v", err)
	}
}
