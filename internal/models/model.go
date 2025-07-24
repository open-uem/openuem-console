package models

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/migrate"
	"github.com/open-uem/ent/orgmetadata"
	"github.com/open-uem/ent/profile"
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

	return m.Client.Tag.Update().Where(tag.Not(tag.HasTenant())).SetTenantID(t.ID).Exec(context.Background())
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

	return m.Client.Profile.Update().Where(profile.Not(profile.HasSite())).SetSiteID(s.ID).Exec(context.Background())
}

func (m *Model) AssociateMetadataToDefaultTenant() error {
	t, err := m.GetDefaultTenant()
	if err != nil {
		return fmt.Errorf("could not find default tenant")
	}

	return m.Client.OrgMetadata.Update().Where(orgmetadata.Not(orgmetadata.HasTenant())).SetTenantID(t.ID).Exec(context.Background())
}

func (m *Model) AssociateDomainToDefaultSite(domain string) error {
	t, err := m.GetDefaultTenant()
	if err != nil {
		return fmt.Errorf("could not find default tenant")
	}

	s, err := m.GetDefaultSite(t)
	if err != nil {
		return fmt.Errorf("could not find default site")
	}

	return m.Client.Site.Update().SetDomain(domain).Where(site.ID(s.ID), site.HasTenantWith(tenant.ID(t.ID))).Exec(context.Background())
}

func (m *Model) SetDefaultNickname() error {
	// look for agents that has no nickname and set it to the hostname
	migrateAgents, err := m.Client.Agent.Query().Where(agent.Or(agent.Nickname(""), agent.NicknameIsNil())).All(context.Background())
	if err != nil {
		return fmt.Errorf("could not find agents without nickname")
	}

	for _, a := range migrateAgents {
		if err := m.Client.Agent.Update().Where(agent.ID(a.ID)).SetNickname(a.Hostname).Exec(context.Background()); err != nil {
			log.Printf("[ERROR]: could not set default nickname to agent: %s, reason: %v", a.Hostname, err)
		}
	}

	return nil
}
