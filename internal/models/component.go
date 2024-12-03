package models

import (
	"context"
	"os"
	"runtime"

	"github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/component"
)

func (m *Model) SetComponent(c component.Component, version string, channel component.Channel) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	exists := true
	s, err := m.Client.Component.Query().Where(component.Hostname(hostname), component.ComponentEQ(c), component.Arch(runtime.GOARCH), component.Os(runtime.GOOS), component.Version(version), component.ChannelEQ(channel)).Only(context.Background())
	if err != nil {
		if !openuem_ent.IsNotFound(err) {
			return err
		}
		exists = false
	}

	if !exists {
		return m.Client.Component.Create().SetHostname(hostname).SetComponent(c).SetArch(runtime.GOARCH).SetOs(runtime.GOOS).SetVersion(version).SetChannel(channel).Exec(context.Background())
	}
	return m.Client.Component.Update().SetHostname(hostname).SetComponent(c).SetArch(runtime.GOARCH).SetOs(runtime.GOOS).SetVersion(version).SetChannel(channel).Where(component.ID(s.ID)).Exec(context.Background())
}
