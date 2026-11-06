package handlers

import (
	"github.com/labstack/echo/v4"
)

func (h *Handler) Register(e *echo.Echo) {
	e.GET("/auth", h.Auth)
	e.GET("/agent-files/:profile/:hash/:name", h.AgentFile)

	// Private winget repository (WinGet REST source) for the agents
	winget := e.Group("/winget/:agent/:token", h.WinGetSourceAuth)
	winget.GET("/information", h.WinGetInformation)
	winget.POST("/manifestSearch", h.WinGetManifestSearch)
	winget.GET("/packageManifests/:id", h.WinGetPackageManifest)
	winget.GET("/installers/:id/:version/:file", h.WinGetInstaller)
}
