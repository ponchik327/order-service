package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env" env-default:"local"`
	Database   `yaml:"data_base"`
	Cache      `yaml:"cache"`
	Kafka      `yaml:"kafka"`
	HttpServer `yaml:"http_server"`
}

type Database struct {
	Adress   string `yaml:"adress" env-default:"localhost:5432"`
	User     string `yaml:"user" env-default:"user"`
	Password string `yaml:"password" env-default:"password"`
	Name     string `yaml:"name" env-default:"orders_db"`
	SslMode  bool   `yaml:"ssl_mode" env-default:"false"`
}

type Cache struct {
	Adress string        `yaml:"adress" env-default:"redis:6379"`
	Ttl    time.Duration `yaml:"ttl" env-default:"10m"`
}

type Kafka struct {
	Adress      string `yaml:"adress" env-default:"localhost:29092"`
	GroupId     string `yaml:"group_id" env-default:"order-consumer-group"`
	OffsetReset string `yaml:"offset_reset" env-default:"earliest"`
	Topic       string `yaml:"topic" env-default:"orders"`
}

type HttpServer struct {
	Adress      string        `yaml:"adress" env-default:"localhost:8081"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
