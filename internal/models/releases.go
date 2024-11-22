package models

import (
	"context"

	"github.com/doncicuto/openuem-console/internal/views/admin_views"
	"github.com/doncicuto/openuem_ent"
	ent "github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/release"
)

func (m *Model) GetAgentsReleases() ([]*openuem_ent.Release, error) {
	return m.Client.Release.Query().Order(ent.Desc(release.FieldVersion)).All(context.Background())
}

func (m *Model) GetAgentsReleaseByType(channel, os, arch, version string) (*openuem_ent.Release, error) {
	return m.Client.Release.Query().Where(release.Channel(channel), release.Os(os), release.Arch(arch), release.Version(version)).Only(context.Background())
}

func (m *Model) GetHigherAgentReleaseInstalled() (*ent.Release, error) {
	data, err := m.Client.Release.Query().Where(release.HasAgents()).Order(ent.Desc(release.FieldVersion)).First(context.Background())
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
	return m.Client.Release.Query().Where(release.VersionLT(version)).Count(context.Background())
}

func (m *Model) SaveNewReleaseAvailable(newRelease admin_views.LatestRelease) error {
	exists, err := m.Client.Release.Query().Where(release.Version(newRelease.Version)).Exist(context.Background())
	if err != nil {
		return err
	}

	if !exists {
		for _, file := range newRelease.Files {
			err := m.Client.Release.Create().
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
