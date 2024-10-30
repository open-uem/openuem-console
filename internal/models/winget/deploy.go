package models

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/doncicuto/openuem_utils"
	_ "github.com/mattn/go-sqlite3"
)

type DeployPackage struct {
	ID   string
	Name string
}

func SearchPackages(packageName string, p partials.PaginationAndSort) ([]DeployPackage, error) {
	var rows *sql.Rows
	var err error

	// Open Winget DB
	db, err := openWingetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Query the SQLite database
	switch p.SortBy {
	case "name":
		if p.SortOrder == "asc" {
			rows, err = db.Query(`
			SELECT DISTINCT ids.id as id, names.name AS name FROM manifest 
			LEFT JOIN ids ON manifest.id = ids.rowid 
			LEFT JOIN names ON manifest.name = names.rowid 
			LEFT JOIN versions ON manifest.version = versions.rowid
			WHERE names.name LIKE ?	ORDER BY name ASC LIMIT ? OFFSET ?
		`, "%"+packageName+"%", p.PageSize, (p.CurrentPage-1)*p.PageSize)
		} else {
			rows, err = db.Query(`
			SELECT DISTINCT ids.id as id, names.name AS name FROM manifest 
			LEFT JOIN ids ON manifest.id = ids.rowid 
			LEFT JOIN names ON manifest.name = names.rowid 
			LEFT JOIN versions ON manifest.version = versions.rowid
			WHERE names.name LIKE ?	ORDER BY name DESC LIMIT ? OFFSET ?
		`, "%"+packageName+"%", p.PageSize, (p.CurrentPage-1)*p.PageSize)
		}
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Scan our rows
	var packages []DeployPackage
	for rows.Next() {
		var p DeployPackage
		err := rows.Scan(&p.ID, &p.Name)
		if err != nil {
			return nil, err
		}
		packages = append(packages, p)
	}

	return packages, nil
}

func CountPackages(packageName string) (int, error) {

	// Open Winget DB
	db, err := openWingetDB()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	// Query the SQLite database
	rows, err := db.Query(`
        SELECT DISTINCT ids.id as id, names.name AS name FROM manifest 
		LEFT JOIN ids ON manifest.id = ids.rowid 
		LEFT JOIN names ON manifest.name = names.rowid 
		LEFT JOIN versions ON manifest.version = versions.rowid
		WHERE names.name LIKE ?
	`, "%"+packageName+"%")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	// Scan our rows
	count := 0
	for rows.Next() {
		var p DeployPackage
		err := rows.Scan(&p.ID, &p.Name)
		if err != nil {
			return 0, err
		}
		count++
	}

	return count, nil
}

func openWingetDB() (*sql.DB, error) {
	// Create the file
	cwd, err := openuem_utils.GetWd()
	if err != nil {
		return nil, err
	}

	tmpDir := filepath.Join(cwd, "tmp")
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
