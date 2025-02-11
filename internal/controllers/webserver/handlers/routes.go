package handlers

import (
	"github.com/labstack/echo/v4"
)

func (h *Handler) Register(e *echo.Echo) {
	e.GET("/", h.Dashboard, h.IsAuthenticated)

	e.GET("/auth", h.Auth)
	e.GET("/auth/confirm/:token", h.ConfirmEmail)

	e.GET("/agents", func(c echo.Context) error { return h.ListAgents(c, "", "", false) }, h.IsAuthenticated)
	e.POST("/agents", func(c echo.Context) error { return h.ListAgents(c, "", "", false) }, h.IsAuthenticated)
	e.DELETE("/agents", func(c echo.Context) error { return h.ListAgents(c, "", "", false) }, h.IsAuthenticated)
	e.GET("/agents/admit", h.AgentsAdmit, h.IsAuthenticated)
	e.POST("/agents/admit", h.AgentsAdmit, h.IsAuthenticated)
	e.GET("/agents/enable", h.AgentsEnable, h.IsAuthenticated)
	e.POST("/agents/enable", h.AgentsEnable, h.IsAuthenticated)
	e.GET("/agents/disable", h.AgentsDisable, h.IsAuthenticated)
	e.POST("/agents/disable", h.AgentsDisable, h.IsAuthenticated)
	e.GET("/agents/:uuid/delete", h.AgentDelete, h.IsAuthenticated)
	e.GET("/agents/:uuid/disable", h.AgentDisable, h.IsAuthenticated)
	e.GET("/agents/:uuid/admit", h.AgentAdmit, h.IsAuthenticated)
	e.POST("/agents/:uuid/enabled", h.AgentEnable, h.IsAuthenticated)
	e.POST("/agents/:uuid/forcereport", h.AgentForceRun, h.IsAuthenticated)
	e.POST("/agents/:uuid/disable", h.AgentConfirmDisable, h.IsAuthenticated)
	e.POST("/agents/:uuid/admit", func(c echo.Context) error { return h.AgentConfirmAdmission(c, false) }, h.IsAuthenticated)
	e.GET("/agents/:uuid/startvnc", h.AgentStartVNC, h.IsAuthenticated)
	e.POST("/agents/:uuid/startvnc", h.AgentStartVNC, h.IsAuthenticated)
	e.POST("/agents/:uuid/stopvnc", h.AgentStopVNC, h.IsAuthenticated)
	e.POST("/agents/:uuid/forcerestart", h.AgentForceRestart, h.IsAuthenticated)
	e.POST("/agents/:uuid/regeneratecerts", func(c echo.Context) error { return h.AgentConfirmAdmission(c, true) }, h.IsAuthenticated)
	e.DELETE("/agents/:uuid", h.AgentConfirmDelete, h.IsAuthenticated)

	e.GET("/admin", func(c echo.Context) error { return h.ListUsers(c, "", "") }, h.IsAuthenticated)
	e.POST("/admin", func(c echo.Context) error { return h.ListUsers(c, "", "") }, h.IsAuthenticated)
	e.GET("/admin/users", func(c echo.Context) error { return h.ListUsers(c, "", "") }, h.IsAuthenticated)
	e.POST("/admin/users", func(c echo.Context) error { return h.ListUsers(c, "", "") }, h.IsAuthenticated)
	e.GET("/admin/users/new", h.NewUser, h.IsAuthenticated)
	e.POST("/admin/users/import", h.ImportUsers, h.IsAuthenticated)
	e.GET("/admin/users/:uid/profile", h.EditUser, h.IsAuthenticated)
	e.POST("/admin/users/:uid/profile", h.EditUser, h.IsAuthenticated)
	e.POST("/admin/users/:uid/certificate", h.RequestUserCertificate, h.IsAuthenticated)
	e.POST("/admin/users/:uid/renewcertificate", h.RenewUserCertificate, h.IsAuthenticated)
	e.POST("/admin/users/new", h.AddUser, h.IsAuthenticated)
	e.POST("/admin/users/:uid/askconfirm", h.AskForConfirmation, h.IsAuthenticated)
	e.POST("/admin/users/:uid/confirmemail", h.SetEmailConfirmed, h.IsAuthenticated)
	e.DELETE("/admin/users/:uid", h.DeleteUser, h.IsAuthenticated)
	e.GET("/admin/sessions", func(c echo.Context) error { successMessage := ""; return h.ListSessions(c, successMessage) }, h.IsAuthenticated)
	e.GET("/admin/sessions/:token/delete", h.SessionDelete)
	e.DELETE("/admin/sessions/:token", h.SessionConfirmDelete, h.IsAuthenticated)
	e.GET("/admin/tags", h.TagManager, h.IsAuthenticated)
	e.POST("/admin/tags", h.TagManager, h.IsAuthenticated)
	e.DELETE("/admin/tags", h.TagManager, h.IsAuthenticated)
	e.GET("/admin/metadata", h.OrgMetadataManager, h.IsAuthenticated)
	e.POST("/admin/metadata", h.OrgMetadataManager, h.IsAuthenticated)
	e.DELETE("/admin/metadata", h.OrgMetadataManager, h.IsAuthenticated)
	e.GET("/admin/smtp", h.SMTPSettings, h.IsAuthenticated)
	e.POST("/admin/smtp", h.SMTPSettings, h.IsAuthenticated)
	e.POST("/admin/smtp/test", h.TestSMTPSettings, h.IsAuthenticated)
	e.GET("/admin/settings", h.GeneralSettings, h.IsAuthenticated)
	e.POST("/admin/settings", h.GeneralSettings, h.IsAuthenticated)
	e.GET("/admin/update-agents", h.UpdateAgents, h.IsAuthenticated)
	e.POST("/admin/update-agents", h.UpdateAgents, h.IsAuthenticated)
	e.DELETE("/admin/update-agents", h.UpdateAgents, h.IsAuthenticated)
	e.POST("/admin/update-agents/confirm", h.UpdateAgentsConfirm, h.IsAuthenticated)
	e.GET("/admin/certificates", h.ListCertificates, h.IsAuthenticated)
	e.POST("/admin/certificates", h.CertificateConfirmRevocation, h.IsAuthenticated)
	e.DELETE("/admin/certificates", h.RevocateCertificate, h.IsAuthenticated)
	e.GET("/admin/restore", h.Restore, h.IsAuthenticated)
	e.GET("/admin/restore-messenger", h.RestoreMessenger, h.IsAuthenticated)
	e.POST("/admin/restore-messenger", h.RestoreMessenger, h.IsAuthenticated)
	e.GET("/admin/restore-updater", h.RestoreUpdater, h.IsAuthenticated)
	e.POST("/admin/restore-updater", h.RestoreUpdater, h.IsAuthenticated)
	e.GET("/admin/restore-agents", h.RestoreAgents, h.IsAuthenticated)
	e.POST("/admin/restore-agents", h.RestoreAgents, h.IsAuthenticated)
	e.GET("/admin/restore-database", h.RestoreDatabase, h.IsAuthenticated)
	e.POST("/admin/restore-database", h.RestoreDatabase, h.IsAuthenticated)
	e.GET("/admin/update-servers", h.UpdateServers, h.IsAuthenticated)
	e.POST("/admin/update-servers", h.UpdateServers, h.IsAuthenticated)
	e.DELETE("/admin/update-servers/:serverId", h.UpdateServers, h.IsAuthenticated)
	e.POST("/admin/update-servers/confirm", h.UpdateServersConfirm, h.IsAuthenticated)
	e.POST("/admin/confirm-delete-server/:serverId", h.DeleteServerConfirm, h.IsAuthenticated)

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

	e.GET("/computers", func(c echo.Context) error { return h.ComputersList(c, "") }, h.IsAuthenticated)
	e.POST("/computers", func(c echo.Context) error { return h.ComputersList(c, "") }, h.IsAuthenticated)
	e.DELETE("/computers", func(c echo.Context) error { return h.ComputersList(c, "") }, h.IsAuthenticated)
	e.GET("/computers/:uuid", h.Computer, h.IsAuthenticated)
	e.DELETE("/computers/:uuid", h.ComputerConfirmDelete, h.IsAuthenticated)
	e.GET("/computers/:uuid/software", h.Apps, h.IsAuthenticated)
	e.POST("/computers/:uuid/software", h.Apps, h.IsAuthenticated)
	e.GET("/computers/:uuid/hardware", h.Computer, h.IsAuthenticated)
	e.GET("/computers/:uuid/logical-disks", h.LogicalDisks, h.IsAuthenticated)
	e.POST("/computers/:uuid/logical-disks", h.BrowseLogicalDisk, h.IsAuthenticated)
	e.POST("/computers/:uuid/logical-disks/file", h.UploadFile, h.IsAuthenticated)
	e.PUT("/computers/:uuid/logical-disks/file", h.RenameItem, h.IsAuthenticated)
	e.POST("/computers/:uuid/logical-disks/downloadfile", h.DownloadFile, h.IsAuthenticated)
	e.POST("/computers/:uuid/logical-disks/downloadfolder", h.DownloadFolderAsZIP, h.IsAuthenticated)
	e.POST("/computers/:uuid/logical-disks/downloadmany", h.DownloadManyAsZIP, h.IsAuthenticated)
	e.DELETE("/computers/:uuid/logical-disks/file", h.DeleteItem, h.IsAuthenticated)
	e.POST("/computers/:uuid/logical-disks/folder", h.NewFolder, h.IsAuthenticated)
	e.PUT("/computers/:uuid/logical-disks/folder", h.RenameItem, h.IsAuthenticated)
	e.DELETE("/computers/:uuid/logical-disks/folder", h.DeleteItem, h.IsAuthenticated)
	e.DELETE("/computers/:uuid/logical-disks/many", h.DeleteMany, h.IsAuthenticated)
	e.GET("/computers/:uuid/monitors", h.Monitors, h.IsAuthenticated)
	e.GET("/computers/:uuid/network-adapters", h.NetworkAdapters, h.IsAuthenticated)
	e.GET("/computers/:uuid/os", h.OperatingSystem, h.IsAuthenticated)
	e.GET("/computers/:uuid/printers", h.Printers, h.IsAuthenticated)
	e.GET("/computers/:uuid/shares", h.Shares, h.IsAuthenticated)
	e.GET("/computers/:uuid/remote-assistance", h.RemoteAssistance, h.IsAuthenticated)
	e.GET("/computers/:uuid/power", h.PowerManagement, h.IsAuthenticated)
	e.POST("/computers/:uuid/power/:action", h.PowerManagement, h.IsAuthenticated)
	e.GET("/computers/:uuid/notes", h.Notes, h.IsAuthenticated)
	e.POST("/computers/:uuid/notes", h.Notes, h.IsAuthenticated)
	e.GET("/computers/:uuid/deploy", func(c echo.Context) error { return h.ComputerDeploy(c, "") }, h.IsAuthenticated)
	e.POST("/computers/:uuid/deploy", func(c echo.Context) error { return h.ComputerDeploy(c, "") }, h.IsAuthenticated)
	e.GET("/computers/:uuid/deploy/searchinstall", func(c echo.Context) error { return h.ComputerDeploy(c, "") }, h.IsAuthenticated)
	e.POST("/computers/:uuid/deploy/searchinstall", h.ComputerDeploySearchPackagesInstall, h.IsAuthenticated)
	e.POST("/computers/:uuid/deploy/install", h.ComputerDeployInstall, h.IsAuthenticated)
	e.POST("/computers/:uuid/deploy/update", h.ComputerDeployUpdate, h.IsAuthenticated)
	e.POST("/computers/:uuid/deploy/uninstall", h.ComputerDeployUninstall, h.IsAuthenticated)
	e.GET("/computers/:uuid/metadata", h.ComputerMetadata, h.IsAuthenticated)
	e.POST("/computers/:uuid/metadata", h.ComputerMetadata, h.IsAuthenticated)
	e.DELETE("/computers/:uuid/metadata", h.ComputerMetadata, h.IsAuthenticated)

	e.POST("/logout", h.Logout, h.IsAuthenticated)

	e.GET("/network-printers", h.NetworkPrinters, h.IsAuthenticated)

	e.GET("/remote-workers", h.RemoteWorkers, h.IsAuthenticated)

	e.GET("/register", h.SignIn)
	e.POST("/register", h.SendRegister)
	e.GET("/user/:uid/exists", h.UIDExists)
	e.GET("/email/:email/exists", h.EmailExists)

	e.GET("/reports", func(c echo.Context) error { return h.Reports(c, "") }, h.IsAuthenticated)
	e.POST("/reports", func(c echo.Context) error { return h.GenerateAgentsReport(c, "") }, h.IsAuthenticated)

	e.GET("/security", h.ListAntivirusStatus, h.IsAuthenticated)
	e.POST("/security", h.ListAntivirusStatus, h.IsAuthenticated)
	e.GET("/security/:uuid/updates", h.ListLatestUpdates, h.IsAuthenticated)
	e.POST("/security/:uuid/updates", h.ListLatestUpdates, h.IsAuthenticated)
	e.GET("/security/antivirus", h.ListAntivirusStatus, h.IsAuthenticated)
	e.POST("/security/antivirus", h.ListAntivirusStatus, h.IsAuthenticated)
	e.GET("/security/updates", h.ListSecurityUpdatesStatus, h.IsAuthenticated)
	e.POST("/security/updates", h.ListSecurityUpdatesStatus, h.IsAuthenticated)

	e.GET("/software", h.Software, h.IsAuthenticated)
	e.POST("/software", h.Software, h.IsAuthenticated)

	e.GET("/download/:filename", h.Download, h.IsAuthenticated)

	e.POST("/render-markdown", h.RenderMarkdown, h.IsAuthenticated)
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
