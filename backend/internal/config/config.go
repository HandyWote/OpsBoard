package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config 描述了后端服务所需的全部配置。
type Config struct {
	Server ServerConfig
	DB     DBConfig
	Auth   AuthConfig
}

// ServerConfig 控制 HTTP 服务以及中间件参数。
type ServerConfig struct {
	Addr               string
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	ReadHeaderTimeout  time.Duration
	AllowOrigins       []string
	TrustedProxyHeader string
}

// DBConfig 描述数据库连接设置。
type DBConfig struct {
	DSN             string
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// AuthConfig 控制 Token 签发与密码策略。
type AuthConfig struct {
	AccessTokenTTL        time.Duration
	RefreshTokenTTL       time.Duration
	JWTSecret             string
	PasswordHashCost      int
	RefreshTokenHashKey   string
	AllowAutoUserCreation bool
}

// Load 从环境变量构建配置，未设置的值使用默认值。
func Load() (Config, error) {
	cfg := Config{
		Server: ServerConfig{
			Addr:               lookupString("PORT", ":9012"),
			ReadTimeout:        lookupDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:       lookupDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:        lookupDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
			ReadHeaderTimeout:  lookupDuration("SERVER_READ_HEADER_TIMEOUT", 10*time.Second),
			AllowOrigins:       splitAndTrim(os.Getenv("CORS_ALLOW_ORIGINS")),
			TrustedProxyHeader: lookupString("TRUSTED_PROXY_HEADER", "X-Real-IP"),
		},
		DB: DBConfig{
			DSN:             os.Getenv("DATABASE_URL"),
			Host:            lookupString("DB_HOST", "127.0.0.1"),
			Port:            lookupInt("DB_PORT", 5432),
			User:            lookupString("DB_USER", "opsboard"),
			Password:        lookupString("DB_PASSWORD", "admin"),
			Name:            lookupString("DB_NAME", "opsboard"),
			SSLMode:         lookupString("DB_SSLMODE", "disable"),
			MaxOpenConns:    lookupInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    lookupInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: lookupDuration("DB_CONN_MAX_LIFETIME", time.Hour),
		},
		Auth: AuthConfig{
			AccessTokenTTL:        lookupDuration("AUTH_ACCESS_TOKEN_TTL", time.Hour),
			RefreshTokenTTL:       lookupDuration("AUTH_REFRESH_TOKEN_TTL", 7*24*time.Hour),
			JWTSecret:             lookupString("AUTH_JWT_SECRET", ""),
			PasswordHashCost:      lookupInt("AUTH_PASSWORD_COST", 12),
			RefreshTokenHashKey:   lookupString("AUTH_REFRESH_HASH_KEY", ""),
			AllowAutoUserCreation: lookupBool("AUTH_ALLOW_AUTO_USER_CREATION", true),
		},
	}

	if !strings.HasPrefix(cfg.Server.Addr, ":") && !strings.Contains(cfg.Server.Addr, ":") {
		cfg.Server.Addr = ":" + cfg.Server.Addr
	}

	if cfg.DB.DSN == "" {
		cfg.DB.DSN = fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.DB.Host,
			cfg.DB.Port,
			cfg.DB.User,
			cfg.DB.Password,
			cfg.DB.Name,
			cfg.DB.SSLMode,
		)
	}

	if cfg.Auth.JWTSecret == "" {
		return Config{}, errors.New("AUTH_JWT_SECRET 未配置")
	}

	if cfg.Auth.RefreshTokenHashKey == "" {
		return Config{}, errors.New("AUTH_REFRESH_HASH_KEY 未配置")
	}

	return cfg, nil
}

func lookupString(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func lookupInt(key string, fallback int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if num, err := strconv.Atoi(v); err == nil {
			return num
		}
	}
	return fallback
}

func lookupBool(key string, fallback bool) bool {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			return parsed
		}
	}
	return fallback
}

func lookupDuration(key string, fallback time.Duration) time.Duration {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func splitAndTrim(input string) []string {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
