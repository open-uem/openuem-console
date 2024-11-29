package handlers

import (
	"github.com/doncicuto/openuem-console/internal/views/admin_views"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
)

func (h *Handler) Restore(c echo.Context) error {
	return RenderView(c, admin_views.RestoreIndex("| Restore", admin_views.Restore(c, h.SessionManager, "")))
}

func (h *Handler) RestoreMessenger(c echo.Context) error {
	if c.Request().Method == "POST" {
		if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
		}

		if err := h.NATSConnection.Publish("agent.rollback.messenger", nil); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.no_responders"), false))
		}
		return RenderView(c, admin_views.RestoreIndex("| Restore", admin_views.Restore(c, h.SessionManager, i18n.T(c.Request().Context(), "restore.restore_requested"))))
	}
	return RenderConfirm(c, partials.Confirm(i18n.T(c.Request().Context(), "restore.confirm_restore"), "/admin/restore-messenger", "/admin/restore", true))
}

func (h *Handler) RestoreUpdater(c echo.Context) error {
	if c.Request().Method == "POST" {
		if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
		}

		if err := h.NATSConnection.Publish("agent.rollback.updater", nil); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.no_responders"), false))
		}
		return RenderView(c, admin_views.RestoreIndex("| Restore", admin_views.Restore(c, h.SessionManager, i18n.T(c.Request().Context(), "restore.restore_requested"))))
	}
	return RenderConfirm(c, partials.Confirm(i18n.T(c.Request().Context(), "restore.confirm_restore"), "/admin/restore-updater", "/admin/restore", true))
}

func (h *Handler) RestoreAgents(c echo.Context) error {
	if c.Request().Method == "POST" {
		if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
		}

		if err := h.NATSConnection.Publish("agent.rollback", nil); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.no_responders"), false))
		}
		return RenderView(c, admin_views.RestoreIndex("| Restore", admin_views.Restore(c, h.SessionManager, i18n.T(c.Request().Context(), "restore.restore_requested"))))
	}
	return RenderConfirm(c, partials.Confirm(i18n.T(c.Request().Context(), "restore.confirm_restore"), "/admin/restore-agents", "/admin/restore", true))
}

func (h *Handler) RestoreDatabase(c echo.Context) error {
	if c.Request().Method == "POST" {
		_, err := h.Model.DeleteAllAgents()
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.no_responders"), false))
		}
		return RenderView(c, admin_views.RestoreIndex("| Delete", admin_views.Restore(c, h.SessionManager, i18n.T(c.Request().Context(), "restore.delete_database_requested"))))
	}
	return RenderConfirm(c, partials.Confirm(i18n.T(c.Request().Context(), "restore.confirm_delete_database"), "/admin/restore-database", "/admin/restore", true))
}
