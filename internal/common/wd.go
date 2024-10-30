package common

import (
	"log"
	"os"
	"path/filepath"
)

func GetWd() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		log.Printf("[ERROR]:could not get executable info: %v", err)
		return "", err
	}
	return filepath.Dir(ex), nil
}
