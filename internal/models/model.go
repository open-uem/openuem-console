package models

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	ent "github.com/doncicuto/openuem_ent"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Model struct {
	Client *ent.Client
}

func New(dbUrl string) (*Model, error) {
	model := Model{}

	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		return nil, fmt.Errorf("could not connect with Postgres database: %v", err)
	}

	model.Client = ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))

	// TODO Automatic migrations only in development
	ctx := context.Background()
	if os.Getenv("ENV") != "prod" {
		if err := model.Client.Schema.Create(ctx); err != nil {
			return nil, err
		}
	}

	return &model, nil
}

func (m *Model) Close() {
	m.Client.Close()
}
