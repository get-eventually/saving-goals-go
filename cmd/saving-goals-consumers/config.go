package main

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Database Database
	Jaeger   Jaeger
	Kafka    Kafka
}

type Kafka struct {
	Host string `default:"localhost"`
	Port uint16 `default:"29092"`
}

func (k Kafka) Addr() string {
	return fmt.Sprintf("%s:%d", k.Host, k.Port)
}

type Jaeger struct {
	Host string `default:"localhost"`
	Port uint16 `default:"6831"`
}

func (j Jaeger) Addr() string {
	return fmt.Sprintf("%s:%d", j.Host, j.Port)
}

type Database struct {
	Username string `default:"saving_goals_api"`
	Password string `default:"saving_goals_api"`
	Database string `default:"saving_goals"`
	Host     string `default:"localhost"`
	Port     uint16 `default:"5432"`
}

func (db Database) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		db.Username,
		db.Password,
		db.Host,
		db.Port,
		db.Database,
	)
}

func ParseConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("", &config); err != nil {
		return Config{}, fmt.Errorf("config: failed to parse: %w", err)
	}

	return config, nil
}
