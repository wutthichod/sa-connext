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
	RABBITMQ() RABBITMQ
	Database() Database
	JWT() JWT
	Notification() Notification
	String() string
}

type App struct {
	Gateway      string
	User         string
	Chat         string
	Notification string
	Event        string
	Organizer    string
}

type RABBITMQ struct {
	URI string
}

type Database struct {
	DSN  string
	Name string
}

type JWT struct {
	Token string
}

type Notification struct {
	Email   string
	EmailPW string
}

type config struct {
	AppCfg      App
	DatabaseCfg Database
	RabbitMqCfg RABBITMQ
	JwtCfg      JWT
	NotiCfg     Notification
}

func (c *config) App() App                   { return c.AppCfg }
func (c *config) Database() Database         { return c.DatabaseCfg }
func (c *config) JWT() JWT                   { return c.JwtCfg }
func (c *config) RABBITMQ() RABBITMQ         { return c.RabbitMqCfg }
func (c *config) Notification() Notification { return c.NotiCfg }

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
			Gateway:      getEnv("GATEWAY_ADDR", ""),
			User:         getEnv("USER_ADDR", ""),
			Chat:         getEnv("CHAT_ADDR", ""),
			Notification: getEnv("NOTI_ADDR", ""),
			Event:        getEnv("EVENT_ADDR", ""),
			Organizer:    getEnv("ORGANIZER_ADDR", ""),
		},
		DatabaseCfg: Database{
			DSN:  getEnv("DATABASE_DSN", ""),
			Name: getEnv("DATABASE_NAME", ""),
		},
		JwtCfg: JWT{
			Token: getEnv("JWT_SECRET", ""),
		},
		RabbitMqCfg: RABBITMQ{
			URI: getEnv("RABBITMQ_URI", ""),
		},
		NotiCfg: Notification{
			Email:   getEnv("EMAIL", ""),
			EmailPW: getEnv("EMAIL_PW", ""),
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
