package models

import (
	"context"
	"fmt"
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

	if tenantID == -1 {
		query = query.Where(profile.And(profile.Not(profile.HasSite()), profile.Not(profile.HasTenant())))
	} else {
		if siteID == -1 {
			query = query.Where(profile.HasTenantWith(tenant.ID(tenantID)), profile.Not(profile.HasSite()))
		} else {
			query = query.Where(profile.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))
		}
	}

	return query.Count(context.Background())
}

func (m *Model) GetProfilesByPage(p partials.PaginationAndSort, c *partials.CommonInfo) ([]*ent.Profile, error) {
	var err error
	var profiles []*ent.Profile

	query := m.Client.Profile.Query().WithTasks().WithTags().WithIssues(func(q *ent.ProfileIssueQuery) {
		q.WithTasksreports(func(q *ent.TaskReportQuery) { q.WithTask().All(context.Background()) }).All(context.Background())
	}).Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}

	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if tenantID == -1 {
		query = query.Where(profile.And(profile.Not(profile.HasSite()), profile.Not(profile.HasTenant())))
	} else {
		if siteID == -1 {
			query = query.Where(profile.HasTenantWith(tenant.ID(tenantID)), profile.Not(profile.HasSite()))
		} else {
			query = query.Where(profile.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))
		}
	}

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

func (m *Model) AddProfile(siteID int, tenantID int, description string) (*ent.Profile, error) {
	query := m.Client.Profile.Create().SetName(description)

	if tenantID != -1 {
		query.AddTenantIDs(tenantID)
		if siteID != -1 {
			query.AddSiteIDs(siteID)
		}
	}

	profile, err := query.Save(context.Background())
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func (m *Model) UpdateProfile(profileID int, description string, apply string, c *partials.CommonInfo) error {

	switch apply {
	case "applyToAll":
		return m.Client.Profile.Update().Where(profile.ID(profileID)).SetName(description).ClearTags().SetApplyToAll(true).Exec(context.Background())
	case "useTags":
		return m.Client.Profile.Update().Where(profile.ID(profileID)).SetName(description).SetApplyToAll(false).Exec(context.Background())
	}
	return m.Client.Profile.Update().Where(profile.ID(profileID)).SetName(description).ClearTags().SetApplyToAll(false).Exec(context.Background())
}

func (m *Model) GetProfileById(profileId int, c *partials.CommonInfo) (*ent.Profile, error) {

	return m.Client.Profile.Query().WithTags().WithTasks().WithIssues().Where(profile.ID(profileId)).First(context.Background())
}

func (m *Model) DeleteProfile(profileID int, c *partials.CommonInfo) error {
	_, err := m.Client.Task.Delete().Where(task.HasProfileWith(profile.ID(profileID))).Exec(context.Background())
	if err != nil {
		return err
	}

	_, err = m.Client.Profile.Delete().Where(profile.ID(profileID)).Exec(context.Background())
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

	return m.Client.ProfileIssue.Query().
		WithAgents().
		WithTasksreports(func(q *ent.TaskReportQuery) { q.WithTask().All(context.Background()) }).
		Where(profileissue.HasProfileWith(profile.ID(profileID))).
		Order(ent.Desc(profileissue.FieldWhen)).
		Limit(p.PageSize).
		Offset((p.CurrentPage - 1) * p.PageSize).All(context.Background())
}

func (m *Model) EnableProfile(profiledID int, enabled bool) error {
	return m.Client.Profile.Update().SetDisabled(!enabled).Where(profile.ID(profiledID)).Exec(context.Background())
}

func (m *Model) SetProfileAsGlobal(profiledID int) error {
	return m.Client.Profile.Update().ClearSite().ClearTenant().Where(profile.ID(profiledID)).Exec(context.Background())
}

func (m *Model) SetProfileAsTenantProfile(profiledID int, tenantID int) error {
	return m.Client.Profile.Update().AddTenantIDs(tenantID).ClearSite().Where(profile.ID(profiledID)).Exec(context.Background())
}

// TODO-Steve we should check which profiles can be listed based on user's role
func (m *Model) GetAllProfiles() ([]*ent.Profile, error) {
	return m.Client.Profile.Query().All(context.Background())
}

func (m *Model) CloneProfile(profileID int, description string, tenantID int, siteID int) error {
	// 1. Get the profile and its tasks
	profile, err := m.Client.Profile.Query().WithTasks().Where(profile.ID(profileID)).Only(context.Background())
	if err != nil {
		return err
	}

	// 2. Initiate transaction
	tx, err := m.Client.Tx(context.Background())
	if err != nil {
		return err
	}

	// 3. Create the new profile
	query := tx.Profile.Create().SetName(description)
	if tenantID != -1 {
		query.AddTenantIDs(tenantID)

		if siteID != -1 {
			query.AddSiteIDs(siteID)
		}
	}

	newProfile, err := query.Save(context.Background())
	if err != nil {
		return rollback(tx, err)
	}

	// 4. Clone the tasks to the profile
	for index, t := range profile.Edges.Tasks {
		if err := m.CloneTaskInProfileTransaction(tx, t.ID, t.Name, newProfile.ID, index+1); err != nil {
			return rollback(tx, err)
		}
	}

	// 5. Commit the transaction.
	return tx.Commit()
}

func rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		err = fmt.Errorf("%w: %v", err, rerr)
	}
	return err
}
