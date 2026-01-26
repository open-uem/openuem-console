package models

import (
	"context"
	"strconv"

	"github.com/open-uem/ent"
	"github.com/open-uem/ent/profile"
	"github.com/open-uem/ent/profileissue"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/task"
	"github.com/open-uem/ent/tenant"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (m *Model) CountAllProfiles(c *partials.CommonInfo) (int, error) {
	query := m.Client.Profile.Query()

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return -1, err
	}

	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return -1, err
	}

	if siteID == -1 {
		return -1, err
	}

	query = query.Where(profile.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))

	return query.Count(context.Background())
}

func (m *Model) GetProfilesByPage(p partials.PaginationAndSort, c *partials.CommonInfo) ([]*ent.Profile, error) {
	var err error
	var profiles []*ent.Profile

	query := m.Client.Profile.Query().WithTasks().WithTags().WithIssues(func(q *ent.ProfileIssueQuery) { q.WithTasksreports().All(context.Background()) }).Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}

	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		return nil, err
	}

	query = query.Where(profile.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))

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

func (m *Model) AddProfile(siteID int, description string) (*ent.Profile, error) {
	profile, err := m.Client.Profile.Create().SetName(description).SetSiteID(siteID).Save(context.Background())
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func (m *Model) UpdateProfile(profileId int, description string, apply string, c *partials.CommonInfo) error {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return err
	}

	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	if siteID == -1 {
		return err
	}

	switch apply {
	case "applyToAll":
		return m.Client.Profile.Update().Where(profile.ID(profileId), profile.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).SetName(description).ClearTags().SetApplyToAll(true).Exec(context.Background())
	case "useTags":
		return m.Client.Profile.Update().Where(profile.ID(profileId), profile.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).SetName(description).SetApplyToAll(false).Exec(context.Background())
	}
	return m.Client.Profile.Update().Where(profile.ID(profileId), profile.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).SetName(description).ClearTags().SetApplyToAll(false).Exec(context.Background())
}

func (m *Model) GetProfileById(profileId int, c *partials.CommonInfo) (*ent.Profile, error) {

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}

	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		return nil, err
	}

	return m.Client.Profile.Query().WithTags().WithTasks().WithIssues().Where(profile.ID(profileId), profile.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).First(context.Background())
}

func (m *Model) DeleteProfile(profileID int, c *partials.CommonInfo) error {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return err
	}

	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	if siteID == -1 {
		return err
	}

	_, err = m.Client.Task.Delete().Where(task.HasProfileWith(profile.ID(profileID))).Exec(context.Background())
	if err != nil {
		return err
	}

	_, err = m.Client.Profile.Delete().Where(profile.ID(profileID), profile.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).Exec(context.Background())
	if err != nil {
		return err
	}

	return nil
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
	// Remove issues that has no agents associated
	nDeleted, err := m.Client.ProfileIssue.Delete().Where(profileissue.Not(profileissue.HasAgents())).Exec(context.Background())
	if err != nil {
		return nDeleted, err
	}

	return m.Client.ProfileIssue.Query().Where(profileissue.HasProfileWith(profile.ID(profileID))).Count(context.Background())
}

func (m *Model) GetProfileIssuesByPage(p partials.PaginationAndSort, profileID int) ([]*ent.ProfileIssue, error) {
	// Remove issues that has no agents associated
	_, err := m.Client.ProfileIssue.Delete().Where(profileissue.Not(profileissue.HasAgents())).Exec(context.Background())
	if err != nil {
		return nil, err
	}

	return m.Client.ProfileIssue.Query().WithAgents().WithTasksreports(func(q *ent.TaskReportQuery) { q.WithTask().All(context.Background()) }).Where(profileissue.HasProfileWith(profile.ID(profileID))).Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize).All(context.Background())
}

func (m *Model) EnableProfile(profiledID int, enabled bool) error {
	return m.Client.Profile.Update().SetDisabled(!enabled).Where(profile.ID(profiledID)).Exec(context.Background())
}
