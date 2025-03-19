package models

import (
	"context"

	"github.com/open-uem/ent"
	"github.com/open-uem/ent/profile"
	"github.com/open-uem/ent/profileissue"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (m *Model) CountAllProfiles() (int, error) {
	query := m.Client.Profile.Query()

	return query.Count(context.Background())
}

func (m *Model) GetProfilesByPage(p partials.PaginationAndSort) ([]*ent.Profile, error) {
	var err error
	var profiles []*ent.Profile

	query := m.Client.Profile.Query().WithTasks().WithTags().WithIssues().Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)

	switch p.SortBy {
	case "name":
		if p.SortOrder == "asc" {
			profiles, err = query.Order(ent.Asc(profile.FieldName)).All(context.Background())
		} else {
			profiles, err = query.Order(ent.Desc(profile.FieldName)).All(context.Background())
		}
	default:
		profiles, err = query.Order(ent.Desc(profile.FieldName)).All(context.Background())
	}

	if err != nil {
		return nil, err
	}
	return profiles, nil
}

func (m *Model) AddProfile(description string) (*ent.Profile, error) {
	profile, err := m.Client.Profile.Create().SetName(description).Save(context.Background())
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func (m *Model) UpdateProfile(profileId int, description string, apply string) error {
	switch apply {
	case "applyToAll":
		return m.Client.Profile.UpdateOneID(profileId).SetName(description).ClearTags().SetApplyToAll(true).Exec(context.Background())
	case "useTags":
		return m.Client.Profile.UpdateOneID(profileId).SetName(description).SetApplyToAll(false).Exec(context.Background())
	}
	return m.Client.Profile.UpdateOneID(profileId).SetName(description).ClearTags().SetApplyToAll(false).Exec(context.Background())
}

func (m *Model) GetProfileById(profileId int) (*ent.Profile, error) {
	return m.Client.Profile.Query().WithTags().WithTasks().WithIssues().Where(profile.ID(profileId)).First(context.Background())
}

func (m *Model) DeleteProfile(profileId int) error {
	return m.Client.Profile.DeleteOneID(profileId).Exec(context.Background())
}

func (m *Model) AddTagToProfile(profileId int, tagId int) error {
	_, err := m.Client.Profile.UpdateOneID(profileId).SetApplyToAll(false).AddTagIDs(tagId).Save(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (m *Model) RemoveTagFromProfile(profileId int, tagId int) error {
	_, err := m.Client.Profile.UpdateOneID(profileId).RemoveTagIDs(tagId).Save(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (m *Model) CountAllProfileIssues(profileID int) (int, error) {
	return m.Client.ProfileIssue.Query().Where(profileissue.HasProfileWith(profile.ID(profileID))).Count(context.Background())
}

func (m *Model) GetProfileIssuesByPage(p partials.PaginationAndSort, profileID int) ([]*ent.ProfileIssue, error) {
	return m.Client.ProfileIssue.Query().WithAgents().Where(profileissue.HasProfileWith(profile.ID(profileID))).Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize).All(context.Background())
}
