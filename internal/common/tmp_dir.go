package common

import (
	"log"
	"os"
)

func (w *Worker) CreateDowloadTempDir() error {
	if _, err := os.Stat(w.DownloadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(w.DownloadDir, 0770); err != nil {
			log.Printf("[ERROR]: could not create temp download directory, reason: %v", err)
			return err
		}
	}

	return nil
}

func (w *Worker) CreateWingetDBDir() error {
	if _, err := os.Stat(w.WinGetDBFolder); os.IsNotExist(err) {
		if err := os.MkdirAll(w.WinGetDBFolder, 0770); err != nil {
			log.Printf("[ERROR]: could not create temp download directory, reason: %v", err)
			return err
		}
	}

	return nil
}
