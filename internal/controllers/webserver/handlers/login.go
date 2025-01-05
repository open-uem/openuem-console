package handlers

import (
	"github.com/open-uem/openuem-console/internal/views/login_views"
	"github.com/labstack/echo/v4"
)

func (h *Handler) Login(c echo.Context) error {
	return RenderLogin(c, login_views.LoginIndex(login_views.Login()))
}
