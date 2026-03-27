package handlers

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/open-uem/ent"
	"github.com/open-uem/ent/softwarepackage"
	"github.com/open-uem/nats"
	"github.com/open-uem/utils"
)

type Component struct {
	ID     string `xml:"id"`
	Names  []Name `xml:"name"`
	Bundle string `xml:"bundle"`
	Custom Custom `xml:"custom"`
}

type CustomValue struct {
	Key  string `xml:"key,attr"`
	Text string `xml:",chardata"`
}

type Custom struct {
	Values []CustomValue `xml:"value"`
}

type Name struct {
	Lang string `xml:"http://www.w3.org/XML/1998/namespace lang,attr"`
	Text string `xml:",chardata"`
}

type FlatpakXML struct {
	Components []Component `xml:"component"`
}

type Cask struct {
	Token string   `json:"token"`
	Name  []string `json:"name"`
}

type Formula struct {
	Name string `json:"name"`
}

func (h *Handler) StartCommonPackagesDBJob() error {
	var err error

	if err := h.UpdateSoftwarePackageTable(); err != nil {
		return err
	}

	h.CommonAppsJob, err = h.TaskScheduler.NewJob(
		gocron.DurationJob(
			time.Duration(24*time.Hour),
		),
		gocron.NewTask(
			func() {
				if err := h.UpdateSoftwarePackageTable(); err != nil {
					log.Printf("[ERROR]: cron job could not update the software packages table, reason: %v", err)
				}
			},
		),
	)
	if err != nil {
		log.Printf("[ERROR]: could not schedule job that updates the software packages table, reason: %v", err)
		return err
	}

	return nil
}

func (h *Handler) UpdateSoftwarePackageTable() error {

	// 1. Download Winget msix that contains the database file
	if err := h.DownloadWingetDB(); err != nil {
		log.Printf("[ERROR]: the Winget database has not been downloaded from Microsoft: %v", err)
	} else {
		log.Println("[INFO]: the Winget database has been downloaded from Microsoft")
		if err := h.UpdateWingetApps(); err != nil {
			log.Printf("[ERROR]: could not add Winget apps to our database: %v", err)
			return err
		}
		log.Println("[INFO]: Winget apps have been added to our database")
	}

	// 2. Download flatpak XML file
	if err := h.DownloadFlatpakXML("x86_64"); err != nil {
		log.Printf("[ERROR]: the Flatpak apps XML has not been downloaded from Flatpak: %v", err)
	} else {
		log.Println("[INFO]: the Flatpak apps XML has been downloaded from Flatpak")
		if err := h.UpdateFlatpakApps("x86_64"); err != nil {
			log.Printf("[ERROR]: could not add Flatpak apps for x86_64 to our database: %v", err)
			return err
		}
		log.Println("[INFO]: Flatpak apps have been added to our database")
	}

	// 4. Download flatpak brew cask file
	if err := h.DownloadBrewJSON("cask"); err != nil {
		log.Printf("[ERROR]: the Brew casks JSON file for casks has not been downloaded from Brew: %v", err)
	} else {
		log.Println("[INFO]: the Brew casks JSON file for casks has been downloaded from Brew")
		if err := h.UpdateBrewCasks(); err != nil {
			return err
		}
		log.Println("[INFO]: Brew casks have been added to our database")
	}

	// 5. Download flatpak brew formula file
	if err := h.DownloadBrewJSON("formula"); err != nil {
		log.Printf("[ERROR]: the Brew formula JSON file for formulae has not been downloaded from Brew: %v", err)
	} else {
		log.Println("[INFO]: the Brew formula JSON file for formulae has been downloaded from Brew")
		if err := h.UpdateBrewFormulae(); err != nil {
			return err
		}
		log.Println("[INFO]: Brew formulae has been added to our database")

	}

	return nil
}

func (h *Handler) DownloadWingetDB() error {
	url := "https://cdn.winget.microsoft.com/cache/source2.msix"

	zipPath := filepath.Join(getCommonAppsFolder(), "winget.zip")
	out, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil

}

func (h *Handler) DownloadFlatpakXML(arch string) error {
	url := fmt.Sprintf("https://dl.flathub.org/repo/appstream/%s/appstream.xml.gz", arch)

	dbPath := filepath.Join(getCommonAppsFolder(), fmt.Sprintf("flatpak_%s.xml", arch))
	out, err := os.Create(dbPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// gunzip file
	r, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	// Write the body to file
	_, err = io.Copy(out, r)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) DownloadBrewJSON(brewType string) error {

	url := fmt.Sprintf("https://formulae.brew.sh/api/%s.json", brewType)

	dbPath := filepath.Join(getCommonAppsFolder(), fmt.Sprintf("brew_%s.json", brewType))
	out, err := os.Create(dbPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) UpdateWingetApps() error {
	var rows *sql.Rows

	zipPath := filepath.Join(getCommonAppsFolder(), "winget.zip")
	wingetSQLiteDB := filepath.Join(getCommonAppsFolder(), "index.db")

	// Open ZIP reader
	archive, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer archive.Close()

	// look for the index.db
	for _, f := range archive.File {
		if f.Name == "Public/index.db" {
			src, err := f.Open()
			if err != nil {
				return err
			}
			defer src.Close()

			dst, err := os.Create(wingetSQLiteDB)
			if err != nil {
				return err
			}
			defer dst.Close()

			_, err = io.Copy(dst, src)
			if err != nil {
				return err
			}
			break
		}
	}
	archive.Close()

	// Remove temp ZIP file
	if err := os.Remove(zipPath); err != nil {
		return err
	}

	// Open SQLite database
	db, err := sql.Open("sqlite3", wingetSQLiteDB)
	if err != nil {
		return err
	}
	defer db.Close()

	// Initiate transaction
	tx, err := h.Model.Client.Tx(context.Background())
	if err != nil {
		return err
	}

	// Delete existing winget apps
	if _, err = h.Model.Client.SoftwarePackage.Delete().Where(softwarepackage.Source("winget")).Exec(context.Background()); err != nil {
		return rollback(tx, err)
	}

	// Scan Winget DB rows
	rows, err = db.Query(`SELECT DISTINCT id, name FROM packages`)
	if err != nil {
		return rollback(tx, err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var p nats.SoftwarePackage
			err := rows.Scan(&p.ID, &p.Name)
			if err != nil {
				return rollback(tx, err)
			}

			if err := tx.SoftwarePackage.Create().SetPackageID(p.ID).SetName(p.Name).SetSource("winget").Exec(context.Background()); err != nil {
				return rollback(tx, err)
			}
		}
	}

	// Commit the transaction.
	return tx.Commit()
}

func (h *Handler) UpdateFlatpakApps(arch string) error {
	xmlPath := filepath.Join(getCommonAppsFolder(), fmt.Sprintf("flatpak_%s.xml", arch))

	data, err := os.ReadFile(xmlPath)
	if err != nil {
		return err
	}

	fXML := FlatpakXML{}
	if err := xml.Unmarshal(data, &fXML); err != nil {
		return err
	}

	// Initiate transaction
	tx, err := h.Model.Client.Tx(context.Background())
	if err != nil {
		return err
	}

	// Delete existing winget apps
	if _, err = h.Model.Client.SoftwarePackage.Delete().Where(softwarepackage.Source("flatpak")).Exec(context.Background()); err != nil {
		return rollback(tx, err)
	}

	for _, c := range fXML.Components {
		packageID := ""
		packageName := ""
		packageBranch := ""
		packageVerified := false

		// Check if removing .desktop suffix only names have only 1 periods (Names must contain at least 2 periods)
		packageID = strings.TrimSuffix(c.ID, ".desktop")

		if len(strings.Split(packageID, ".")) <= 2 {
			packageID = c.ID
		}

		for _, n := range c.Names {
			if n.Lang == "" {
				packageName = n.Text
				break
			}
		}

		for _, n := range c.Custom.Values {
			if n.Key == "flathub::verification::verified" {
				isVerified, err := strconv.ParseBool(n.Text)
				if err == nil {
					packageVerified = isVerified
				}
				break
			}
		}

		if c.Bundle != "" {
			bundleText := strings.Split(c.Bundle, "/")
			packageBranch = bundleText[len(bundleText)-1]
		}

		if err := tx.SoftwarePackage.Create().SetPackageID(packageID).SetName(packageName).SetSource("flatpak").SetBranch(packageBranch).SetVerified(packageVerified).Exec(context.Background()); err != nil {
			return rollback(tx, err)
		}

	}

	// Commit the transaction.
	return tx.Commit()
}

func (h *Handler) UpdateBrewFormulae() error {
	jsonPath := filepath.Join(getCommonAppsFolder(), "brew_formula.json")

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}

	formulae := []Formula{}
	if err := json.Unmarshal(data, &formulae); err != nil {
		return err
	}

	// Initiate transaction
	tx, err := h.Model.Client.Tx(context.Background())
	if err != nil {
		return err
	}

	// Delete existing winget apps
	if _, err := h.Model.Client.SoftwarePackage.Delete().Where(softwarepackage.Source("brew"), softwarepackage.BrewType("formula")).Exec(context.Background()); err != nil {
		return rollback(tx, err)
	}

	for _, f := range formulae {
		packageID := f.Name
		packageName := f.Name

		if err := tx.SoftwarePackage.Create().SetPackageID(packageID).SetName(packageName).SetSource("brew").SetBrewType("formula").Exec(context.Background()); err != nil {
			return rollback(tx, err)
		}
	}

	// Commit the transaction.
	return tx.Commit()

}

func (h *Handler) UpdateBrewCasks() error {
	jsonPath := filepath.Join(getCommonAppsFolder(), "brew_cask.json")

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}

	casks := []Cask{}
	if err := json.Unmarshal(data, &casks); err != nil {
		return err
	}

	// Initiate transaction
	tx, err := h.Model.Client.Tx(context.Background())
	if err != nil {
		return err
	}

	// Delete existing winget apps
	if _, err := h.Model.Client.SoftwarePackage.Delete().Where(softwarepackage.Source("brew"), softwarepackage.BrewType("cask")).Exec(context.Background()); err != nil {
		return rollback(tx, err)
	}

	for _, c := range casks {
		packageID := "cask-" + c.Token
		if len(c.Name) == 0 {
			continue
		}
		packageName := c.Name[0]
		tokenSplitted := strings.Split(c.Token, "@")
		if len(tokenSplitted) > 1 {
			packageName = fmt.Sprintf("%s (%s)", c.Name[0], tokenSplitted[1])
		}

		if err := tx.SoftwarePackage.Create().SetPackageID(packageID).SetName(packageName).SetSource("brew").SetBrewType("cask").Exec(context.Background()); err != nil {
			return rollback(tx, err)
		}
	}

	// Commit the transaction.
	return tx.Commit()
}

func rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		err = fmt.Errorf("%w: %v", err, rerr)
	}
	return err
}

func getCommonAppsFolder() string {
	// Get working directory
	cwd, err := utils.GetWd()
	if err != nil {
		log.Fatal("[FATAL]: could not get working directory")
	}

	folder := filepath.Join(cwd, "tmp", "commondb")
	if strings.HasSuffix(cwd, "tmp") {
		folder = filepath.Join(cwd, "commondb")
	}

	return folder
}
