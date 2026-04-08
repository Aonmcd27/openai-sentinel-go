package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	sentinel "openai-sentinel-go"
)

type appConfig struct {
	listenAddr             string
	port                   string
	apiBearerToken         string
	clientTimeoutMs        int
	sentinelBaseURL        string
	sentinelTimeoutMs      int
	sentinelMaxAttempts    int
	sentinelDirectFallback bool
	turnstileStaticToken   string
}

func main() {
	cfg := loadConfig()
	service := sentinel.NewService(sentinel.Config{
		SentinelBaseURL:        cfg.sentinelBaseURL,
		SentinelTimeout:        time.Duration(cfg.sentinelTimeoutMs) * time.Millisecond,
		SentinelMaxAttempts:    cfg.sentinelMaxAttempts,
		SentinelDirectFallback: cfg.sentinelDirectFallback,
		TurnstileStaticToken:   cfg.turnstileStaticToken,
	})

	srv := server{
		apiBearerToken:  cfg.apiBearerToken,
		clientTimeoutMs: cfg.clientTimeoutMs,
		buildToken: func(ctx context.Context, session *sentinel.Session, flow, referer, turnstileToken string) (sentinel.Token, error) {
			return service.Build(ctx, session, flow, referer, turnstileToken)
		},
	}

	addr := cfg.listenAddr + ":" + cfg.port
	log.Printf("render-api listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.routes()); err != nil {
		log.Fatal(err)
	}
}

func loadConfig() appConfig {
	return appConfig{
		listenAddr:             getEnv("LISTEN_ADDR", "0.0.0.0"),
		port:                   getEnv("PORT", "10000"),
		apiBearerToken:         os.Getenv("API_BEARER_TOKEN"),
		clientTimeoutMs:        getEnvInt("CLIENT_TIMEOUT_MS", 10000),
		sentinelBaseURL:        getEnv("SENTINEL_BASE_URL", "https://sentinel.openai.com"),
		sentinelTimeoutMs:      getEnvInt("SENTINEL_TIMEOUT_MS", 10000),
		sentinelMaxAttempts:    getEnvInt("SENTINEL_MAX_ATTEMPTS", 2),
		sentinelDirectFallback: getEnvBool("SENTINEL_DIRECT_FALLBACK", false),
		turnstileStaticToken:   os.Getenv("TURNSTILE_STATIC_TOKEN"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
