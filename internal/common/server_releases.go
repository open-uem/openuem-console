package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/doncicuto/openuem_utils"
)

func (w *Worker) GetServerReleases() error {
	settings, err := w.Model.GetGeneralSettings()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://releases.openuem.eu/api?action=latestServerRelease&channel=%s", settings.UpdateChannel)

	body, err := openuem_utils.QueryReleasesEndpoint(url)
	if err != nil {
		return err
	}

	latestServerReleaseFilePath := filepath.Join(w.ServerReleasesFolder, "latest.json")

	if err := os.WriteFile(latestServerReleaseFilePath, body, 0660); err != nil {
		return err
	}

	url = fmt.Sprintf("https://releases.openuem.eu/api?action=allServerReleases&channel=%s", settings.UpdateChannel)

	body, err = openuem_utils.QueryReleasesEndpoint(url)
	if err != nil {
		return err
	}

	allServerReleasesFilePath := filepath.Join(w.ServerReleasesFolder, "releases.json")

	if err := os.WriteFile(allServerReleasesFilePath, body, 0660); err != nil {
		return err
	}

	return nil
}
