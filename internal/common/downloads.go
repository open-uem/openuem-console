package common

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/go-co-op/gocron/v2"
)

func (w *Worker) StartDownloadCleanJob() error {
	var err error

	// Create task
	_, err = w.TaskScheduler.NewJob(
		gocron.DurationJob(
			time.Duration(time.Duration(5*time.Minute)),
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
						fileName := path.Join(w.DownloadDir, d.Name())
						info, err := os.Stat(fileName)
						if err != nil {
							log.Printf("[ERROR]: could not read the download directory contents: %v", err)
							continue
						}
						if info.ModTime().Before(time.Now().Add(-1 * time.Minute)) {
							os.RemoveAll(fileName)
						} else {
							fmt.Println(info.ModTime(), info.ModTime().Before(time.Now().Add(-1*time.Minute)))
						}
					}
				}
			},
		),
	)
	if err != nil {
		log.Printf("[FATAL]: could not start the download directory clean job: %v", err)
		return err
	}
	log.Println("[INFO]: download directory clean job has been scheduled every 5 minutes")
	return nil
}
