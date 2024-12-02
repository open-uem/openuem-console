package router

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	session "github.com/canidam/echo-scs-session"
	"github.com/doncicuto/openuem-console/internal/controllers/router/middleware"
	"github.com/doncicuto/openuem-console/internal/controllers/sessions"
	"github.com/doncicuto/openuem-console/internal/views"
	"github.com/doncicuto/openuem-console/internal/views/locales"
	"github.com/doncicuto/openuem_utils"
	"github.com/invopop/ctxi18n"
	"github.com/labstack/echo/v4"
	mw "github.com/labstack/echo/v4/middleware"
)

func New(s *sessions.SessionManager, server, port, maxUploadSize string) *echo.Echo {

	e := echo.New()

	cwd, err := openuem_utils.GetWd()
	if err != nil {
		log.Fatalf("[FATAL]: could not get working directory: %v", err)
	}

	// Static Assets
	assetsPath := staticAssets(e, cwd)

	// Favicon
	faviconHandler(e, assetsPath)

	// Add i18n middleware
	if err := ctxi18n.LoadWithDefault(locales.Content, "en"); err != nil {
		log.Fatalf("[FATAL]: could not load translations: %v", err)
	}
	e.Use(middleware.GetLocale)

	// Limit uploads
	e.Use(mw.BodyLimit(maxUploadSize))

	// Add CORS middleware
	e.Use(mw.CORSWithConfig(mw.CORSConfig{
		AllowOrigins: []string{fmt.Sprintf("https://%s:%s", server, port)},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	// Add sessions middleware
	e.Use(session.LoadAndSave(s.Manager))

	// Custom HTTP Error Handler
	e.HTTPErrorHandler = customHTTPErrorHandler

	// Debug - enable logger
	// e.Use(mw.Logger())

	return e
}

func faviconHandler(e *echo.Echo, assetsPath string) string {

	// TODO - Replace with a better cache approach like immutable
	// nosniff + no-cache
	customHeaderMw := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set(echo.HeaderXContentTypeOptions, "nosniff")
			c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
			return next(c)
		}
	}

	e.Add(http.MethodGet, "/favicon.ico", echo.StaticFileHandler("favicon.ico", os.DirFS(assetsPath)), customHeaderMw)

	return assetsPath
}

func staticAssets(e *echo.Echo, cwd string) string {
	var assetsPath string

	// Static assets + Headers (Ref: https://github.com/labstack/echo/issues/1902#issuecomment-2435145166)
	if strings.HasSuffix(cwd, "tmp") {
		// DEVEL
		assetsPath = filepath.Join(filepath.Dir(cwd), "assets")
	} else {
		assetsPath = filepath.Join(cwd, "assets")
	}

	// TODO Etag middleware so no-cache can make use of no-cache
	// Ref: https://github.com/pablor21/echo-etag
	// There's an issue when using Caddy as reverse proxy ("unexpected EOF")
	// Disable until we find if the issue is with the middleware or with Caddy
	// e.Use(etag.Etag())

	// nosniff + no-cache
	customHeaderMw := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set(echo.HeaderXContentTypeOptions, "nosniff")
			c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
			return next(c)
		}
	}

	e.Add(http.MethodGet, "/assets*", echo.StaticDirectoryHandler(os.DirFS(assetsPath), false), customHeaderMw)

	return assetsPath
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
