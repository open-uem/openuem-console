package common

import (
	"os"
)

func (w *Worker) CreateDowloadTempDir() error {
	if _, err := os.Stat(w.DownloadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(w.DownloadDir, 0770); err != nil {
			return err
		}
	}

	return nil
}

func (w *Worker) CreateWingetDBDir() error {
	if _, err := os.Stat(w.WinGetDBFolder); os.IsNotExist(err) {
		if err := os.MkdirAll(w.WinGetDBFolder, 0770); err != nil {
			return err
		}
	}

	return nil
}
