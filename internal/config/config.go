package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Server   ServerConfig   `mapstructure:"server"`
	Runner   RunnerConfig   `mapstructure:"runner"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnectTimeout  int           `mapstructure:"connect_timeout"`
}

type ServerConfig struct {
	Listen         string        `mapstructure:"listen"`
	SecretKey      string        `mapstructure:"secret_key"`
	SessionTimeout time.Duration `mapstructure:"session_timeout"`
}

type RunnerConfig struct {
	MaxConcurrent int    `mapstructure:"max_concurrent"`
	Timeout       string `mapstructure:"timeout"`
	MemoryLimitMB int    `mapstructure:"memory_limit_mb"`
	CPULimit      int    `mapstructure:"cpu_limit"`
}

func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Set default values
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "5m")
	viper.SetDefault("database.connect_timeout", 5)

	viper.SetDefault("server.listen", ":8080")
	viper.SetDefault("server.session_timeout", "24h")

	viper.SetDefault("runner.max_concurrent", 5)
	viper.SetDefault("runner.timeout", "30s")
	viper.SetDefault("runner.memory_limit_mb", 256)
	viper.SetDefault("runner.cpu_limit", 1)

	// Read environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("OJ") // Environment variables will be prefixed with OJ_
	viper.BindEnv("database.host", "OJ_DB_HOST")
	viper.BindEnv("database.port", "OJ_DB_PORT")
	viper.BindEnv("database.user", "OJ_DB_USER")
	viper.BindEnv("database.password", "OJ_DB_PASSWORD")
	viper.BindEnv("database.dbname", "OJ_DB_NAME")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if we have environment variables
			fmt.Println("Config file not found, using environment variables and defaults")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}
