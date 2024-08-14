package config

import "github.com/spf13/viper"

type Config struct {
	DatabaseURL string `mapstructure:"database_url"`
	HTTPPort    string `mapstructure:"http_port"`
	NatsURL     string `mapstructure:"nats_url"`
	NatsSubject string `mapstructure:"nats_subject"`
	LogLevel    string `mapstructure:"log_level"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath("cmd/config")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
