package handlers

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/enttest"
	"github.com/open-uem/ent/task"
	"github.com/open-uem/openuem-console/internal/models"
	"github.com/open-uem/openuem-console/internal/taskfiles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCertificate(t *testing.T, cn string, usage x509.ExtKeyUsage, parent *x509.Certificate, parentKey *rsa.PrivateKey) (*x509.Certificate, *rsa.PrivateKey) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	if parent == nil {
		template.IsCA = true
		template.BasicConstraintsValid = true
		parent, parentKey = template, key
	} else {
		template.ExtKeyUsage = []x509.ExtKeyUsage{usage}
	}

	der, err := x509.CreateCertificate(rand.Reader, template, parent, &key.PublicKey, parentKey)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	return cert, key
}

func TestAgentFile(t *testing.T) {
	ctx := context.Background()
	client := enttest.Open(t, "sqlite3", "file:agentfilehandler?mode=memory&_fk=1")
	defer client.Close()

	tenant := client.Tenant.Create().SetDescription("Tenant").SaveX(ctx)
	site := client.Site.Create().SetDescription("Site").SetTenant(tenant).SaveX(ctx)
	client.Agent.Create().SetID("agent1").SetHostname("pc1").SetOs("windows").SetNickname("pc1").
		SetAgentStatus(agent.AgentStatusEnabled).AddSite(site).ExecX(ctx)
	client.Agent.Create().SetID("agent2").SetHostname("pc2").SetOs("windows").SetNickname("pc2").
		SetAgentStatus(agent.AgentStatusEnabled).ExecX(ctx)
	profile := client.Profile.Create().SetName("All").SetApplyToAll(true).AddSite(site).SaveX(ctx)

	store, err := taskfiles.New(t.TempDir())
	require.NoError(t, err)
	file, err := store.Save(profile.ID, "app.msi", strings.NewReader("msi content"))
	require.NoError(t, err)

	caCert, caKey := newTestCertificate(t, "OpenUEM CA", 0, nil, nil)
	agentCert, _ := newTestCertificate(t, agentCertificateCommonName, x509.ExtKeyUsageClientAuth, caCert, caKey)
	userCert, _ := newTestCertificate(t, "admin", x509.ExtKeyUsageClientAuth, caCert, caKey)
	foreignCA, foreignKey := newTestCertificate(t, "Other CA", 0, nil, nil)
	foreignAgentCert, _ := newTestCertificate(t, agentCertificateCommonName, x509.ExtKeyUsageClientAuth, foreignCA, foreignKey)

	h := &Handler{Model: &models.Model{Client: client}, CACert: caCert, TaskFiles: store}
	e := echo.New()

	download := func(cert *x509.Certificate, agentID string, f taskfiles.File) (int, string) {
		req := httptest.NewRequest(http.MethodGet, f.URLPath(), nil)
		req.Header.Set(AgentIDHeader, agentID)
		if cert != nil {
			req.TLS = &tls.ConnectionState{PeerCertificates: []*x509.Certificate{cert}}
		}
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("profile", "hash", "name")
		c.SetParamValues(strconv.Itoa(f.ProfileID), f.Hash, f.Name)

		err := h.AgentFile(c)
		var httpErr *echo.HTTPError
		if errors.As(err, &httpErr) {
			return httpErr.Code, ""
		}
		require.NoError(t, err)
		return rec.Code, rec.Body.String()
	}

	code, body := download(agentCert, "agent1", file)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "msi content", body)

	// Only agents, with a certificate issued by OpenUEM's CA
	code, _ = download(nil, "agent1", file)
	assert.Equal(t, http.StatusUnauthorized, code, "no certificate")
	code, _ = download(userCert, "agent1", file)
	assert.Equal(t, http.StatusUnauthorized, code, "user certificate")
	code, _ = download(foreignAgentCert, "agent1", file)
	assert.Equal(t, http.StatusUnauthorized, code, "certificate from another CA")

	// Only agents the profile applies to
	code, _ = download(agentCert, "agent2", file)
	assert.Equal(t, http.StatusForbidden, code, "agent in another site")
	code, _ = download(agentCert, "", file)
	assert.Equal(t, http.StatusForbidden, code, "no agent id")

	// Only existing files of that profile
	code, _ = download(agentCert, "agent1", taskfiles.File{ProfileID: profile.ID, Hash: strings.Repeat("0", 64), Name: "app.msi"})
	assert.Equal(t, http.StatusNotFound, code)
	code, _ = download(agentCert, "agent1", taskfiles.File{ProfileID: profile.ID, Hash: file.Hash, Name: "../app.msi"})
	assert.Equal(t, http.StatusNotFound, code)

	// A task cloned into another profile can use the file for that profile's agents
	site2 := client.Site.Create().SetDescription("Site 2").SetTenant(tenant).SaveX(ctx)
	client.Agent.Create().SetID("agent3").SetHostname("pc3").SetOs("windows").SetNickname("pc3").
		SetAgentStatus(agent.AgentStatusEnabled).AddSite(site2).ExecX(ctx)
	profile2 := client.Profile.Create().SetName("Site 2").SetApplyToAll(true).AddSite(site2).SaveX(ctx)
	code, _ = download(agentCert, "agent3", file)
	assert.Equal(t, http.StatusForbidden, code, "agent of another profile without a task using the file")
	client.Task.Create().SetName("Install app").SetType(task.TypeMsiInstall).SetMsiPath("https://console.example.com:1324" + file.URLPath()).SetProfile(profile2).ExecX(ctx)
	code, _ = download(agentCert, "agent3", file)
	assert.Equal(t, http.StatusOK, code, "agent of a profile with a cloned task")

	// The Client-Cert header proves nothing, it's only used behind a reverse proxy
	headerDownload := func(reverseProxyAuthPort string) int {
		h.ReverseProxyAuthPort = reverseProxyAuthPort
		defer func() { h.ReverseProxyAuthPort = "" }()

		req := httptest.NewRequest(http.MethodGet, file.URLPath(), nil)
		req.Header.Set(AgentIDHeader, "agent1")
		req.Header.Set("Client-Cert", ":"+base64.StdEncoding.EncodeToString(agentCert.Raw)+":")
		c := e.NewContext(req, httptest.NewRecorder())
		c.SetParamNames("profile", "hash", "name")
		c.SetParamValues(strconv.Itoa(file.ProfileID), file.Hash, file.Name)

		var httpErr *echo.HTTPError
		if err := h.AgentFile(c); errors.As(err, &httpErr) {
			return httpErr.Code
		}
		return http.StatusOK
	}
	assert.Equal(t, http.StatusUnauthorized, headerDownload(""), "header without reverse proxy")
	assert.Equal(t, http.StatusOK, headerDownload("443"), "header behind a reverse proxy")
}
