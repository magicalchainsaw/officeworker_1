package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Log      LogConfig
}

type ServerConfig struct {
	Port         string
	Mode         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	Secret            string
	Expiration        time.Duration
	RefreshExpiration time.Duration
}

type LogConfig struct {
	Level  string
	Format string
}

func setDefaults() {
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SERVER_MODE", "debug")
	viper.SetDefault("SERVER_READ_TIMEOUT", "15s")
	viper.SetDefault("SERVER_WRITE_TIMEOUT", "15s")

	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 3306)
	viper.SetDefault("DB_USER", "root")
	viper.SetDefault("DB_PASSWORD", "")
	viper.SetDefault("DB_NAME", "officeworker")
	viper.SetDefault("DB_MAX_IDLE_CONNS", 10)
	viper.SetDefault("DB_MAX_OPEN_CONNS", 100)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", "3600s")

	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", 6379)
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)

	viper.SetDefault("JWT_SECRET", "your-secret-key-change-in-production")
	viper.SetDefault("JWT_EXPIRATION", "24h")
	viper.SetDefault("JWT_REFRESH_EXPIRATION", "168h")

	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "json")
}

func Load() (*Config, error) {
	setDefaults()

	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:         strings.TrimSpace(viper.GetString("SERVER_PORT")),
			Mode:         strings.TrimSpace(viper.GetString("SERVER_MODE")),
		},
		Database: DatabaseConfig{
			Host:            strings.TrimSpace(viper.GetString("DB_HOST")),
			Port:            viper.GetInt("DB_PORT"),
			User:            strings.TrimSpace(viper.GetString("DB_USER")),
			Password:        viper.GetString("DB_PASSWORD"),
			DBName:          strings.TrimSpace(viper.GetString("DB_NAME")),
			MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
			MaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
		},
		Redis: RedisConfig{
			Host:     strings.TrimSpace(viper.GetString("REDIS_HOST")),
			Port:     viper.GetInt("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		JWT: JWTConfig{
			Secret: strings.TrimSpace(viper.GetString("JWT_SECRET")),
		},
		Log: LogConfig{
			Level:  strings.TrimSpace(viper.GetString("LOG_LEVEL")),
			Format: strings.TrimSpace(viper.GetString("LOG_FORMAT")),
		},
	}

	var err error
	if cfg.Server.ReadTimeout, err = getDuration("SERVER_READ_TIMEOUT"); err != nil {
		return nil, err
	}
	if cfg.Server.WriteTimeout, err = getDuration("SERVER_WRITE_TIMEOUT"); err != nil {
		return nil, err
	}
	if cfg.Database.ConnMaxLifetime, err = getDuration("DB_CONN_MAX_LIFETIME"); err != nil {
		return nil, err
	}
	if cfg.JWT.Expiration, err = getDuration("JWT_EXPIRATION"); err != nil {
		return nil, err
	}
	if cfg.JWT.RefreshExpiration, err = getDuration("JWT_REFRESH_EXPIRATION"); err != nil {
		return nil, err
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getDuration(key string) (time.Duration, error) {
	raw := strings.TrimSpace(viper.GetString(key))
	if raw == "" {
		return 0, nil
	}

	duration, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for %s: %w", key, err)
	}

	return duration, nil
}

func validate(cfg *Config) error {
	switch {
	case cfg.Server.Port == "":
		return fmt.Errorf("server port is required")
	case cfg.Server.ReadTimeout <= 0:
		return fmt.Errorf("server read timeout must be greater than zero")
	case cfg.Server.WriteTimeout <= 0:
		return fmt.Errorf("server write timeout must be greater than zero")
	case cfg.Database.Host == "":
		return fmt.Errorf("database host is required")
	case cfg.Database.Port <= 0:
		return fmt.Errorf("database port must be greater than zero")
	case cfg.Database.User == "":
		return fmt.Errorf("database user is required")
	case cfg.Database.DBName == "":
		return fmt.Errorf("database name is required")
	case cfg.Database.MaxIdleConns < 0:
		return fmt.Errorf("database max idle conns cannot be negative")
	case cfg.Database.MaxOpenConns <= 0:
		return fmt.Errorf("database max open conns must be greater than zero")
	case cfg.Database.ConnMaxLifetime <= 0:
		return fmt.Errorf("database connection max lifetime must be greater than zero")
	case cfg.Redis.Host == "":
		return fmt.Errorf("redis host is required")
	case cfg.Redis.Port <= 0:
		return fmt.Errorf("redis port must be greater than zero")
	case cfg.JWT.Secret == "":
		return fmt.Errorf("jwt secret is required")
	case cfg.JWT.Expiration <= 0:
		return fmt.Errorf("jwt expiration must be greater than zero")
	case cfg.JWT.RefreshExpiration <= 0:
		return fmt.Errorf("jwt refresh expiration must be greater than zero")
	case cfg.Log.Level == "":
		return fmt.Errorf("log level is required")
	case cfg.Log.Format == "":
		return fmt.Errorf("log format is required")
	default:
		return nil
	}
}
