package config

import "crypto/subtle"

type AMPQConfig struct {
	URI       string `env:"URI,notEmpty"`
	QueueName string `env:"QUEUE_NAME,notEmpty"`
}

type ImageSizes struct {
	Large     uint `env:"LARGE" envDefault:"2500"`
	Medium    uint `env:"MED" envDefault:"1000"`
	Small     uint `env:"SMALL" envDefault:"400"`
	Thumbnail uint `env:"THUMBNAIL" envDefault:"100"`
}

type PublisherConfig struct {
	Enable        bool          `env:"ENABLE" envDefault:"false"`
	ScanBefore    bool          `env:"SCAN_BEFORE" envDefault:"false"`
	MaxUploadSize string        `env:"MAX_UPLOAD_SIZE" envDefault:"16M"`
	ApiKey        Base64Decoded `env:"API_KEY"`
}

func (c *PublisherConfig) CompareApiKey(otherKey Base64Decoded) bool {
	return subtle.ConstantTimeCompare(c.ApiKey, otherKey) == 1
}

type ConsumerConfig struct {
	Enable bool `env:"ENABLE" envDefault:"false"`
}

type AppConfig struct {
	AMPQConfig    AMPQConfig      `envPrefix:"AMPQ__"`
	ImageSizes    ImageSizes      `envPrefix:"IMAGE_SIZES__"`
	Publisher     PublisherConfig `envPrefix:"PUBLISHER__"`
	Consumer      ConsumerConfig  `envPrefix:"CONSUMER__"`
	OriginalsPath string          `env:"ORIGINALS_PATH,notEmpty"`
	OptimizedPath string          `env:"OPTIMIZED_PATH,notEmpty"`
}
