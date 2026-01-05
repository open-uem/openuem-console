package models

import (
	"context"
	"sort"

	openuem_ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/release"
	openuem_nats "github.com/open-uem/nats"
	"golang.org/x/mod/semver"
)

func (m *Model) GetLatestServerRelease(channel string) (*openuem_ent.Release, error) {
	data, err := m.Client.Release.Query().Where(release.Channel(channel), release.ReleaseTypeEQ(release.ReleaseTypeServer)).All(context.Background())

	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	sort.Slice(data, func(i, j int) bool {
		return semver.Compare("v"+data[i].Version, "v"+data[j].Version) > 0
	})

	return data[0], nil
}

func (m *Model) GetServerReleases() ([]string, error) {
	data, err := m.Client.Release.Query().Unique(true).Order(openuem_ent.Desc(release.FieldVersion)).Where(release.ReleaseTypeEQ(release.ReleaseTypeServer)).Select(release.FieldVersion).Strings(context.Background())
	if err != nil {
		return []string{}, err
	}

	sort.Slice(data, func(i, j int) bool {
		return semver.Compare("v"+data[i], "v"+data[j]) > 0
	})

	return data, nil
}

func (m *Model) GetLatestAgentRelease(channel string) (*openuem_ent.Release, error) {
	data, err := m.Client.Release.Query().Where(release.Channel(channel), release.ReleaseTypeEQ(release.ReleaseTypeAgent)).All(context.Background())
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	sort.Slice(data, func(i, j int) bool {
		return semver.Compare("v"+data[i].Version, "v"+data[j].Version) > 0
	})

	return data[0], nil
}

func (m *Model) GetAgentsReleases() ([]string, error) {
	data, err := m.Client.Release.Query().Unique(true).Where(release.ReleaseTypeEQ(release.ReleaseTypeAgent)).Select(release.FieldVersion).Strings(context.Background())
	if err != nil {
		return []string{}, err
	}

	sort.Slice(data, func(i, j int) bool {
		return semver.Compare("v"+data[i], "v"+data[j]) > 0
	})

	return data, nil
}

func (m *Model) GetAgentsReleaseByType(release_type release.ReleaseType, channel, os, arch, version string) (*openuem_ent.Release, error) {
	return m.Client.Release.Query().Where(release.ReleaseTypeEQ(release_type), release.Channel(channel), release.Os(os), release.Arch(arch), release.Version(version)).Only(context.Background())
}

func (m *Model) GetServersReleaseByType(release_type release.ReleaseType, channel, os, arch, version string) (*openuem_ent.Release, error) {
	return m.Client.Release.Query().Where(release.ReleaseTypeEQ(release_type), release.Channel(channel), release.Os(os), release.Arch(arch), release.Version(version)).Only(context.Background())
}

func (m *Model) GetHigherAgentReleaseInstalled() (*openuem_ent.Release, error) {
	data, err := m.Client.Release.Query().Where(release.ReleaseTypeEQ(release.ReleaseTypeAgent), release.HasAgentsWith(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission))).All(context.Background())
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	sort.Slice(data, func(i, j int) bool {
		return semver.Compare("v"+data[i].Version, "v"+data[j].Version) > 0
	})

	return data[0], nil
}

func (m *Model) CountOutdatedAgents() (int, error) {
	release, err := m.GetHigherAgentReleaseInstalled()
	if err != nil || release == nil {
		return 0, err
	}

	return m.CountUpgradableAgents(release.Version)
}

func (m *Model) CountUpgradableAgents(version string) (int, error) {
	count := 0
	data, err := m.Client.Agent.Query().WithRelease().Where(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).All(context.Background())
	if err != nil {
		return count, err
	}

	for _, item := range data {
		if item.Edges.Release != nil && semver.Compare("v"+item.Edges.Release.Version, "v"+version) < 0 {
			count += 1
		}
	}

	return count, nil
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
