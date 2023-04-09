package main

import (
	"os"
	"path/filepath"

	"github.com/h2non/bimg"
)

// Represents a single image processing job
type ImageJob struct {
	OriginalPath     string
	OptimizedPath    string
	OptimizedMaxSize uint
	OptimizedType    bimg.ImageType
	OptimizedQuality uint
}

func (job *ImageJob) Run() error {
	// read image into memory
	image, err := bimg.Read(job.OriginalPath)
	if err != nil {
		return err
	}

	// ensure directory structure exists
	dstBasePath := filepath.Dir(job.OptimizedPath)
	if err := os.MkdirAll(dstBasePath, os.ModePerm); err != nil {
		return err
	}

	loadedImage := bimg.NewImage(image)

	// get original image size, used later to determine if it needs resizing
	originalSize, err := loadedImage.Size()
	if err != nil {
		return err
	}

	// process image
	optimisedImage, err := loadedImage.Process(bimg.Options{
		Width:         intMin(originalSize.Width, int(job.OptimizedMaxSize)),
		Type:          job.OptimizedType,
		StripMetadata: true,
		Quality:       int(job.OptimizedQuality),
	})
	if err != nil {
		return err
	}

	// write image to disk
	return bimg.Write(job.OptimizedPath, optimisedImage)
}
