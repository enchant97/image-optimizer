package config

import "crypto/subtle"

type AMPQConfig struct {
	URI       string `yaml:"uri" validate:"required,uri"`
	QueueName string `yaml:"queueName" validate:"required"`
}

type StorageConfig struct {
	Originals string `yaml:"originals" validate:"required"`
	Optimized string `yaml:"optimized" validate:"required"`
}

type ConsumerConfig struct {
	Enable bool `yaml:"enable"`
}

type OptimizationJobFormatConfig struct {
	Enable  bool `yaml:"enable"`
	Quality uint `yaml:"quality" validate:"required"`
}

type OptimizationJobFormatsConfig struct {
	JPEG OptimizationJobFormatConfig `yaml:"jpeg"`
	WebP OptimizationJobFormatConfig `yaml:"webp"`
	AVIF OptimizationJobFormatConfig `yaml:"avif"`
}

type OptimizationJobConfig struct {
	Name     string                       `yaml:"name" validate:"required"`
	MaxWidth uint                         `yaml:"maxWidth" validate:"required"`
	Formats  OptimizationJobFormatsConfig `yaml:"formats" validate:"required"`
}

type PublisherConfig struct {
	Enable        bool                    `yaml:"enable"`
	ScanBefore    bool                    `yaml:"scanBefore"`
	MaxUploadSize string                  `yaml:"maxUploadSize" validate:"required"`
	ApiKey        Base64Decoded           `yaml:"apiKey" validate:"required"`
	Optimizations []OptimizationJobConfig `yaml:"optimizations"`
}

func (c *PublisherConfig) CompareApiKey(otherKey Base64Decoded) bool {
	return subtle.ConstantTimeCompare(c.ApiKey, otherKey) == 1
}

type AppConfig struct {
	AMPQConfig AMPQConfig      `yaml:"ampq" validate:"required"`
	Storage    StorageConfig   `yaml:"storage" validate:"required"`
	Consumer   ConsumerConfig  `yaml:"consumer" validate:"required"`
	Publisher  PublisherConfig `yaml:"publisher" validate:"required"`
}
