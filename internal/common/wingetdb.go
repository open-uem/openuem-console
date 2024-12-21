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
	// Try to download at start
	if err := w.DownloadWgetDB(); err != nil {
		log.Printf("[ERROR]: could not get index.db, reason: %v", err)
		w.DownloadWingetJobDuration = 5 * time.Second
	} else {
		log.Println("[INFO]: winget index.db has been downloaded")
		w.DownloadWingetJobDuration = 24 * time.Hour
	}

	// Create task
	if err := w.StartDownloadWingetDBJob(); err != nil {
		log.Printf("[ERROR]: could not start the winget download job: %v", err)
		return err
	}
	log.Println("[INFO]: download index.db job has been scheduled every ", w.DownloadWingetJobDuration.String())
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

func (w *Worker) StartDownloadWingetDBJob() error {
	var err error
	var jobDuration time.Duration
	w.DownloadWingetDBJob, err = w.TaskScheduler.NewJob(
		gocron.DurationJob(
			time.Duration(w.DownloadWingetJobDuration),
		),
		gocron.NewTask(
			func() {
				if err := w.DownloadWgetDB(); err != nil {
					log.Printf("[ERROR]: could not get index.db, reason: %v", err)
					jobDuration = 2 * time.Minute
				} else {
					jobDuration = 24 * time.Hour
				}

				if jobDuration.String() == w.DownloadWingetJobDuration.String() {
					return
				}

				w.DownloadWingetJobDuration = jobDuration
				w.TaskScheduler.RemoveJob(w.DownloadWingetDBJob.ID())
				if err := w.StartDownloadWingetDBJob(); err == nil {
					log.Println("[INFO]: download winget db job has been re-scheduled every " + w.DownloadWingetJobDuration.String())
				}
			},
		),
	)
	if err != nil {
		log.Printf("[ERROR]: could not schedule winget db Job, reason: %v", err)
		return err
	}
	return nil
}
