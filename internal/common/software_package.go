package common

import (
	"log"

	"github.com/open-uem/nats"
	models "github.com/open-uem/openuem-console/internal/models/winget"
)

func (w *Worker) StartCommonPackagesDBJob() error {
	commonDB, err := models.OpenCommonDB(w.CommonSoftwareDBFolder)
	if err != nil {
		log.Printf("[ERROR]: could not get common software database, reason: %v", err)
		return err
	}
	defer func() {
		if err := commonDB.Close(); err != nil {
			log.Printf("[ERROR]: could not close common software database, reason: %v", err)
		}
	}()

	models.CreateCommonSoftwareTable(commonDB)

	if err := models.DeleteCommonSoftwareTable(commonDB); err != nil {
		log.Println("[ERROR]: could not delete common software table")
		return err
	}

	flatpakDB, err := models.OpenFlatpakDB(w.FlatpakDBFolder)
	if err != nil {
		log.Println("[INFO]: could not get flatpak database")
	} else {
		rows, err := flatpakDB.Query(`SELECT DISTINCT id, name FROM apps`)
		if err != nil {
			log.Println("[INFO]: could not query flatpak apps")
		} else {
			packages := []nats.SoftwarePackage{}
			defer rows.Close()
			for rows.Next() {
				var p nats.SoftwarePackage
				err := rows.Scan(&p.ID, &p.Name)
				if err != nil {
					log.Printf("[ERROR]: could not get flatpak db row, reason: %v", err)
					break
				}
				packages = append(packages, p)
			}

			if err := models.InsertCommonSoftware(commonDB, packages, "flatpak"); err != nil {
				log.Printf("[ERROR]: could not insert flatpak apps to common software database, reason: %v", err)
			}
		}
	}

	wingetDB, err := models.OpenWingetDB(w.WinGetDBFolder)
	if err != nil {
		log.Println("[INFO]: could not get winget database")
	} else {
		rows, err := wingetDB.Query(`SELECT DISTINCT ids.id as id, names.name AS name FROM manifest LEFT JOIN ids ON manifest.id = ids.rowid LEFT JOIN names ON manifest.name = names.rowid`)
		if err != nil {
			log.Println("[INFO]: could not query winget apps")
		} else {
			packages := []nats.SoftwarePackage{}
			defer rows.Close()
			for rows.Next() {
				var p nats.SoftwarePackage
				err := rows.Scan(&p.ID, &p.Name)
				if err != nil {
					log.Printf("[ERROR]: could not get winget db row, reason: %v", err)
					break
				}
				packages = append(packages, p)
			}

			if err := models.InsertCommonSoftware(commonDB, packages, "winget"); err != nil {
				log.Printf("[ERROR]: could not insert winget apps to common software database, reason: %v", err)
				return err
			}
		}
	}

	log.Println("[INFO]: the common software database has been created")
	return nil
}
