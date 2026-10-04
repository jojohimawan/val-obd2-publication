package config

import (
	"os"

	"github.com/Netflix/go-env"
	"github.com/joho/godotenv"
)

type Config struct {
	KafkaBrokerURL      string `env:"KAFKA_BROKER_URL,required=true"`
	Vin                 string `env:"VIN,default=C304TEST"`
	SchemaRegistryURL   string `env:"SCHEMA_REGISTRY_URL,required=true"`
	SerialPort          string `env:"SERIAL_PORT,required=true"`
	CanNetwork          string `env:"CAN_NETWORK, required=true"`
	CanNetworkAddress   string `env:"CAN_NETWORK_ADDRESS, required=true"`
	MongoDBURI          string `env:"MONGODB_URI, required=true"`
	MongoDBDatabase     string `env:"MONGODB_DATABASE, required=true"`
	MongoDBCollection   string `env:"MONGODB_COLLECTION, required=true"`
	MongoDBAuthUser     string `env:"MONGODB_AUTH_USER, required=true"`
	MongoDBAuthPassword string `env:"MONGODB_AUTH_PASSWORD, required=true"`

	Environment string `env:"APP_ENV,default=development"`

	Extras env.EnvSet
}

func Load() (*Config, error) {
	var cfg Config
	log := ComponentLogger("config")

	if os.Getenv("APP_ENV") == "development" || os.Getenv("APP_ENV") == "" {
		if err := godotenv.Load(); err != nil {
			log.Error("no env file found", "message", err)
		}
	}

	extras, err := env.UnmarshalFromEnviron(&cfg)
	if err != nil {
		return nil, err
	}
	cfg.Extras = extras

	return &cfg, nil
}
