package handlers

import (
	"archive/zip"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/views/admin_views"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (h *Handler) ListEnrollmentTokens(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.invalid_tenant_id"), false))
	}

	tokens, err := h.Model.GetEnrollmentTokens(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	sites, err := h.Model.GetSites(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	agentsExists, err := h.Model.AgentsExists(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return RenderView(c, admin_views.EnrollmentTokensIndex(" | Enrollment",
		admin_views.EnrollmentTokens(c, tokens, sites, "", agentsExists, serversExists, commonInfo),
		commonInfo))
}

func (h *Handler) CreateEnrollmentToken(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.invalid_tenant_id"), true))
	}

	description := c.FormValue("description")
	tokenValue := uuid.New().String()

	maxUses := 0
	if v := c.FormValue("max_uses"); v != "" {
		maxUses, _ = strconv.Atoi(v)
	}

	var siteID *int
	if v := c.FormValue("site_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err == nil && id > 0 {
			siteID = &id
		}
	}

	var expiresAt *time.Time
	if v := c.FormValue("expires_at"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err == nil {
			expiresAt = &t
		}
	}

	_, err = h.Model.CreateEnrollmentToken(tenantID, siteID, description, tokenValue, maxUses, expiresAt)
	if err != nil {
		log.Printf("[ERROR]: could not create enrollment token: %v", err)
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return h.ListEnrollmentTokens(c)
}

func (h *Handler) DeleteEnrollmentToken(c echo.Context) error {
	tokenID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return RenderError(c, partials.ErrorMessage("Invalid token ID", true))
	}

	err = h.Model.DeleteEnrollmentToken(tokenID)
	if err != nil {
		log.Printf("[ERROR]: could not delete enrollment token: %v", err)
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return h.ListEnrollmentTokens(c)
}

func (h *Handler) ToggleEnrollmentToken(c echo.Context) error {
	tokenID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return RenderError(c, partials.ErrorMessage("Invalid token ID", true))
	}

	active := c.FormValue("active") == "true"

	err = h.Model.ToggleEnrollmentToken(tokenID, active)
	if err != nil {
		log.Printf("[ERROR]: could not toggle enrollment token: %v", err)
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return h.ListEnrollmentTokens(c)
}

// buildConfigZIP creates an in-memory ZIP with openuem.ini and all certificates.
func (h *Handler) buildConfigZIP(iniContent string) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Add openuem.ini
	fw, err := zw.Create("openuem.ini")
	if err != nil {
		return nil, fmt.Errorf("could not create ZIP entry: %w", err)
	}
	if _, err := fw.Write([]byte(iniContent)); err != nil {
		return nil, fmt.Errorf("could not write config: %w", err)
	}

	// Add certificate files
	certFiles := map[string]string{
		"certificates/ca.cer":    h.CACertPath,
		"certificates/agent.cer": h.AgentCertPath,
		"certificates/agent.key": h.AgentKeyPath,
		"certificates/sftp.cer":  h.SFTPCertPath,
	}

	for zipPath, filePath := range certFiles {
		if filePath == "" {
			continue
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("[WARN]: could not read %s: %v", filePath, err)
			continue
		}
		fw, err := zw.Create(zipPath)
		if err != nil {
			return nil, fmt.Errorf("could not create ZIP entry %s: %w", zipPath, err)
		}
		if _, err := fw.Write(data); err != nil {
			return nil, fmt.Errorf("could not write %s: %w", zipPath, err)
		}
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("could not finalize ZIP: %w", err)
	}

	return buf.Bytes(), nil
}

func (h *Handler) DownloadConfigZIP(c echo.Context) error {
	tokenID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return RenderError(c, partials.ErrorMessage("Invalid token ID", true))
	}

	token, err := h.Model.GetEnrollmentTokenByID(tokenID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	externalNATS := agentNATSURL(h.NATSServers)
	iniContent := generateConfigINI(externalNATS, token.Token)

	zipData, err := h.buildConfigZIP(iniContent)
	if err != nil {
		log.Printf("[ERROR]: could not build config ZIP: %v", err)
		return RenderError(c, partials.ErrorMessage("Could not create ZIP file", true))
	}

	filename := fmt.Sprintf("openuem-config-%s.zip", token.Token[:8])
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	return c.Blob(200, "application/zip", zipData)
}

// PublicDownloadConfig serves config ZIP without session auth.
// The enrollment token value in the URL acts as authentication.
func (h *Handler) PublicDownloadConfig(c echo.Context) error {
	tokenValue := c.Param("token")
	if tokenValue == "" {
		return c.String(http.StatusBadRequest, "missing token")
	}

	token, err := h.Model.GetEnrollmentTokenByValue(tokenValue)
	if err != nil {
		return c.String(http.StatusNotFound, "invalid token")
	}

	if !token.Active {
		return c.String(http.StatusForbidden, "token is inactive")
	}
	if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
		return c.String(http.StatusForbidden, "token has expired")
	}
	if token.MaxUses > 0 && token.CurrentUses >= token.MaxUses {
		return c.String(http.StatusForbidden, "token usage limit reached")
	}

	platform := c.QueryParam("platform")
	switch platform {
	case "linux", "macos", "windows":
	default:
		platform = "linux"
	}

	externalNATS := agentNATSURL(h.NATSServers)
	iniContent := generatePlatformConfigINI(platform, externalNATS, token.Token)

	zipData, err := h.buildConfigZIP(iniContent)
	if err != nil {
		log.Printf("[ERROR]: could not build config ZIP: %v", err)
		return c.String(http.StatusInternalServerError, "could not create config package")
	}

	if err := h.Model.IncrementEnrollmentTokenUses(tokenValue); err != nil {
		log.Printf("[WARN]: could not increment token usage count: %v", err)
	}

	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="openuem-config-%s.zip"`, tokenValue[:8]))
	return c.Blob(http.StatusOK, "application/zip", zipData)
}

func (h *Handler) GetInstallCommand(c echo.Context) error {
	tokenID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return RenderError(c, partials.ErrorMessage("Invalid token ID", true))
	}

	platform := c.QueryParam("platform")
	switch platform {
	case "linux", "macos-amd64", "macos-arm64", "windows":
	default:
		platform = "linux"
	}

	token, err := h.Model.GetEnrollmentTokenByID(tokenID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	consoleURL := fmt.Sprintf("https://%s", c.Request().Host)

	var command string
	var platformLabel string

	switch platform {
	case "linux":
		command = fmt.Sprintf(`curl -fsSL "%s/api/enroll/%s/install?platform=linux" | sudo bash`, consoleURL, token.Token)
		platformLabel = "Linux"
	case "macos-amd64":
		command = fmt.Sprintf(`curl -fsSL "%s/api/enroll/%s/install?platform=macos-amd64" | sudo bash`, consoleURL, token.Token)
		platformLabel = "macOS Intel"
	case "macos-arm64":
		command = fmt.Sprintf(`curl -fsSL "%s/api/enroll/%s/install?platform=macos-arm64" | sudo bash`, consoleURL, token.Token)
		platformLabel = "macOS ARM"
	case "windows":
		command = fmt.Sprintf(`irm "%s/api/enroll/%s/install?platform=windows" | iex`, consoleURL, token.Token)
		platformLabel = "Windows"
	}

	return RenderView(c, admin_views.InstallCommand(command, platformLabel))
}

// PublicInstallScript serves a platform-specific install script.
// The enrollment token value in the URL acts as authentication.
func (h *Handler) PublicInstallScript(c echo.Context) error {
	tokenValue := c.Param("token")
	if tokenValue == "" {
		return c.String(http.StatusBadRequest, "missing token")
	}

	token, err := h.Model.GetEnrollmentTokenByValue(tokenValue)
	if err != nil {
		return c.String(http.StatusNotFound, "invalid token")
	}

	if !token.Active {
		return c.String(http.StatusForbidden, "token is inactive")
	}
	if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
		return c.String(http.StatusForbidden, "token has expired")
	}
	if token.MaxUses > 0 && token.CurrentUses >= token.MaxUses {
		return c.String(http.StatusForbidden, "token usage limit reached")
	}

	platform := c.QueryParam("platform")
	switch platform {
	case "linux", "macos-amd64", "macos-arm64", "windows":
	default:
		platform = "linux"
	}

	consoleURL := fmt.Sprintf("https://%s", c.Request().Host)

	var script string
	var contentType string

	switch platform {
	case "linux":
		script = generateLinuxScript(consoleURL, tokenValue)
		contentType = "text/x-shellscript"
	case "macos-amd64":
		script = generateMacOSScript(consoleURL, tokenValue, "amd64")
		contentType = "text/x-shellscript"
	case "macos-arm64":
		script = generateMacOSScript(consoleURL, tokenValue, "arm64")
		contentType = "text/x-shellscript"
	case "windows":
		script = generateWindowsScript(consoleURL, tokenValue)
		contentType = "text/plain"
	}

	return c.Blob(http.StatusOK, contentType, []byte(script))
}

const agentReleaseBaseURL = "https://github.com/open-uem/openuem-agent/releases/latest/download"

func generateLinuxScript(consoleURL, token string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

CONFIG_DIR="/etc/openuem-agent"
RELEASE_URL="%s"

echo "Installing OpenUEM Agent..."

# Download and extract config + certificates
mkdir -p "$CONFIG_DIR"
curl -fsSL "%s/api/enroll/%s/config?platform=linux" -o /tmp/openuem-config.zip
unzip -o /tmp/openuem-config.zip -d "$CONFIG_DIR"
rm /tmp/openuem-config.zip

# Download and install agent
curl -fsSL "$RELEASE_URL/openuem-agent-linux-amd64.deb" -o /tmp/openuem-agent.deb
dpkg -i /tmp/openuem-agent.deb
rm /tmp/openuem-agent.deb

echo "OpenUEM Agent installed successfully."
`, agentReleaseBaseURL, consoleURL, token)
}

func generateMacOSScript(consoleURL, token, arch string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

CONFIG_DIR="/Library/OpenUEMAgent/etc/openuem-agent"
RELEASE_URL="%s"

echo "Installing OpenUEM Agent..."

# Download and extract config + certificates
mkdir -p "$CONFIG_DIR"
curl -fsSL "%s/api/enroll/%s/config?platform=macos" -o /tmp/openuem-config.zip
unzip -o /tmp/openuem-config.zip -d "$CONFIG_DIR"
rm /tmp/openuem-config.zip

# Download and install agent
curl -fsSL "$RELEASE_URL/openuem-agent-darwin-%s.pkg" -o /tmp/openuem-agent.pkg
installer -pkg /tmp/openuem-agent.pkg -target /
rm /tmp/openuem-agent.pkg

echo "OpenUEM Agent installed successfully."
`, agentReleaseBaseURL, consoleURL, token, arch)
}

func generateWindowsScript(consoleURL, token string) string {
	return fmt.Sprintf(`$ErrorActionPreference = 'Stop'

$InstallDir = "$env:ProgramFiles\OpenUEM\Agent"
$ReleaseURL = "%s"

Write-Host "Installing OpenUEM Agent..."

# Download and extract config + certificates
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Invoke-WebRequest "%s/api/enroll/%s/config?platform=windows" -OutFile "$env:TEMP\openuem-config.zip"
Expand-Archive "$env:TEMP\openuem-config.zip" $InstallDir -Force
Remove-Item "$env:TEMP\openuem-config.zip"

# Download and install agent
Invoke-WebRequest "$ReleaseURL/openuem-agent-windows-amd64.msi" -OutFile "$env:TEMP\openuem-agent.msi"
Start-Process msiexec -ArgumentList "/i `+"\""+`$env:TEMP\openuem-agent.msi`+"\""+` /qn" -Wait
Remove-Item "$env:TEMP\openuem-agent.msi"

Write-Host "OpenUEM Agent installed successfully."
`, agentReleaseBaseURL, consoleURL, token)
}

func generatePlatformConfigINI(platform, natsServers, token string) string {
	var sb strings.Builder
	sb.WriteString("[Agent]\n")
	sb.WriteString("UUID=\n")
	sb.WriteString("Enabled=true\n")
	sb.WriteString("ExecuteTaskEveryXMinutes=5\n")
	sb.WriteString("Debug=false\n")
	sb.WriteString("DefaultFrequency=5\n")
	sb.WriteString("SFTPPort=2022\n")
	sb.WriteString("VNCProxyPort=5900\n")
	sb.WriteString("SFTPDisabled=false\n")
	sb.WriteString("RemoteAssistanceDisabled=false\n")
	sb.WriteString(fmt.Sprintf("EnrollmentToken=%s\n", token))
	sb.WriteString("\n[NATS]\n")
	sb.WriteString(fmt.Sprintf("NATSServers=%s\n", natsServers))
	sb.WriteString("\n[Certificates]\n")
	if platform == "windows" {
		sb.WriteString("CACert=C:\\Program Files\\OpenUEM\\Agent\\certificates\\ca.cer\n")
		sb.WriteString("AgentCert=C:\\Program Files\\OpenUEM\\Agent\\certificates\\agent.cer\n")
		sb.WriteString("AgentKey=C:\\Program Files\\OpenUEM\\Agent\\certificates\\agent.key\n")
		sb.WriteString("SFTPCert=C:\\Program Files\\OpenUEM\\Agent\\certificates\\sftp.cer\n")
	} else {
		sb.WriteString("CACert=certificates/ca.cer\n")
		sb.WriteString("AgentCert=certificates/agent.cer\n")
		sb.WriteString("AgentKey=certificates/agent.key\n")
		sb.WriteString("SFTPCert=certificates/sftp.cer\n")
	}
	return sb.String()
}

func generateConfigINI(natsServers, token string) string {
	var sb strings.Builder
	sb.WriteString("[Agent]\n")
	sb.WriteString("UUID=\n")
	sb.WriteString("Enabled=true\n")
	sb.WriteString("ExecuteTaskEveryXMinutes=5\n")
	sb.WriteString("Debug=false\n")
	sb.WriteString("DefaultFrequency=5\n")
	sb.WriteString("SFTPPort=2022\n")
	sb.WriteString("VNCProxyPort=5900\n")
	sb.WriteString("SFTPDisabled=false\n")
	sb.WriteString("RemoteAssistanceDisabled=false\n")
	sb.WriteString(fmt.Sprintf("EnrollmentToken=%s\n", token))
	sb.WriteString("\n[NATS]\n")
	sb.WriteString(fmt.Sprintf("NATSServers=%s\n", natsServers))
	sb.WriteString("\n[Certificates]\n")
	sb.WriteString("CACert=certificates/ca.cer\n")
	sb.WriteString("AgentCert=certificates/agent.cer\n")
	sb.WriteString("AgentKey=certificates/agent.key\n")
	sb.WriteString("SFTPCert=certificates/sftp.cer\n")
	return sb.String()
}

// agentNATSURL returns the external NATS URL for agent configs.
// It combines NATS_SERVER (external host) and NATS_PORT (external port),
// falling back to the internal NATS_SERVERS value.
func agentNATSURL(fallback string) string {
	server := os.Getenv("NATS_SERVER")
	port := os.Getenv("NATS_PORT")

	if server == "" {
		return fallback
	}

	// Strip scheme if present, we'll add tls:// ourselves
	host := strings.TrimPrefix(strings.TrimPrefix(server, "tls://"), "nats://")

	if port != "" {
		return "tls://" + host + ":" + port
	}
	return "tls://" + host
}

func (h *Handler) listEnrollmentTokensWithError(c echo.Context, commonInfo *partials.CommonInfo, errMsg string) error {
	tenantID, _ := strconv.Atoi(commonInfo.TenantID)
	tokens, _ := h.Model.GetEnrollmentTokens(tenantID)
	sites, _ := h.Model.GetSites(tenantID)
	agentsExists, _ := h.Model.AgentsExists(commonInfo)
	serversExists, _ := h.Model.ServersExists()

	return RenderView(c, admin_views.EnrollmentTokensIndex(" | Enrollment",
		admin_views.EnrollmentTokens(c, tokens, sites, errMsg, agentsExists, serversExists, commonInfo),
		commonInfo))
}
