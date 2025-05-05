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
)

type Model struct {
	Client *ent.Client
}

func New(dbUrl string, driverName string) (*Model, error) {
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
		tenant, err := m.CreateDefaultTenant()
		if err != nil {
			return fmt.Errorf("could not create default tenant")
		}
		nSites, err := m.CountSites()
		if err != nil {
			return fmt.Errorf("could not count existing sites")
		}

		if nSites == 0 {
			_, err := m.CreateDefaultSite(tenant)
			if err != nil {
				return fmt.Errorf("could not create default site")
			}
		}
	}

	return nil
}

func (m *Model) AssociateAgentsToDefaultTenantAndSite() error {

	tenant, err := m.GetDefaultTenant()
	if err != nil {
		return fmt.Errorf("could not find default tenant")
	}

	site, err := m.GetDefaultSite(tenant)
	if err != nil {
		return fmt.Errorf("coulf not find default site")
	}

	if err := m.AssociateDefaultSiteToAgents(site); err != nil {
		return fmt.Errorf("could not associate agents to default site")
	}

	return nil
}
