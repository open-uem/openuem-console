package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) Logout(c echo.Context) error {
	if err := h.SessionManager.Manager.Destroy(c.Request().Context()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return h.Login(c)
}
