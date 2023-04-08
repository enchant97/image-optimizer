package main

import "github.com/h2non/bimg"

type ImageJob struct {
	OriginalPath     string
	OptimizedPath    string
	OptimizedMaxSize uint
	OptimizedType    bimg.ImageType
	OptimizedQuality uint
}

func (job *ImageJob) Run() error {
	image, err := bimg.Read(job.OriginalPath)
	if err != nil {
		return err
	}
	loadedImage := bimg.NewImage(image)

	originalSize, err := loadedImage.Size()
	if err != nil {
		return err
	}

	optimisedImage, err := loadedImage.Process(bimg.Options{
		Width:         intMin(originalSize.Width, int(job.OptimizedMaxSize)),
		Type:          job.OptimizedType,
		StripMetadata: true,
		Quality:       int(job.OptimizedQuality),
	})
	if err != nil {
		return err
	}

	return bimg.Write(job.OptimizedPath, optimisedImage)
}
