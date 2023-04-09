package consumer

import (
	"encoding/json"
	"log"

	"github.com/enchant97/image-optimizer/config"
	"github.com/enchant97/image-optimizer/core"
)

func Run(appConfig config.AppConfig, rabbitMQ core.RabbitMQ) error {
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
			job.Nack(false, false)
			continue
		}
		log.Println("picked up new job:", imageJob)
		if err := imageJob.Run(); err != nil {
			log.Println("error running job, requeuing:", err)
			job.Nack(false, true)
			continue
		}
		core.PanicOnError(job.Ack(false))
		log.Println("finished job:", imageJob)
	}

	return nil
}
