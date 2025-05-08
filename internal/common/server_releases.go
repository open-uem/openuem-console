package common

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/open-uem/utils"
)

func (w *Worker) GetServerReleases() error {
	settings, err := w.Model.GetGeneralSettings("-1")
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://releases.openuem.eu/api?action=latestServerRelease&channel=%s", settings.UpdateChannel)

	body, err := utils.QueryReleasesEndpoint(url)
	if err != nil {
		return err
	}

	latestServerReleaseFilePath := filepath.Join(w.ServerReleasesFolder, "latest.json")

	if err := os.WriteFile(latestServerReleaseFilePath, body, 0660); err != nil {
		return err
	}

	url = fmt.Sprintf("https://releases.openuem.eu/api?action=allServerReleases&channel=%s", settings.UpdateChannel)

	body, err = utils.QueryReleasesEndpoint(url)
	if err != nil {
		return err
	}

	allServerReleasesFilePath := filepath.Join(w.ServerReleasesFolder, "releases.json")

	if err := os.WriteFile(allServerReleasesFilePath, body, 0660); err != nil {
		return err
	}

	return nil
}

func (w *Worker) StartServerReleasesDownloadJob() error {
	// Try to download server releases at start
	if err := w.GetServerReleases(); err != nil {
		log.Printf("[ERROR]: could not get server releases, reason: %v", err)
		w.DownloadServerReleasesJobDuration = 10 * time.Minute
	} else {
		log.Println("[INFO]: server releases files have been downloaded")
		w.DownloadServerReleasesJobDuration = 6 * time.Hour
	}

	// Create task
	if err := w.StartDownloadServerReleasesJob(); err == nil {
		log.Printf("[INFO]: download server releases job has been scheduled every %d minutes", 10)
	}
	return nil
}

func (w *Worker) StartDownloadServerReleasesJob() error {
	var err error
	var jobDuration time.Duration
	w.DownloadServerReleasesJob, err = w.TaskScheduler.NewJob(
		gocron.DurationJob(
			time.Duration(w.DownloadServerReleasesJobDuration),
		),
		gocron.NewTask(
			func() {
				if err := w.GetServerReleases(); err != nil {
					log.Printf("[ERROR]: could not get server releases, reason: %v", err)
					jobDuration = 2 * time.Minute
				} else {
					jobDuration = 6 * time.Hour
				}

				if jobDuration.String() == w.DownloadServerReleasesJobDuration.String() {
					return
				}

				w.DownloadServerReleasesJobDuration = jobDuration
				w.TaskScheduler.RemoveJob(w.DownloadServerReleasesJob.ID())
				if err := w.StartDownloadServerReleasesJob(); err == nil {
					log.Println("[INFO]: download server releases job has been re-scheduled every " + w.DownloadServerReleasesJobDuration.String())
				}
			},
		),
	)
	if err != nil {
		log.Printf("[ERROR]: could not schedule Server Releases Job, reason: %v", err)
		return err
	}
	return nil
}
