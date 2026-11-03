package handlers

import (
	"github.com/labstack/echo/v4"
)

func (h *Handler) Register(e *echo.Echo) {
	e.GET("/auth", h.Auth)
	e.GET("/agent-files/:profile/:hash/:name", h.AgentFile)
}
