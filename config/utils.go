package config

import (
	"encoding/base64"
	"errors"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

func (appConfig *ConsumerAppConfig) ParseConfig() error {
	var rawConfig []byte
	var err error

	if customPath, isSet := os.LookupEnv("IO_CONSUMER_CONFIG"); isSet {
		rawConfig, err = os.ReadFile(customPath)
		if err != nil {
			return err
		}
	} else {
		rawConfig, err = os.ReadFile("config.yaml")
		if err != nil {
			return err
		}
	}

	if err := yaml.Unmarshal(rawConfig, &appConfig); err != nil {
		return err
	}

	validate := validator.New()
	return validate.Struct(appConfig)
}

func (appConfig *PublisherAppConfig) ParseConfig() error {
	var rawConfig []byte
	var err error

	if customPath, isSet := os.LookupEnv("IO_PRODUCER_CONFIG"); isSet {
		rawConfig, err = os.ReadFile(customPath)
		if err != nil {
			return err
		}
	} else {
		rawConfig, err = os.ReadFile("config.yaml")
		if err != nil {
			return err
		}
	}

	if err := yaml.Unmarshal(rawConfig, &appConfig); err != nil {
		return err
	}

	validate := validator.New()
	return validate.Struct(appConfig)
}

type Base64Decoded []byte

func (b *Base64Decoded) UnmarshalText(text []byte) error {
	decoded, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return errors.New("cannot decode base64 string")
	}
	*b = decoded
	return nil
}
