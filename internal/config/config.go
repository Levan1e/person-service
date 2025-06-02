package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Database содержит настройки для подключения к базе данных.
type Database struct {
	Host     string `mapstructure:"db_host"`
	Port     string `mapstructure:"db_port"`
	User     string `mapstructure:"db_user"`
	Password string `mapstructure:"db_password"`
	Name     string `mapstructure:"db_name"`
}

// APIs содержит URL внешних API для обогащения данных.
type APIs struct {
	Agify       string `mapstructure:"agify_api_url"`
	Genderize   string `mapstructure:"genderize_api_url"`
	Nationalize string `mapstructure:"nationalize_api_url"`
}

// Server содержит настройки сервера.
type Server struct {
	Port string `mapstructure:"server_port"`
}

// Config содержит все настройки приложения.
type Config struct {
	Server   Server   `mapstructure:"server"`
	Database Database `mapstructure:"database"`
	APIs     APIs     `mapstructure:"apis"`
	LogLevel string   `mapstructure:"log_level"`
}

// LoadConfig загружает конфигурацию из .env файла или переменных окружения.
func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	fmt.Println("Попытка чтения .env файла...")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Ошибка чтения .env: %v\n", err)
	} else {
		fmt.Println("Файл .env успешно прочитан")
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("не удалось десериализовать конфигурацию: %w", err)
	}

	fmt.Printf("Загруженная конфигурация: %+v\n", cfg)

	if cfg.Server.Port == "" {
		return nil, fmt.Errorf("SERVER_PORT обязателен")
	}
	if cfg.Database.Name == "" {
		return nil, fmt.Errorf("DB_NAME обязателен")
	}
	if cfg.APIs.Agify == "" || cfg.APIs.Genderize == "" || cfg.APIs.Nationalize == "" {
		return nil, fmt.Errorf("URL внешних API обязательны")
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	return &cfg, nil
}
