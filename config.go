package main

import (
	"crypto/sha256"
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Environment string `mapstructure:"ENVIRONMENT"`
	Host        string `mapstructure:"HOST"`
	Port        string `mapstructure:"PORT"`

	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBUsername string `mapstructure:"DB_USERNAME"`
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBName     string `mapstructure:"DB_NAME"`
	DBOptions  string `mapstructure:"DB_OPTIONS"`
	DBUrl      string `mapstructure:"DB_URL"`

	ChrysalisAPIKey string `mapstructure:"CHRYSALIS_API_KEY"`


	DecodedAPIKey []byte
}

func (c *Config) GetDBUrl() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?%s",
		c.DBUsername,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
		c.DBOptions,
	)
}

func LoadConfig(v *viper.Viper, path string) (config Config) {
	v.AddConfigPath(".")
	v.SetConfigName(path)
	v.SetConfigType("env")

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("config: %v\n", err)
		return
	}
	if err := v.Unmarshal(&config); err != nil {
		log.Fatalf("config: %v\n", err)
		return
	}

	key := sha256.Sum256([]byte(config.ChrysalisAPIKey))
	config.DecodedAPIKey = key[:]

	return
}
