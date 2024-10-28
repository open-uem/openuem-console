package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/doncicuto/openuem-console/internal/views/admin_views"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_nats"
	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/ocsp"
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

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	p.NItems, err = h.Model.CountAllUsers()
	if err != nil {
		successMessage = ""
		errMessage = err.Error()
	}

	users, err := h.Model.GetUsersByPage(p)
	if err != nil {
		successMessage = ""
		errMessage = err.Error()
	}

	return renderView(c, admin_views.UsersIndex(" | Users", admin_views.Users(c, users, p, successMessage, errMessage)))
}

func (h *Handler) NewUser(c echo.Context) error {
	return renderView(c, admin_views.UsersIndex(" | Users", admin_views.NewUser(c)))
}

func (h *Handler) AddUser(c echo.Context) error {
	u := NewUser{}
	successMessage := ""
	errMessage := ""

	decoder := form.NewDecoder()
	if err := c.Request().ParseForm(); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}
	err := decoder.Decode(&u, c.Request().Form)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(u); err != nil {
		// TODO Try to translate and create a nice error message
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	err = h.Model.AddUser(u.UID, u.Name, u.Email, u.Phone, u.Country)
	if err != nil {
		// TODO manage duplicate key error
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	addedUser, err := h.Model.GetUserById(u.UID)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := sendConfirmationEmail(h, c, addedUser); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	successMessage = i18n.T(c.Request().Context(), "new.user.success")
	return h.ListUsers(c, successMessage, errMessage)
}

func (h *Handler) RequestUserCertificate(c echo.Context) error {

	uid := c.Param("uid")

	user, err := h.Model.GetUserById(uid)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// TODO many of the following values should be set as settings in database and not hardcoded
	certRequest := openuem_nats.CertificateRequest{
		Username:   user.ID,
		FullName:   user.Name,
		Email:      user.Email,
		Country:    user.Country,
		Password:   user.CertClearPassword,
		YearsValid: 1,
	}

	data, err := json.Marshal(certRequest)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := h.NATSConnection.Publish("certificates.new", data); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	successMessage := i18n.T(c.Request().Context(), "users.certificate_requested")
	return h.ListUsers(c, successMessage, "")
}

func (h *Handler) DeleteUser(c echo.Context) error {
	uid := c.Param("uid")
	_, err := h.Model.GetUserById(uid)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// Delete user
	if err := h.Model.DeleteUser(uid); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// Revoke certificate
	cert, err := h.Model.GetCertificateByUID(uid)
	if err != nil {
		if !openuem_ent.IsNotFound(err) {
			return renderError(c, partials.ErrorMessage(err.Error(), false))
		}
		successMessage := i18n.T(c.Request().Context(), "users.deleted")
		return h.ListUsers(c, successMessage, "")
	}

	if err := h.Model.RevokeCertificate(cert, "user has been deleted", ocsp.CessationOfOperation); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// Delete certificate information
	if err := h.Model.DeleteCertificate(cert.ID); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	successMessage := i18n.T(c.Request().Context(), "users.deleted")
	return h.ListUsers(c, successMessage, "")
}

func (h *Handler) RenewUserCertificate(c echo.Context) error {
	uid := c.Param("uid")
	user, err := h.Model.GetUserById(uid)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// First revoke certificate
	cert, err := h.Model.GetCertificateByUID(uid)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := h.Model.RevokeCertificate(cert, "a new certificate has been requested", ocsp.CessationOfOperation); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// Now delete certificate
	if err := h.Model.DeleteCertificate(cert.ID); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// Now request a new certificate
	certRequest := openuem_nats.CertificateRequest{
		Username:   user.ID,
		FullName:   user.Name,
		Email:      user.Email,
		Country:    user.Country,
		YearsValid: 1,
	}

	data, err := json.Marshal(certRequest)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := h.NATSConnection.Publish("certificates.new", data); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	successMessage := i18n.T(c.Request().Context(), "users.certificate_requested")
	return h.ListUsers(c, successMessage, "")
}

func (h *Handler) SetEmailConfirmed(c echo.Context) error {
	uid := c.Param("uid")
	exists, err := h.Model.UserExists(uid)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if !exists {
		return renderError(c, partials.ErrorMessage("user doesn't exist", false))
	}

	err = h.Model.Client.User.UpdateOneID(uid).SetEmailVerified(true).SetRegister("users.review_request").Exec(context.Background())
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return h.ListUsers(c, "Email has been confirmed", "")
}

func (h *Handler) AskForConfirmation(c echo.Context) error {
	uid := c.Param("uid")
	user, err := h.Model.GetUserById(uid)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := sendConfirmationEmail(h, c, user); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return h.ListUsers(c, "A new confirmation email has been sent to "+user.Email, "")
}

func (h *Handler) EditUser(c echo.Context) error {
	uid := c.Param("uid")
	user, err := h.Model.GetUserById(uid)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if c.Request().Method == "POST" {
		if err := h.Model.UpdateUser(uid, c.FormValue("name"), c.FormValue("email"), c.FormValue("phone"), c.FormValue("country")); err != nil {
			return renderError(c, partials.ErrorMessage(err.Error(), false))
		}

		return renderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "users.edit.success")))
	}

	return renderView(c, admin_views.UsersIndex(" | Users", admin_views.EditUser(c, user)))
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

	if err := h.NATSConnection.Publish("notification.confirm_email", data); err != nil {
		return err
	}

	return nil
}
