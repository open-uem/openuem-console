package common

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-co-op/gocron/v2"
)

func (w *Worker) StartFlatpakDBDownloadJob() error {
	// Try to download at start
	if err := w.DownloaFlatpakDB(); err != nil {
		log.Printf("[ERROR]: could not get flatpak.db, reason: %v", err)
		w.DownloadFlatpakJobDuration = 30 * time.Minute
	} else {
		log.Println("[INFO]: flatpak.db has been downloaded")
		w.DownloadFlatpakJobDuration = 24 * time.Hour
	}

	// Create task
	if err := w.StartDownloadFlatpakDBJob(); err != nil {
		log.Printf("[ERROR]: could not start the flatpak download job: %v", err)
		return err
	}
	log.Printf("[INFO]: download flatpak.db job has been scheduled every %d minutes", 30)
	return nil
}

func (w *Worker) DownloaFlatpakDB() error {
	url := "https://downloads.openuem.eu/flatpak/flatpak.db"

	// If we're in development don't download
	if os.Getenv("DEVEL") == "true" {
		return nil
	}

	dbPath := filepath.Join(w.FlatpakDBFolder, "flatpak.db")
	out, err := os.Create(dbPath)
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

	return nil
}

func (w *Worker) StartDownloadFlatpakDBJob() error {
	var err error
	var jobDuration time.Duration
	w.DownloadFlatpakDBJob, err = w.TaskScheduler.NewJob(
		gocron.DurationJob(
			time.Duration(w.DownloadFlatpakJobDuration),
		),
		gocron.NewTask(
			func() {
				if err := w.DownloaFlatpakDB(); err != nil {
					log.Printf("[ERROR]: could not get flatpak.db, reason: %v", err)
					jobDuration = 2 * time.Minute
				} else {
					jobDuration = 24 * time.Hour
				}

				if jobDuration.String() == w.DownloadFlatpakJobDuration.String() {
					return
				}

				w.DownloadFlatpakJobDuration = jobDuration
				w.TaskScheduler.RemoveJob(w.DownloadFlatpakDBJob.ID())
				if err := w.StartDownloadFlatpakDBJob(); err == nil {
					log.Println("[INFO]: download flatpak db job has been re-scheduled every " + w.DownloadFlatpakJobDuration.String())
				}
			},
		),
	)
	if err != nil {
		log.Printf("[ERROR]: could not schedule flatpak db Job, reason: %v", err)
		return err
	}
	return nil
}
