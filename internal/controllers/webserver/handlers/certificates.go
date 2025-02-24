package handlers

import (
	"fmt"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	model "github.com/open-uem/openuem-console/internal/models/servers"
	"github.com/open-uem/openuem-console/internal/views"
	"github.com/open-uem/openuem-console/internal/views/admin_views"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"golang.org/x/crypto/ocsp"
)

func (h *Handler) ListCertificates(c echo.Context) error {
	return h.GetCertificates(c, "")
}

func (h *Handler) GetCertificates(c echo.Context, successMessage string) error {
	f := filters.CertificateFilter{}

	f.Description = c.FormValue("filterByDescription")
	f.Username = c.FormValue("filterByUsername")

	certTypes, err := h.Model.GetCertificatesTypes()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	filteredCertTypesOptions := []string{}
	for index := range certTypes {
		value := c.FormValue(fmt.Sprintf("filterByType%d", index))
		if value != "" {
			filteredCertTypesOptions = append(filteredCertTypesOptions, value)
		}
	}
	f.TypeOptions = filteredCertTypesOptions

	expiryFrom := c.FormValue("filterByExpiryDateFrom")
	if expiryFrom != "" {
		f.ExpiryFrom = expiryFrom
	}
	expiryTo := c.FormValue("filterByExpiryDateTo")
	if expiryTo != "" {
		f.ExpiryTo = expiryTo
	}

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

	p.NItems, err = h.Model.CountAllCertificates(f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	certificates, err := h.Model.GetCertificatesByPage(p, f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	l := views.GetTranslatorForDates(c)

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	agentsExists, err := h.Model.AgentsExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, admin_views.CertificatesIndex(" | Certificates", admin_views.Certificates(c, p, f, h.SessionManager, h.Version, latestServerRelease.Version, l, certTypes, certificates, successMessage, agentsExists, serversExists)))
}

func (h *Handler) CertificateConfirmRevocation(c echo.Context) error {
	serial := c.FormValue("serial")
	if serial == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "certificates.no_serial"), true))
	}

	return RenderConfirm(c, partials.ConfirmCertRevocation(serial))
}

func (h *Handler) RevocateCertificate(c echo.Context) error {
	serial := c.FormValue("serial")
	if serial == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "certificates.no_serial"), true))
	}

	// First revoke certificate
	cert, err := h.Model.GetCertificateBySerial(serial)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := h.Model.RevokeCertificate(cert, "the certificate has been revoked", ocsp.Revoked); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if cert.UID != "" {
		if err := h.Model.UserSetRevokedCertificate(cert.UID); err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}
	}

	// Now delete certificate
	if err := h.Model.DeleteCertificate(cert.ID); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return h.GetCertificates(c, i18n.T(c.Request().Context(), "certificates.revocation_success"))
}
