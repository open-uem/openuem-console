package models

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/alexedwards/argon2id"
	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/recoverycode"
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

func (m *Model) ConsumeRecoveryCode(uid string, code string) bool {
	hashes, err := m.Client.RecoveryCode.Query().Where(recoverycode.HasUserWith(user.ID(uid))).All(context.Background())
	if err != nil {
		log.Println("[ERROR]: could not find recovery codes for this user")
		return false
	}

	for _, hash := range hashes {
		match, err := argon2id.ComparePasswordAndHash(code, hash.Code)
		if err == nil && match {
			if err := m.Client.RecoveryCode.Update().SetUsed(true).Where(recoverycode.ID(hash.ID)).Exec(context.Background()); err != nil {
				log.Printf("[ERROR]: could not invalidate recovery code %s, reason: %v", code, err)
				return false
			}
			return true
		}
	}

	log.Println("[ERROR]: could not find the recovery code")
	return false
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

func (m *Model) CreateDefaultAdminPassword() error {
	exist, err := m.Client.User.Query().Where(user.ID("openuem")).Exist(context.Background())
	if err != nil {
		return err
	}

	if !exist {
		hash, err := argon2id.CreateHash("pa$$word", argon2id.DefaultParams)
		if err != nil {
			return err
		}

		return m.Client.User.Create().
			SetID("openuem").
			SetRegister("users.force_change_password").
			SetName("OpenUEM Administrator").
			SetPasswd(true).
			SetHash(hash).
			Exec(context.Background())
	}

	return nil
}

func (m *Model) ChangePassword(username string, password string) error {
	exist, err := m.Client.User.Query().Where(user.ID(username)).Exist(context.Background())
	if err != nil {
		return err
	}

	if exist {
		hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
		if err != nil {
			return err
		}

		// Save password
		return m.Client.User.Update().Where(user.ID(username)).SetRegister("users.completed").SetHash(hash).Exec(context.Background())
	} else {
		return errors.New("user not found")
	}
}

func (m *Model) SaveTOTPSecretKey(username string, secret string) error {
	exist, err := m.Client.User.Query().Where(user.ID(username)).Exist(context.Background())
	if err != nil {
		return err
	}

	if exist {
		return m.Client.User.Update().Where(user.ID(username)).SetTotpSecret(secret).Exec(context.Background())
	} else {
		return errors.New("user not found")
	}
}

func (m *Model) SaveRecoveryCodes(username string, codes []string) error {
	exist, err := m.Client.User.Query().Where(user.ID(username)).Exist(context.Background())
	if err != nil {
		return err
	}

	if exist {
		// Check for existing recovery codes
		hasCodes, err := m.Client.RecoveryCode.Query().Where(recoverycode.HasUserWith(user.ID(username))).Exist(context.Background())
		if err != nil {
			return err
		}

		// Delete existing codes
		if hasCodes {
			if _, err := m.Client.RecoveryCode.Delete().Where(recoverycode.HasUserWith(user.ID(username))).Exec(context.Background()); err != nil {
				return err
			}
		}

		// Generate hashes
		for _, c := range codes {
			hash, err := argon2id.CreateHash(c, argon2id.DefaultParams)
			if err != nil {
				return err
			}

			if err := m.Client.RecoveryCode.Create().SetUserID(username).SetCode(hash).Exec(context.Background()); err != nil {
				return err
			}
		}

		return m.Client.User.Update().SetUse2fa(true).SetTotpSecretConfirmed(true).Where(user.ID(username)).Exec(context.Background())
	} else {
		return errors.New("user not found")
	}
}

func (m *Model) GetUserHash(username string) (*ent.User, error) {
	return m.Client.User.Query().Select(user.FieldHash).Where(user.ID(username)).First(context.Background())
}

func (m *Model) GetUserTOTPSecret(username string) (*ent.User, error) {
	return m.Client.User.Query().Select(user.FieldTotpSecret).Where(user.ID(username)).First(context.Background())
}

func (m *Model) Disable2FA(username string) error {
	// Delete recovery codes
	_, err := m.Client.RecoveryCode.Delete().Where(recoverycode.HasUserWith(user.ID(username))).Exec(context.Background())
	if err != nil {
		return err
	}

	// Disable 2FA and remove TOTP secret
	return m.Client.User.UpdateOneID(username).SetUse2fa(false).SetTotpSecret("").SetTotpSecretConfirmed(false).Exec(context.Background())
}
