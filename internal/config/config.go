package config

import (
	"net"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

type Config struct {
	Addr          string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	AdminUser     string
	AdminPassword string

	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	SMTPFrom string
	SMTPTo   string

	TurnstileSiteKey string
	TurnstileSecret  string
	TurnstileSkip    bool
}

func Load() Config {
	return Config{
		Addr:          env("APP_ADDR", ":8080"),
		DBHost:        env("DB_HOST", "127.0.0.1"),
		DBPort:        env("DB_PORT", "3306"),
		DBUser:        env("DB_USER", "klowna"),
		DBPassword:    env("DB_PASSWORD", "klownapass"),
		DBName:        env("DB_NAME", "klowna"),
		AdminUser:     env("ADMIN_USER", "admin"),
		AdminPassword: env("ADMIN_PASSWORD", "admin"),

		SMTPHost: env("SMTP_HOST", ""),
		SMTPPort: env("SMTP_PORT", "587"),
		SMTPUser: env("SMTP_USER", ""),
		SMTPPass: env("SMTP_PASS", ""),
		SMTPFrom: env("SMTP_FROM", ""),
		SMTPTo:   env("SMTP_TO", ""),

		TurnstileSiteKey: env("TURNSTILE_SITE_KEY", "0x4AAAAAAFsU9zRaPwnUXFLc"),
		TurnstileSecret:  env("TURNSTILE_SECRET", ""),
		TurnstileSkip:    envBool("TURNSTILE_SKIP", false),
	}
}

func (c Config) DSN() string {
	m := mysql.NewConfig()
	m.User = c.DBUser
	m.Passwd = c.DBPassword
	m.Net = "tcp"
	m.Addr = net.JoinHostPort(c.DBHost, c.DBPort)
	m.DBName = c.DBName
	m.ParseTime = true
	m.Loc = time.UTC
	m.MultiStatements = true
	m.Params = map[string]string{"charset": "utf8mb4"}
	return m.FormatDSN()
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
