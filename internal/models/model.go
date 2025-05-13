package models

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/migrate"
	"github.com/open-uem/ent/orgmetadata"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/tag"
	"github.com/open-uem/ent/tenant"
)

type Model struct {
	Client *ent.Client
}

func New(dbUrl string, driverName, domain string) (*Model, error) {
	var db *sql.DB
	var err error

	model := Model{}

	switch driverName {
	case "pgx":
		db, err = sql.Open("pgx", dbUrl)
		if err != nil {
			return nil, fmt.Errorf("could not connect with Postgres database: %v", err)
		}
		model.Client = ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	default:
		return nil, fmt.Errorf("unsupported DB driver")
	}

	// TODO Automatic migrations only in non-stable versions
	ctx := context.Background()
	if os.Getenv("ENV") != "prod" {
		if err := model.Client.Schema.Create(ctx,
			migrate.WithDropIndex(true),
			migrate.WithDropColumn(true)); err != nil {
			return nil, err
		}
	}

	// Create default orgs and sites #feat-119
	if err := model.CreateDefaultTenantAndSite(); err != nil {
		return nil, err
	}

	// Associate agents without site to default site #feat-119
	if err := model.AssociateAgentsToDefaultTenantAndSite(); err != nil {
		return nil, err
	}

	// Associate tags without tenant to default tenant #feat-119
	if err := model.AssociateTagsToDefaultTenant(); err != nil {
		return nil, err
	}

	// Associate metadata without tenant to default tenant #feat-119
	if err := model.AssociateMetadataToDefaultTenant(); err != nil {
		return nil, err
	}

	// Associate profiles without tenant to default tenant #feat-119
	if err := model.AssociateProfilesToDefaultTenantAndSite(); err != nil {
		return nil, err
	}

	// Associate domain to default site #feat-119
	if err := model.AssociateDomainToDefaultSite(domain); err != nil {
		return nil, err
	}

	return &model, nil
}

func (m *Model) Close() error {
	return m.Client.Close()
}

func (m *Model) CreateDefaultTenantAndSite() error {
	nTenants, err := m.CountTenants()
	if err != nil {
		return fmt.Errorf("could not count existing tenants")
	}

	if nTenants == 0 {
		t, err := m.CreateDefaultTenant()
		if err != nil {
			return fmt.Errorf("could not create default tenant")
		}
		nSites, err := m.CountSites(t.ID)
		if err != nil {
			return fmt.Errorf("could not count existing sites")
		}

		if nSites == 0 {
			_, err := m.CreateDefaultSite(t)
			if err != nil {
				return fmt.Errorf("could not create default site")
			}

			// Create copy of global settings
			if err := m.CloneGlobalSettings(t.ID); err != nil {
				return fmt.Errorf("could not clone global settings, reason: %v", err)
			}
		}
	}

	return nil
}

func (m *Model) AssociateAgentsToDefaultTenantAndSite() error {

	t, err := m.GetDefaultTenant()
	if err != nil {
		return fmt.Errorf("could not find default tenant")
	}

	s, err := m.GetDefaultSite(t)
	if err != nil {
		return fmt.Errorf("coulf not find default site")
	}

	if err := m.AssociateDefaultSiteToAgents(s); err != nil {
		return fmt.Errorf("could not associate agents to default site")
	}

	return nil
}

func (m *Model) AssociateTagsToDefaultTenant() error {
	t, err := m.GetDefaultTenant()
	if err != nil {
		return fmt.Errorf("could not find default tenant")
	}

	return m.Client.Tag.Update().SetTenantID(t.ID).Where(tag.Not(tag.HasTenant())).Exec(context.Background())
}

func (m *Model) AssociateProfilesToDefaultTenantAndSite() error {
	t, err := m.GetDefaultTenant()
	if err != nil {
		return fmt.Errorf("could not find default tenant")
	}

	s, err := m.GetDefaultSite(t)
	if err != nil {
		return fmt.Errorf("coulf not find default site")
	}

	return m.Client.Profile.Update().SetSiteID(s.ID).Exec(context.Background())
}

func (m *Model) AssociateMetadataToDefaultTenant() error {
	t, err := m.GetDefaultTenant()
	if err != nil {
		return fmt.Errorf("could not find default tenant")
	}

	return m.Client.OrgMetadata.Update().SetTenantID(t.ID).Where(orgmetadata.Not(orgmetadata.HasTenant())).Exec(context.Background())
}

func (m *Model) AssociateDomainToDefaultSite(domain string) error {
	t, err := m.GetDefaultTenant()
	if err != nil {
		return fmt.Errorf("could not find default tenant")
	}

	s, err := m.GetDefaultSite(t)
	if err != nil {
		return fmt.Errorf("coulf not find default site")
	}

	return m.Client.Site.Update().SetDomain(domain).Where(site.ID(s.ID), site.HasTenantWith(tenant.ID(t.ID))).Exec(context.Background())
}
