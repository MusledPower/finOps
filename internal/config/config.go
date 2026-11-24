package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr, DatabaseURL, BearerToken string
	LLMBaseURL, LLMAPIKey, LLMModel    string
	LLMTimeout                         time.Duration
	MaxRequestBytes                    int64
}

func Load() Config {
	return Config{
		HTTPAddr:        env("HTTP_ADDR", ":8080"),
		DatabaseURL:     env("DATABASE_URL", "postgres://decomposer:decomposer@localhost:5432/decomposer?sslmode=disable"),
		BearerToken:     env("API_BEARER_TOKEN", "change-me"),
		LLMBaseURL:      env("LLM_BASE_URL", "https://api.openai.com/v1"),
		LLMAPIKey:       os.Getenv("LLM_API_KEY"),
		LLMModel:        os.Getenv("LLM_MODEL"),
		LLMTimeout:      duration("LLM_TIMEOUT", 90*time.Second),
		MaxRequestBytes: int64env("MAX_REQUEST_BYTES", 262144),
	}
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func duration(k string, d time.Duration) time.Duration {
	v, e := time.ParseDuration(os.Getenv(k))
	if e != nil {
		return d
	}
	return v
}
func int64env(k string, d int64) int64 {
	v, e := strconv.ParseInt(os.Getenv(k), 10, 64)
	if e != nil {
		return d
	}
	return v
}
