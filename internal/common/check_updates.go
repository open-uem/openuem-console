package common

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/doncicuto/openuem_ent/release"
	"github.com/doncicuto/openuem_nats"
	"github.com/doncicuto/openuem_utils"
	"github.com/go-co-op/gocron/v2"
)

func (w *Worker) StartCheckLatestReleasesJob(channel string) error {
	if err := w.GetLatestReleases(channel); err != nil {
		log.Printf("[ERROR]: could not get latest agent releases, reason: %v", err)
	} else {
		log.Println("[INFO]: latest agent releases have been checked")
	}

	// Create task
	_, err := w.TaskScheduler.NewJob(
		gocron.DurationJob(
			time.Duration(time.Duration(6*time.Hour)),
		),
		gocron.NewTask(
			func() {
				if err := w.GetLatestReleases(channel); err != nil {
					log.Printf("[ERROR]: could not get latest agent releases, reason: %v", err)
					return
				}
			},
		),
	)
	if err != nil {
		return err
	}
	log.Println("[INFO]: check latest releases job has been scheduled every 6 hours")
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

	body, err := openuem_utils.QueryReleasesEndpoint(url)
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

	body, err := openuem_utils.QueryReleasesEndpoint(url)
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
