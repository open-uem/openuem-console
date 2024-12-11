package servers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/doncicuto/openuem_nats"
	"github.com/doncicuto/openuem_utils"
)

func GetLatestServerRelease() (*openuem_nats.OpenUEMRelease, error) {
	cwd, err := openuem_utils.GetWd()
	if err != nil {
		return nil, err
	}

	tmpDir := filepath.Join(cwd, "tmp", "server-releases")
	if strings.HasSuffix(cwd, "tmp") {
		tmpDir = filepath.Join(cwd, "server-releases")
	}

	latestServerReleasePath := filepath.Join(tmpDir, "latest.json")

	if _, err = os.Stat(latestServerReleasePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("latest server releases json file doesn't exist, reason: %v", err)
	}

	data, err := os.ReadFile(latestServerReleasePath)
	if err != nil {
		return nil, err
	}

	r := openuem_nats.OpenUEMRelease{}
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

func GetServerReleases() (*openuem_nats.OpenUEMRelease, error) {
	cwd, err := openuem_utils.GetWd()
	if err != nil {
		return nil, err
	}

	tmpDir := filepath.Join(cwd, "tmp", "server-releases")
	if strings.HasSuffix(cwd, "tmp") {
		tmpDir = filepath.Join(cwd, "server-releases")
	}

	serverReleasesPath := filepath.Join(tmpDir, "releases.json")

	if _, err = os.Stat(serverReleasesPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("server releases json file doesn't exist, reason: %v", err)
	}

	data, err := os.ReadFile(serverReleasesPath)
	if err != nil {
		return nil, err
	}

	var r = openuem_nats.OpenUEMRelease{}
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}

	return &r, nil
}
