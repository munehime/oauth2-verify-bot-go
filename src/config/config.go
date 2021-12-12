package config

import (
	"log"

	"github.com/spf13/viper"
)

var config *viper.Viper

type Config struct {
}

func Setup(env string) {
	config = viper.New()
	config.SetConfigType("yaml")
	config.SetConfigName(env)
	config.AddConfigPath("config/")
	config.AddConfigPath(".")
	err := config.ReadInConfig()
	if err != nil {
		log.Fatalf("Error on parsing configuration file %+v", err)
	}
}

func GetConfig() *viper.Viper {
	return config
}
