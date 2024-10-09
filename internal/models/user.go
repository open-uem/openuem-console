package models

import (
	"context"
	"time"

	"github.com/doncicuto/openuem-console/internal/views/partials"
	ent "github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/user"
)

func (m *Model) GetAllUsers() ([]*ent.User, error) {
	users, err := m.Client.User.Query().All(context.Background())
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (m *Model) CountAllUsers() (int, error) {
	count, err := m.Client.User.Query().Count(context.Background())
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *Model) GetUsersByPage(p partials.PaginationAndSort) ([]*ent.User, error) {
	var err error
	var users []*ent.User

	query := m.Client.User.Query().Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)

	switch p.SortBy {
	case "uid":
		if p.SortOrder == "asc" {
			users, err = query.Order(ent.Asc(user.FieldID)).All(context.Background())
		} else {
			users, err = query.Order(ent.Desc(user.FieldID)).All(context.Background())
		}
	case "name":
		if p.SortOrder == "asc" {
			users, err = query.Order(ent.Asc(user.FieldName)).All(context.Background())
		} else {
			users, err = query.Order(ent.Desc(user.FieldName)).All(context.Background())
		}
	case "email":
		if p.SortOrder == "asc" {
			users, err = query.Order(ent.Asc(user.FieldEmail)).All(context.Background())
		} else {
			users, err = query.Order(ent.Desc(user.FieldEmail)).All(context.Background())
		}
	case "phone":
		if p.SortOrder == "asc" {
			users, err = query.Order(ent.Asc(user.FieldPhone)).All(context.Background())
		} else {
			users, err = query.Order(ent.Desc(user.FieldPhone)).All(context.Background())
		}
	case "country":
		if p.SortOrder == "asc" {
			users, err = query.Order(ent.Asc(user.FieldCountry)).All(context.Background())
		} else {
			users, err = query.Order(ent.Desc(user.FieldCountry)).All(context.Background())
		}
	case "register":
		if p.SortOrder == "asc" {
			users, err = query.Order(ent.Asc(user.FieldRegister)).All(context.Background())
		} else {
			users, err = query.Order(ent.Desc(user.FieldRegister)).All(context.Background())
		}
	case "created":
		if p.SortOrder == "asc" {
			users, err = query.Order(ent.Asc(user.FieldCreated)).All(context.Background())
		} else {
			users, err = query.Order(ent.Desc(user.FieldCreated)).All(context.Background())
		}
	case "modified":
		if p.SortOrder == "asc" {
			users, err = query.Order(ent.Asc(user.FieldModified)).All(context.Background())
		} else {
			users, err = query.Order(ent.Desc(user.FieldModified)).All(context.Background())
		}

	default:
		users, err = query.Order(ent.Desc(user.FieldID)).All(context.Background())
	}

	if err != nil {
		return nil, err
	}
	return users, nil
}

func (m *Model) UserExists(uid string) (bool, error) {
	return m.Client.User.Query().Where(user.ID(uid)).Exist(context.Background())
}

func (m *Model) EmailExists(email string) (bool, error) {
	return m.Client.User.Query().Where(user.Email(email)).Exist(context.Background())
}

func (m *Model) AddUser(uid, name, email, phone string) error {
	_, err := m.Client.User.Create().SetID(uid).SetName(name).SetEmail(email).SetPhone(phone).SetCreated(time.Now()).Save(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (m *Model) RegisterUser(uid, name, email, phone, country, password string) error {
	_, err := m.Client.User.Create().SetID(uid).SetName(name).SetEmail(email).SetPhone(phone).SetCountry(country).SetCertClearPassword(password).SetCreated(time.Now()).Save(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (m *Model) GetUserById(uid string) (*ent.User, error) {
	return m.Client.User.Get(context.Background(), uid)
}

func (m *Model) ConfirmEmail(uid string) error {
	return m.Client.User.Update().SetEmailVerified(true).SetRegister("users.review_request").Where(user.ID(uid)).Exec(context.Background())
}

func (m *Model) ConfirmLogIn(uid string) error {
	return m.Client.User.Update().SetRegister("users.completed").SetCertClearPassword("").Where(user.ID(uid)).Exec(context.Background())
}

func (m *Model) DeleteUser(uid string) error {
	return m.Client.User.DeleteOneID(uid).Exec(context.Background())
}
