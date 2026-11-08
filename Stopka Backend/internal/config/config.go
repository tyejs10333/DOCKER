
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	AppEnv      string
	HTTPAddr    string
	StaticDir   string
	DatabaseURL string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:      os.Getenv("APP_ENV"),
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		StaticDir:   getEnv("STATIC_DIR", "../Stopka Frontend"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	if cfg.AppEnv == "" {
		return Config{}, fmt.Errorf("не задан APP_ENV: укажи dev или prod")
	}

	if cfg.AppEnv != "dev" && cfg.AppEnv != "prod" {
		return Config{}, fmt.Errorf("APP_ENV должен быть dev или prod")
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("не задан DATABASE_URL в файле .env")
	}

	absolutePath, err := filepath.Abs(cfg.StaticDir)
	if err != nil {
		return Config{}, fmt.Errorf("не удалось определить путь к фронтенду: %w", err)
	}

	info, err := os.Stat(absolutePath)
	if err != nil {
		return Config{}, fmt.Errorf("не найдена папка фронтенда %s: %w", absolutePath, err)
	}

	if !info.IsDir() {
		return Config{}, fmt.Errorf("путь к фронтенду должен указывать на папку")
	}

	cfg.StaticDir = absolutePath

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
