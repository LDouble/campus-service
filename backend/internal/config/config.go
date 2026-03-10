package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	WeChat   WeChatConfig
}

type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Mode         string
}

type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	MaxIdleConns int
	MaxOpenConns int
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	AccessTokenSecret   string
	RefreshTokenSecret  string
	AccessTokenExpiry   time.Duration
	RefreshTokenExpiry  time.Duration
	Issuer              string
}

type WeChatConfig struct {
	AppID     string
	AppSecret string
}

var Cfg *Config

func Init(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Set defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "10s")
	viper.SetDefault("server.write_timeout", "10s")
	viper.SetDefault("server.mode", "debug")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)

	viper.SetDefault("jwt.access_token_expiry", "2h")
	viper.SetDefault("jwt.refresh_token_expiry", "168h") // 7 days

	// Allow environment variable overrides
	viper.AutomaticEnv()
	if err := viper.BindEnv("DB_PASSWORD"); err != nil {
		return fmt.Errorf("failed to bind env DB_PASSWORD: %w", err)
	}
	if err := viper.BindEnv("REDIS_PASSWORD"); err != nil {
		return fmt.Errorf("failed to bind env REDIS_PASSWORD: %w", err)
	}
	if err := viper.BindEnv("JWT_ACCESS_TOKEN_SECRET"); err != nil {
		return fmt.Errorf("failed to bind env JWT_ACCESS_TOKEN_SECRET: %w", err)
	}
	if err := viper.BindEnv("JWT_REFRESH_TOKEN_SECRET"); err != nil {
		return fmt.Errorf("failed to bind env JWT_REFRESH_TOKEN_SECRET: %w", err)
	}

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	Cfg = &Config{}

	// 验证敏感配置环境变量
	if os.Getenv("DB_PASSWORD") == "" {
		return fmt.Errorf("DB_PASSWORD environment variable is required")
	}
	if os.Getenv("JWT_ACCESS_TOKEN_SECRET") == "" {
		return fmt.Errorf("JWT_ACCESS_TOKEN_SECRET environment variable is required")
	}
	if os.Getenv("JWT_REFRESH_TOKEN_SECRET") == "" {
		return fmt.Errorf("JWT_REFRESH_TOKEN_SECRET environment variable is required")
	}

	// Server config
	Cfg.Server.Port = viper.GetInt("server.port")
	Cfg.Server.ReadTimeout = viper.GetDuration("server.read_timeout")
	Cfg.Server.WriteTimeout = viper.GetDuration("server.write_timeout")
	Cfg.Server.Mode = viper.GetString("server.mode")

	// Database config
	Cfg.Database.Host = viper.GetString("database.host")
	Cfg.Database.Port = viper.GetInt("database.port")
	Cfg.Database.User = viper.GetString("database.user")
	Cfg.Database.Password = os.Getenv("DB_PASSWORD")
	Cfg.Database.DBName = viper.GetString("database.dbname")
	Cfg.Database.MaxIdleConns = viper.GetInt("database.max_idle_conns")
	Cfg.Database.MaxOpenConns = viper.GetInt("database.max_open_conns")

	// Redis config
	Cfg.Redis.Host = viper.GetString("redis.host")
	Cfg.Redis.Port = viper.GetInt("redis.port")
	Cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")
	Cfg.Redis.DB = viper.GetInt("redis.db")

	// JWT config
	Cfg.JWT.AccessTokenSecret = os.Getenv("JWT_ACCESS_TOKEN_SECRET")
	Cfg.JWT.RefreshTokenSecret = os.Getenv("JWT_REFRESH_TOKEN_SECRET")
	Cfg.JWT.AccessTokenExpiry = viper.GetDuration("jwt.access_token_expiry")
	Cfg.JWT.RefreshTokenExpiry = viper.GetDuration("jwt.refresh_token_expiry")
	Cfg.JWT.Issuer = viper.GetString("jwt.issuer")

	// WeChat config
	Cfg.WeChat.AppID = viper.GetString("wechat.app_id")
	Cfg.WeChat.AppSecret = viper.GetString("wechat.app_secret")

	return nil
}

// GetDSN returns MySQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}

// GetRedisAddr returns Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}