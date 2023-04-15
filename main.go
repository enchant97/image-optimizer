package main

import (
	"log"
	"os"
	"strings"

	"github.com/enchant97/image-optimizer/config"
	"github.com/enchant97/image-optimizer/consumer"
	"github.com/enchant97/image-optimizer/core"
	"github.com/enchant97/image-optimizer/publisher"
)

func main() {
	var consumerConfig *config.ConsumerAppConfig
	var publisherConfig *config.PublisherAppConfig

	if mode, isSet := os.LookupEnv("IO_MODE"); isSet {
		mode = strings.ToUpper(mode)
		switch mode {
		case "BOTH":
			consumerConfig = &config.ConsumerAppConfig{}
			publisherConfig = &config.PublisherAppConfig{}
			core.PanicOnError(consumerConfig.ParseConfig())
			core.PanicOnError(publisherConfig.ParseConfig())
		case "CONSUMER":
			consumerConfig = &config.ConsumerAppConfig{}
			core.PanicOnError(consumerConfig.ParseConfig())
		case "PUBLISHER":
			publisherConfig = &config.PublisherAppConfig{}
			core.PanicOnError(publisherConfig.ParseConfig())
		}
	} else {
		log.Fatalln("IO_MODE must be set, either 'CONSUMER', 'PUBLISHER' or 'BOTH'")
	}

	if publisherConfig != nil {
		go func() {
			rabbitMQ := core.RabbitMQ{}
			core.PanicOnError(rabbitMQ.Connect(publisherConfig.AMPQConfig))
			defer rabbitMQ.Close()
			core.PanicOnError(publisher.Run(*publisherConfig, rabbitMQ))
		}()
	}

	if consumerConfig != nil {
		go func() {
			rabbitMQ := core.RabbitMQ{}
			core.PanicOnError(rabbitMQ.Connect(consumerConfig.AMPQConfig))
			defer rabbitMQ.Close()
			core.PanicOnError(consumer.Run(*consumerConfig, rabbitMQ))
		}()
	}

	var waitForever chan struct{}
	log.Printf("running. To exit press CTRL+C")
	<-waitForever
}
