package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/wingetrepo"
)

// tenantsContextKey keeps the organizations of the agent that uses the winget source
const tenantsContextKey = "winget-agent-tenants"

// WinGetSourceAuth checks the agent token of the winget source URL and that the agent is
// enabled. winget can't present client certificates, so each agent gets its own URL
// with a token derived from its id.
func (h *Handler) WinGetSourceAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if h.WinGetLibrary == nil {
			return wingetError(c, http.StatusNotFound)
		}

		agentID := c.Param("agent")
		if !wingetrepo.ValidAgentToken(h.WinGetKey, agentID, c.Param("token")) {
			return wingetError(c, http.StatusUnauthorized)
		}

		tenantIDs, ok, err := h.Model.EnabledAgentTenants(agentID)
		if err != nil {
			log.Printf("[ERROR]: could not check the agent %s of the winget source, reason: %v", agentID, err)
			return wingetError(c, http.StatusInternalServerError)
		}
		if !ok {
			return wingetError(c, http.StatusForbidden)
		}

		c.Set(tenantsContextKey, tenantIDs)
		return next(c)
	}
}

// wingetError answers with a JSON error. The console's error handler renders HTML pages,
// which winget can't read.
func wingetError(c echo.Context, code int) error {
	return c.JSON(code, wingetrepo.ErrorResponse{ErrorCode: code, ErrorMessage: http.StatusText(code)})
}

// visiblePackages returns the packages the agent's organizations can install
func (h *Handler) visiblePackages(c echo.Context) ([]wingetrepo.Package, error) {
	all, err := h.WinGetLibrary.List()
	if err != nil {
		return nil, err
	}

	tenantIDs, _ := c.Get(tenantsContextKey).([]int)
	visible := []wingetrepo.Package{}
	for _, p := range all {
		if p.VisibleTo(tenantIDs) {
			visible = append(visible, p)
		}
	}
	return visible, nil
}

// WinGetInformation describes the source (GET /information)
func (h *Handler) WinGetInformation(c echo.Context) error {
	return c.JSON(http.StatusOK, wingetrepo.NewInformation())
}

// WinGetManifestSearch searches packages (POST /manifestSearch). No match is answered
// with 204 No Content, as winget expects.
func (h *Handler) WinGetManifestSearch(c echo.Context) error {
	req := wingetrepo.SearchRequest{}
	if err := c.Bind(&req); err != nil {
		return wingetError(c, http.StatusBadRequest)
	}

	packages, err := h.visiblePackages(c)
	if err != nil {
		return wingetError(c, http.StatusInternalServerError)
	}

	result := wingetrepo.Search(packages, req)
	if len(result.Data) == 0 {
		return c.NoContent(http.StatusNoContent)
	}
	return c.JSON(http.StatusOK, result)
}

// WinGetPackageManifest describes the versions of a package
// (GET /packageManifests/{PackageIdentifier}?Version=...)
func (h *Handler) WinGetPackageManifest(c echo.Context) error {
	packages, err := h.visiblePackages(c)
	if err != nil {
		return wingetError(c, http.StatusInternalServerError)
	}

	version := c.QueryParam("Version")
	versions := []wingetrepo.Package{}
	for _, p := range packages {
		if p.PackageIdentifier == c.Param("id") && (version == "" || p.PackageVersion == version) {
			versions = append(versions, p)
		}
	}
	if len(versions) == 0 {
		return c.NoContent(http.StatusNoContent)
	}

	installersURL := "https://" + c.Request().Host + "/winget/" + c.Param("agent") + "/" + c.Param("token") + "/installers"
	return c.JSON(http.StatusOK, wingetrepo.BuildManifest(versions, installersURL))
}

// WinGetInstaller serves the installer of a package version the agent can install
func (h *Handler) WinGetInstaller(c echo.Context) error {
	path, p, err := h.WinGetLibrary.InstallerPath(c.Param("id"), c.Param("version"))
	if errors.Is(err, wingetrepo.ErrPackageNotFound) || (err == nil && p.InstallerFile != c.Param("file")) {
		return wingetError(c, http.StatusNotFound)
	}
	if err != nil {
		return wingetError(c, http.StatusInternalServerError)
	}

	tenantIDs, _ := c.Get(tenantsContextKey).([]int)
	if !p.VisibleTo(tenantIDs) {
		return wingetError(c, http.StatusNotFound)
	}
	return c.Attachment(path, p.InstallerFile)
}
