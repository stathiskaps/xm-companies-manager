package config

import (
	"fmt"
	"os"
)

type Config struct {
	Database DatabaseConfig
	JWT      JWTConfig
	Kafka    KafkaConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret string
}

type KafkaConfig struct {
	Brokers string
}

func Load() (*Config, error) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret: os.Getenv("JWT_SECRET"),
		},
		Kafka: KafkaConfig{
			Brokers: os.Getenv("KAFKA_BROKERS"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	required := map[string]string{
		"DB_HOST":     c.Database.Host,
		"DB_PORT":     c.Database.Port,
		"DB_USER":     c.Database.User,
		"DB_PASSWORD": c.Database.Password,
		"DB_NAME":     c.Database.Name,
		"JWT_SECRET":  c.JWT.Secret,
	}

	for name, value := range required {
		if value == "" {
			return fmt.Errorf("%s environment variable is required", name)
		}
	}

	return nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
