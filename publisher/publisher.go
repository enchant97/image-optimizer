package publisher

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/enchant97/image-optimizer/config"
	"github.com/enchant97/image-optimizer/core"
	"github.com/h2non/bimg"
	"github.com/labstack/echo/v4"
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
				if _, err := os.Stat(jobResult.Job.OptimizedPath); os.IsNotExist(err) {
					if err := dispatchJob(jobResult.Job); err != nil {
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

	e := echo.New()
	e.POST("/api/optimize/:path", func(c echo.Context) error {
		path := c.Param("path")
		originalPath := filepath.Join(appConfig.OriginalsPath, path)
		for job := range createJobsForOriginal(appConfig, originalPath) {
			if err := dispatchJob(job); err != nil {
				log.Println("error publishing job:", err)
				return c.NoContent(http.StatusInternalServerError)
			}
			log.Println("published job:", job)
		}
		return c.NoContent(http.StatusNoContent)
	})

	return e.Start(":8000")
}
