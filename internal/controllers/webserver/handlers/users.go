package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"slices"
	"strings"

	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	openuem_ent "github.com/open-uem/ent"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views/admin_views"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"golang.org/x/crypto/ocsp"
	"software.sslmate.com/src/go-pkcs12"
)

type NewUser struct {
	UID      string `form:"uid" validate:"required"`
	Name     string `form:"name"`
	Email    string `form:"email" validate:"required,email"`
	Phone    string `form:"phone"`
	Country  string `form:"country"`
	AuthType string `form:"auth-type" validate:"required"`
}

func (h *Handler) ListUsers(c echo.Context, successMessage, errMessage string) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	f := filters.UserFilter{}

	usernameFilter := c.FormValue("filterByUsername")
	if usernameFilter != "" {
		f.Username = usernameFilter
	}

	nameFilter := c.FormValue("filterByName")
	if nameFilter != "" {
		f.Name = nameFilter
	}

	emailFilter := c.FormValue("filterByEmail")
	if emailFilter != "" {
		f.Email = emailFilter
	}

	phoneFilter := c.FormValue("filterByPhone")
	if phoneFilter != "" {
		f.Phone = phoneFilter
	}

	createdFrom := c.FormValue("filterByCreatedDateFrom")
	if createdFrom != "" {
		f.CreatedFrom = createdFrom
	}
	createdTo := c.FormValue("filterByCreatedDateTo")
	if createdTo != "" {
		f.CreatedTo = createdTo
	}

	modifiedFrom := c.FormValue("filterByModifiedDateFrom")
	if modifiedFrom != "" {
		f.ModifiedFrom = modifiedFrom
	}
	modifiedTo := c.FormValue("filterByModifiedDateTo")
	if modifiedTo != "" {
		f.ModifiedTo = modifiedTo
	}

	filteredRegisterStatus := []string{}
	for index := range openuem_nats.RegisterPossibleStatus() {
		value := c.FormValue(fmt.Sprintf("filterByRegisterStatus%d", index))
		if value != "" {
			filteredRegisterStatus = append(filteredRegisterStatus, value)
		}
	}
	f.RegisterOptions = filteredRegisterStatus

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

	p.NItems, err = h.Model.CountAllUsers(f)
	if err != nil {
		successMessage = ""
		errMessage = err.Error()
	}

	users, err := h.Model.GetUsersByPage(p, f)
	if err != nil {
		successMessage = ""
		errMessage = err.Error()
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	agentsExists, err := h.Model.AgentsExists(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	warnAboutSMTP := h.Model.IsPasswdAuthEnabled() && !h.Model.IsSMTPConfigured()

	return RenderView(c, admin_views.UsersIndex(" | Users", admin_views.Users(c, p, f, users, successMessage, errMessage, refreshTime, agentsExists, serversExists, warnAboutSMTP, commonInfo), commonInfo))
}

func (h *Handler) NewUser(c echo.Context) error {
	var err error

	settings, err := h.Model.GetAuthenticationSettings()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.could_not_get_settings", err.Error()), true))
	}

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	defaultCountry, err := h.Model.GetDefaultCountry()
	if err != nil {
		return err
	}

	agentsExists, err := h.Model.AgentsExists(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return RenderView(c, admin_views.UsersIndex(" | Users", admin_views.NewUser(c, defaultCountry, agentsExists, serversExists, commonInfo, settings), commonInfo))
}

func (h *Handler) AddUser(c echo.Context) error {
	u := NewUser{}
	successMessage := ""
	errMessage := ""

	decoder := form.NewDecoder()
	if err := c.Request().ParseForm(); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}
	err := decoder.Decode(&u, c.Request().Form)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(u); err != nil {
		// TODO Try to translate and create a nice error message
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	exists, err := h.Model.UserExists(u.UID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if exists {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.user_id_exists"), true))
	}

	// TODO - use constants for certificate types
	if !slices.Contains([]string{"certificate", "passwd", "oidc"}, u.AuthType) {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.invalid_type", u.AuthType), true))
	}

	addedUser, err := h.Model.AddUser(u.UID, u.Name, u.Email, u.Phone, u.Country, u.AuthType)
	if err != nil {
		// TODO manage duplicate key error
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	switch u.AuthType {
	case "certificate":
		if err := h.sendConfirmationEmail(c, addedUser); err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}
		successMessage = i18n.T(c.Request().Context(), "new.user.success")
	case "oidc":
		successMessage = i18n.T(c.Request().Context(), "new.user.success_oidc")
	case "passwd":
		if err := h.sendLinkToGeneratePassword(c, addedUser); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.could_not_send_new_account_email"), false))
		}
		successMessage = i18n.T(c.Request().Context(), "new.user.success_passwd")
	}

	return h.ListUsers(c, successMessage, errMessage)
}

func (h *Handler) RequestUserCertificate(c echo.Context) error {

	uid := c.Param("uid")

	user, err := h.Model.GetUserById(uid)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := h.SendCertificateRequestToNATS(c, user); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	successMessage := i18n.T(c.Request().Context(), "users.certificate_requested")
	return h.ListUsers(c, successMessage, "")
}

func (h *Handler) SendCertificateRequestToNATS(c echo.Context, user *openuem_ent.User) error {
	userCertYears, err := h.Model.GetDefaultUserCertDuration()
	if err != nil {
		return err
	}

	consoleUrl := c.Request().Header.Get("Origin")
	if h.ReverseProxyServer != "" {
		consoleUrl = h.ReverseProxyServer
	}

	certRequest := openuem_nats.CertificateRequest{
		Username:   user.ID,
		FullName:   user.Name,
		Email:      user.Email,
		Country:    user.Country,
		Password:   user.CertClearPassword,
		YearsValid: userCertYears,
		ConsoleURL: consoleUrl,
	}

	data, err := json.Marshal(certRequest)
	if err != nil {
		return err
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return fmt.Errorf("%s", i18n.T(c.Request().Context(), "nats.not_connected"))
	}

	if err := h.NATSConnection.Publish("certificates.user", data); err != nil {
		return err
	}
	return nil
}

func (h *Handler) DeleteUser(c echo.Context) error {
	uid := c.Param("uid")

	if uid == "admin" || uid == "openuem" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.admin_cannot_be_removed"), false))
	}
	_, err := h.Model.GetUserById(uid)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// Delete user
	if err := h.Model.DeleteUser(uid); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// Revoke certificate
	cert, err := h.Model.GetCertificateByUID(uid)
	if err != nil {
		if !openuem_ent.IsNotFound(err) {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}
		successMessage := i18n.T(c.Request().Context(), "users.deleted")
		return h.ListUsers(c, successMessage, "")
	}

	if err := h.Model.RevokeCertificate(cert, "user has been deleted", ocsp.CessationOfOperation); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// Delete certificate information
	if err := h.Model.DeleteCertificate(cert.ID); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	successMessage := i18n.T(c.Request().Context(), "users.deleted")
	return h.ListUsers(c, successMessage, "")
}

func (h *Handler) RenewUserCertificate(c echo.Context) error {
	uid := c.Param("uid")
	user, err := h.Model.GetUserById(uid)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// First revoke certificate
	cert, err := h.Model.GetCertificateByUID(uid)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := h.Model.RevokeCertificate(cert, "a new certificate has been requested", ocsp.CessationOfOperation); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// Now delete certificate
	if err := h.Model.DeleteCertificate(cert.ID); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// Now request a new certificate
	consoleUrl := c.Request().Header.Get("Origin")
	if h.ReverseProxyServer != "" {
		consoleUrl = h.ReverseProxyServer
	}
	certRequest := openuem_nats.CertificateRequest{
		Username:   user.ID,
		FullName:   user.Name,
		Email:      user.Email,
		Country:    user.Country,
		ConsoleURL: consoleUrl,
		YearsValid: 1,
	}

	data, err := json.Marshal(certRequest)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
	}

	if err := h.NATSConnection.Publish("certificates.user", data); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	successMessage := i18n.T(c.Request().Context(), "users.certificate_requested")
	return h.ListUsers(c, successMessage, "")
}

func (h *Handler) SetEmailConfirmed(c echo.Context) error {
	uid := c.Param("uid")
	exists, err := h.Model.UserExists(uid)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if !exists {
		return RenderError(c, partials.ErrorMessage("user doesn't exist", false))
	}

	err = h.Model.Client.User.UpdateOneID(uid).SetEmailVerified(true).SetRegister(openuem_nats.REGISTER_IN_REVIEW).Exec(context.Background())
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return h.ListUsers(c, i18n.T(c.Request().Context(), "users.email_confirmed"), "")
}

func (h *Handler) ApproveAccount(c echo.Context) error {
	uid := c.Param("uid")
	exists, err := h.Model.UserExists(uid)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if !exists {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.user_not_found"), true))
	}

	err = h.Model.Client.User.UpdateOneID(uid).SetRegister(openuem_nats.REGISTER_APPROVED).Exec(context.Background())
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.could_not_update_register", err.Error()), false))
	}

	return h.ListUsers(c, i18n.T(c.Request().Context(), "users.approved"), "")
}

func (h *Handler) ResendPasswordLink(c echo.Context) error {
	uid := c.Param("uid")
	u, err := h.Model.GetUserById(uid)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.user_not_found"), true))
	}

	err = h.Model.Client.User.UpdateOneID(uid).SetRegister(openuem_nats.REGISTER_PASSWORD_LINK_SENT).Exec(context.Background())
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.could_not_update_register", err.Error()), true))
	}

	if err := h.sendLinkToGeneratePassword(c, u); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.could_not_update_register", err.Error()), true))
	}

	return h.ListUsers(c, i18n.T(c.Request().Context(), "users.new_password_link_sent"), "")
}

func (h *Handler) AskForConfirmation(c echo.Context) error {
	uid := c.Param("uid")
	user, err := h.Model.GetUserById(uid)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := h.sendConfirmationEmail(c, user); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return h.ListUsers(c, i18n.T(c.Request().Context(), "users.new_confirmation_email_sent", user.Email), "")
}

func (h *Handler) sendConfirmationEmail(c echo.Context, user *openuem_ent.User) error {
	token, err := h.generateEmailToken(user.ID, "Email Confirmation", 24)
	if err != nil {
		return err
	}

	notification := openuem_nats.Notification{
		To:               user.Email,
		Subject:          "Please, confirm your email address",
		MessageTitle:     "OpenUEM | Verify your email address",
		MessageText:      "Please, confirm your email address so that it can be used to receive emails from OpenUEM",
		MessageGreeting:  fmt.Sprintf("Hi %s", user.Name),
		MessageAction:    "Confirm email",
		MessageActionURL: c.Request().Header.Get("Origin") + "/auth/confirm/" + token,
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return fmt.Errorf("%s", i18n.T(c.Request().Context(), "nats.not_connected"))
	}

	if err := h.NATSConnection.Publish("notification.confirm_email", data); err != nil {
		return err
	}

	return nil
}

func (h *Handler) sendLinkToGeneratePassword(c echo.Context, user *openuem_ent.User) error {
	token, err := h.generateEmailToken(user.ID, "New password", 1)
	if err != nil {
		return err
	}

	if err := h.Model.SaveNewAccountToken(user.ID, token); err != nil {
		return err
	}

	notification := openuem_nats.Notification{
		To:               user.Email,
		Subject:          "A new account has been created",
		MessageTitle:     "OpenUEM | A new account has been created",
		MessageText:      "You must set a password to log into OpenUEM. Click the link below to set your initial password. NOTE: the following link will only be valid for one hour",
		MessageGreeting:  fmt.Sprintf("Hi %s, a new OpenUEM account with username %s has been created for you", user.Name, user.ID),
		MessageAction:    "Generate a password",
		MessageActionURL: c.Request().Header.Get("Origin") + fmt.Sprintf("/login/new?token=%s", token),
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return fmt.Errorf("%s", i18n.T(c.Request().Context(), "nats.not_connected"))
	}

	if err := h.NATSConnection.Publish("notification.confirm_email", data); err != nil {
		return err
	}

	return nil
}

func (h *Handler) ImportUsers(c echo.Context) error {
	// Source
	file, err := c.FormFile("csvFile")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	r := csv.NewReader(src)

	validate := validator.New()
	index := 1

	var errors = []string{}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}

		user := openuem_ent.User{}

		if len(record) != 6 {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.import_error_wrong_format", index), false))
		}

		if record[0] == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.import_required", "userid", index), false))
		}
		user.ID = record[0]
		user.Name = record[1]

		if record[2] == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.import_required", "email", index), false))
		}

		if record[2] != "" {
			if errs := validate.Var(record[2], "email"); errs != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.import_wrong_email", index), false))
			}
		}
		user.Email = record[2]

		if record[3] != "" {
			if errs := validate.Var(strings.ToUpper(record[3]), "iso3166_1_alpha2"); errs != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.import_wrong_country", index, record[3]), false))
			}
		}
		user.Country = strings.ToUpper(record[3])
		user.Phone = record[4]

		authType := record[5]

		if !slices.Contains([]string{"certificate", "passwd", "oidc"}, authType) {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.import_wrong_oidc"), false))
		}

		index++

		u, err := h.Model.AddUser(user.ID, user.Name, user.Email, user.Phone, user.Country, authType)
		if err != nil {
			// TODO manage duplicate key error
			errors = append(errors, err.Error())
			continue
		}

		switch authType {
		case "certificate":
			u.CertClearPassword = pkcs12.DefaultPassword
			if err := h.SendCertificateRequestToNATS(c, u); err != nil {
				errors = append(errors, err.Error())
			}
		case "passwd":
			if err := h.sendLinkToGeneratePassword(c, u); err != nil {
				errors = append(errors, err.Error())
			}
		}

	}

	if len(errors) > 0 {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.import_wrong_users", strings.Join(errors, ",")), false))
	}

	return h.ListUsers(c, i18n.T(c.Request().Context(), "users.import_success"), "")
}
