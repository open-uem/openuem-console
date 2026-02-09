package helpers

import (
	"fmt"
	"strconv"
	"strings"
)

// HexToHSL converts a hex color string to HSL format without the "hsl()" wrapper
// Input: "#6d28d9" or "6d28d9"
// Output: "263 70% 50%" (format used by CSS variables in the theme)
func HexToHSL(hex string) string {
	// Remove # if present
	hex = strings.TrimPrefix(hex, "#")

	if len(hex) != 6 {
		return ""
	}

	// Parse RGB values
	r, err := strconv.ParseInt(hex[0:2], 16, 64)
	if err != nil {
		return ""
	}
	g, err := strconv.ParseInt(hex[2:4], 16, 64)
	if err != nil {
		return ""
	}
	b, err := strconv.ParseInt(hex[4:6], 16, 64)
	if err != nil {
		return ""
	}

	// Convert to 0-1 range
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	// Find min and max
	max := rf
	if gf > max {
		max = gf
	}
	if bf > max {
		max = bf
	}

	min := rf
	if gf < min {
		min = gf
	}
	if bf < min {
		min = bf
	}

	// Calculate lightness
	l := (max + min) / 2.0

	var h, s float64

	if max == min {
		// Achromatic
		h = 0
		s = 0
	} else {
		d := max - min

		// Calculate saturation
		if l > 0.5 {
			s = d / (2.0 - max - min)
		} else {
			s = d / (max + min)
		}

		// Calculate hue
		switch max {
		case rf:
			h = (gf - bf) / d
			if gf < bf {
				h += 6.0
			}
		case gf:
			h = (bf-rf)/d + 2.0
		case bf:
			h = (rf-gf)/d + 4.0
		}
		h /= 6.0
	}

	// Convert to degrees and percentages
	hDeg := h * 360.0
	sPct := s * 100.0
	lPct := l * 100.0

	return fmt.Sprintf("%.1f %.1f%% %.1f%%", hDeg, sPct, lPct)
}

// GetContrastColor returns a contrasting foreground color (white or black in HSL format)
// based on the luminance of the background color
func GetContrastColor(hex string) string {
	hex = strings.TrimPrefix(hex, "#")

	if len(hex) != 6 {
		return "0 0% 100%" // Default to white
	}

	r, _ := strconv.ParseInt(hex[0:2], 16, 64)
	g, _ := strconv.ParseInt(hex[2:4], 16, 64)
	b, _ := strconv.ParseInt(hex[4:6], 16, 64)

	// Calculate relative luminance
	// Using the formula: 0.299*R + 0.587*G + 0.114*B
	luminance := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 255.0

	if luminance > 0.5 {
		// Dark text for light backgrounds
		return "0 0% 9%" // Near black
	}
	// Light text for dark backgrounds
	return "0 0% 98%" // Near white
}

// GenerateBrandingCSS generates the complete CSS style tag content for the primary color
// Uses high specificity selectors to override theme defaults
func GenerateBrandingCSS(primaryColor string) string {
	if primaryColor == "" {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<style>")

	// Generate the select dropdown arrow SVG with the primary color
	// The color needs to be URL-encoded (# becomes %23)
	arrowSVG := fmt.Sprintf("url(\"data:image/svg+xml;charset=utf-8,%%3Csvg xmlns='http://www.w3.org/2000/svg' width='24' height='16'%%3E%%3Cpath fill='%%23%s' d='M12 1 9 6h6zM12 13 9 8h6z'/%%3E%%3C/svg%%3E\")",
		strings.TrimPrefix(primaryColor, "#"))

	// Build CSS properties string for primary color
	cssProps := fmt.Sprintf("--primary:%s !important;--primary-foreground:%s !important;--ring:%s !important;--uk-form-list-image:%s !important;",
		HexToHSL(primaryColor), GetContrastColor(primaryColor), HexToHSL(primaryColor), arrowSVG)

	// Use specific theme class selectors for highest specificity
	// This directly targets .uk-theme-openuem and other theme classes
	sb.WriteString(":root{" + cssProps + "}")
	sb.WriteString(".uk-theme-openuem,.uk-theme-openuem.dark{" + cssProps + "}")
	sb.WriteString(".uk-theme-zinc,.uk-theme-zinc.dark{" + cssProps + "}")
	sb.WriteString(".uk-theme-slate,.uk-theme-slate.dark{" + cssProps + "}")
	sb.WriteString(".uk-theme-stone,.uk-theme-stone.dark{" + cssProps + "}")
	sb.WriteString(".uk-theme-gray,.uk-theme-gray.dark{" + cssProps + "}")
	sb.WriteString(".uk-theme-neutral,.uk-theme-neutral.dark{" + cssProps + "}")
	sb.WriteString(".uk-theme-red,.uk-theme-red.dark{" + cssProps + "}")
	sb.WriteString(".uk-theme-rose,.uk-theme-rose.dark{" + cssProps + "}")
	sb.WriteString(".uk-theme-orange,.uk-theme-orange.dark{" + cssProps + "}")
	sb.WriteString(".uk-theme-green,.uk-theme-green.dark{" + cssProps + "}")
	sb.WriteString(".uk-theme-blue,.uk-theme-blue.dark{" + cssProps + "}")
	sb.WriteString(".uk-theme-yellow,.uk-theme-yellow.dark{" + cssProps + "}")
	sb.WriteString(".uk-theme-violet,.uk-theme-violet.dark{" + cssProps + "}")

	sb.WriteString("</style>")

	return sb.String()
}
