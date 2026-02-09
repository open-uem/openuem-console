package models

import (
	"context"

	"github.com/open-uem/ent"
)

// GetBranding retrieves the global branding settings.
// There should only be one branding record (singleton pattern).
func (m *Model) GetBranding() (*ent.Branding, error) {
	return m.Client.Branding.Query().First(context.Background())
}

// GetOrCreateBranding retrieves branding settings or creates default if not exists.
func (m *Model) GetOrCreateBranding() (*ent.Branding, error) {
	b, err := m.Client.Branding.Query().First(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			// Create default branding
			return m.Client.Branding.Create().
				SetProductName("OpenUEM").
				SetPrimaryColor("#16a34a").
				Save(context.Background())
		}
		return nil, err
	}
	return b, nil
}

// UpdateBranding updates the global branding settings.
func (m *Model) UpdateBranding(b *ent.Branding) error {
	update := m.Client.Branding.UpdateOneID(b.ID)

	// Logo settings
	if b.LogoLight != "" {
		update = update.SetLogoLight(b.LogoLight)
	} else {
		update = update.ClearLogoLight()
	}
	if b.LogoSmall != "" {
		update = update.SetLogoSmall(b.LogoSmall)
	} else {
		update = update.ClearLogoSmall()
	}

	// Colors
	if b.PrimaryColor != "" {
		update = update.SetPrimaryColor(b.PrimaryColor)
	}

	// Text settings
	if b.ProductName != "" {
		update = update.SetProductName(b.ProductName)
	}

	// Login page
	if b.LoginBackgroundImage != "" {
		update = update.SetLoginBackgroundImage(b.LoginBackgroundImage)
	} else {
		update = update.ClearLoginBackgroundImage()
	}
	if b.LoginWelcomeText != "" {
		update = update.SetLoginWelcomeText(b.LoginWelcomeText)
	} else {
		update = update.ClearLoginWelcomeText()
	}

	return update.Exec(context.Background())
}

// SaveLogoLight saves the light mode logo.
func (m *Model) SaveLogoLight(logoData string) error {
	b, err := m.GetOrCreateBranding()
	if err != nil {
		return err
	}
	return m.Client.Branding.UpdateOneID(b.ID).
		SetLogoLight(logoData).
		Exec(context.Background())
}

// SaveLogoSmall saves the small logo/favicon.
func (m *Model) SaveLogoSmall(logoData string) error {
	b, err := m.GetOrCreateBranding()
	if err != nil {
		return err
	}
	return m.Client.Branding.UpdateOneID(b.ID).
		SetLogoSmall(logoData).
		Exec(context.Background())
}

// UpdatePrimaryColor updates the primary color.
func (m *Model) UpdatePrimaryColor(primary string) error {
	b, err := m.GetOrCreateBranding()
	if err != nil {
		return err
	}

	return m.Client.Branding.UpdateOneID(b.ID).
		SetPrimaryColor(primary).
		Exec(context.Background())
}

// SaveLoginBackgroundImage saves the login page background image.
func (m *Model) SaveLoginBackgroundImage(imageData string) error {
	b, err := m.GetOrCreateBranding()
	if err != nil {
		return err
	}
	return m.Client.Branding.UpdateOneID(b.ID).
		SetLoginBackgroundImage(imageData).
		Exec(context.Background())
}

// SaveLoginWelcomeText saves the login page welcome text.
func (m *Model) SaveLoginWelcomeText(text string) error {
	b, err := m.GetOrCreateBranding()
	if err != nil {
		return err
	}
	return m.Client.Branding.UpdateOneID(b.ID).
		SetLoginWelcomeText(text).
		Exec(context.Background())
}

// BrandingExists checks if branding settings exist.
func (m *Model) BrandingExists() (bool, error) {
	return m.Client.Branding.Query().Exist(context.Background())
}

// DeleteLogoLight removes the light mode logo.
func (m *Model) DeleteLogoLight() error {
	b, err := m.GetBranding()
	if err != nil {
		return err
	}
	return m.Client.Branding.UpdateOneID(b.ID).
		ClearLogoLight().
		Exec(context.Background())
}

// DeleteLogoSmall removes the small logo.
func (m *Model) DeleteLogoSmall() error {
	b, err := m.GetBranding()
	if err != nil {
		return err
	}
	return m.Client.Branding.UpdateOneID(b.ID).
		ClearLogoSmall().
		Exec(context.Background())
}

// DeleteLoginBackgroundImage removes the login background image.
func (m *Model) DeleteLoginBackgroundImage() error {
	b, err := m.GetBranding()
	if err != nil {
		return err
	}
	return m.Client.Branding.UpdateOneID(b.ID).
		ClearLoginBackgroundImage().
		Exec(context.Background())
}

func (m *Model) UpdateShowVersion(show bool) error {
	b, err := m.GetOrCreateBranding()
	if err != nil {
		return err
	}
	return m.Client.Branding.UpdateOneID(b.ID).
		SetShowVersion(show).
		Exec(context.Background())
}

func (m *Model) UpdateBugReportLink(link string) error {
	b, err := m.GetOrCreateBranding()
	if err != nil {
		return err
	}
	update := m.Client.Branding.UpdateOneID(b.ID)
	if link == "" {
		update = update.ClearBugReportLink()
	} else {
		update = update.SetBugReportLink(link)
	}
	return update.Exec(context.Background())
}

func (m *Model) UpdateHelpLink(link string) error {
	b, err := m.GetOrCreateBranding()
	if err != nil {
		return err
	}
	update := m.Client.Branding.UpdateOneID(b.ID)
	if link == "" {
		update = update.ClearHelpLink()
	} else {
		update = update.SetHelpLink(link)
	}
	return update.Exec(context.Background())
}
