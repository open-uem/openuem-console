//go:build windows

package models

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-uem/utils"
)

func OpenWingetDB() (*sql.DB, error) {
	cwd, err := utils.GetWd()
	if err != nil {
		return nil, err
	}

	tmpDir := filepath.Join(cwd, "tmp", "winget")
	if strings.HasSuffix(cwd, "tmp") {
		tmpDir = cwd
	}

	indexPath := filepath.Join(tmpDir, "index.db")

	// Open Winget Community Repository index database
	if _, err = os.Stat(indexPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database doesn't exist, reason: %v", err)
	}

	return sql.Open("sqlite3", indexPath)
}
