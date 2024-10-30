package models

import (
	"context"

	"github.com/doncicuto/openuem-console/internal/views/partials"
	ent "github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/sessions"
)

func (m *Model) CountAllSessions() (int, error) {
	count, err := m.Client.Sessions.Query().Count(context.Background())
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *Model) GetSessionsByPage(p partials.PaginationAndSort) ([]*ent.Sessions, error) {
	var err error
	var s []*ent.Sessions

	query := m.Client.Sessions.Query().WithOwner().Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)

	switch p.SortBy {
	case "token":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(sessions.FieldID))
		} else {
			query = query.Order(ent.Desc(sessions.FieldID))
		}
	case "uid":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(sessions.OwnerColumn))
		} else {
			query = query.Order(ent.Desc(sessions.OwnerColumn))
		}
	case "expiry":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(sessions.FieldExpiry))
		} else {
			query = query.Order(ent.Desc(sessions.FieldExpiry))
		}
	default:
		query = query.Order(ent.Desc(sessions.OwnerColumn))
	}

	s, err = query.All(context.Background())
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (m *Model) DeleteSession(token string) error {
	if err := m.Client.Sessions.DeleteOneID(token).Exec(context.Background()); err != nil {
		return err
	}
	return nil
}
