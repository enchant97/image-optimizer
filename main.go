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

	panicOnError(filepath.WalkDir(appConfig.OriginalsPath, func(path string, info os.DirEntry, err error) error {
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

			jobs = append(jobs,
				ImageJob{
					OriginalPath:     path,
					OptimizedPath:    filepath.Join(optimizedPath, info.Name()+"@large.webp"),
					OptimizedType:    bimg.WEBP,
					OptimizedQuality: 80,
					OptimizedMaxSize: appConfig.ImageSizes.Large,
				},
				ImageJob{
					OriginalPath:     path,
					OptimizedPath:    filepath.Join(optimizedPath, info.Name()+"@medium.webp"),
					OptimizedType:    bimg.WEBP,
					OptimizedQuality: 80,
					OptimizedMaxSize: appConfig.ImageSizes.Medium,
				},
				ImageJob{
					OriginalPath:     path,
					OptimizedPath:    filepath.Join(optimizedPath, info.Name()+"@small.webp"),
					OptimizedType:    bimg.WEBP,
					OptimizedQuality: 60,
					OptimizedMaxSize: appConfig.ImageSizes.Small,
				},
				ImageJob{
					OriginalPath:     path,
					OptimizedPath:    filepath.Join(optimizedPath, info.Name()+"@thumbnail.webp"),
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
