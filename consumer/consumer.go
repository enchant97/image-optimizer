package consumer

import (
	"encoding/json"
	"log"

	"github.com/enchant97/image-optimizer/config"
	"github.com/enchant97/image-optimizer/core"
)

func Run(appConfig config.ConsumerAppConfig, rabbitMQ core.RabbitMQ) error {
	jobs, err := rabbitMQ.Ch.Consume(
		rabbitMQ.Queue.Name, // queue
		"",                  // consumer
		false,               // auto-ack
		false,               // exclusive
		false,               // no-local
		false,               // no-wait
		nil,                 // args
	)
	if err != nil {
		return err
	}

	for job := range jobs {
		imageJob := core.ImageJob{}
		if err := json.Unmarshal(job.Body, &imageJob); err != nil {
			log.Println("error unmarshalling job:", err)
			core.PanicOnError(job.Nack(false, false))
			continue
		}
		log.Println("picked up new job:", imageJob)

		if !core.DoesFileExist(imageJob.OriginalPath) {
			// original not found, needed for job
			log.Println("error running job, original not found:", imageJob.OriginalPath)
			core.PanicOnError(job.Nack(false, false))
		} else if imageJob.Overwrite || !core.DoesFileExist(imageJob.OptimizedPath) {
			// optimized does not exist (or overwrite is set), so optimization can happen
			if err := imageJob.Run(); err != nil {
				// job failed
				log.Println("error running job:", err)
				core.PanicOnError(job.Nack(false, false))
			} else {
				// job finished ok
				core.PanicOnError(job.Ack(false))
				log.Println("finished job:", imageJob)
			}
		} else {
			// optimized already found, so skip
			log.Println("skipping job, optimized already exists:", imageJob.OptimizedPath)
			core.PanicOnError(job.Ack(false))
		}
	}

	return nil
}
