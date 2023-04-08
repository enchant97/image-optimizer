package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/enchant97/image-optimizer/config"
	"github.com/h2non/bimg"
)

func main() {
	// Parse config
	var appConfig config.AppConfig
	panicOnError(appConfig.ParseConfig())

	jobs := make([]ImageJob, 0)

	panicOnError(filepath.Walk(appConfig.OriginalsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			log.Println("adding job(s) for", info.Name())
			jobs = append(jobs,
				ImageJob{
					OriginalPath:     path,
					OptimizedPath:    filepath.Join(appConfig.OptimizedPath, info.Name()+"@large.webp"),
					OptimizedType:    bimg.WEBP,
					OptimizedQuality: 80,
					OptimizedMaxSize: appConfig.ImageSizes.Large,
				},
				ImageJob{
					OriginalPath:     path,
					OptimizedPath:    filepath.Join(appConfig.OptimizedPath, info.Name()+"@medium.webp"),
					OptimizedType:    bimg.WEBP,
					OptimizedQuality: 80,
					OptimizedMaxSize: appConfig.ImageSizes.Medium,
				},
				ImageJob{
					OriginalPath:     path,
					OptimizedPath:    filepath.Join(appConfig.OptimizedPath, info.Name()+"@small.webp"),
					OptimizedType:    bimg.WEBP,
					OptimizedQuality: 60,
					OptimizedMaxSize: appConfig.ImageSizes.Small,
				},
				ImageJob{
					OriginalPath:     path,
					OptimizedPath:    filepath.Join(appConfig.OptimizedPath, info.Name()+"@thumbnail.webp"),
					OptimizedType:    bimg.WEBP,
					OptimizedQuality: 20,
					OptimizedMaxSize: appConfig.ImageSizes.Thumbnail,
				},
			)
		}

		return nil
	}))

	for _, job := range jobs {
		panicOnError(job.Run())
	}
}
