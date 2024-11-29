package common

import (
	"archive/zip"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-co-op/gocron/v2"
)

func (w *Worker) StartWinGetDBDownloadJob() error {
	var err error

	// Try to download at start
	if err := w.DownloadWgetDB(); err != nil {
		log.Printf("[ERROR]: could not get index.db, reason: %v", err)
	} else {
		log.Println("[INFO]: winget index.db has been downloaded")
	}

	// Create task
	_, err = w.TaskScheduler.NewJob(
		gocron.DurationJob(
			time.Duration(time.Duration(24*time.Hour)),
		),
		gocron.NewTask(
			func() {
				if err := w.DownloadWgetDB(); err != nil {
					log.Printf("[ERROR]: could not get index.db, reason: %v", err)
					return
				}
			},
		),
	)
	if err != nil {
		log.Printf("[FATAL]: could not start the download directory clean job: %v", err)
		return err
	}
	log.Println("[INFO]: download index.db job has been scheduled every day")
	return nil
}

func (w *Worker) DownloadWgetDB() error {
	url := "https://cdn.winget.microsoft.com/cache/source.msix"

	// If we're in development don't download
	if os.Getenv("DEVEL") == "true" {
		return nil
	}

	zipPath := filepath.Join(w.WinGetDBFolder, "winget.zip")
	out, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	out.Close()

	// Open ZIP reader
	archive, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, f := range archive.File {
		if f.Name == "Public/index.db" {
			src, err := f.Open()
			if err != nil {
				return err
			}
			defer src.Close()

			dst, err := os.Create(filepath.Join(w.WinGetDBFolder, "index.db"))
			if err != nil {
				return err
			}
			defer dst.Close()

			_, err = io.Copy(dst, src)
			if err != nil {
				return err
			}
			break
		}
	}
	archive.Close()

	// Remove temp ZIP file
	return os.Remove(zipPath)
}
