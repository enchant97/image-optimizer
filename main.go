package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/enchant97/image-optimizer/config"
	"github.com/h2non/bimg"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Conn  *amqp.Connection
	Ch    *amqp.Channel
	Queue amqp.Queue
}

func (r *RabbitMQ) Close() error {
	if err := r.Ch.Close(); err != nil {
		return err
	}
	if err := r.Conn.Close(); err != nil {
		return err
	}
	return nil
}

func (r *RabbitMQ) Connect(config config.AMPQConfig) error {
	conn, err := amqp.Dial(config.URI)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	q, err := ch.QueueDeclare(
		config.QueueName, // name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return err
	}
	r.Conn = conn
	r.Ch = ch
	r.Queue = q
	return nil
}

func runPublisher(appConfig config.AppConfig, rabbitMQ RabbitMQ) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dispatchJob := func(job ImageJob) error {
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

		return filepath.WalkDir(appConfig.OriginalsPath, func(path string, info os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				srcBasePath, err := filepath.Rel(appConfig.OriginalsPath, path)
				if err != nil {
					return err
				}
				srcBasePath = filepath.Dir(srcBasePath)
				optimizedPath := filepath.Join(appConfig.OptimizedPath, srcBasePath)

				log.Println("adding job(s) for", path)

				dispatchJob(ImageJob{
					OriginalPath:     path,
					OptimizedPath:    filepath.Join(optimizedPath, info.Name()+"@large.webp"),
					OptimizedType:    bimg.WEBP,
					OptimizedQuality: 80,
					OptimizedMaxSize: appConfig.ImageSizes.Large,
				})
				dispatchJob(ImageJob{
					OriginalPath:     path,
					OptimizedPath:    filepath.Join(optimizedPath, info.Name()+"@medium.webp"),
					OptimizedType:    bimg.WEBP,
					OptimizedQuality: 80,
					OptimizedMaxSize: appConfig.ImageSizes.Medium,
				})
				dispatchJob(ImageJob{
					OriginalPath:     path,
					OptimizedPath:    filepath.Join(optimizedPath, info.Name()+"@small.webp"),
					OptimizedType:    bimg.WEBP,
					OptimizedQuality: 60,
					OptimizedMaxSize: appConfig.ImageSizes.Small,
				})
				dispatchJob(ImageJob{
					OriginalPath:     path,
					OptimizedPath:    filepath.Join(optimizedPath, info.Name()+"@thumbnail.webp"),
					OptimizedType:    bimg.WEBP,
					OptimizedQuality: 20,
					OptimizedMaxSize: appConfig.ImageSizes.Thumbnail,
				})
			}

			return nil
		})
	}

	return nil
}

func runConsumer(appConfig config.AppConfig, rabbitMQ RabbitMQ) error {
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
		imageJob := ImageJob{}
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
		panicOnError(job.Ack(false))
		log.Println("finished job:", imageJob)
	}

	return nil
}

func main() {
	// Parse config
	var appConfig config.AppConfig
	panicOnError(appConfig.ParseConfig())

	if !appConfig.Consumer.Enable && !appConfig.Publisher.Enable {
		log.Fatalln("either (or both) 'CONSUMER' or 'PUBLISHER' must be enabled")
	}

	if appConfig.Publisher.Enable {
		go func() {
			rabbitMQ := RabbitMQ{}
			panicOnError(rabbitMQ.Connect(appConfig.AMPQConfig))
			defer rabbitMQ.Close()
			panicOnError(runPublisher(appConfig, rabbitMQ))
		}()
	}

	if appConfig.Consumer.Enable {
		go func() {
			rabbitMQ := RabbitMQ{}
			panicOnError(rabbitMQ.Connect(appConfig.AMPQConfig))
			defer rabbitMQ.Close()
			panicOnError(runConsumer(appConfig, rabbitMQ))
		}()
	}

	var waitForever chan struct{}
	log.Printf("running. To exit press CTRL+C")
	<-waitForever
}
