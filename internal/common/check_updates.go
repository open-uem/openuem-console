package common

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/doncicuto/openuem_ent/release"
	"github.com/doncicuto/openuem_nats"
	"github.com/go-co-op/gocron/v2"
)

func (w *Worker) StartCheckLatestReleasesJob(channel string) error {
	// Try to do some checks at start
	if err := w.GetLatestReleases(channel); err != nil {
		log.Printf("[ERROR]: could not get latest releases, reason: %v", err)
	} else {
		log.Println("[INFO]: latest releases have been checked")
	}

	// Create task
	_, err := w.TaskScheduler.NewJob(
		gocron.DurationJob(
			time.Duration(time.Duration(6*time.Hour)),
		),
		gocron.NewTask(
			func() {
				if err := w.GetLatestReleases(channel); err != nil {
					log.Printf("[ERROR]: could not get latest releases, reason: %v", err)
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

	if err := conn.Publish("agent.update.updater", data); err != nil {
		log.Printf("[ERROR]: could not send update updater message to agents, reason: %v", err)
		return err
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

	if err := conn.Publish("agent.update.messenger", data); err != nil {
		log.Printf("[ERROR]: could not send update messenger message to agents, reason: %v", err)
		return err
	}

	return nil
}

func (w *Worker) CheckAgentLatestReleases(channel string) error {
	// Check agent release against our API
	url := fmt.Sprintf("https://releases.openuem.eu/api?action=latestAgentRelease&channel=%s", channel)

	body, err := QueryReleasesEndpoint(url, channel)
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

	body, err := QueryReleasesEndpoint(url, "")
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

	body, err := QueryReleasesEndpoint(url, "")
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

func QueryReleasesEndpoint(url, channel string) ([]byte, error) {
	client := http.Client{
		Timeout: time.Second * 8,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "openuem-console")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return nil, err
	}

	return body, nil
}
