package models

import (
	"context"

	openuem_ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/release"
	openuem_nats "github.com/open-uem/nats"
)

func (m *Model) GetLatestServerRelease(channel string) (*openuem_ent.Release, error) {
	return m.Client.Release.Query().Where(release.Channel(channel), release.ReleaseTypeEQ(release.ReleaseTypeServer)).Order(openuem_ent.Desc(release.FieldVersion)).First(context.Background())
}

func (m *Model) GetServerReleases() ([]string, error) {
	return m.Client.Release.Query().Unique(true).Order(openuem_ent.Desc(release.FieldVersion)).Where(release.ReleaseTypeEQ(release.ReleaseTypeServer)).Select(release.FieldVersion).Strings(context.Background())
}

func (m *Model) GetLatestAgentRelease(channel string) (*openuem_ent.Release, error) {
	return m.Client.Release.Query().Where(release.Channel(channel), release.ReleaseTypeEQ(release.ReleaseTypeAgent)).Order(openuem_ent.Desc(release.FieldVersion)).First(context.Background())
}

func (m *Model) GetAgentsReleases() ([]string, error) {
	return m.Client.Release.Query().Unique(true).Where(release.ReleaseTypeEQ(release.ReleaseTypeAgent)).Order(openuem_ent.Desc(release.FieldVersion)).Select(release.FieldVersion).Strings(context.Background())
}

func (m *Model) GetAgentsReleaseByType(release_type release.ReleaseType, channel, os, arch, version string) (*openuem_ent.Release, error) {
	return m.Client.Release.Query().Where(release.ReleaseTypeEQ(release_type), release.Channel(channel), release.Os(os), release.Arch(arch), release.Version(version)).Only(context.Background())
}

func (m *Model) GetServersReleaseByType(release_type release.ReleaseType, channel, os, arch, version string) (*openuem_ent.Release, error) {
	return m.Client.Release.Query().Where(release.ReleaseTypeEQ(release_type), release.Channel(channel), release.Os(os), release.Arch(arch), release.Version(version)).Only(context.Background())
}

func (m *Model) GetHigherAgentReleaseInstalled() (*openuem_ent.Release, error) {
	data, err := m.Client.Release.Query().Where(release.ReleaseTypeEQ(release.ReleaseTypeAgent), release.HasAgentsWith(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission))).Order(openuem_ent.Desc(release.FieldVersion)).First(context.Background())
	if err != nil {
		if openuem_ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return data, nil
}

func (m *Model) CountOutdatedAgents() (int, error) {
	release, err := m.GetHigherAgentReleaseInstalled()
	if err != nil || release == nil {
		return 0, err
	}

	return m.CountUpgradableAgents(release.Version)
}

func (m *Model) CountUpgradableAgents(version string) (int, error) {
	return m.Client.Agent.Query().
		Where(
			agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission),
			agent.HasReleaseWith(release.VersionLT(version)),
		).Count(context.Background())
}

func (m *Model) SaveNewReleaseAvailable(releaseType release.ReleaseType, newRelease openuem_nats.OpenUEMRelease) error {
	for _, file := range newRelease.Files {
		exists, err := m.Client.Release.Query().Where(release.ReleaseTypeEQ(releaseType), release.Os(file.Os), release.Arch(file.Arch), release.Version(newRelease.Version)).Exist(context.Background())
		if err != nil {
			return err
		}

		if !exists {
			err := m.Client.Release.Create().
				SetReleaseType(releaseType).
				SetVersion(newRelease.Version).
				SetSummary(newRelease.Summary).
				SetChannel(newRelease.Channel).
				SetReleaseNotes(newRelease.ReleaseNotesURL).
				SetReleaseDate(newRelease.ReleaseDate).
				SetIsCritical(newRelease.IsCritical).
				SetArch(file.Arch).
				SetOs(file.Os).
				SetFileURL(file.FileURL).
				SetChecksum(file.Checksum).
				Exec(context.Background())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
