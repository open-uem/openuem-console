package deploy_views

import (
	"context"
	"encoding/json"

	"github.com/invopop/ctxi18n/i18n"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/wingetrepo"
)

// Actions that can be requested from the Deploy section
const (
	ActionInstall   = "install"
	ActionUpdate    = "update"
	ActionUninstall = "uninstall"
	// ActionUpdateAll updates every package that has an available update
	ActionUpdateAll = "updateall"

	// TabUpdateStatus isn't an action: it shows the status of the agents' update runs
	TabUpdateStatus = "updatestatus"
	// TabLibrary isn't an action: it shows the packages of the private winget repository
	TabLibrary = "library"
)

// ParseAction validates the deploy action sent in a form. The "true" and "false" values
// of the former installation type boolean are still accepted, as they may be bookmarked.
func ParseAction(value string) (string, bool) {
	switch value {
	case "true":
		return ActionInstall, true
	case "false":
		return ActionUninstall, true
	case ActionInstall, ActionUpdate, ActionUninstall, ActionUpdateAll:
		return value, true
	}
	return "", false
}

// actionPath returns the path of the Deploy section page for an action
func actionPath(action string) string {
	return "/deploy/" + action
}

// searchPath returns the path used to search packages for an action
func searchPath(action string) string {
	return "/deploy/search" + action
}

// selectAgentsPath returns the path of the page where agents are selected for an action
func selectAgentsPath(action string) string {
	if action == ActionUpdateAll {
		return actionPath(ActionUpdateAll)
	}
	return "/deploy/selectpackagedeployment"
}

// actionTitle returns the translation key of the title of an action
func actionTitle(action string) string {
	switch action {
	case ActionUpdate:
		return "Update"
	case ActionUninstall:
		return "Uninstall"
	case ActionUpdateAll:
		return "updateall.tab"
	case TabUpdateStatus:
		return "updatestatus.tab"
	case TabLibrary:
		return "library.tab"
	default:
		return "Install"
	}
}

// actionIcon returns the icon used for an action
func actionIcon(action string) string {
	switch action {
	case ActionUpdate, ActionUpdateAll:
		return "package-check"
	case ActionUninstall:
		return "package-minus"
	default:
		return "package-plus"
	}
}

// UpdateRequestFailure describes a computer whose update request couldn't be sent
type UpdateRequestFailure struct {
	Computer string
	Reason   string
}

// UpdateAllResult summarizes a request to update all software on several computers
type UpdateAllResult struct {
	// Started is the number of computers that are online and start updating now
	Started int
	// Queued lists the computers that are offline and will update when they connect
	Queued   []string
	Failures []UpdateRequestFailure
}

// UpdateStatusRow is the status of a computer's software update runs
type UpdateStatusRow struct {
	AgentID  string
	Nickname string
	OS       string
	// Reachable is set if the agent answered with the status of its last run
	Reachable bool
	// Online is set if the agent didn't answer the status request but answers pings,
	// i.e. its version doesn't support software updates
	Online bool
	// Queued is set if an update request is waiting for the agent
	Queued bool
	Run    *openuem_nats.UpdateRun
}

// updateRunCounts describes how many packages of a run were updated, failed or were up to date
func updateRunCounts(ctx context.Context, run *openuem_nats.UpdateRun) string {
	updated, upToDate, failed := run.Counts()
	return i18n.T(ctx, "updatestatus.result_counts", updated, failed, upToDate)
}

// libraryPackageValues identifies a package version of the library in a request
func libraryPackageValues(p wingetrepo.Package) string {
	data, _ := json.Marshal(map[string]string{"library-id": p.PackageIdentifier, "library-version": p.PackageVersion})
	return string(data)
}
