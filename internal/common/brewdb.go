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

func (w *Worker) StartBrewDBDownloadJob() error {
	// Try to download at start
	if err := w.DownloaBrewDB(); err != nil {
		log.Printf("[ERROR]: could not get brew.db, reason: %v", err)
		w.DownloadBrewJobDuration = 30 * time.Minute
	} else {
		log.Println("[INFO]: brew.db has been downloaded")
		w.DownloadBrewJobDuration = 24 * time.Hour
	}

	// Create task
	if err := w.StartDownloadBrewDBJob(); err != nil {
		log.Printf("[ERROR]: could not start the brew.db download job: %v", err)
		return err
	}
	log.Printf("[INFO]: download brew.db job has been scheduled every %d minutes", 30)
	return nil
}

func (w *Worker) DownloaBrewDB() error {
	url := "https://downloads.openuem.eu/brew/brew.db"

	// If we're in development don't download
	if os.Getenv("DEVEL") == "true" {
		return nil
	}

	dbPath := filepath.Join(w.BrewDBFolder, "brew.db")
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

func (w *Worker) StartDownloadBrewDBJob() error {
	var err error
	var jobDuration time.Duration
	w.DownloadBrewDBJob, err = w.TaskScheduler.NewJob(
		gocron.DurationJob(
			time.Duration(w.DownloadBrewJobDuration),
		),
		gocron.NewTask(
			func() {
				if err := w.DownloaBrewDB(); err != nil {
					log.Printf("[ERROR]: could not get brew.db, reason: %v", err)
					jobDuration = 2 * time.Minute
				} else {
					jobDuration = 24 * time.Hour
				}

				if jobDuration.String() == w.DownloadBrewJobDuration.String() {
					return
				}

				w.DownloadBrewJobDuration = jobDuration
				w.TaskScheduler.RemoveJob(w.DownloadBrewDBJob.ID())
				if err := w.StartDownloadBrewDBJob(); err == nil {
					log.Println("[INFO]: download brew db job has been re-scheduled every " + w.DownloadBrewJobDuration.String())
				}
			},
		),
	)
	if err != nil {
		log.Printf("[ERROR]: could not schedule brew db Job, reason: %v", err)
		return err
	}
	return nil
}
