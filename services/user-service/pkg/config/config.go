package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
)

type Config interface {
	App() App
	Database() Database
	String() string
}

type Database struct {
	DSN  string `validate:"required"`
	Name string `validate:"required"`
}

type App struct {
	Port string `validate:"required"`
}

type config struct {
	AppCfg         App
	DatabaseCfg Database
}

func (c *config) App() App      { return c.AppCfg }
func (c *config) Database() Database { return c.DatabaseCfg }

func (c *config) String() string {
	jsonBytes, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		log.Fatalf("Failed to convert config to JSON: %v", err)
	}
	return string(jsonBytes)
}

func InitConfig() (Config, error) {
	cfg := &config{
		AppCfg: App{
			Port: getEnv("GRPC_ADDR", ""),
		},
		DatabaseCfg: Database{
			DSN:  getEnv("DATABASE_DSN", ""),
			Name: getEnv("DATABASE_NAME", ""),
		},
	}

	if err := validator.New().Struct(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
