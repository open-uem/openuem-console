package handlers

import (
	"github.com/doncicuto/openuem-console/internal/views/deploy_views"
	"github.com/labstack/echo/v4"
)

func (h *Handler) Deploy(c echo.Context) error {
	return renderView(c, deploy_views.DeployIndex("| Deploy", deploy_views.Deploy()))
}
