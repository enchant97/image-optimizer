package main

import (
	"log"

	"github.com/enchant97/image-optimizer/config"
	"github.com/enchant97/image-optimizer/consumer"
	"github.com/enchant97/image-optimizer/core"
	"github.com/enchant97/image-optimizer/publisher"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err == nil {
		log.Println("loaded environment variables from .env file")
	}
	// Parse config
	var appConfig config.AppConfig
	core.PanicOnError(appConfig.ParseConfig())

	if !appConfig.Consumer.Enable && !appConfig.Publisher.Enable {
		log.Fatalln("either (or both) 'CONSUMER' or 'PUBLISHER' must be enabled")
	}

	if appConfig.Publisher.Enable {
		go func() {
			rabbitMQ := core.RabbitMQ{}
			core.PanicOnError(rabbitMQ.Connect(appConfig.AMPQConfig))
			defer rabbitMQ.Close()
			core.PanicOnError(publisher.Run(appConfig, rabbitMQ))
		}()
	}

	if appConfig.Consumer.Enable {
		go func() {
			rabbitMQ := core.RabbitMQ{}
			core.PanicOnError(rabbitMQ.Connect(appConfig.AMPQConfig))
			defer rabbitMQ.Close()
			core.PanicOnError(consumer.Run(appConfig, rabbitMQ))
		}()
	}

	var waitForever chan struct{}
	log.Printf("running. To exit press CTRL+C")
	<-waitForever
}
