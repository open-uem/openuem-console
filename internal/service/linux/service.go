//go:build linux

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/doncicuto/openuem-console/internal/common"
)

func main() {
	w := common.NewWorker("openuem-console-service")
	if err := w.GenerateConsoleConfig(); err != nil {
		log.Fatalf("[FATAL]: could not generate config for OpenUEM console: %v", err)
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

	// Start the worker
	w.StartWorker()

	// Keep the connection alive
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	log.Println("[INFO]: the Console and Auth servers are ready and waiting for requests")
	<-done

	w.StopWorker()
	log.Printf("[INFO]: the Console and Auth servers have stopped\n")
}
