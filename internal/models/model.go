package models

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
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
	case "sqlite3":
		db, err = sql.Open("sqlite3", dbUrl)
		if err != nil {
			return nil, fmt.Errorf("could not connect with SQLite database: %v", err)
		}
		model.Client = ent.NewClient(ent.Driver(entsql.OpenDB(dialect.SQLite, db)))
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
