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
}

type ApplicationsFilter struct {
	AppName string
	Vendor  string
}
