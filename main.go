package main

import (
	"encoding/json"
	"log"

	"github.com/enchant97/image-optimizer/config"
	"github.com/enchant97/image-optimizer/core"
	"github.com/enchant97/image-optimizer/publisher"
)

func runConsumer(appConfig config.AppConfig, rabbitMQ core.RabbitMQ) error {
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

func main() {
	// Parse config
	var appConfig config.AppConfig
	core.PanicOnError(appConfig.ParseConfig())

	if !appConfig.Consumer.Enable && !appConfig.Publisher.Enable {
		log.Fatalln("either (or both) 'CONSUMER' or 'PUBLISHER' must be enabled")
	}

	if appConfig.Publisher.Enable {
		go func() {
			rabbitMQ := core.RabbitMQ{}
			core.PanicOnError(rabbitMQ.Connect(appConfig.AMPQConfig))
			defer rabbitMQ.Close()
			core.PanicOnError(publisher.Run(appConfig, rabbitMQ))
		}()
	}

	if appConfig.Consumer.Enable {
		go func() {
			rabbitMQ := core.RabbitMQ{}
			core.PanicOnError(rabbitMQ.Connect(appConfig.AMPQConfig))
			defer rabbitMQ.Close()
			core.PanicOnError(runConsumer(appConfig, rabbitMQ))
		}()
	}

	var waitForever chan struct{}
	log.Printf("running. To exit press CTRL+C")
	<-waitForever
}
