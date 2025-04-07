package models

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func SearchPackages(packageName string, p partials.PaginationAndSort, dbFolder string, f filters.DeployPackageFilter) ([]nats.SoftwarePackage, error) {
	var rows *sql.Rows
	var err error

	sources := []string{}
	for _, s := range f.Sources {
		sources = append(sources, "'"+s+"'")
	}

	// Open DB
	db, err := OpenCommonDB(dbFolder)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Query the SQLite database
	switch p.SortBy {
	case "name":
		if p.SortOrder == "asc" {
			if len(f.Sources) == 0 {
				rows, err = db.Query(`SELECT DISTINCT id, name, source FROM apps WHERE name LIKE ? ORDER BY name ASC LIMIT ? OFFSET ? `, "%"+packageName+"%", p.PageSize, (p.CurrentPage-1)*p.PageSize)
			} else {
				q := fmt.Sprintf("SELECT DISTINCT id, name, source FROM apps WHERE name LIKE ? AND source IN (%s) ORDER BY name ASC LIMIT ? OFFSET ?", strings.Join(sources, ","))
				rows, err = db.Query(q, "%"+packageName+"%", p.PageSize, (p.CurrentPage-1)*p.PageSize)
			}
		} else {
			if len(f.Sources) == 0 {
				rows, err = db.Query(`SELECT DISTINCT id, name, source FROM apps WHERE name LIKE ? ORDER BY name DESC LIMIT ? OFFSET ? `, "%"+packageName+"%", p.PageSize, (p.CurrentPage-1)*p.PageSize)
			} else {
				q := fmt.Sprintf("SELECT DISTINCT id, name, source FROM apps WHERE name LIKE ? AND source IN (%s) ORDER BY name DESC LIMIT ? OFFSET ?", strings.Join(sources, ","))
				rows, err = db.Query(q, "%"+packageName+"%", p.PageSize, (p.CurrentPage-1)*p.PageSize)
			}
		}
	case "source":
		if p.SortOrder == "asc" {
			if len(f.Sources) == 0 {
				rows, err = db.Query(`SELECT DISTINCT id, name, source FROM apps WHERE name LIKE ? ORDER BY source ASC LIMIT ? OFFSET ? `, "%"+packageName+"%", p.PageSize, (p.CurrentPage-1)*p.PageSize)
			} else {
				q := fmt.Sprintf("SELECT DISTINCT id, name, source FROM apps WHERE name LIKE ? AND source IN (%s) ORDER BY sources ASC LIMIT ? OFFSET ?", strings.Join(sources, ","))
				rows, err = db.Query(q, "%"+packageName+"%", p.PageSize, (p.CurrentPage-1)*p.PageSize)
			}
		} else {
			if len(f.Sources) == 0 {
				rows, err = db.Query(`SELECT DISTINCT id, name, source FROM apps WHERE name LIKE ? ORDER BY source DESC LIMIT ? OFFSET ? `, "%"+packageName+"%", p.PageSize, (p.CurrentPage-1)*p.PageSize)
			} else {
				q := fmt.Sprintf("SELECT DISTINCT id, name, source FROM apps WHERE name LIKE ? AND source IN (%s) ORDER BY sources DESC LIMIT ? OFFSET ?", strings.Join(sources, ","))
				rows, err = db.Query(q, "%"+packageName+"%", p.PageSize, (p.CurrentPage-1)*p.PageSize)
			}
		}
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Scan our rows
	var packages []nats.SoftwarePackage
	for rows.Next() {
		var p nats.SoftwarePackage
		err := rows.Scan(&p.ID, &p.Name, &p.Source)
		if err != nil {
			return nil, err
		}
		packages = append(packages, p)
	}

	return packages, nil
}

func CountPackages(packageName string, indexPath string, f filters.DeployPackageFilter) (int, error) {
	var err error
	var rows *sql.Rows

	db, err := OpenCommonDB(indexPath)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	sources := []string{}
	for _, s := range f.Sources {
		sources = append(sources, "'"+s+"'")
	}

	// Query the SQLite database
	if len(f.Sources) == 0 {
		rows, err = db.Query(`
        SELECT DISTINCT id, name, source FROM apps
		WHERE name LIKE ?
	`, "%"+packageName+"%")
		if err != nil {
			return 0, err
		}
	} else {
		q := fmt.Sprintf("SELECT DISTINCT id, name, source FROM apps WHERE name LIKE ? AND source IN (%s)", strings.Join(sources, ","))
		rows, err = db.Query(q, "%"+packageName+"%")
		if err != nil {
			return 0, err
		}
	}

	defer rows.Close()

	// Scan our rows
	count := 0
	for rows.Next() {
		var p nats.SoftwarePackage
		err := rows.Scan(&p.ID, &p.Name, &p.Source)
		if err != nil {
			return 0, err
		}
		count++
	}

	return count, nil
}

func OpenWingetDB(indexPath string) (*sql.DB, error) {
	// Open Winget Community Repository index database
	dbPath := filepath.Join(indexPath, "index.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database doesn't exist, reason: %v", err)
	}

	return sql.Open("sqlite3", dbPath)
}

func SearchAllPackages(packageName string, wingetFolder string) ([]nats.SoftwarePackage, error) {
	var rows *sql.Rows
	var err error

	// Open Winget DB
	db, err := OpenWingetDB(wingetFolder)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Query the SQLite database
	rows, err = db.Query(`
			SELECT DISTINCT ids.id as id, names.name AS name FROM manifest 
			LEFT JOIN ids ON manifest.id = ids.rowid 
			LEFT JOIN names ON manifest.name = names.rowid 
			LEFT JOIN versions ON manifest.version = versions.rowid
			WHERE names.name LIKE ?	ORDER BY name ASC
		`, "%"+packageName+"%")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Scan our rows
	var packages []nats.SoftwarePackage
	for rows.Next() {
		var p nats.SoftwarePackage
		err := rows.Scan(&p.ID, &p.Name)
		if err != nil {
			return nil, err
		}
		packages = append(packages, p)
	}

	return packages, nil
}

func OpenFlatpakDB(indexPath string) (*sql.DB, error) {
	dbPath := filepath.Join(indexPath, "flatpak.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database doesn't exist, reason: %v", err)
	}

	return sql.Open("sqlite3", dbPath)
}

func OpenCommonDB(indexPath string) (*sql.DB, error) {
	dbPath := filepath.Join(indexPath, "common.db")

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		// create database if it doesn't exist
		f, err := os.Create(dbPath)
		if err != nil {
			return nil, fmt.Errorf("could not create common.db, reason: %v", err)
		}
		f.Close()
	}

	return sql.Open("sqlite3", dbPath)
}

func CreateCommonSoftwareTable(db *sql.DB) {
	sqlStmt := `create table apps (id text not null primary key, name text, source text)`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Println("[INFO]: could not create table apps for commondb")
	}
}

func DeleteCommonSoftwareTable(db *sql.DB) error {
	sqlStmt := `delete from apps`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		return err
	}
	return nil
}

func InsertCommonSoftware(db *sql.DB, apps []nats.SoftwarePackage, source string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("insert into apps(id, name, source) values(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, app := range apps {
		_, err = stmt.Exec(app.ID, app.Name, source)
		if err != nil {
			continue
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
