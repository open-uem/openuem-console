package main

import (
	"log"

	"github.com/doncicuto/openuem-console/internal/common"
	"github.com/doncicuto/openuem_utils"
	"golang.org/x/sys/windows/svc"
)

func main() {
	w := common.NewWorker("openuem-console-service.txt")
	if err := w.GenerateConsoleConfig(); err != nil {
		log.Printf("[ERROR]: could not generate config for OpenUEM console: %v", err)
	}

	// Create temp directory for downloads
	if err := w.CreateDowloadTempDir(); err != nil {
		log.Printf("[ERROR]: could not create download temp dir: %v", err)
	}

	s := openuem_utils.NewOpenUEMWindowsService()
	s.ServiceStart = w.StartWorker
	s.ServiceStop = w.StopWorker

	// Run service
	err := svc.Run("openuem-console-service", s)
	if err != nil {
		log.Printf("[ERROR]: could not run service: %v", err)
	}
}
