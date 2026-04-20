package models

import (
	"context"
	"errors"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/sessions"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/utils"
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

func (m *Model) GetSessionsTokens() ([]*ent.Sessions, error) {
	return m.Client.Sessions.Query().Select(sessions.FieldID).All(context.Background())
}

func (m *Model) UpdateSessionToken(tokenID string, newTokenID string) error {
	session, err := m.Client.Sessions.Query().WithOwner().Where(sessions.ID(tokenID)).First(context.Background())
	if err != nil {
		return err
	}

	if session.Edges.Owner != nil {
		if err := m.Client.Sessions.Create().SetID(newTokenID).SetExpiry(session.Expiry).SetData(session.Data).SetOwnerID(session.Edges.Owner.ID).Exec(context.Background()); err != nil {
			return err
		}
	} else {
		if err := m.Client.Sessions.Create().SetID(newTokenID).SetExpiry(session.Expiry).SetData(session.Data).Exec(context.Background()); err != nil {
			return err
		}
	}

	if err := m.Client.Sessions.DeleteOneID(tokenID).Exec(context.Background()); err != nil {
		return err
	}

	return nil
}

func (m *Model) AddUserToSession(token string, userID string, encryptionMasterKey string) error {
	if encryptionMasterKey != "" {

		allSessions, err := m.Client.Sessions.Query().All(context.Background())
		if err != nil {
			return err
		}

		for _, s := range allSessions {
			isTokenEncrypted, err := utils.IsSensitiveFieldEncrypted(s.ID, encryptionMasterKey)
			if err != nil {
				return err
			}

			if isTokenEncrypted {
				tokenInClear, err := utils.DecryptSensitiveField(s.ID, encryptionMasterKey)
				if err != nil {
					return err
				}

				if tokenInClear != token {
					continue
				}

				return m.Client.Sessions.UpdateOneID(s.ID).SetOwnerID(userID).Exec(context.Background())
			}

			return errors.New("token is not encrypted in database")
		}

	} else {
		return m.Client.Sessions.UpdateOneID(token).SetOwnerID(userID).Exec(context.Background())
	}

	return nil
}
