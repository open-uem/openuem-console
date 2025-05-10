package filters

import (
	"net/url"

	"github.com/labstack/echo/v4"
)

type AgentFilter struct {
	Hostname              string
	Versions              []string
	AgentStatusOptions    []string
	AgentOSVersions       []string
	Tags                  []int
	OSVersions            []string
	ComputerManufacturers []string
	ComputerModels        []string
	Username              string
	ContactFrom           string
	ContactTo             string
	WithApplication       string
	SelectedItems         int
	SelectedAllAgents     string
	SelectedRelease       string
	IsRemote              []string
	NoContact             bool
}

type ApplicationsFilter struct {
	AppName string
	Vendor  string
	Version string
}

type UserFilter struct {
	Username        string
	Name            string
	Email           string
	Country         string
	Phone           string
	Status          string
	CreatedFrom     string
	CreatedTo       string
	ModifiedFrom    string
	ModifiedTo      string
	RegisterOptions []string
}

type TenantFilter struct {
	Name           string
	DefaultOptions []string
	CreatedFrom    string
	CreatedTo      string
	ModifiedFrom   string
	ModifiedTo     string
}

type SiteFilter struct {
	Name           string
	DefaultOptions []string
	CreatedFrom    string
	CreatedTo      string
	ModifiedFrom   string
	ModifiedTo     string
}

type AntivirusFilter struct {
	Hostname                string
	AntivirusNameOptions    []string
	AntivirusUpdatedOptions []string
	AntivirusEnabledOptions []string
	AgentOSVersions         []string
}

type SystemUpdatesFilter struct {
	Hostname             string
	AgentOSVersions      []string
	UpdateStatus         []string
	LastSearchFrom       string
	LastSearchTo         string
	LastInstallFrom      string
	LastInstallTo        string
	PendingUpdateOptions []string
}

type CertificateFilter struct {
	Serial      string
	TypeOptions []string
	Description string
	ExpiryFrom  string
	ExpiryTo    string
	Username    string
}

type UpdateAgentsFilter struct {
	Hostname              string
	Releases              []string
	Tags                  []int
	TaskStatus            []string
	TaskResult            string
	TaskLastExecutionFrom string
	TaskLastExecutionTo   string
	SelectedItems         int
	SelectedAllAgents     string
	SelectedRelease       string
}

type UpdateServersFilter struct {
	Hostname           string
	Releases           []string
	UpdateStatus       []string
	UpdateMessage      string
	UpdateWhenFrom     string
	UpdateWhenTo       string
	SelectedItems      int
	SelectedAllServers string
	SelectedRelease    string
}

type DeployPackageFilter struct {
	Sources []string
}

func GetPaginationUrl(c echo.Context) string {
	// If Hx-Replace-Url is set in the header that means that we come from a dialog
	// and that we force to go to page 1, to avoid going to a non-existent page
	// due to a filter and we must keep the filters, also we add a check
	// to be sure that we're in the right url associated with the pagination

	replaceUrl := c.Response().Header().Get("Hx-Replace-Url")
	if replaceUrl != "" {
		if u, err := url.Parse(replaceUrl); err == nil {
			q := u.Query()
			q.Del("page")
			q.Del("pageSize")
			u.RawQuery = q.Encode()
			return u.String()
		}
	}

	return c.Request().URL.Path
}
