package models

import (
	"context"
	"time"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/scheduledevent"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/tenant"
)

// Actions that a scheduled event can run on the agents of an organization (or site)
const (
	ScheduledEventUpdateAll = "updateall"
	ScheduledEventPowerOff  = "poweroff"
	ScheduledEventRestart   = "reboot"
	ScheduledEventWakeOnLAN = "wakeonlan"
)

// ScheduledEventTypes returns the list of tasks that can be scheduled
func ScheduledEventTypes() []string {
	return []string{ScheduledEventUpdateAll, ScheduledEventPowerOff, ScheduledEventRestart, ScheduledEventWakeOnLAN}
}

// GetScheduledEvents returns the scheduled events of a tenant, optionally limited to one site
func (m *Model) GetScheduledEvents(tenantID, siteID int) ([]*ent.ScheduledEvent, error) {
	query := m.Client.ScheduledEvent.Query().Where(scheduledevent.HasTenantWith(tenant.ID(tenantID)))

	if siteID != -1 {
		query = query.Where(scheduledevent.HasSiteWith(site.ID(siteID)))
	}

	return query.WithTenant().WithSite().Order(ent.Asc(scheduledevent.FieldCreatedAt)).All(context.Background())
}

// GetScheduledEvent returns a scheduled event by its id
func (m *Model) GetScheduledEvent(id int) (*ent.ScheduledEvent, error) {
	return m.Client.ScheduledEvent.Query().Where(scheduledevent.ID(id)).WithTenant().WithSite().Only(context.Background())
}

// GetEnabledScheduledEvents returns every enabled scheduled event, regardless of tenant and site
func (m *Model) GetEnabledScheduledEvents() ([]*ent.ScheduledEvent, error) {
	return m.Client.ScheduledEvent.Query().Where(scheduledevent.EnabledEQ(true)).WithTenant().WithSite().All(context.Background())
}

// CreateScheduledEvent stores a new scheduled event. A recurring event defines
// a cron expression, an event that happens only once defines the date and time
// when it must run.
func (m *Model) CreateScheduledEvent(tenantID, siteID int, description, eventType, cron string, when *time.Time) (*ent.ScheduledEvent, error) {
	query := m.Client.ScheduledEvent.Create().
		SetDescription(description).
		SetType(eventType).
		SetTenantID(tenantID)

	if cron != "" {
		query.SetCron(cron)
	}
	if when != nil {
		query.SetWhen(*when)
	}
	if siteID != -1 {
		query.SetSiteID(siteID)
	}

	return query.Save(context.Background())
}

// UpdateScheduledEvent updates the definition of a scheduled event. A previously
// defined recurrence or single date is replaced by the new values.
func (m *Model) UpdateScheduledEvent(id int, description, eventType, cron string, when *time.Time) error {
	query := m.Client.ScheduledEvent.UpdateOneID(id).
		SetDescription(description).
		SetType(eventType)

	if cron != "" {
		query = query.SetCron(cron).ClearWhen()
	} else {
		query = query.ClearCron()
	}

	if when != nil {
		query = query.SetWhen(*when)
	} else {
		query = query.ClearWhen()
	}

	return query.Exec(context.Background())
}

// SetScheduledEventEnabled enables or disables a scheduled event and resets
// its last run info when it is enabled again
func (m *Model) SetScheduledEventEnabled(id int, enabled bool) error {
	query := m.Client.ScheduledEvent.UpdateOneID(id).SetEnabled(enabled)
	if enabled {
		query = query.ClearLastRun().SetLastRunFailed(false).ClearLastResult()
	}
	return query.Exec(context.Background())
}

// SetScheduledEventResult records when a scheduled event last ran and what the
// outcome was
func (m *Model) SetScheduledEventResult(id int, failed bool, result string) error {
	return m.Client.ScheduledEvent.UpdateOneID(id).
		SetLastRun(time.Now()).
		SetLastRunFailed(failed).
		SetLastResult(result).
		Exec(context.Background())
}

// DeleteScheduledEvent removes a scheduled event
func (m *Model) DeleteScheduledEvent(id int) error {
	return m.Client.ScheduledEvent.DeleteOneID(id).Exec(context.Background())
}

// GetScheduledEventAgents returns the enabled agents that a scheduled event
// of a tenant (optionally of one site) applies to. Agents waiting for approval
// are not included.
func (m *Model) GetScheduledEventAgents(tenantID, siteID int) ([]*ent.Agent, error) {
	query := m.Client.Agent.Query().Where(
		agent.AgentStatusEQ(agent.AgentStatusEnabled),
		agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))),
	)

	if siteID != -1 {
		query = query.Where(agent.HasSiteWith(site.ID(siteID)))
	}

	return query.All(context.Background())
}
