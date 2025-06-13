package handlers

import (
	"bytes"
	"context"
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/ocsp"
)

func (h *Handler) Auth(c echo.Context) error {
	var err error
	var cert *x509.Certificate
	certs := c.Request().TLS.PeerCertificates

	if len(certs) != 1 {
		clientEncodedCert := c.Request().Header.Get("Client-Cert")
		if clientEncodedCert != "" {
			cleanBase64 := ""
			if strings.Contains(clientEncodedCert, "BEGIN CERTIFICATE") {
				// NGINX
				cleanBase64 = strings.TrimPrefix(clientEncodedCert, ":-----BEGIN CERTIFICATE----- ")
				cleanBase64 = strings.TrimSuffix(cleanBase64, " -----END CERTIFICATE-----:")
				cleanBase64 = strings.ReplaceAll(cleanBase64, " ", "")
			} else {
				// Caddy
				cleanBase64 = strings.Trim(clientEncodedCert, ":")
			}
			decoded, err := base64.StdEncoding.DecodeString(cleanBase64)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "The certificate could not be decoded")
			}
			cert, err = x509.ParseCertificate(decoded)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Could not parse client certificate")
			}
		} else {
			return echo.NewHTTPError(http.StatusUnauthorized, "Please provide valid credentials")
		}
	} else {
		cert = certs[0]
	}

	caCert := h.CACert

	uid := cert.Subject.CommonName
	if uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "Wrong certificate")
	}

	if len(cert.OCSPServer) == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "No OCSP responders found in certificate")
	}
	ocspServer := cert.OCSPServer[0]

	// Verify cert
	ocspURL, err := url.Parse(ocspServer)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not parse OCSP Responder URL")
	}

	issuer, err := getIssuerFromCert(cert, caCert)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Certificate did not pass verification")
	}

	ocspRequest, err := ocsp.CreateRequest(cert, issuer, &ocsp.RequestOptions{Hash: crypto.SHA256})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not create OCSP Request")
	}

	httpRequest, err := http.NewRequest(http.MethodPost, ocspServer, bytes.NewBuffer(ocspRequest))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not create request to OCSP Responder")
	}

	httpRequest.Header.Add("Content-Type", "application/ocsp-request")
	httpRequest.Header.Add("Accept", "application/ocsp-response")
	httpRequest.Header.Add("host", ocspURL.Host)

	httpClient := &http.Client{}
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not send request to OCSP Responder")
	}
	defer httpResponse.Body.Close()
	output, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not read response from OCSP Responder")
	}

	ocspResponse, err := ocsp.ParseResponse(output, issuer)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not parse OCSP Response")
	}

	if ocspResponse.Status == 2 {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not check OCSP status, try again later")
	}

	if ocspResponse.Status == 1 {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized. Your certificate has been revoked")
	}

	// Check if uid exists in database
	user, err := h.Model.GetUserById(uid)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not check if user exists")
	}
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Access is denied")
	}

	msg := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
	if msg != uid {
		err := h.SessionManager.Manager.RenewToken(c.Request().Context())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		h.SessionManager.Manager.Put(c.Request().Context(), "uid", uid)
		h.SessionManager.Manager.Put(c.Request().Context(), "username", user.Name)
		h.SessionManager.Manager.Put(c.Request().Context(), "user-agent", c.Request().UserAgent())
		h.SessionManager.Manager.Put(c.Request().Context(), "ip-address", c.Request().RemoteAddr)
		token, expiry, err := h.SessionManager.Manager.Commit(c.Request().Context())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		h.SessionManager.Manager.WriteSessionCookie(c.Request().Context(), c.Response().Writer, token, expiry)

		_, err = h.Model.Client.Sessions.UpdateOneID(token).SetOwnerID(uid).Save(context.Background())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		// if it's the first time let's confirm login and remove the cert password
		if err := h.Model.ConfirmLogIn(uid); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	// TODO - Get user's default tenant and site
	myTenant, err := h.Model.GetDefaultTenant()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	mySite, err := h.Model.GetDefaultSite(myTenant)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if h.ReverseProxyAuthPort != "" {
		url := strings.TrimSuffix(c.Request().Referer(), "/")
		url += fmt.Sprintf("/tenant/%d/site/%d/dashboard", myTenant.ID, mySite.ID)
		return c.Redirect(http.StatusFound, url)
	} else {
		return c.Redirect(http.StatusFound, fmt.Sprintf("https://%s:%s/tenant/%d/site/%d/dashboard", h.ServerName, h.ConsolePort, myTenant.ID, mySite.ID))
	}
}

func getIssuerFromCert(cert, caCert *x509.Certificate) (*x509.Certificate, error) {

	// Check if current certificate is valid for client auth and is issued by our CA
	trustedCAPool := x509.NewCertPool()
	trustedCAPool.AddCert(caCert)
	vOpts := x509.VerifyOptions{
		Roots:     trustedCAPool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	chains, err := cert.Verify(vOpts)
	if err != nil || len(chains) == 0 {
		return nil, err
	} else {
		return chains[0][1], err
	}
}
