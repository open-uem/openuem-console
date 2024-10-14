package handlers

import (
	"github.com/labstack/echo/v4"
)

func (h *Handler) Register(e *echo.Echo) {
	e.GET("/", h.Dashboard, h.IsAuthenticated)

	e.GET("/auth", h.Auth)
	e.GET("/auth/confirm/:token", h.ConfirmEmail)

	e.GET("/agents", func(c echo.Context) error { return h.ListAgents(c, "", "") }, h.IsAuthenticated)
	e.POST("/agents", func(c echo.Context) error { return h.ListAgents(c, "", "") }, h.IsAuthenticated)
	e.GET("/agents/:uuid/delete", h.AgentDelete, h.IsAuthenticated)
	e.GET("/agents/:uuid/disable", h.AgentDisable, h.IsAuthenticated)
	e.POST("/agents/:uuid/enabled", h.AgentEnable, h.IsAuthenticated)
	e.POST("/agents/:uuid/forcereport", h.AgentForceRun, h.IsAuthenticated)
	e.POST("/agents/:uuid/disable", h.AgentConfirmDisable, h.IsAuthenticated)
	e.POST("/agents/:uuid/startvnc", h.AgentStartVNC, h.IsAuthenticated)
	e.POST("/agents/:uuid/stopvnc", h.AgentStopVNC, h.IsAuthenticated)
	e.DELETE("/agents/:uuid", h.AgentConfirmDelete, h.IsAuthenticated)

	e.GET("/config", func(c echo.Context) error { return h.ListUsers(c, "", "") }, h.IsAuthenticated)
	e.GET("/config/users", func(c echo.Context) error { return h.ListUsers(c, "", "") }, h.IsAuthenticated)
	e.GET("/config/users/new", h.NewUser, h.IsAuthenticated)
	e.GET("/config/users/:uid/certificate", h.RequestUserCertificate, h.IsAuthenticated)
	e.GET("/config/users/:uid/renewcertificate", h.RenewUserCertificate, h.IsAuthenticated)
	e.GET("/config/sessions", func(c echo.Context) error { successMessage := ""; return h.ListSessions(c, successMessage) }, h.IsAuthenticated)
	e.POST("/config/users/new", h.AddUser, h.IsAuthenticated)
	e.POST("/config/users/:uid/askconfirm", h.AskForConfirmation, h.IsAuthenticated)
	e.POST("/config/users/:uid/confirmemail", h.SetEmailConfirmed, h.IsAuthenticated)
	e.DELETE("/config/users/:uid", h.DeleteUser, h.IsAuthenticated)

	e.GET("/config/sessions/:token/delete", h.SessionDelete)
	e.DELETE("/config/sessions/:token", h.SessionConfirmDelete, h.IsAuthenticated)

	e.GET("/dashboard", h.Dashboard, h.IsAuthenticated)

	e.GET("/deploy", h.DeployInstall, h.IsAuthenticated)
	e.GET("/deploy/install", h.DeployInstall, h.IsAuthenticated)
	e.GET("/deploy/uninstall", h.DeployUninstall, h.IsAuthenticated)
	e.GET("/deploy/searchinstall", h.DeployInstall, h.IsAuthenticated)
	e.POST("/deploy/searchinstall", func(c echo.Context) error { return h.SearchPackagesAction(c, true) }, h.IsAuthenticated)
	e.GET("/deploy/searchuninstall", h.DeployUninstall, h.IsAuthenticated)
	e.POST("/deploy/searchuninstall", func(c echo.Context) error { return h.SearchPackagesAction(c, false) }, h.IsAuthenticated)
	e.GET("/deploy/selectpackagedeployment", h.SelectPackageDeployment, h.IsAuthenticated)
	e.POST("/deploy/selectpackagedeployment", h.DeployPackageToSelectedAgents, h.IsAuthenticated)

	e.GET("/desktops", h.Desktops, h.IsAuthenticated)
	e.POST("/desktops", h.Desktops, h.IsAuthenticated)
	e.GET("/desktops/:uuid", h.Computer, h.IsAuthenticated)
	e.DELETE("/desktops/:uuid", h.AgentConfirmDelete, h.IsAuthenticated)
	e.GET("/desktops/:uuid/software", h.Apps, h.IsAuthenticated)
	e.POST("/desktops/:uuid/software", h.Apps, h.IsAuthenticated)
	e.GET("/desktops/:uuid/hardware", h.Computer, h.IsAuthenticated)
	e.GET("/desktops/:uuid/logical-disks", h.LogicalDisks, h.IsAuthenticated)
	e.GET("/desktops/:uuid/monitors", h.Monitors, h.IsAuthenticated)
	e.GET("/desktops/:uuid/network-adapters", h.NetworkAdapters, h.IsAuthenticated)
	e.GET("/desktops/:uuid/os", h.OperatingSystem, h.IsAuthenticated)
	e.GET("/desktops/:uuid/printers", h.Printers, h.IsAuthenticated)
	e.GET("/desktops/:uuid/shares", h.Shares, h.IsAuthenticated)
	e.GET("/desktops/:uuid/remote-assistance", h.RemoteAssistance, h.IsAuthenticated)
	e.GET("/desktops/:uuid/wol", h.WakeOnLan, h.IsAuthenticated)
	e.POST("/desktops/:uuid/wol", h.WakeOnLan, h.IsAuthenticated)
	e.GET("/desktops/:uuid/deploy", h.DesktopDeploy, h.IsAuthenticated)
	e.POST("/desktops/:uuid/deploy", h.DesktopDeploy, h.IsAuthenticated)
	e.GET("/desktops/:uuid/deploy/searchinstall", h.DesktopDeploy, h.IsAuthenticated)
	e.POST("/desktops/:uuid/deploy/searchinstall", h.DesktopDeploySearchPackagesInstall, h.IsAuthenticated)
	e.POST("/desktops/:uuid/deploy/install", h.DesktopDeployInstall, h.IsAuthenticated)
	e.POST("/desktops/:uuid/deploy/update", h.DesktopDeployUpdate, h.IsAuthenticated)
	e.POST("/desktops/:uuid/deploy/uninstall", h.DesktopDeployUninstall, h.IsAuthenticated)

	e.POST("/logout", h.Logout, h.IsAuthenticated)

	e.GET("/network-printers", h.NetworkPrinters, h.IsAuthenticated)

	e.GET("/remote-workers", h.RemoteWorkers, h.IsAuthenticated)

	e.GET("/register", h.SignIn)
	e.POST("/register", h.SendRegister)
	e.GET("/user/:uid/exists", h.UIDExists)
	e.GET("/email/:email/exists", h.EmailExists)

	e.GET("/security", h.ListAntivirusStatus, h.IsAuthenticated)
	e.GET("/security/antivirus", h.ListAntivirusStatus, h.IsAuthenticated)
	e.GET("/security/updates", h.ListSecurityUpdatesStatus, h.IsAuthenticated)

	e.GET("/software", h.Software, h.IsAuthenticated)
	e.POST("/software", h.Software, h.IsAuthenticated)

}

func (h *Handler) IsAuthenticated(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Redirect to Login if user has no session
		if !h.SessionManager.Manager.Exists(c.Request().Context(), "uid") {
			return h.Login(c)
		}

		return next(c)
	}
}
