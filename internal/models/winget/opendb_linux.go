//go:build linux

package models

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

func OpenWingetDB() (*sql.DB, error) {
	tmpDir := "/tmp/winget"

	indexPath := filepath.Join(tmpDir, "index.db")

	// Open Winget Community Repository index database
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database doesn't exist, reason: %v", err)
	}

	return sql.Open("sqlite", indexPath)
}
