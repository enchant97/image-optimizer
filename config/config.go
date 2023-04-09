package config

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

type AppConfig struct {
	AMPQConfig    AMPQConfig `envPrefix:"AMPQ__"`
	ImageSizes    ImageSizes `envPrefix:"IMAGE_SIZES__"`
	OriginalsPath string     `env:"ORIGINALS_PATH,notEmpty"`
	OptimizedPath string     `env:"OPTIMIZED_PATH,notEmpty"`
	Publisher     bool       `env:"PUBLISHER" envDefault:"false"`
	Consumer      bool       `env:"CONSUMER" envDefault:"false"`
	PublisherScan bool       `env:"PUBLISHER_SCAN" envDefault:"false"`
}
