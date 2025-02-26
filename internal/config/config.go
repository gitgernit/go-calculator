package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	TimeAdditionMS        int    `env:"TIME_ADDITION_MS" env-default:"100"`
	TimeSubtractionMS     int    `env:"TIME_SUBTRACTION_MS" env-default:"100"`
	TimeMultiplicationsMS int    `env:"TIME_MULTIPLICATIONS_MS" env-default:"100"`
	TimeDivisionsMS       int    `env:"TIME_DIVISIONS_MS" env-default:"100"`
	ComputingPower        int    `env:"COMPUTING_POWER" env-default:"4"`
	OrchestratorPort      int    `env:"ORCHESTRATOR_PORT" env-default:"8080"`
	OrchestratorHost      string `env:"ORCHESTRATOR_HOST" env-default:"0.0.0.0"`
	PollingIntervalMS     int    `env:"POLLING_INTERVAL" env-default:"250"`
}

func New() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	cfg := Config{}
	err := cleanenv.ReadEnv(&cfg)

	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
