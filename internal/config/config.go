package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBURL        string
	Port         string
	CookieName   string
	CookieSecure bool
}

func Load() Config {
	return Config{
		DBURL:        getEnv("DB_URL", "postgres://postgres:postgres@localhost:5432/chat?sslmode=disable"),
		Port:         getEnv("PORT", "8080"),
		CookieName:   getEnv("COOKIE_NAME", "session_token"),
		CookieSecure: getEnvBool("COOKIE_SECURE", false),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
