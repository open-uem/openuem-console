package common

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/open-uem/ent/release"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/utils"
)

func (w *Worker) StartCheckLatestReleasesJob(channel string) error {
	if err := w.GetLatestReleases(channel); err != nil {
		log.Printf("[ERROR]: could not get latest agent releases, reason: %v", err)
		w.DownloadLatestReleaseJobDuration = 2 * time.Minute
	} else {
		log.Println("[INFO]: latest agent releases have been checked")
		w.DownloadLatestReleaseJobDuration = 6 * time.Hour
	}

	// Create task
	if err := w.StartDownloadLatestAgentReleaseJob(channel); err != nil {
		return err
	}
	log.Printf("[INFO]: check latest releases job has been scheduled every %d hours", 6)
	return nil
}

func (w *Worker) GetLatestReleases(channel string) error {
	if err := w.CheckAgentLatestReleases(channel); err != nil {
		return err
	}

	if err := w.CheckServerLatestReleases(channel); err != nil {
		return err
	}

	return nil
}

func (w *Worker) CheckAgentLatestReleases(channel string) error {
	// Check agent release against our API
	url := fmt.Sprintf("https://releases.openuem.eu/api?action=latestAgentRelease&channel=%s", channel)

	body, err := utils.QueryReleasesEndpoint(url)
	if err != nil {
		return err
	}

	latestAgentRelease := openuem_nats.OpenUEMRelease{}
	if err := json.Unmarshal(body, &latestAgentRelease); err != nil {
		return err
	}

	if err := w.Model.SaveNewReleaseAvailable(release.ReleaseTypeAgent, latestAgentRelease); err != nil {
		return err
	}

	return nil
}

func (w *Worker) CheckServerLatestReleases(channel string) error {
	// Check server release against our API
	url := fmt.Sprintf("https://releases.openuem.eu/api?action=latestServerRelease&channel=%s", channel)

	body, err := utils.QueryReleasesEndpoint(url)
	if err != nil {
		return err
	}

	latestServerRelease := openuem_nats.OpenUEMRelease{}
	if err := json.Unmarshal(body, &latestServerRelease); err != nil {
		return err
	}

	if err := w.Model.SaveNewReleaseAvailable(release.ReleaseTypeServer, latestServerRelease); err != nil {
		return err
	}

	return nil
}

func (w *Worker) StartDownloadLatestAgentReleaseJob(channel string) error {
	var err error
	var jobDuration time.Duration
	w.DownloadLatestReleaseJob, err = w.TaskScheduler.NewJob(
		gocron.DurationJob(
			time.Duration(w.DownloadLatestReleaseJobDuration.Nanoseconds()),
		),
		gocron.NewTask(
			func() {
				if err := w.GetLatestReleases(channel); err != nil {
					log.Printf("[ERROR]: could not get latest agent releases, reason: %v", err)
					jobDuration = 2 * time.Minute
				} else {
					jobDuration = 6 * time.Hour
				}

				if jobDuration.String() == w.DownloadLatestReleaseJobDuration.String() {
					return
				}

				w.DownloadLatestReleaseJobDuration = jobDuration
				w.TaskScheduler.RemoveJob(w.DownloadLatestReleaseJob.ID())
				if err := w.StartDownloadLatestAgentReleaseJob(channel); err == nil {
					log.Println("[INFO]: download latest agent releases job has been re-scheduled every " + w.DownloadLatestReleaseJobDuration.String())
				}
			},
		),
	)
	if err != nil {
		log.Printf("[ERROR]: could not schedule latest agent releases job, reason: %v", err)
		return err
	}
	return nil
}
