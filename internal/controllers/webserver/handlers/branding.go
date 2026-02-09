package handlers

import (
	"encoding/base64"
	"io"
	"net/http"
	"strings"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent"
	"github.com/open-uem/openuem-console/internal/views/admin_views"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

const (
	// maxLogoSize is the maximum file size for logo uploads (2MB)
	maxLogoSize = 2 * 1024 * 1024
	// maxBackgroundSize is the maximum file size for background images (5MB)
	maxBackgroundSize = 5 * 1024 * 1024
)

// GetBrandingSettings handles GET /admin/branding
func (h *Handler) GetBrandingSettings(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	branding, err := h.Model.GetOrCreateBranding()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, admin_views.BrandingSettingsIndex(" | Branding", admin_views.BrandingSettings(c, branding, commonInfo, ""), commonInfo))
}

// PostBrandingLogo handles POST /admin/branding/logo (single logo)
func (h *Handler) PostBrandingLogo(c echo.Context) error {
	return h.handleLogoUpload(c, "light")
}

// DeleteBrandingLogo handles DELETE /admin/branding/logo
func (h *Handler) DeleteBrandingLogo(c echo.Context) error {
	if err := h.Model.DeleteLogoLight(); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}
	return h.renderBrandingWithSuccess(c, i18n.T(c.Request().Context(), "branding.logo_deleted"))
}

// PostBrandingFavicon handles POST /admin/branding/favicon
func (h *Handler) PostBrandingFavicon(c echo.Context) error {
	return h.handleLogoUpload(c, "small")
}

// DeleteBrandingFavicon handles DELETE /admin/branding/favicon
func (h *Handler) DeleteBrandingFavicon(c echo.Context) error {
	if err := h.Model.DeleteLogoSmall(); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}
	return h.renderBrandingWithSuccess(c, i18n.T(c.Request().Context(), "branding.favicon_deleted"))
}

// PostBrandingProductName handles POST /admin/branding/product-name
func (h *Handler) PostBrandingProductName(c echo.Context) error {
	productName := c.FormValue("product_name")
	if productName == "" {
		productName = "OpenUEM"
	}

	branding, err := h.Model.GetOrCreateBranding()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	branding.ProductName = productName
	if err := h.Model.UpdateBranding(branding); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	// Force a full page reload to update the header
	c.Response().Header().Set("HX-Redirect", "/admin/branding")
	return c.NoContent(http.StatusOK)
}

// PostBrandingColors handles POST /admin/branding/colors
func (h *Handler) PostBrandingColors(c echo.Context) error {
	// The text input is synced with the color picker via JavaScript
	// Use the text input value as it is always up-to-date
	primary := c.FormValue("primary_color_text")

	// Fallback to color picker value if text input is empty
	if primary == "" {
		primary = c.FormValue("primary_color")
	}

	if err := h.Model.UpdatePrimaryColor(primary); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	// Force a full page reload by redirecting to the same page
	// This ensures the new CSS in <head> is loaded
	c.Response().Header().Set("HX-Redirect", "/admin/branding")
	return c.NoContent(http.StatusOK)
}

// PostBrandingLogin handles POST /admin/branding/login (welcome text only)
func (h *Handler) PostBrandingLogin(c echo.Context) error {
	branding, err := h.Model.GetOrCreateBranding()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	branding.LoginWelcomeText = c.FormValue("login_welcome_text")

	if err := h.Model.UpdateBranding(branding); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return h.renderBrandingWithSuccess(c, i18n.T(c.Request().Context(), "branding.saved"))
}

// PostBrandingLoginBackground handles POST /admin/branding/login-background
func (h *Handler) PostBrandingLoginBackground(c echo.Context) error {
	branding, err := h.Model.GetOrCreateBranding()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	// Handle background image upload
	file, err := c.FormFile("login_background")
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "branding.no_file_selected"), true))
	}

	// Check file size limit
	if file.Size > maxBackgroundSize {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "branding.file_too_large"), true))
	}

	src, err := file.Open()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	mimeType := http.DetectContentType(data)
	if !strings.HasPrefix(mimeType, "image/") {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "branding.invalid_image"), true))
	}

	base64Data := base64.StdEncoding.EncodeToString(data)
	branding.LoginBackgroundImage = "data:" + mimeType + ";base64," + base64Data

	if err := h.Model.UpdateBranding(branding); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return h.renderBrandingWithSuccess(c, i18n.T(c.Request().Context(), "branding.saved"))
}

// DeleteBrandingLoginBackground handles DELETE /admin/branding/login-background
func (h *Handler) DeleteBrandingLoginBackground(c echo.Context) error {
	branding, err := h.Model.GetBranding()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	branding.LoginBackgroundImage = ""
	if err := h.Model.UpdateBranding(branding); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return h.renderBrandingWithSuccess(c, i18n.T(c.Request().Context(), "branding.logo_deleted"))
}

// handleLogoUpload processes logo file uploads
func (h *Handler) handleLogoUpload(c echo.Context, logoType string) error {
	fieldName := "logo"
	if logoType == "small" {
		fieldName = "favicon"
	}

	file, err := c.FormFile(fieldName)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "branding.no_file_selected"), true))
	}

	if file.Size > maxLogoSize {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "branding.file_too_large"), true))
	}

	src, err := file.Open()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	mimeType := http.DetectContentType(data)
	if !strings.HasPrefix(mimeType, "image/") {
		if strings.HasSuffix(strings.ToLower(file.Filename), ".svg") {
			mimeType = "image/svg+xml"
		} else {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "branding.invalid_image"), true))
		}
	}

	base64Data := base64.StdEncoding.EncodeToString(data)
	dataURL := "data:" + mimeType + ";base64," + base64Data

	var saveErr error
	switch logoType {
	case "light":
		saveErr = h.Model.SaveLogoLight(dataURL)
	case "small":
		saveErr = h.Model.SaveLogoSmall(dataURL)
	}

	if saveErr != nil {
		return RenderError(c, partials.ErrorMessage(saveErr.Error(), true))
	}

	return h.renderBrandingWithSuccess(c, i18n.T(c.Request().Context(), "branding.logo_uploaded"))
}

// renderBrandingWithSuccess renders the branding page with a success message
func (h *Handler) renderBrandingWithSuccess(c echo.Context, message string) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	branding, err := h.Model.GetOrCreateBranding()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, admin_views.BrandingSettingsIndex(" | Branding", admin_views.BrandingSettings(c, branding, commonInfo, message), commonInfo))
}

// GetBrandingForViews returns branding data for use in views
func (h *Handler) GetBrandingForViews() (*ent.Branding, error) {
	return h.Model.GetOrCreateBranding()
}
