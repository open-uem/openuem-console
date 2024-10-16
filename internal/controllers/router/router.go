package router

import (
	"log"
	"net/http"
	"strconv"

	session "github.com/canidam/echo-scs-session"
	"github.com/doncicuto/openuem-console/internal/controllers/router/middleware"
	"github.com/doncicuto/openuem-console/internal/controllers/sessions"
	"github.com/doncicuto/openuem-console/internal/views"
	"github.com/doncicuto/openuem-console/internal/views/locales"
	"github.com/invopop/ctxi18n"
	"github.com/labstack/echo/v4"
	mw "github.com/labstack/echo/v4/middleware"
)

func New(s *sessions.SessionManager) *echo.Echo {
	e := echo.New()

	// Static assets
	e.Static("/static", "assets")
	e.File("/favicon.ico", "assets/favicon.ico")

	// Add i18n middleware
	if err := ctxi18n.LoadWithDefault(locales.Content, "en"); err != nil {
		log.Fatalf("could not load translations: %v", err)
	}
	e.Use(middleware.GetLocale)

	// Limit uploads
	// TODO - This should be a setting!
	e.Use(mw.BodyLimit("512M"))

	// Add CORS middleware
	e.Use(mw.CORSWithConfig(mw.CORSConfig{
		AllowOrigins: []string{"https://localhost:1323"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	// Add sessions middleware
	e.Use(session.LoadAndSave(s.Manager))

	// Custom HTTP Error Handler
	e.HTTPErrorHandler = customHTTPErrorHandler

	return e
}

func customHTTPErrorHandler(err error, c echo.Context) {
	if he, ok := err.(*echo.HTTPError); ok {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
		switch he.Code {
		case http.StatusNotFound:
			if err := views.ErrorPage("404", "Page Not Found").Render(c.Request().Context(), c.Response().Writer); err != nil {
				c.Logger().Error(err)
			}
		case http.StatusInternalServerError:
			message := "Internal server error"
			if he.Message != nil {
				message = he.Message.(string)
			}

			if err := views.ErrorPage("503", message).Render(c.Request().Context(), c.Response().Writer); err != nil {
				c.Logger().Error(err)
			}
		case http.StatusUnauthorized:
			message := "Unauthorized Access"
			if he.Message != nil {
				message = he.Message.(string)
			}

			if err := views.ErrorPage("401", message).Render(c.Request().Context(), c.Response().Writer); err != nil {
				c.Logger().Error(err)
			}
		case http.StatusMethodNotAllowed:
			if err := views.ErrorPage("405", "Method Not Allowed").Render(c.Request().Context(), c.Response().Writer); err != nil {
				c.Logger().Error(err)
			}
		case http.StatusRequestEntityTooLarge:
			if err := views.ErrorPage("413", "Request Entity Too Large").Render(c.Request().Context(), c.Response().Writer); err != nil {
				c.Logger().Error(err)
			}
		default:
			if err := views.ErrorPage(strconv.Itoa(he.Code), "Error found").Render(c.Request().Context(), c.Response().Writer); err != nil {
				c.Logger().Error(err)
			}
		}
	} else {
		c.Logger().Error(err)
	}
}
