package common

import (
	"log"
	"os"
	"path"
	"time"

	"github.com/go-co-op/gocron/v2"
)

func (w *Worker) StartDownloadCleanJob() error {
	var err error

	// Create task for running the agent
	_, err = w.TaskScheduler.NewJob(
		gocron.DurationJob(
			time.Duration(time.Duration(60*time.Minute)),
		),
		gocron.NewTask(
			func() {
				if _, err := os.Stat(w.DownloadDir); !os.IsNotExist(err) {
					dir, err := os.ReadDir(w.DownloadDir)
					if err != nil {
						log.Printf("[ERROR]: could not read the download directory contents: %v", err)
						return
					}
					for _, d := range dir {
						os.RemoveAll(path.Join([]string{w.DownloadDir, d.Name()}...))
					}
				}
			},
		),
	)
	if err != nil {
		log.Printf("[FATAL]: could not start the download directory clean job: %v", err)
		return err
	}
	log.Println("[INFO]: download directory clean job has been scheduled every 60 minutes")
	return nil
}
