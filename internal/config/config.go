package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Logger        `yaml:"logger"`
	Postgres      `yaml:"postgres"`
	HTTPServer    `yaml:"http_server"`
	AuthService   `yaml:"auth_service"`
	BannerService `yaml:"banner_service"`
	Redis         `yaml:"redis"`
	RabbitMQ      `yaml:"rabbitmq"`
}

type Logger struct {
	LogLevel string `yaml:"log_level" env-default:"debug"`
}

type HTTPServer struct {
	Address      string        `yaml:"address"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

type Postgres struct {
	MaxPoolSize int           `yaml:"max_pool_size" env-default:"2"`
	ConnTimeout time.Duration `yaml:"conn_timeout" env-default:"3s"`
	User        string        `yaml:"user" env:"DB_USER"`
	Password    string        `yaml:"password" env:"DB_PASSWORD"`
	Host        string        `yaml:"host"`
	Port        int           `yaml:"port"`
	DBName      string        `yaml:"db_name" env:"DB_NAME"`
}

type AuthService struct {
	SecretKey string `yaml:"secret_key" env:"AUTH_SECRET_KEY"`
}

type BannerService struct {
	CacheTTL         time.Duration `yaml:"cached_ttl" env-default:"300s"`
	DeleteWorkersNum int           `yaml:"delete_workers_num" env-default:"1"`
	QueueName        string        `yaml:"queue_name"`
}

type Redis struct {
	Address     string        `yaml:"address"`
	Port        int           `yaml:"port"`
	Password    string        `yaml:"password" env:"REDIS_PASSWORD"`
	ConnTimeout time.Duration `yaml:"conn_timeout" env-default:"3s"`
}

type RabbitMQ struct {
	Address  string `yaml:"address"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user" env:"RABBITMQ_DEFAULT_USER"`
	Password string `yaml:"password" env:"RABBITMQ_DEFAULT_PASS"`
	Name     string `yaml:"name"`
}

func Init() (*Config, error) {
	// Попытка считать путь файла с конфигами
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		return nil, fmt.Errorf("config.Init CONFIG_PATH is not set")
	}

	// Проверка существует ли файл
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config.Init config file does not exist: %s", configPath)
	}

	// Чтение конфига
	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("config.Init cannot read config: %s", err)
	}

	return &cfg, nil
}
