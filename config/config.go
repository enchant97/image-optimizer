package config

type ImageSizes struct {
	Large     uint `env:"LARGE" envDefault:"2500"`
	Medium    uint `env:"MED" envDefault:"1000"`
	Small     uint `env:"SMALL" envDefault:"400"`
	Thumbnail uint `env:"THUMBNAIL" envDefault:"100"`
}

type AppConfig struct {
	ImageSizes    ImageSizes `envPrefix:"IMAGE_SIZES__"`
	OriginalsPath string     `env:"ORIGINALS_PATH,notEmpty"`
	OptimizedPath string     `env:"OPTIMIZED_PATH,notEmpty"`
}
