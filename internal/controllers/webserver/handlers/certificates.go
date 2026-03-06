package handlers

import (
	"fmt"
	"log"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/views/admin_views"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"golang.org/x/crypto/ocsp"
)

func (h *Handler) ListCertificates(c echo.Context) error {
	return h.GetCertificates(c, "")
}

func (h *Handler) GetCertificates(c echo.Context, successMessage string) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

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

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.NewPaginationAndSort(itemsPerPage)
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"), itemsPerPage)

	p.NItems, err = h.Model.CountAllCertificates(f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	certificates, err := h.Model.GetCertificatesByPage(p, f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	agentsExists, err := h.Model.AgentsExists(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return RenderView(c, admin_views.CertificatesIndex(" | Certificates", admin_views.Certificates(c, p, f, certTypes, certificates, successMessage, agentsExists, serversExists, itemsPerPage, commonInfo), commonInfo))
}

func (h *Handler) CertificateConfirmRevocation(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	serial := c.FormValue("serial")
	if serial == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "certificates.no_serial"), true))
	}

	return RenderConfirm(c, partials.ConfirmCertRevocation(c, serial, commonInfo))
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
