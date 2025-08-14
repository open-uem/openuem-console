package models

import (
	"context"
	"time"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/user"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (m *Model) CountAllUsers(f filters.UserFilter) (int, error) {
	query := m.Client.User.Query()

	applyUsersFilter(query, f)

	count, err := query.Count(context.Background())
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *Model) GetUsersByPage(p partials.PaginationAndSort, f filters.UserFilter) ([]*ent.User, error) {
	query := m.Client.User.Query()

	applyUsersFilter(query, f)

	switch p.SortBy {
	case "uid":
		if p.SortOrder == "asc" {
			query.Order(ent.Asc(user.FieldID))
		} else {
			query.Order(ent.Desc(user.FieldID))
		}
	case "name":
		if p.SortOrder == "asc" {
			query.Order(ent.Asc(user.FieldName))
		} else {
			query.Order(ent.Desc(user.FieldName))
		}
	case "email":
		if p.SortOrder == "asc" {
			query.Order(ent.Asc(user.FieldEmail))
		} else {
			query.Order(ent.Desc(user.FieldEmail))
		}
	case "phone":
		if p.SortOrder == "asc" {
			query.Order(ent.Asc(user.FieldPhone))
		} else {
			query.Order(ent.Desc(user.FieldPhone))
		}
	case "country":
		if p.SortOrder == "asc" {
			query.Order(ent.Asc(user.FieldCountry))
		} else {
			query.Order(ent.Desc(user.FieldCountry))
		}
	case "register":
		if p.SortOrder == "asc" {
			query.Order(ent.Asc(user.FieldRegister))
		} else {
			query.Order(ent.Desc(user.FieldRegister))
		}
	case "created":
		if p.SortOrder == "asc" {
			query.Order(ent.Asc(user.FieldCreated))
		} else {
			query.Order(ent.Desc(user.FieldCreated))
		}
	case "modified":
		if p.SortOrder == "asc" {
			query.Order(ent.Asc(user.FieldModified))
		} else {
			query.Order(ent.Desc(user.FieldModified))
		}

	default:
		query.Order(ent.Desc(user.FieldID))
	}

	return query.Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize).All(context.Background())
}

func (m *Model) UserExists(uid string) (bool, error) {
	return m.Client.User.Query().Where(user.ID(uid)).Exist(context.Background())
}

func (m *Model) EmailExists(email string) (bool, error) {
	return m.Client.User.Query().Where(user.Email(email)).Exist(context.Background())
}

func (m *Model) AddUser(uid, name, email, phone, country string, oidc bool) error {
	query := m.Client.User.Create().SetID(uid).SetName(name).SetEmail(email).SetPhone(phone).SetCountry(country).SetOpenid(oidc).SetCreated(time.Now())

	if oidc {
		query.SetEmailVerified(true).SetRegister(openuem_nats.REGISTER_IN_REVIEW)
	}

	return query.Exec(context.Background())
}

func (m *Model) AddImportedUser(uid, name, email, phone, country string, oidc bool) error {
	query := m.Client.User.Create().SetID(uid).SetName(name).SetEmail(email).SetPhone(phone).SetCountry(country).SetOpenid(oidc).SetCreated(time.Now())

	if oidc {
		query.SetRegister(openuem_nats.REGISTER_IN_REVIEW)
	} else {
		query.SetRegister(openuem_nats.REGISTER_CERTIFICATE_SENT)
	}

	return query.Exec(context.Background())
}

func (m *Model) AddOIDCUser(uid, name, email, phone string, emailVerified bool) error {
	_, err := m.Client.User.Create().SetID(uid).SetName(name).SetEmail(email).SetPhone(phone).SetEmailVerified(emailVerified).SetCreated(time.Now()).SetRegister(openuem_nats.REGISTER_APPROVED).SetOpenid(true).Save(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (m *Model) UpdateUser(uid, name, email, phone, country string) error {
	query := m.Client.User.UpdateOneID(uid).SetName(name).SetEmail(email).SetPhone(phone).SetCountry(country).SetModified(time.Now())

	return query.Exec(context.Background())
}

func (m *Model) RegisterUser(uid, name, email, phone, country, password string, oidc bool) error {
	_, err := m.Client.User.Create().SetID(uid).SetName(name).SetEmail(email).SetPhone(phone).SetCountry(country).SetCertClearPassword(password).SetOpenid(oidc).SetCreated(time.Now()).Save(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (m *Model) GetUserById(uid string) (*ent.User, error) {
	return m.Client.User.Get(context.Background(), uid)
}

func (m *Model) ConfirmEmail(uid string) error {
	return m.Client.User.Update().SetEmailVerified(true).SetRegister(openuem_nats.REGISTER_IN_REVIEW).Where(user.ID(uid)).Exec(context.Background())
}

func (m *Model) UserSetRevokedCertificate(uid string) error {
	return m.Client.User.Update().SetRegister(openuem_nats.REGISTER_REVOKED).Where(user.ID(uid)).Exec(context.Background())
}

func (m *Model) ConfirmLogIn(uid string) error {
	return m.Client.User.Update().SetRegister(openuem_nats.REGISTER_COMPLETE).SetCertClearPassword("").Where(user.ID(uid)).Exec(context.Background())
}

func (m *Model) DeleteUser(uid string) error {
	return m.Client.User.DeleteOneID(uid).Exec(context.Background())
}

func applyUsersFilter(query *ent.UserQuery, f filters.UserFilter) {

	if len(f.Username) > 0 {
		query.Where(user.IDContainsFold(f.Username))
	}

	if len(f.Name) > 0 {
		query.Where(user.NameContainsFold(f.Name))
	}

	if len(f.Email) > 0 {
		query.Where(user.EmailContainsFold(f.Email))
	}

	if len(f.Phone) > 0 {
		query.Where(user.PhoneContainsFold(f.Phone))
	}

	if len(f.CreatedFrom) > 0 {
		dateFrom, err := time.Parse("2006-01-02", f.CreatedFrom)
		if err == nil {
			query.Where(user.CreatedGTE(dateFrom))
		}
	}

	if len(f.CreatedTo) > 0 {
		dateTo, err := time.Parse("2006-01-02", f.CreatedTo)
		if err == nil {
			query.Where(user.CreatedLTE(dateTo))
		}
	}

	if len(f.ModifiedFrom) > 0 {
		dateFrom, err := time.Parse("2006-01-02", f.ModifiedFrom)
		if err == nil {
			query.Where(user.ModifiedGTE(dateFrom))
		}
	}

	if len(f.ModifiedTo) > 0 {
		dateTo, err := time.Parse("2006-01-02", f.ModifiedTo)
		if err == nil {
			query.Where(user.ModifiedLTE(dateTo))
		}
	}

	if len(f.RegisterOptions) > 0 {
		query.Where(user.RegisterIn(f.RegisterOptions...))
	}
}

func (m *Model) SaveOIDCTokenInfo(uid string, accessToken string, refreshToken string, idToken string, tokenType string, expiry int) error {
	return m.Client.User.UpdateOneID(uid).
		SetAccessToken(accessToken).
		SetRefreshToken(refreshToken).
		SetIDToken(idToken).
		SetTokenType(tokenType).
		SetTokenExpiry(expiry).
		Exec(context.Background())
}
