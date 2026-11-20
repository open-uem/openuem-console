package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/linde12/gowol"
	ent "github.com/open-uem/ent"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/models"
	"github.com/open-uem/openuem-console/internal/views/deploy_views"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/robfig/cron/v3"
)

// How often the leader reconciles the jobs of the scheduled tasks with the
// database, so tasks created or edited while it was busy are picked up
const scheduledEventsReconcileInterval = 1 * time.Minute

// scheduledCronPatterns maps the cron expressions the form builds with the
// frequency they belong to, so a task can be edited with its frequency again
var scheduledCronPatterns = map[string]*regexp.Regexp{
	deploy_views.FrequencyDaily:   regexp.MustCompile(`^\d+ \d+ \* \* \*$`),
	deploy_views.FrequencyWeekly:  regexp.MustCompile(`^\d+ \d+ \* \* [0-7]$`),
	deploy_views.FrequencyMonthly: regexp.MustCompile(`^\d+ \d+ (?:[1-9]|[12]\d|3[01]) \* \*$`),
}

// ScheduledEventData groups the form values that define a scheduled event
type ScheduledEventData struct {
	Description string
	Type        string
	Cron        string
	When        *time.Time
}

// EnableScheduledEventsJobs starts running the scheduled tasks: the console
// instance that wins the NATS leader election registers the job of every
// enabled event, and a reconciliation job keeps the registrations in sync, so
// tasks created or edited while the leader was busy (or by a follower instance)
// are picked up within a minute. It runs in the election's OnPromote callback.
func (h *Handler) EnableScheduledEventsJobs() {
	h.scheduledEventMu.Lock()
	h.scheduledEventsLeader = true
	h.scheduledEventMu.Unlock()

	h.StartScheduledEventsJob()

	if job, err := h.TaskScheduler.NewJob(
		gocron.DurationJob(scheduledEventsReconcileInterval),
		gocron.NewTask(func() { h.StartScheduledEventsJob() }),
	); err != nil {
		log.Printf("[ERROR]: could not start the scheduled tasks reconciliation job, reason: %v", err)
	} else {
		h.scheduledEventMu.Lock()
		h.scheduledEventsReconcileJob = job
		h.scheduledEventMu.Unlock()
	}
}

// DisableScheduledEventsJobs stops running the scheduled tasks: callbacks and
// jobs are unregistered by the election's OnDemote (and jobs of events deleted
// meanwhile are gone together with them)
func (h *Handler) DisableScheduledEventsJobs() {
	h.scheduledEventMu.Lock()
	defer h.scheduledEventMu.Unlock()

	h.scheduledEventsLeader = false
	if h.scheduledEventsReconcileJob != nil {
		if err := h.TaskScheduler.RemoveJob(h.scheduledEventsReconcileJob.ID()); err != nil {
			log.Printf("[ERROR]: could not remove the scheduled tasks reconciliation job, reason: %v", err)
		}
		h.scheduledEventsReconcileJob = nil
	}
	for eventID := range h.scheduledEventJobs {
		h.UnregisterScheduledEvent(eventID)
	}
}

// StartScheduledEventsJob registers the enabled events found in the database,
// unregistering stale jobs first. The leader runs it at startup, after it is
// promoted, and once a minute afterwards.
func (h *Handler) StartScheduledEventsJob() {
	events, err := h.Model.GetEnabledScheduledEvents()
	if err != nil {
		log.Printf("[ERROR]: could not get the enabled scheduled tasks, reason: %v", err)
		return
	}

	enabled := map[int]bool{}
	for _, event := range events {
		enabled[event.ID] = true
		h.RegisterScheduledEvent(event)
	}

	// Tasks disabled or deleted since the last reconciliation, on this or
	// another console instance, are unregistered: a follower's request can't
	// touch the leader's jobs
	h.scheduledEventMu.Lock()
	for eventID := range h.scheduledEventJobs {
		if !enabled[eventID] {
			h.UnregisterScheduledEvent(eventID)
		}
	}
	h.scheduledEventMu.Unlock()
}

// RegisterScheduledEvent adds the job of a scheduled event to the task scheduler.
// Recurring events define a cron expression, events that happen only once run
// at the date and time they were configured with. One-time events in the past
// are disabled instead: their chance to run is gone.
func (h *Handler) RegisterScheduledEvent(event *ent.ScheduledEvent) {
	h.scheduledEventMu.Lock()
	defer h.scheduledEventMu.Unlock()

	// Only the elected leader runs scheduled tasks: every replica would run
	// them once otherwise. Followers keep the task in the database only; the
	// leader reconciles from there
	if !h.scheduledEventsLeader {
		return
	}

	// A job may already exist for this event, e.g. after editing it. The
	// running guard of a run in flight is left alone: it is cleaned up by the
	// run itself, and the reconciliation could delete it any moment
	h.UnregisterScheduledEvent(event.ID)

	var definition gocron.JobDefinition
	var options []gocron.JobOption

	switch {
	case event.Cron != "":
		definition = gocron.CronJob(event.Cron, false)
		options = append(options, gocron.WithSingletonMode(gocron.LimitModeReschedule))
	case event.When != nil:
		if event.When.Before(time.Now()) {
			log.Printf("[INFO]: the scheduled task %q would run in the past, so it has been disabled", event.Description)
			if err := h.Model.SetScheduledEventEnabled(event.ID, false); err != nil {
				log.Printf("[ERROR]: could not disable the task scheduled in the past, reason: %v", err)
			}
			return
		}
		definition = gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(*event.When))
	default:
		log.Printf("[ERROR]: the scheduled task %q has no schedule, so it has been disabled", event.Description)
		if err := h.Model.SetScheduledEventEnabled(event.ID, false); err != nil {
			log.Printf("[ERROR]: could not disable the task without a schedule, reason: %v", err)
		}
		return
	}

	job, err := h.TaskScheduler.NewJob(
		definition,
		gocron.NewTask(func() { h.RunScheduledEvent(event.ID) }),
		options...,
	)
	if err != nil {
		log.Printf("[ERROR]: could not schedule the task %q, reason: %v", event.Description, err)
		return
	}

	h.scheduledEventJobs[event.ID] = job
}

// UnregisterScheduledEvent removes the job of a scheduled event so it stops
// running. Callers must hold the scheduledEventMu mutex
func (h *Handler) UnregisterScheduledEvent(eventID int) {
	if job, ok := h.scheduledEventJobs[eventID]; ok {
		if err := h.TaskScheduler.RemoveJob(job.ID()); err != nil {
			log.Printf("[ERROR]: could not unschedule the task %d, reason: %v", eventID, err)
		}
		delete(h.scheduledEventJobs, eventID)
	}
}

// ScheduledEventNextRun returns when the job of a scheduled event will run next
func (h *Handler) ScheduledEventNextRun(eventID int) *time.Time {
	h.scheduledEventMu.Lock()
	job, ok := h.scheduledEventJobs[eventID]
	h.scheduledEventMu.Unlock()

	if !ok {
		return nil
	}
	nextRun, err := job.NextRun()
	if err != nil || nextRun.IsZero() {
		return nil
	}
	return &nextRun
}

// RunScheduledEvent runs the task of a scheduled event against every enabled agent
// of its organization (or site) and records the outcome. It runs in the task
// scheduler, without a request context, so the summary stored in the database is
// always written in English.
func (h *Handler) RunScheduledEvent(eventID int) {
	// A run that is still going on must not start a second one, e.g. the run
	// now button while the scheduler's job of the same task is firing
	h.scheduledEventMu.Lock()
	if h.scheduledEventRunning[eventID] {
		h.scheduledEventMu.Unlock()
		log.Printf("[INFO]: the scheduled task %d runs already, so this request does nothing", eventID)
		return
	}
	h.scheduledEventRunning[eventID] = true
	h.scheduledEventMu.Unlock()
	defer func() {
		h.scheduledEventMu.Lock()
		delete(h.scheduledEventRunning, eventID)
		h.scheduledEventMu.Unlock()
	}()

	event, err := h.Model.GetScheduledEvent(eventID)
	if err != nil {
		log.Printf("[ERROR]: could not get the scheduled task %d, reason: %v", eventID, err)
		return
	}

	tenantID := -1
	if event.Edges.Tenant != nil {
		tenantID = event.Edges.Tenant.ID
	}
	siteID := -1
	if event.Edges.Site != nil {
		siteID = event.Edges.Site.ID
	}

	agents, err := h.Model.GetScheduledEventAgents(tenantID, siteID)
	if err != nil {
		if err := h.Model.SetScheduledEventResult(event.ID, true, fmt.Sprintf("Could not get the computers of the organization, reason: %v", err)); err != nil {
			log.Printf("[ERROR]: could not record the failed run of the scheduled task %d, reason: %v", event.ID, err)
		}
		h.finishScheduledEvent(event)
		return
	}

	failed := false
	var summary string
	switch event.Type {
	case models.ScheduledEventUpdateAll:
		failed, summary = h.runScheduledUpdateAll(agents)
	case models.ScheduledEventPowerOff:
		failed, summary = h.runScheduledPowerAction(agents, "poweroff")
	case models.ScheduledEventRestart:
		failed, summary = h.runScheduledPowerAction(agents, "restart")
	case models.ScheduledEventWakeOnLAN:
		failed, summary = h.runScheduledWakeOnLAN(agents)
	default:
		failed, summary = true, fmt.Sprintf("Unknown task type %q", event.Type)
	}

	if err := h.Model.SetScheduledEventResult(event.ID, failed, summary); err != nil {
		log.Printf("[ERROR]: could not record the run of the scheduled task %d, reason: %v", event.ID, err)
	}
	h.finishScheduledEvent(event)
}

// finishScheduledEvent disables a one-time event after its single run
func (h *Handler) finishScheduledEvent(event *ent.ScheduledEvent) {
	if event.Cron != "" {
		return
	}

	h.scheduledEventMu.Lock()
	h.UnregisterScheduledEvent(event.ID)
	h.scheduledEventMu.Unlock()

	if err := h.Model.SetScheduledEventEnabled(event.ID, false); err != nil {
		log.Printf("[ERROR]: could not disable the one-time task %q after running it, reason: %v", event.Description, err)
	}
}

// runScheduledUpdateAll queues an update run for every agent, so even computers
// that are offline update all their software when they connect again
func (h *Handler) runScheduledUpdateAll(agents []*ent.Agent) (bool, string) {
	if h.JetStream == nil || h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return true, "The console was not connected to the messaging server, so no computer started to update all its software"
	}

	agentIDs := []string{}
	for _, a := range agents {
		agentIDs = append(agentIDs, a.ID)
	}

	if len(agentIDs) == 0 {
		return false, "No enabled computer belongs to the organization of this task"
	}

	started, queued, failures := 0, 0, 0
	var failureReasons []string
	for agentID, outcome := range h.queuePackageUpdates(agentIDs) {
		if outcome.err != nil {
			failures++
			if len(failureReasons) < 5 {
				failureReasons = append(failureReasons, fmt.Sprintf("%s: %v", agentID, outcome.err))
			}
			continue
		}
		if outcome.online {
			started++
		} else {
			queued++
		}
	}

	summary := fmt.Sprintf("%d computer(s) started to update all their software, %d will update when they connect", started, queued)
	if failures > 0 {
		summary += fmt.Sprintf(", %d update request(s) failed", failures)
		for _, reason := range failureReasons {
			summary += " - " + reason
		}
		return failures == len(agentIDs), summary
	}
	return false, summary
}

// runScheduledPowerAction asks every agent to power off or restart now, as the
// schedule has just fired. Only agents that are online can be asked
func (h *Handler) runScheduledPowerAction(agents []*ent.Agent, action string) (bool, string) {
	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return true, "The console was not connected to the messaging server, so no computer could be asked to " + action
	}

	// Both actions expect a RebootOrRestart message and no date, because the
	// schedule has already fired: the agent must act right now. The subject is
	// the one the console uses for the per-computer actions, so an agent always
	// answers it: a restart reboots the machine (not the agent service)
	subject := "agent.poweroff."
	if action == "restart" {
		subject = "agent.reboot."
	}

	request, err := json.Marshal(openuem_nats.RebootOrRestart{})
	if err != nil {
		return true, fmt.Sprintf("Could not prepare the request, reason: %v", err)
	}

	asked, failures := 0, 0
	limit := make(chan struct{}, updatePackagesConcurrency)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, a := range agents {
		wg.Add(1)
		limit <- struct{}{}
		go func(agentID string) {
			defer wg.Done()
			defer func() { <-limit }()

			mu.Lock()
			asked++
			mu.Unlock()

			if _, err := h.NATSConnection.Request(subject+agentID, request, time.Duration(h.NATSTimeout)*time.Second); err != nil {
				mu.Lock()
				failures++
				mu.Unlock()
			}
		}(a.ID)
	}
	wg.Wait()

	if asked == 0 {
		return false, "No enabled computer belongs to the organization of this task"
	}
	if failures > 0 {
		return true, fmt.Sprintf("%d computer(s) were asked to %s, %d did not answer", asked-failures, action, failures)
	}
	return false, fmt.Sprintf("%d computer(s) were asked to %s", asked, action)
}

// runScheduledWakeOnLAN sends a Wake on LAN magic packet to every agent of the event
func (h *Handler) runScheduledWakeOnLAN(agents []*ent.Agent) (bool, string) {
	sent, invalid, failures := 0, 0, 0

	for _, a := range agents {
		if _, err := net.ParseMAC(a.MAC); err != nil {
			invalid++
			continue
		}
		packet, err := gowol.NewMagicPacket(a.MAC)
		if err != nil {
			failures++
			continue
		}
		if err := packet.Send("255.255.255.255"); err != nil {
			failures++
			continue
		}
		sent++
	}

	switch {
	case sent == 0 && invalid == 0 && failures == 0:
		return false, "No enabled computer belongs to the organization of this task"
	case sent == 0:
		return true, fmt.Sprintf("No Wake on LAN packet could be sent, %d computer(s) have no valid MAC address, %d failed", invalid, failures)
	case failures > 0 || invalid > 0:
		return false, fmt.Sprintf("%d Wake on LAN packet(s) were sent, %d computer(s) have no valid MAC address, %d failed", sent, invalid, failures)
	default:
		return false, fmt.Sprintf("%d Wake on LAN packet(s) were sent successfully", sent)
	}
}

// ScheduledEvents renders the list of scheduled tasks of the organization and site
func (h *Handler) ScheduledEvents(c echo.Context) error {
	return h.renderScheduledEvents(c, "")
}

// renderScheduledEvents builds the list of scheduled tasks with the next run of
// every registered job
func (h *Handler) renderScheduledEvents(c echo.Context, successMessage string) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	siteID, err := strconv.Atoi(commonInfo.SiteID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	events, err := h.Model.GetScheduledEvents(tenantID, siteID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	nextRuns := []deploy_views.ScheduledEventNextRun{}
	for _, event := range events {
		if nextRun := h.ScheduledEventNextRun(event.ID); nextRun != nil {
			nextRuns = append(nextRuns, deploy_views.ScheduledEventNextRun{EventID: event.ID, NextRun: *nextRun})
		}
	}

	return RenderView(c, deploy_views.DeployIndex("| Deploy", deploy_views.ScheduledEvents(c, events, nextRuns, successMessage, commonInfo), commonInfo))
}

// NewScheduledEvent renders the form to create a new scheduled task
func (h *Handler) NewScheduledEvent(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	return RenderView(c, deploy_views.DeployIndex("| Deploy", deploy_views.ScheduledEventForm(c, nil, deploy_views.FrequencyDaily, commonInfo), commonInfo))
}

// CreateScheduledEvent validates the submitted form, stores the new task in the
// database and registers its job
func (h *Handler) CreateScheduledEvent(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	data, err := h.parseScheduledEventForm(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "scheduled.wrong_schedule", err.Error()), false))
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "scheduled.wrong_schedule", err.Error()), false))
	}
	siteID, err := strconv.Atoi(commonInfo.SiteID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "scheduled.wrong_schedule", err.Error()), false))
	}

	event, err := h.Model.CreateScheduledEvent(tenantID, siteID, data.Description, data.Type, data.Cron, data.When)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	h.RegisterScheduledEvent(event)

	return h.renderScheduledEvents(c, i18n.T(c.Request().Context(), "scheduled.created"))
}

// EditScheduledEvent renders the form to edit a scheduled task
func (h *Handler) EditScheduledEvent(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "scheduled.wrong_id"), false))
	}

	event, err := h.Model.GetScheduledEvent(id)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if !h.IsScheduledEventInScope(event, commonInfo) {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "scheduled.wrong_id"), false))
	}

	frequency := deploy_views.FrequencyCustom
	switch {
	case event.Cron != "":
		if match, ok := matchScheduledCron(event.Cron); ok {
			frequency = match
		}
	case event.When != nil:
		frequency = deploy_views.FrequencyOnce
	}

	return RenderView(c, deploy_views.DeployIndex("| Deploy", deploy_views.ScheduledEventForm(c, event, frequency, commonInfo), commonInfo))
}

// SaveScheduledEvent validates the submitted form, updates the task in the
// database and registers its job again with the new schedule
func (h *Handler) SaveScheduledEvent(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "scheduled.wrong_id"), false))
	}

	event, err := h.Model.GetScheduledEvent(id)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if !h.IsScheduledEventInScope(event, commonInfo) {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "scheduled.wrong_id"), false))
	}

	data, err := h.parseScheduledEventForm(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "scheduled.wrong_schedule", err.Error()), false))
	}

	if err := h.Model.UpdateScheduledEvent(id, data.Description, data.Type, data.Cron, data.When); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// Register the job again, even if the task is disabled, so the edit is not
	// lost when it is toggled on later
	if event, err := h.Model.GetScheduledEvent(id); err == nil {
		if event.Enabled {
			h.RegisterScheduledEvent(event)
		}
	}

	return h.renderScheduledEvents(c, i18n.T(c.Request().Context(), "scheduled.updated"))
}

// ToggleScheduledEvent enables or disables a scheduled task. Enabling a one-time
// task scheduled in the past is not allowed on the form, but restarting the
// console can find one, so registering disables it instead of running it
func (h *Handler) ToggleScheduledEvent(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "scheduled.wrong_id"), false))
	}

	event, err := h.Model.GetScheduledEvent(id)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if !h.IsScheduledEventInScope(event, commonInfo) {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "scheduled.wrong_id"), false))
	}

	enable := !event.Enabled
	if err := h.Model.SetScheduledEventEnabled(id, enable); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if enable {
		h.RegisterScheduledEvent(event)

		// A one-time task scheduled in the past is disabled again by
		// registering it, so say what really happened
		if refreshed, err := h.Model.GetScheduledEvent(id); err == nil && refreshed.Enabled {
			return h.renderScheduledEvents(c, i18n.T(c.Request().Context(), "scheduled.enabled"))
		}
		return h.renderScheduledEvents(c, i18n.T(c.Request().Context(), "scheduled.disabled"))
	}

	h.scheduledEventMu.Lock()
	h.UnregisterScheduledEvent(id)
	h.scheduledEventMu.Unlock()
	return h.renderScheduledEvents(c, i18n.T(c.Request().Context(), "scheduled.disabled"))
}

// RunScheduledEventNow runs the task of a scheduled event right away instead of
// waiting for its schedule. It answers immediately, the run is recorded when it
// finishes
func (h *Handler) RunScheduledEventNow(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "scheduled.wrong_id"), false))
	}

	event, err := h.Model.GetScheduledEvent(id)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if !h.IsScheduledEventInScope(event, commonInfo) {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "scheduled.wrong_id"), false))
	}

	go h.RunScheduledEvent(event.ID)

	return h.renderScheduledEvents(c, i18n.T(c.Request().Context(), "scheduled.running_now"))
}

// DeleteScheduledEvent removes a scheduled task and unregisters its job
func (h *Handler) DeleteScheduledEvent(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "scheduled.wrong_id"), false))
	}

	event, err := h.Model.GetScheduledEvent(id)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if !h.IsScheduledEventInScope(event, commonInfo) {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "scheduled.wrong_id"), false))
	}

	if err := h.Model.DeleteScheduledEvent(id); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	h.scheduledEventMu.Lock()
	h.UnregisterScheduledEvent(id)
	h.scheduledEventMu.Unlock()

	return h.renderScheduledEvents(c, i18n.T(c.Request().Context(), "scheduled.deleted"))
}

// IsScheduledEventInScope checks that the logged in user is working with the
// organization and site a scheduled task belongs to
func (h *Handler) IsScheduledEventInScope(event *ent.ScheduledEvent, commonInfo *partials.CommonInfo) bool {
	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return false
	}

	siteID, err := strconv.Atoi(commonInfo.SiteID)
	if err != nil {
		return false
	}

	if event.Edges.Tenant == nil || event.Edges.Tenant.ID != tenantID {
		return false
	}

	// An organization-wide task is visible from every site of the organization
	if event.Edges.Site == nil {
		return true
	}

	// From the organization level every site task can be managed
	if siteID == -1 {
		return true
	}

	return event.Edges.Site.ID == siteID
}

// parseScheduledEventForm validates the form of a scheduled task. Simple
// frequencies (once, daily, weekly, monthly) build the cron expression of the
// task; the custom frequency expects a cron expression directly. The errors are
// already translated, with the reason included where it helps.
func (h *Handler) parseScheduledEventForm(c echo.Context) (ScheduledEventData, error) {
	data := ScheduledEventData{}
	ctx := c.Request().Context()

	data.Description = strings.TrimSpace(c.FormValue("description"))
	if data.Description == "" {
		return data, errors.New(i18n.T(ctx, "scheduled.no_description"))
	}
	if len(data.Description) > 250 {
		return data, errors.New(i18n.T(ctx, "scheduled.wrong_description"))
	}

	data.Type = c.FormValue("type")
	if !slices.Contains(models.ScheduledEventTypes(), data.Type) {
		return data, errors.New(i18n.T(ctx, "scheduled.wrong_type"))
	}

	frequency := c.FormValue("frequency")
	switch frequency {
	case deploy_views.FrequencyOnce:
		when, err := time.ParseInLocation("2006-01-02T15:04", c.FormValue("when"), time.Local)
		if err != nil {
			return data, errors.New(i18n.T(ctx, "scheduled.could_not_parse_time"))
		}
		if !when.After(time.Now()) {
			return data, errors.New(i18n.T(ctx, "scheduled.when_past"))
		}
		data.When = &when
	case deploy_views.FrequencyDaily:
		hour, minute, ok := parseScheduledTime(c.FormValue("time"))
		if !ok {
			return data, errors.New(i18n.T(ctx, "scheduled.could_not_parse_time"))
		}
		data.Cron = fmt.Sprintf("%d %d * * *", minute, hour)
	case deploy_views.FrequencyWeekly:
		hour, minute, ok := parseScheduledTime(c.FormValue("time"))
		if !ok {
			return data, errors.New(i18n.T(ctx, "scheduled.could_not_parse_time"))
		}
		weekday, err := strconv.Atoi(c.FormValue("weekday"))
		if err != nil || weekday < 0 || weekday > 6 {
			return data, errors.New(i18n.T(ctx, "scheduled.wrong_weekday"))
		}
		data.Cron = fmt.Sprintf("%d %d * * %d", minute, hour, weekday)
	case deploy_views.FrequencyMonthly:
		hour, minute, ok := parseScheduledTime(c.FormValue("time"))
		if !ok {
			return data, errors.New(i18n.T(ctx, "scheduled.could_not_parse_time"))
		}
		monthday, err := strconv.Atoi(c.FormValue("monthday"))
		if err != nil || monthday < 1 || monthday > 31 {
			return data, errors.New(i18n.T(ctx, "scheduled.wrong_monthday"))
		}
		data.Cron = fmt.Sprintf("%d %d %d * *", minute, hour, monthday)
	case deploy_views.FrequencyCustom:
		data.Cron = c.FormValue("cron")
		schedule, err := cron.ParseStandard(data.Cron)
		if err != nil {
			return data, errors.New(i18n.T(ctx, "scheduled.wrong_cron", err.Error()))
		}
		// A valid expression can still never fire, e.g. "0 0 31 2 *": there is
		// no February 31st. The task would be saved and quietly do nothing
		if nextFire := schedule.Next(time.Now()); nextFire.IsZero() {
			return data, errors.New(i18n.T(ctx, "scheduled.cron_never_fires"))
		}
	default:
		return data, errors.New(i18n.T(ctx, "scheduled.wrong_frequency"))
	}

	return data, nil
}

// parseScheduledTime parses the HH:MM time the form submits
func parseScheduledTime(value string) (int, int, bool) {
	hourMinute, err := time.Parse("15:04", value)
	if err != nil {
		return 0, 0, false
	}
	return hourMinute.Hour(), hourMinute.Minute(), true
}

// matchScheduledCron returns the simple frequency a cron expression was built
// with by the form, if any
func matchScheduledCron(cronExpression string) (string, bool) {
	for frequency, pattern := range scheduledCronPatterns {
		if pattern.MatchString(cronExpression) {
			return frequency, true
		}
	}
	return "", false
}
