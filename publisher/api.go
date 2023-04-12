package publisher

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"

	"github.com/enchant97/image-optimizer/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func appConfigMiddleware(appConfig config.AppConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("AppConfig", appConfig)
			return next(c)
		}
	}
}

func jobPublisherMiddleware(jobPublisher JobPublisher) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("JobPublisher", jobPublisher)
			return next(c)
		}
	}
}

func postOptimiseOriginal(c echo.Context) error {
	appConfig := c.Get("AppConfig").(config.AppConfig)
	jobPublisher := c.Get("JobPublisher").(*JobPublisher)
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
	jobsChannel := createJobsForOriginal(appConfig, originalPath)
	if err := jobPublisher.PublishJobs(jobsChannel); err != nil {
		c.Logger().Error("error publishing job:", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusNoContent)
}

func RunApiServer(appConfig config.AppConfig, jobPublisher JobPublisher) error {
	e := echo.New()

	e.POST(
		"/api/optimize/:path",
		postOptimiseOriginal,
		appConfigMiddleware(appConfig),
		jobPublisherMiddleware(jobPublisher),
		middleware.BodyLimit(appConfig.Publisher.MaxUploadSize),
	)

	return e.Start(":8000")
}
