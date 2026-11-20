package deploy_views

import (
	"fmt"
	"strconv"
	"strings"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

// scheduledEventHourMinute returns the HH:MM time the form built the cron
// expression of a scheduled event with, so it can be edited again
func scheduledEventHourMinute(event *ent.ScheduledEvent) string {
	if event == nil {
		return ""
	}
	fields := strings.Fields(event.Cron)
	if len(fields) < 2 {
		return ""
	}
	minute, errMinute := strconv.Atoi(fields[0])
	hour, errHour := strconv.Atoi(fields[1])
	if errMinute != nil || errHour != nil {
		return ""
	}
	return fmt.Sprintf("%02d:%02d", hour, minute)
}

// scheduledEventWeekday returns the weekday of a weekly cron expression, where
// 0 is Sunday like in the cron expression itself
func scheduledEventWeekday(event *ent.ScheduledEvent) int {
	if event == nil {
		return 0
	}
	fields := strings.Fields(event.Cron)
	if len(fields) < 5 {
		return 0
	}
	weekday, err := strconv.Atoi(fields[4])
	if err != nil || weekday < 0 || weekday > 7 {
		return 0
	}
	if weekday == 7 {
		return 0
	}
	return weekday
}

// scheduledEventMonthday returns the day of the month of a monthly cron expression
func scheduledEventMonthday(event *ent.ScheduledEvent) string {
	if event == nil {
		return ""
	}
	fields := strings.Fields(event.Cron)
	if len(fields) < 3 {
		return ""
	}
	if _, err := strconv.Atoi(fields[2]); err != nil {
		return ""
	}
	return fields[2]
}

// scheduledEventDescription returns the description of the task or an empty string
func scheduledEventDescription(event *ent.ScheduledEvent) string {
	if event == nil {
		return ""
	}
	return event.Description
}

// scheduledEventSavePath returns the path the form of a scheduled task posts to
func scheduledEventSavePath(event *ent.ScheduledEvent, commonInfo *partials.CommonInfo) string {
	if event == nil {
		return scheduledEventsPath("/new", commonInfo)
	}
	return scheduledEventsPath("/edit/"+fmt.Sprintf("%d", event.ID), commonInfo)
}

// scheduledEventWhenValue returns the date and time of a one-time task formatted
// for a datetime-local input
func scheduledEventWhenValue(event *ent.ScheduledEvent) string {
	if event == nil || event.When == nil {
		return ""
	}
	return event.When.Local().Format("2006-01-02T15:04")
}

// scheduledEventCronValue returns the cron expression of a task defined with a
// custom cron schedule
func scheduledEventCronValue(event *ent.ScheduledEvent, frequency string) string {
	if event == nil || event.Cron == "" || frequency != FrequencyCustom {
		return ""
	}
	return event.Cron
}
