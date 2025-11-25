package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/views/register_views"
)

type MyCustomClaims struct {
	jwt.RegisteredClaims
}

func (h *Handler) Auth(c echo.Context) error {
	if h.ReverseProxyAuthPort != "" {
		u, err := url.Parse(c.Request().Referer())
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not parse url")
		}
		return c.Redirect(http.StatusFound, fmt.Sprintf("https://%s:%s/auth", u.Hostname(), h.ReverseProxyAuthPort))
	} else {
		return c.Redirect(http.StatusFound, fmt.Sprintf("https://%s:%s/auth", h.ServerName, h.AuthPort))
	}

}

func (h *Handler) ConfirmEmail(c echo.Context) error {
	tokenString := c.Param("token")

	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.JWTKey), nil
	})

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if claims, ok := token.Claims.(*MyCustomClaims); ok {
		if time.Now().After(claims.ExpiresAt.Time) {
			return echo.NewHTTPError(http.StatusBadRequest, "token has expired, please contact your administrator to request a new confirmation email")
		}

		user, err := h.Model.GetUserById(claims.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		if user.EmailVerified {
			return echo.NewHTTPError(http.StatusBadRequest, "you've already confirmed your email")
		}

		if err := h.Model.ConfirmEmail(user.ID); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		csrfToken, ok := c.Get("csrf").(string)
		if !ok || csrfToken == "" {
			return echo.NewHTTPError(http.StatusForbidden, i18n.T(c.Request().Context(), "authentication.csrf_token_not_found"))
		}

		return RenderView(c, register_views.RegisterIndex(register_views.EmailConfirmed(), csrfToken))

	} else {
		return echo.NewHTTPError(http.StatusBadRequest, "unknown claims type, cannot proceed")
	}
}
