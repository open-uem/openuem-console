package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem-console/internal/views/register_views"
)

type RegisterRequest struct {
	UID      string `form:"uid" validate:"required"`
	Name     string `form:"name" validate:"required"`
	Email    string `form:"email" validate:"required,email"`
	Phone    string `form:"phone" validate:"required,e164"`
	Country  string `form:"country" validate:"required,iso3166_1_alpha2"`
	Password string `form:"password" validate:"required"`
}

func (h *Handler) SignIn(c echo.Context) error {
	validations := register_views.RegisterValidations{}

	defaultCountry, err := h.Model.GetDefaultCountry()
	if err != nil {
		return err
	}

	return RenderView(c, register_views.RegisterIndex(register_views.Register(c, register_views.RegisterValues{}, validations, defaultCountry)))
}

func (h *Handler) SendRegister(c echo.Context) error {
	defaultCountry, err := h.Model.GetDefaultCountry()
	if err != nil {
		return err
	}

	r := RegisterRequest{}
	decoder := form.NewDecoder()
	if err := c.Request().ParseForm(); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	err = decoder.Decode(&r, c.Request().Form)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(r); err != nil {
		validations := register_views.RegisterValidations{}
		values := register_views.RegisterValues{}
		values.UID = r.UID
		values.Name = r.Name
		values.Email = r.Email
		values.Phone = r.Phone
		values.Country = r.Country
		values.Password = r.Password

		errs := validate.Var(r.UID, "required")
		if errs != nil {
			validations.UIDRequired = true
		}

		exists, err := h.Model.UserExists(r.UID)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}

		if exists {
			validations.UIDExists = true
		}

		errs = validate.Var(r.Name, "required")
		if errs != nil {
			validations.NameRequired = true
		}

		errs = validate.Var(r.Email, "required")
		if errs != nil {
			validations.EmailRequired = true
		}

		errs = validate.Var(r.Email, "email")
		if errs != nil {
			validations.EmailInvalid = true
		}

		exists, err = h.Model.EmailExists(r.Email)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}
		if exists {
			validations.EmailExists = true
		}

		errs = validate.Var(r.Country, "required")
		if errs != nil {
			validations.CountryRequired = true
		}

		errs = validate.Var(strings.ToUpper(r.Country), "iso3166_1_alpha2")
		if errs != nil {
			validations.CountryInvalid = true
		}

		errs = validate.Var(r.Phone, "required")
		if errs != nil {
			validations.PhoneRequired = true
		}

		errs = validate.Var(r.Phone, "e164")
		if errs != nil {
			validations.PhoneInvalid = true
		}

		errs = validate.Var(r.Password, "required")
		if errs != nil {
			validations.PasswordRequired = true
		}

		return RenderView(c, register_views.RegisterIndex(register_views.Register(c, values, validations, defaultCountry)))
	}

	if err := h.Model.RegisterUser(r.UID, r.Name, r.Email, r.Phone, r.Country, r.Password); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	token, err := h.generateConfirmEmailToken(r.UID)
	if err != nil {
		// rollback register user
		if err := h.Model.DeleteUser(r.UID); err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	notification := openuem_nats.Notification{
		To:               r.Email,
		Subject:          "Please, confirm your email address",
		MessageTitle:     "OpenUEM | Verify your email address",
		MessageText:      "Please, confirm your email address so that it can be used to receive emails from OpenUEM",
		MessageGreeting:  fmt.Sprintf("Hi %s", r.Name),
		MessageAction:    "Confirm email",
		MessageActionURL: c.Request().Header.Get("Origin") + "/auth/confirm/" + token,
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
	}

	if err := h.NATSConnection.Publish("notification.confirm_email", data); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return RenderView(c, register_views.RegisterIndex(register_views.RegisterSuccesful()))
}

func (h *Handler) UIDExists(c echo.Context) error {
	uid := strings.TrimSpace(c.Param("uid"))
	exists, err := h.Model.UserExists(uid)
	if err != nil {
		return c.String(http.StatusInternalServerError, "error")
	}

	if exists {
		return c.String(http.StatusOK, "true")
	} else {
		return c.String(http.StatusOK, "false")
	}
}

func (h *Handler) EmailExists(c echo.Context) error {
	email := strings.TrimSpace(c.Param("email"))
	exists, err := h.Model.EmailExists(email)
	if err != nil {
		return c.String(http.StatusInternalServerError, "error")
	}

	if exists {
		return c.String(http.StatusOK, "true")
	} else {
		return c.String(http.StatusOK, "false")
	}
}

func (h *Handler) generateConfirmEmailToken(uid string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "OpenUEM",
		Subject:   "Email Confirmation",
		ID:        uid,
	})

	return token.SignedString([]byte(h.JWTKey))
}
