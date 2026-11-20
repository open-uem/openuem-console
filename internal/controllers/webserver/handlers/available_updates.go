package handlers

import (
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/models"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

const (
	// availableUpdatesJobInterval is how often the console asks the computers which
	// software updates they need
	availableUpdatesJobInterval = time.Hour
	// availableUpdatesNATSWait is how long the first run of the job waits for the
	// console to get connected to NATS
	availableUpdatesNATSWait = 15 * time.Minute
)

// agentUpdates is how many software updates a computer needed the last time the console
// asked it, and which organization and site it belongs to, so the dashboard's statistic
// can be scoped to them
type agentUpdates struct {
	Count    int
	TenantID int
	SiteID   int
}

// StartAvailableUpdatesJob asks the computers which software updates they need for the
// dashboard's statistic, and refreshes that count periodically. It's a background job:
// listing the updates of a computer runs its package manager, which takes minutes on
// Windows, so the dashboard itself never waits for it
func (h *Handler) StartAvailableUpdatesJob() error {
	h.availableUpdatesMu.Lock()
	h.AvailableUpdates = map[string]agentUpdates{}
	h.availableUpdatesMu.Unlock()

	job, err := h.TaskScheduler.NewJob(
		gocron.DurationJob(availableUpdatesJobInterval),
		gocron.NewTask(func() { h.RefreshAvailableUpdates() }),
	)
	if err != nil {
		return err
	}
	h.AvailableUpdatesJob = job

	// The first count can't wait an hour, but the console may still be connecting to
	// NATS, so the run starts as soon as the connection is there
	go func() {
		deadline := time.Now().Add(availableUpdatesNATSWait)
		for time.Now().Before(deadline) {
			if h.NATSConnection != nil && h.NATSConnection.IsConnected() {
				h.RefreshAvailableUpdates()
				return
			}
			time.Sleep(15 * time.Second)
		}
	}()

	return nil
}

// RefreshAvailableUpdates asks every computer, concurrently, which software updates it
// needs, and stores the count of each one. Computers that don't answer keep no count
// and aren't shown in the dashboard's statistic until they do
func (h *Handler) RefreshAvailableUpdates() {
	// Listing updates can take longer than the interval, requests would pile up
	if !h.refreshingAvailableUpdates.CompareAndSwap(false, true) {
		return
	}
	defer h.refreshingAvailableUpdates.Store(false)

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return
	}

	sources, err := h.Model.GetAllAgentOrgSites()
	if err != nil {
		log.Printf("[ERROR]: could not get the computers for the software updates statistic, reason: %v", err)
		return
	}

	// The full answer, not StatusOnly: the count is the number of available updates
	request, err := json.Marshal(openuem_nats.ListPackageUpdatesRequest{})
	if err != nil {
		return
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	next := map[string]agentUpdates{}
	limit := make(chan struct{}, updatePackagesConcurrency)

	for _, source := range sources {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limit <- struct{}{}
			defer func() { <-limit }()

			info := agentUpdates{TenantID: source.TenantID, SiteID: source.SiteID}

			msg, err := h.NATSConnection.Request("agent.listpackageupdates."+source.ID, request, listPackageUpdatesTimeout)
			if err != nil {
				// The computer is offline or doesn't support software updates: it
				// gets no entry, so it isn't counted until it answers
				return
			}

			response := openuem_nats.AvailableUpdatesResponse{}
			if err := json.Unmarshal(msg.Data, &response); err != nil {
				return
			}
			info.Count = len(response.Updates)

			mu.Lock()
			next[source.ID] = info
			mu.Unlock()
		}()
	}
	wg.Wait()

	h.availableUpdatesMu.Lock()
	h.AvailableUpdates = next
	h.availableUpdatesMu.Unlock()

	log.Printf("[INFO]: %d computers answered the software updates statistic", len(next))
}

// availableUpdatesCounts classifies the computers of the current organization and site
// by the number of software updates they need, for the dashboard's pie
func (h *Handler) availableUpdatesCounts(commonInfo *partials.CommonInfo) models.AvailableUpdatesCounts {
	counts := models.AvailableUpdatesCounts{}
	tenantID, _ := strconv.Atoi(commonInfo.TenantID)
	siteID, _ := strconv.Atoi(commonInfo.SiteID)

	h.availableUpdatesMu.RLock()
	defer h.availableUpdatesMu.RUnlock()
	for _, info := range h.AvailableUpdates {
		if tenantID != -1 && info.TenantID != tenantID {
			continue
		}
		if siteID != -1 && info.SiteID != siteID {
			continue
		}
		switch {
		case info.Count == 0:
			counts.Zero++
		case info.Count < 5:
			counts.UnderFive++
		case info.Count <= 20:
			counts.FiveToTwenty++
		default:
			counts.OverTwenty++
		}
	}
	return counts
}
