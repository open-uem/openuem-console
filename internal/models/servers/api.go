package servers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/open-uem/openuem_nats"
)

func GetLatestServerReleaseFromAPI(tmpDir string) (*openuem_nats.OpenUEMRelease, error) {
	latestServerReleasePath := filepath.Join(tmpDir, "latest.json")

	if _, err := os.Stat(latestServerReleasePath); os.IsNotExist(err) {
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
