package handlers

import (
	"github.com/doncicuto/openuem-console/internal/views/remote_workers_views"
	"github.com/labstack/echo/v4"
)

func (h *Handler) RemoteWorkers(c echo.Context) error {
	return RenderView(c, remote_workers_views.RemoteWorkersIndex("| Remote Workers", remote_workers_views.RemoteWorkers()))
}
