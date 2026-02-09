package models

import (
	"context"
	"time"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/enrollmenttoken"
	"github.com/open-uem/ent/tenant"
)

func (m *Model) CreateEnrollmentToken(tenantID int, siteID *int, description string, tokenValue string, maxUses int, expiresAt *time.Time) (*ent.EnrollmentToken, error) {
	query := m.Client.EnrollmentToken.Create().
		SetToken(tokenValue).
		SetDescription(description).
		SetMaxUses(maxUses).
		SetActive(true).
		SetTenantID(tenantID)

	if siteID != nil && *siteID > 0 {
		query.SetSiteID(*siteID)
	}

	if expiresAt != nil {
		query.SetExpiresAt(*expiresAt)
	}

	return query.Save(context.Background())
}

func (m *Model) GetEnrollmentTokens(tenantID int) ([]*ent.EnrollmentToken, error) {
	return m.Client.EnrollmentToken.Query().
		Where(enrollmenttoken.HasTenantWith(tenant.ID(tenantID))).
		WithSite().
		Order(ent.Desc(enrollmenttoken.FieldCreated)).
		All(context.Background())
}

func (m *Model) GetEnrollmentTokenByID(tokenID int) (*ent.EnrollmentToken, error) {
	return m.Client.EnrollmentToken.Query().
		Where(enrollmenttoken.ID(tokenID)).
		WithSite().
		WithTenant().
		Only(context.Background())
}

func (m *Model) DeleteEnrollmentToken(tokenID int) error {
	return m.Client.EnrollmentToken.DeleteOneID(tokenID).Exec(context.Background())
}

func (m *Model) ToggleEnrollmentToken(tokenID int, active bool) error {
	return m.Client.EnrollmentToken.UpdateOneID(tokenID).
		SetActive(active).
		Exec(context.Background())
}

func (m *Model) GetEnrollmentTokenByValue(tokenValue string) (*ent.EnrollmentToken, error) {
	return m.Client.EnrollmentToken.Query().
		Where(enrollmenttoken.Token(tokenValue)).
		WithSite().
		WithTenant().
		Only(context.Background())
}

func (m *Model) IncrementEnrollmentTokenUses(tokenValue string) error {
	_, err := m.Client.EnrollmentToken.Update().
		Where(enrollmenttoken.Token(tokenValue)).
		AddCurrentUses(1).
		Save(context.Background())
	return err
}
