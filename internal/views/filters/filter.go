package filters

type AgentFilter struct {
	Hostname              string
	AgentEnabledOptions   []string
	AgentOSVersions       []string
	Tags                  []int
	OSVersions            []string
	ComputerManufacturers []string
	ComputerModels        []string
	Username              string
	ContactFrom           string
	ContactTo             string
	WithApplication       string
	SelectedStatus        string
	SelectedItems         int
	SelectedAllAgents     string
}

type ApplicationsFilter struct {
	AppName string
	Vendor  string
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
