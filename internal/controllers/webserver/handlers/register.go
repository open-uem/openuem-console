package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem-console/internal/views/register_views"
)

type RegisterRequest struct {
	UID      string `form:"uid" validate:"required"`
	Name     string `form:"name" validate:"required"`
	Email    string `form:"email" validate:"required,email"`
	Phone    string `form:"phone" validate:"required,e164"`
	Country  string `form:"country" validate:"required,iso3166_1_alpha2"`
	Password string `form:"password"`
	AuthType string `form:"auth-type" validate:"required"`
}

func (h *Handler) SignIn(c echo.Context) error {
	validations := register_views.RegisterValidations{}

	defaultCountry, err := h.Model.GetDefaultCountry()
	if err != nil {
		return err
	}

	settings, err := h.Model.GetAuthenticationSettings()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, i18n.T(c.Request().Context(), "authentication.could_not_get_settings"))
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok || csrfToken == "" {
		return echo.NewHTTPError(http.StatusForbidden, i18n.T(c.Request().Context(), "authentication.csrf_token_not_found"))
	}

	return RenderView(c, register_views.RegisterIndex(register_views.Register(c, register_views.RegisterValues{}, validations, defaultCountry, settings), csrfToken))
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
		values.AuthType = r.AuthType

		errs := validate.Var(r.UID, "required")
		if errs != nil {
			validations.UIDRequired = true
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

		errs = validate.Var(r.AuthType, "required")
		if errs != nil {
			validations.AuthTypeRequired = true
		}

		if r.AuthType == "certificate" {
			errs = validate.Var(r.Password, "required")
			if errs != nil {
				validations.PasswordRequired = true
			}
		}

		settings, err := h.Model.GetAuthenticationSettings()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, i18n.T(c.Request().Context(), "authentication.could_not_get_settings"))
		}

		csrfToken, ok := c.Get("csrf").(string)
		if !ok || csrfToken == "" {
			return echo.NewHTTPError(http.StatusForbidden, i18n.T(c.Request().Context(), "authentication.csrf_token_not_found"))
		}

		return RenderView(c, register_views.RegisterIndex(register_views.Register(c, values, validations, defaultCountry, settings), csrfToken))
	}

	if err := h.Model.RegisterUser(r.UID, r.Name, r.Email, r.Phone, r.Country, r.Password, r.AuthType); err != nil {
		log.Printf("[ERROR]: could not process the register request, reason: %v", err)
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok || csrfToken == "" {
		return echo.NewHTTPError(http.StatusForbidden, i18n.T(c.Request().Context(), "authentication.csrf_token_not_found"))
	}

	return RenderView(c, register_views.RegisterIndex(register_views.RegisterSuccesful(), csrfToken))
}

func (h *Handler) generateEmailToken(uid string, subject string, hours int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(hours) * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "OpenUEM",
		Subject:   subject,
		ID:        uid,
	})

	return token.SignedString([]byte(h.JWTKey))
}
