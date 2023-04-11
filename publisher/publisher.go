package publisher

import (
	"bytes"
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
	"github.com/labstack/echo/v4/middleware"
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

		if len(appConfig.Publisher.ApiKey) != 0 {
			// if api key is set, validate against request header
			// get base64 encoded key from header
			rawApiKey := c.Request().Header.Get("X-Api-Key")
			if len(rawApiKey) == 0 {
				// no key provided
				return c.NoContent(http.StatusUnauthorized)
			}
			// decode key from base64
			apiKey := config.Base64Decoded{}
			if err := apiKey.UnmarshalText([]byte(rawApiKey)); err != nil {
				// invalid base64
				return c.NoContent(http.StatusUnauthorized)
			}
			// compare key
			if !appConfig.Publisher.CompareApiKey(apiKey) {
				// invalid key
				return c.NoContent(http.StatusUnauthorized)
			}
		}

		originalPath := filepath.Join(appConfig.OriginalsPath, path)

		var bodyBytes bytes.Buffer

		if _, err := bodyBytes.ReadFrom(c.Request().Body); err != nil {
			return c.NoContent(http.StatusBadRequest)
		}

		if bodyBytes.Len() == 0 {
			// if no body is provided, assume the file already exists
			if _, err := os.Stat(originalPath); os.IsNotExist(err) {
				return c.NoContent(http.StatusNotFound)
			}
		} else {
			// if body is provided, assume the file is being uploaded
			// don't allow overwriting existing files
			if _, err := os.Stat(filepath.Dir(originalPath)); os.IsExist(err) {
				return c.NoContent(http.StatusConflict)
			}
			// ensure original directory path exists
			if err := os.MkdirAll(filepath.Dir(originalPath), os.ModePerm); err != nil {
				return c.NoContent(http.StatusInternalServerError)
			}
			// write to disk
			if err := os.WriteFile(originalPath, bodyBytes.Bytes(), 0644); err != nil {
				return c.NoContent(http.StatusInternalServerError)
			}
		}
		// publish optimization jobs
		for job := range createJobsForOriginal(appConfig, originalPath) {
			if err := dispatchJob(job); err != nil {
				log.Println("error publishing job:", err)
				return c.NoContent(http.StatusInternalServerError)
			}
			log.Println("published job:", job)
		}

		return c.NoContent(http.StatusNoContent)
	}, middleware.BodyLimit(appConfig.Publisher.MaxUploadSize))

	return e.Start(":8000")
}
