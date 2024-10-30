package common

import (
	"log"
	"time"

	"github.com/doncicuto/openuem-console/internal/models"
	"github.com/go-co-op/gocron/v2"
)

func (w *Worker) StartDBConnectJob() error {
	var err error

	w.Model, err = models.New(w.DBUrl)
	if err == nil {
		log.Println("[INFO]: connection established with database")
		return nil
	}
	log.Printf("[ERROR]: could not connect with database %v", err)

	// Create task
	w.DBConnectJob, err = w.TaskScheduler.NewJob(
		gocron.DurationJob(
			time.Duration(time.Duration(2*time.Minute)),
		),
		gocron.NewTask(
			func() {
				w.Model, err = models.New(w.DBUrl)
				if err != nil {
					log.Printf("[ERROR]: could not connect with database %v", err)
					return
				}
				log.Println("[INFO]: connection established with database")

				if err := w.TaskScheduler.RemoveJob(w.DBConnectJob.ID()); err != nil {
					return
				}
			},
		),
	)
	if err != nil {
		log.Fatalf("[FATAL]: could not start the DB connect job: %v", err)
		return err
	}
	log.Printf("[INFO]: new DB connect job has been scheduled every %d minutes", 2)
	return nil
}
