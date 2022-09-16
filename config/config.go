package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Database Database
	Server   Server
}

type Server struct {
	Port string `env:"SERVER_PORT"`
}

type Database struct {
	Dsn string `env:"MONGO_DSN"`
}


func NewConfig(mongoDsn, redisDsn, port string) (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	if mongoDsn != "" {
		cfg.Database.Dsn = mongoDsn
	}

	if port != "" {
		cfg.Server.Port = port
	}

	return cfg, nil
}
