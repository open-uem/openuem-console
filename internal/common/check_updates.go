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
	log.Println("[INFO]: check latest agent releases job has been scheduled every 6 hours")
	return nil
}

func (w *Worker) GetLatestReleases(channel string) error {
	if err := w.CheckAgentLatestReleases(channel); err != nil {
		return err
	}

	latestUpdaterRelease, err := w.CheckUpdaterLatestReleases()
	if err != nil {
		return err
	}

	// Connect with NATS
	conn, err := openuem_nats.ConnectWithNATS(w.NATSServers, w.ConsoleCertPath, w.ConsolePrivateKeyPath, w.CACertPath)
	if err != nil {
		log.Printf("[ERROR]: could not connect to NATS, reason: %v", err)
		return err
	}

	data, err := json.Marshal(latestUpdaterRelease)
	if err != nil {
		log.Printf("[ERROR]: could not marshall update request for updater, reason: %v", err)
		return err
	}

	agents, err := w.Model.GetAllAgentsToUpdate()
	if err != nil {
		log.Printf("[ERROR]: could not get agents from database, reason: %v", err)
		return err
	}

	for _, a := range agents {
		if err := conn.Publish("agent.update.updater."+a.ID, data); err != nil {
			continue
		}
	}

	latestMessengerRelease, err := w.CheckMessengerLatestReleases()
	if err != nil {
		return err
	}

	data, err = json.Marshal(latestMessengerRelease)
	if err != nil {
		log.Printf("[ERROR]: could not marshall update request for messenger, reason: %v", err)
		return err
	}

	for _, a := range agents {
		if err := conn.Publish("agent.update.messenger."+a.ID, data); err != nil {
			continue
		}
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

func (w *Worker) CheckUpdaterLatestReleases() (*openuem_nats.OpenUEMRelease, error) {
	// Check agent release against our API
	url := fmt.Sprintf("https://releases.openuem.eu/api?action=latestUpdaterRelease")

	body, err := openuem_utils.QueryReleasesEndpoint(url)
	if err != nil {
		return nil, err
	}

	latestRelease := openuem_nats.OpenUEMRelease{}
	if err := json.Unmarshal(body, &latestRelease); err != nil {
		return nil, err
	}

	if err := w.Model.SaveNewReleaseAvailable(release.ReleaseTypeUpdater, latestRelease); err != nil {
		return nil, err
	}

	return &latestRelease, nil
}

func (w *Worker) CheckMessengerLatestReleases() (*openuem_nats.OpenUEMRelease, error) {
	// Check agent release against our API
	url := fmt.Sprintf("https://releases.openuem.eu/api?action=latestMessengerRelease")

	body, err := openuem_utils.QueryReleasesEndpoint(url)
	if err != nil {
		return nil, err
	}

	latestRelease := openuem_nats.OpenUEMRelease{}
	if err := json.Unmarshal(body, &latestRelease); err != nil {
		return nil, err
	}

	if err := w.Model.SaveNewReleaseAvailable(release.ReleaseTypeMessenger, latestRelease); err != nil {
		return nil, err
	}

	return &latestRelease, nil
}
