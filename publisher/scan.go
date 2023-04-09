package publisher

import (
	"os"
	"path/filepath"

	"github.com/enchant97/image-optimizer/config"
	"github.com/enchant97/image-optimizer/core"
	"github.com/h2non/bimg"
)

type ScannedJobResult struct {
	Job core.ImageJob
	Err error
}

func (r ScannedJobResult) NewFromJob(job core.ImageJob) ScannedJobResult {
	return ScannedJobResult{
		Job: job,
	}
}

func (r ScannedJobResult) NewFromError(err error) ScannedJobResult {
	return ScannedJobResult{
		Err: err,
	}
}

func ScanDirectoryForJobs(appConfig config.AppConfig) <-chan ScannedJobResult {
	jobChan := make(chan ScannedJobResult)

	go func() {
		filepath.WalkDir(appConfig.OriginalsPath, func(path string, info os.DirEntry, err error) error {
			if err != nil {
				jobChan <- ScannedJobResult{}.NewFromError(err)
				return err
			}

			if !info.IsDir() {
				srcBasePath, err := filepath.Rel(appConfig.OriginalsPath, path)
				if err != nil {
					jobChan <- ScannedJobResult{}.NewFromError(err)
					return err
				}
				srcBasePath = filepath.Dir(srcBasePath)
				optimizedPath := filepath.Join(appConfig.OptimizedPath, srcBasePath)

				jobChan <- ScannedJobResult{}.NewFromJob(core.ImageJob{
					OriginalPath:     path,
					OptimizedPath:    filepath.Join(optimizedPath, info.Name()+"@large.webp"),
					OptimizedType:    bimg.WEBP,
					OptimizedQuality: 80,
					OptimizedMaxSize: appConfig.ImageSizes.Large,
				})
				jobChan <- ScannedJobResult{}.NewFromJob(core.ImageJob{
					OriginalPath:     path,
					OptimizedPath:    filepath.Join(optimizedPath, info.Name()+"@medium.webp"),
					OptimizedType:    bimg.WEBP,
					OptimizedQuality: 80,
					OptimizedMaxSize: appConfig.ImageSizes.Medium,
				})
				jobChan <- ScannedJobResult{}.NewFromJob(core.ImageJob{
					OriginalPath:     path,
					OptimizedPath:    filepath.Join(optimizedPath, info.Name()+"@small.webp"),
					OptimizedType:    bimg.WEBP,
					OptimizedQuality: 60,
					OptimizedMaxSize: appConfig.ImageSizes.Small,
				})
				jobChan <- ScannedJobResult{}.NewFromJob(core.ImageJob{
					OriginalPath:     path,
					OptimizedPath:    filepath.Join(optimizedPath, info.Name()+"@thumbnail.webp"),
					OptimizedType:    bimg.WEBP,
					OptimizedQuality: 20,
					OptimizedMaxSize: appConfig.ImageSizes.Thumbnail,
				})
			}
			return nil
		})
		close(jobChan)
	}()

	return jobChan
}
