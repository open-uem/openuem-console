package handlers

import (
	"github.com/labstack/echo/v4"
)

func (h *Handler) Register(e *echo.Echo) {
	e.GET("/auth", h.Auth)
}

// RegisterWinGetSource adds the private winget repository (WinGet REST source) for the
// agents. It's served by the console's web server, not by this server: this server asks
// clients for a certificate, and winget, which can't present one, fails the request then
// (WinHTTP's ERROR_WINHTTP_CLIENT_AUTH_CERT_NEEDED).
func (h *Handler) RegisterWinGetSource(e *echo.Echo) {
	winget := e.Group("/winget/:agent/:token", h.WinGetSourceAuth)
	winget.GET("/information", h.WinGetInformation)
	winget.POST("/manifestSearch", h.WinGetManifestSearch)
	winget.GET("/packageManifests/:id", h.WinGetPackageManifest)
	winget.GET("/installers/:id/:version/:file", h.WinGetInstaller)
}
