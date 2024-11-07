package filters

type AgentFilter struct {
	Hostname              string
	EnabledAgents         bool
	DisabledAgents        bool
	WindowsAgents         bool
	LinuxAgents           bool
	MacAgents             bool
	Tags                  []int
	OSVersions            []string
	ComputerManufacturers []string
	ComputerModels        []string
	Username              string
	ContactFrom           string
	ContactTo             string
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
