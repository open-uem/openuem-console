package admin_views

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

var navbarTests = []string{"users", "sessions", "tags", "smtp", "metadata", "settings", "update-agents", "update-servers", "certificates"}

func TestConfigNavbarTabs(t *testing.T) {
	for _, test := range navbarTests {
		t.Run(test, func(t *testing.T) {
			// Pipe the rendered template into goquery.
			r, w := io.Pipe()
			go func() {
				_ = ConfigNavbar(test, true, true).Render(context.Background(), w)
				_ = w.Close()
			}()
			doc, err := goquery.NewDocumentFromReader(r)
			if err != nil {
				t.Fatalf("failed to read template: %v", err)
			}

			selection := doc.Find(".uk-active")
			assert.Equal(t, 1, selection.Length(), "should get only one active tab")
			val, exists := selection.Find("a").Attr("href")
			assert.Equal(t, true, exists, "should get href")
			assert.Equal(t, fmt.Sprintf("/admin/%s", test), val, "should get active tab")
		})
	}
}
