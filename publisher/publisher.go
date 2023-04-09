package publisher

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/enchant97/image-optimizer/config"
	"github.com/enchant97/image-optimizer/core"
	amqp "github.com/rabbitmq/amqp091-go"
)

func Run(appConfig config.AppConfig, rabbitMQ core.RabbitMQ) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dispatchJob := func(job core.ImageJob) error {
		jobBytes, err := json.Marshal(job)
		if err != nil {
			return err
		}
		return rabbitMQ.Ch.PublishWithContext(ctx,
			"",                  // exchange
			rabbitMQ.Queue.Name, // routing key
			false,               // mandatory
			false,               // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        jobBytes,
			},
		)
	}

	if appConfig.Publisher.ScanBefore {
		log.Println("scanning input path for new images")

		for jobResult := range ScanDirectoryForJobs(appConfig) {
			if jobResult.Err != nil {
				log.Println("error scanning directory:", jobResult.Err)
			} else {
				if err := dispatchJob(jobResult.Job); err != nil {
					log.Println("error dispatching job:", err)
				}
			}
			log.Println("published job:", jobResult.Job)
		}
	}

	return nil
}