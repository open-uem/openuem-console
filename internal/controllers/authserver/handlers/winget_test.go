package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/enttest"
	"github.com/open-uem/openuem-console/internal/models"
	"github.com/open-uem/openuem-console/internal/wingetrepo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWinGetSource(t *testing.T) {
	ctx := context.Background()
	client := enttest.Open(t, "sqlite3", "file:wingetsource?mode=memory&_fk=1")
	defer client.Close()

	tenant1 := client.Tenant.Create().SetDescription("Tenant 1").SaveX(ctx)
	tenant2 := client.Tenant.Create().SetDescription("Tenant 2").SaveX(ctx)
	site1 := client.Site.Create().SetDescription("Site 1").SetTenant(tenant1).SaveX(ctx)
	client.Agent.Create().SetID("agent1").SetHostname("pc1").SetOs("windows").SetNickname("pc1").
		SetAgentStatus(agent.AgentStatusEnabled).AddSite(site1).ExecX(ctx)
	client.Agent.Create().SetID("waiting").SetHostname("pc2").SetOs("windows").SetNickname("pc2").
		SetAgentStatus(agent.AgentStatusWaitingForAdmission).AddSite(site1).ExecX(ctx)

	library, err := wingetrepo.NewLibrary(t.TempDir())
	require.NoError(t, err)
	add := func(id, version string, tenantID int, content string) {
		staged, err := library.Stage(strings.NewReader(content))
		require.NoError(t, err)
		_, err = library.Add(wingetrepo.Package{
			PackageIdentifier: id, PackageVersion: version, PackageName: id, Publisher: "Example Corp",
			ProductCode: "{0D8B6A0E-1F2A-4C3B-9D4E-5F6A7B8C9D0E}", Architecture: "neutral", InstallerFile: "app.msi", TenantID: tenantID,
		}, staged)
		require.NoError(t, err)
	}
	add("Example.Global", "1.0", 0, "global installer")
	add("Example.Tenant1", "2.0", tenant1.ID, "tenant 1 installer")
	add("Example.Tenant2", "3.0", tenant2.ID, "tenant 2 installer")

	key := []byte("server key")
	h := &Handler{Model: &models.Model{Client: client}, WinGetLibrary: library, WinGetKey: key}
	e := echo.New()
	h.Register(e)
	server := httptest.NewServer(e)
	defer server.Close()

	base := server.URL + "/winget/agent1/" + wingetrepo.AgentToken(key, "agent1")
	do := func(method, url, body string) (int, string) {
		req, err := http.NewRequest(method, url, strings.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		return resp.StatusCode, string(data)
	}

	// Access
	code, body := do(http.MethodGet, base+"/information", "")
	assert.Equal(t, http.StatusOK, code)
	assert.JSONEq(t, `{"Data":{"SourceIdentifier":"OpenUEM","ServerSupportedVersions":["1.0.0","1.1.0"]}}`, body)
	code, body = do(http.MethodGet, server.URL+"/winget/agent1/wrongtoken/information", "")
	assert.Equal(t, http.StatusUnauthorized, code, "wrong token")
	assert.JSONEq(t, `{"ErrorCode":401,"ErrorMessage":"Unauthorized"}`, body, "errors are JSON, not the console's HTML error pages")
	code, _ = do(http.MethodGet, server.URL+"/winget/agent2/"+wingetrepo.AgentToken(key, "agent1")+"/information", "")
	assert.Equal(t, http.StatusUnauthorized, code, "token of another agent")
	code, _ = do(http.MethodGet, server.URL+"/winget/waiting/"+wingetrepo.AgentToken(key, "waiting")+"/information", "")
	assert.Equal(t, http.StatusForbidden, code, "agent waiting for admission")

	// Search: only the packages of the agent's organization and the global ones
	code, body = do(http.MethodPost, base+"/manifestSearch", `{"Query":{"KeyWord":"example","MatchType":"Substring"}}`)
	assert.Equal(t, http.StatusOK, code)
	search := wingetrepo.SearchResponse{}
	require.NoError(t, json.Unmarshal([]byte(body), &search))
	found := []string{}
	for _, p := range search.Data {
		found = append(found, p.PackageIdentifier)
	}
	assert.ElementsMatch(t, []string{"Example.Global", "Example.Tenant1"}, found)

	code, _ = do(http.MethodPost, base+"/manifestSearch", `{"Query":{"KeyWord":"nothing","MatchType":"Substring"}}`)
	assert.Equal(t, http.StatusNoContent, code)

	// Manifest with the installer URL of the agent
	code, body = do(http.MethodGet, base+"/packageManifests/Example.Tenant1", "")
	assert.Equal(t, http.StatusOK, code)
	manifest := wingetrepo.ManifestResponse{}
	require.NoError(t, json.Unmarshal([]byte(body), &manifest))
	require.Len(t, manifest.Data.Versions, 1)
	installerURL := manifest.Data.Versions[0].Installers[0].InstallerUrl
	assert.Equal(t, "https://"+strings.TrimPrefix(base, "http://")+"/installers/Example.Tenant1/2.0/app.msi", installerURL)

	code, _ = do(http.MethodGet, base+"/packageManifests/Example.Tenant2", "")
	assert.Equal(t, http.StatusNoContent, code, "package of another organization")

	// Installers
	code, body = do(http.MethodGet, base+"/installers/Example.Tenant1/2.0/app.msi", "")
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "tenant 1 installer", body)
	code, _ = do(http.MethodGet, base+"/installers/Example.Tenant2/3.0/app.msi", "")
	assert.Equal(t, http.StatusNotFound, code, "installer of another organization")
	code, _ = do(http.MethodGet, base+"/installers/Example.Tenant1/2.0/other.msi", "")
	assert.Equal(t, http.StatusNotFound, code, "wrong file name")
}
