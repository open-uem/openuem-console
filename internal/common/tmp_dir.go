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

func (w *Worker) CreateCommonSoftwareDBDir() error {
	if _, err := os.Stat(w.CommonSoftwareDBFolder); os.IsNotExist(err) {
		if err := os.MkdirAll(w.CommonSoftwareDBFolder, 0770); err != nil {
			return err
		}
	}

	return nil
}

func (w *Worker) CreateServerReleasesDir() error {
	if _, err := os.Stat(w.ServerReleasesFolder); os.IsNotExist(err) {
		if err := os.MkdirAll(w.ServerReleasesFolder, 0770); err != nil {
			return err
		}
	}

	return nil
}
