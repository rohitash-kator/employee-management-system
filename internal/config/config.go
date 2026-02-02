package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type Config struct {
	AppEnv string

	HTTPAddr string

	MongoURI            string
	MongoDB             string
	MongoConnectTimeout time.Duration

	RequestTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv: "local",

		HTTPAddr: ":8080",

		MongoDB:             "employee_mgmt",
		MongoConnectTimeout: 10 * time.Second,
		RequestTimeout:      5 * time.Second,
	}

	if v := os.Getenv("APP_ENV"); v != "" {
		cfg.AppEnv = v
	}
	if v := os.Getenv("HTTP_ADDR"); v != "" {
		cfg.HTTPAddr = v
	}

	if v := os.Getenv("MONGO_URI"); v != "" {
		cfg.MongoURI = v
	}
	if v := os.Getenv("MONGO_DB"); v != "" {
		cfg.MongoDB = v
	}
	if v := os.Getenv("MONGO_CONNECT_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("parse MONGO_CONNECT_TIMEOUT: %w", err)
		}
		cfg.MongoConnectTimeout = d
	}
	if v := os.Getenv("REQUEST_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("parse REQUEST_TIMEOUT: %w", err)
		}
		cfg.RequestTimeout = d
	}

	if cfg.MongoURI == "" {
		return Config{}, errors.New("MONGO_URI is required")
	}

	return cfg, nil
}
