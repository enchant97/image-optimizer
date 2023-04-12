package publisher

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/enchant97/image-optimizer/config"
	"github.com/enchant97/image-optimizer/core"
	"github.com/h2non/bimg"
	amqp "github.com/rabbitmq/amqp091-go"
)

func createJobsForOriginal(appConfig config.AppConfig, path string) <-chan core.ImageJob {
	jobChan := make(chan core.ImageJob)

	name := filepath.Base(path)
	srcBasePath, err := filepath.Rel(appConfig.OriginalsPath, path)
	if err != nil {
		log.Panicln("Error while getting relative path", err)
	}

	go func() {
		srcBasePath = filepath.Dir(srcBasePath)
		optimizedPath := filepath.Join(appConfig.OptimizedPath, srcBasePath)
		jobChan <- core.ImageJob{
			OriginalPath:     path,
			OptimizedPath:    filepath.Join(optimizedPath, name+"@large.webp"),
			OptimizedType:    bimg.WEBP,
			OptimizedQuality: 80,
			OptimizedMaxSize: appConfig.ImageSizes.Large,
		}
		jobChan <- core.ImageJob{
			OriginalPath:     path,
			OptimizedPath:    filepath.Join(optimizedPath, name+"@medium.webp"),
			OptimizedType:    bimg.WEBP,
			OptimizedQuality: 80,
			OptimizedMaxSize: appConfig.ImageSizes.Medium,
		}
		jobChan <- core.ImageJob{
			OriginalPath:     path,
			OptimizedPath:    filepath.Join(optimizedPath, name+"@small.webp"),
			OptimizedType:    bimg.WEBP,
			OptimizedQuality: 60,
			OptimizedMaxSize: appConfig.ImageSizes.Small,
		}
		jobChan <- core.ImageJob{
			OriginalPath:     path,
			OptimizedPath:    filepath.Join(optimizedPath, name+"@thumbnail.webp"),
			OptimizedType:    bimg.WEBP,
			OptimizedQuality: 20,
			OptimizedMaxSize: appConfig.ImageSizes.Thumbnail,
		}
		close(jobChan)
	}()

	return jobChan
}

// handle publishing new jobs to rabbitMQ,
type JobPublisher struct {
	rabbitMQ *core.RabbitMQ
	ctx      context.Context
	cancel   context.CancelFunc
}

// create a new JobPublisher
func (jp JobPublisher) New(rabbitMQ core.RabbitMQ) JobPublisher {
	jp.rabbitMQ = &rabbitMQ
	jp.ctx, jp.cancel = context.WithTimeout(context.Background(), 5*time.Second)
	return jp
}

// cancel rabbitMQ context
func (jp *JobPublisher) Cancel() {
	jp.cancel()
}

// publish a new job to rabbitMQ
func (jp *JobPublisher) PublishJob(job core.ImageJob) error {
	jobBytes, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return jp.rabbitMQ.Ch.PublishWithContext(jp.ctx,
		"",                     // exchange
		jp.rabbitMQ.Queue.Name, // routing key
		false,                  // mandatory
		false,                  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jobBytes,
		},
	)
}

// publish multiple jobs to rabbitMQ
func (jp *JobPublisher) PublishJobs(jobs <-chan core.ImageJob) error {
	for job := range jobs {
		if err := jp.PublishJob(job); err != nil {
			return err
		}
	}
	return nil
}

func Run(appConfig config.AppConfig, rabbitMQ core.RabbitMQ) error {
	jobPublisher := JobPublisher{}.New(rabbitMQ)
	defer jobPublisher.Cancel()

	if appConfig.Publisher.ScanBefore {
		log.Println("scanning input path for new images")

		for jobResult := range ScanDirectoryForJobs(appConfig) {
			if jobResult.Err != nil {
				log.Println("error scanning directory:", jobResult.Err)
			} else {
				if _, err := os.Stat(jobResult.Job.OptimizedPath); os.IsNotExist(err) {
					if err := jobPublisher.PublishJob(jobResult.Job); err != nil {
						log.Println("error publishing job:", err)
					} else {
						log.Println("published job:", jobResult.Job)
					}
				} else {
					log.Println("skipping publish of job, as already is optimized:", jobResult.Job)
				}
			}
		}
	}

	return RunApiServer(appConfig, jobPublisher)
}
