package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/doncicuto/openuem-console/internal/views/admin_views"
	"github.com/doncicuto/openuem-console/internal/views/filters"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_nats"
	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/ocsp"
	"software.sslmate.com/src/go-pkcs12"
)

type NewUser struct {
	UID     string `form:"uid" validate:"required"`
	Name    string `form:"name" validate:"required"`
	Email   string `form:"email" validate:"required,email"`
	Phone   string `form:"phone"`
	Country string `form:"country"`
}

func (h *Handler) ListUsers(c echo.Context, successMessage, errMessage string) error {
	var err error

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

	createdFrom := c.FormValue("filterCreatedDateFrom")
	if createdFrom != "" {
		f.CreatedFrom = createdFrom
	}
	createdTo := c.FormValue("filterCreatedDateTo")
	if createdTo != "" {
		f.CreatedTo = createdTo
	}

	modifiedFrom := c.FormValue("filterModifiedDateFrom")
	if modifiedFrom != "" {
		f.ModifiedFrom = modifiedFrom
	}
	modifiedTo := c.FormValue("filterModifiedDateTo")
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
	p.GetPaginationAndSortParams(c)

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

	return RenderView(c, admin_views.UsersIndex(" | Users", admin_views.Users(c, users, p, f, successMessage, errMessage, refreshTime)))
}

func (h *Handler) NewUser(c echo.Context) error {
	defaultCountry, err := h.Model.GetDefaultCountry()
	if err != nil {
		return err
	}

	return RenderView(c, admin_views.UsersIndex(" | Users", admin_views.NewUser(c, defaultCountry)))
}

func (h *Handler) AddUser(c echo.Context) error {
	u := NewUser{}
	successMessage := ""
	errMessage := ""

	decoder := form.NewDecoder()
	if err := c.Request().ParseForm(); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	err := decoder.Decode(&u, c.Request().Form)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(u); err != nil {
		// TODO Try to translate and create a nice error message
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	err = h.Model.AddUser(u.UID, u.Name, u.Email, u.Phone, u.Country)
	if err != nil {
		// TODO manage duplicate key error
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	addedUser, err := h.Model.GetUserById(u.UID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := sendConfirmationEmail(h, c, addedUser); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	successMessage = i18n.T(c.Request().Context(), "new.user.success")
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

	certRequest := openuem_nats.CertificateRequest{
		Username:   user.ID,
		FullName:   user.Name,
		Email:      user.Email,
		Country:    user.Country,
		Password:   user.CertClearPassword,
		YearsValid: userCertYears,
	}

	data, err := json.Marshal(certRequest)
	if err != nil {
		return err
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return fmt.Errorf(i18n.T(c.Request().Context(), "nats.not_connected"))
	}

	if err := h.NATSConnection.Publish("certificates.new", data); err != nil {
		return err
	}
	return nil
}

func (h *Handler) DeleteUser(c echo.Context) error {
	uid := c.Param("uid")
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
	certRequest := openuem_nats.CertificateRequest{
		Username:   user.ID,
		FullName:   user.Name,
		Email:      user.Email,
		Country:    user.Country,
		ConsoleURL: c.Request().Header.Get("Origin"),
		YearsValid: 1,
	}

	data, err := json.Marshal(certRequest)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
	}

	if err := h.NATSConnection.Publish("certificates.new", data); err != nil {
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

	return h.ListUsers(c, "Email has been confirmed", "")
}

func (h *Handler) AskForConfirmation(c echo.Context) error {
	uid := c.Param("uid")
	user, err := h.Model.GetUserById(uid)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := sendConfirmationEmail(h, c, user); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return h.ListUsers(c, "A new confirmation email has been sent to "+user.Email, "")
}

func (h *Handler) EditUser(c echo.Context) error {
	uid := c.Param("uid")
	user, err := h.Model.GetUserById(uid)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if c.Request().Method == "POST" {
		if err := h.Model.UpdateUser(uid, c.FormValue("name"), c.FormValue("email"), c.FormValue("phone"), c.FormValue("country")); err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}

		return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "users.edit.success")))
	}

	defaultCountry, err := h.Model.GetDefaultCountry()
	if err != nil {
		return err
	}

	return RenderView(c, admin_views.UsersIndex(" | Users", admin_views.EditUser(c, user, defaultCountry)))
}

func sendConfirmationEmail(h *Handler, c echo.Context, user *openuem_ent.User) error {
	token, err := h.generateConfirmEmailToken(user.ID)
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
		return fmt.Errorf(i18n.T(c.Request().Context(), "nats.not_connected"))
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

		if len(record) != 5 {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.import_error_wrong_format", index), false))
		}

		if record[0] == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.import_required", "userid", index), false))
		}

		if record[2] == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.import_required", "email", index), false))
		}

		if record[2] != "" {
			if errs := validate.Var(record[2], "email"); errs != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.import_wrong_email", index), false))
			}
		}

		if record[3] != "" {
			if errs := validate.Var(strings.ToUpper(record[3]), "iso3166_1_alpha2"); errs != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.import_wrong_country", index, record[3]), false))
			}
		}

		index++

		user := openuem_ent.User{}
		user.ID = record[0]
		user.Name = record[1]
		user.Email = record[2]
		user.Country = strings.ToUpper(record[3])
		user.Phone = record[4]
		user.CertClearPassword = pkcs12.DefaultPassword

		err = h.Model.AddUser(user.ID, user.Name, user.Email, user.Phone, user.Country)
		if err != nil {
			// TODO manage duplicate key error
			errors = append(errors, err.Error())
			continue
		}

		if err := h.SendCertificateRequestToNATS(c, &user); err != nil {
			errors = append(errors, err.Error())
		}

	}

	if len(errors) > 0 {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.import_wrong_users", strings.Join(errors, ",")), false))
	}

	return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "users.import_success")))
}
