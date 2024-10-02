package handlers

import (
	"bytes"
	"context"
	"crypto"
	"io"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/ocsp"
)

func (h *Handler) Auth(c echo.Context) error {

	certs := c.Request().TLS.PeerCertificates

	if len(certs) != 1 {
		return echo.NewHTTPError(http.StatusUnauthorized, "Please provide valid credentials")
	}

	uid := certs[0].Subject.CommonName
	if uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "Wrong certificate")
	}

	if len(certs[0].OCSPServer) == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "No OCSP responders found in certificate")
	}
	ocspServer := certs[0].OCSPServer[0]

	// Verify cert against OCSP
	ocspURL, err := url.Parse(ocspServer)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not parse OCSP Responder URL")
	}

	cert := certs[0]
	caCert := h.CACert

	ocspRequest, err := ocsp.CreateRequest(cert, caCert, &ocsp.RequestOptions{Hash: crypto.SHA256})
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

	ocspResponse, err := ocsp.ParseResponse(output, caCert)
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

	return c.Redirect(http.StatusFound, "https://localhost:1323/dashboard")
}
