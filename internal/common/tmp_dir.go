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

func (w *Worker) CreateFlatpakDBDir() error {
	if _, err := os.Stat(w.FlatpakDBFolder); os.IsNotExist(err) {
		if err := os.MkdirAll(w.FlatpakDBFolder, 0770); err != nil {
			return err
		}
	}

	return nil
}

func (w *Worker) CreateBrewDBDir() error {
	if _, err := os.Stat(w.BrewDBFolder); os.IsNotExist(err) {
		if err := os.MkdirAll(w.BrewDBFolder, 0770); err != nil {
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
