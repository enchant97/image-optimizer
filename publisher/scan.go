package publisher

import (
	"os"
	"path/filepath"

	"github.com/enchant97/image-optimizer/config"
	"github.com/enchant97/image-optimizer/core"
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
				for job := range createJobsForOriginal(appConfig, path) {
					jobChan <- ScannedJobResult{}.NewFromJob(job)
				}
			}
			return nil
		})
		close(jobChan)
	}()

	return jobChan
}
