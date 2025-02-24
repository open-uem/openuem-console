package handlers

import (
	"github.com/labstack/echo/v4"
	model "github.com/open-uem/openuem-console/internal/models/servers"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem-console/internal/views/remote_workers_views"
)

func (h *Handler) RemoteWorkers(c echo.Context) error {
	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, remote_workers_views.RemoteWorkersIndex("| Remote Workers", remote_workers_views.RemoteWorkers(h.SessionManager, h.Version, latestServerRelease.Version)))
}
