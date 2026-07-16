package config

import "os"

type Config struct {
	DatabaseURL       string
	APIPort           string
	APIHost           string
	OIDCIssuer        string
	OIDCClientID      string
	OIDCClientSecret  string
	JWTSecret         string
	RedisURL          string
	MeilisearchURL    string
	MeilisearchAPIKey string
	S3Endpoint        string
	S3AccessKey       string
	S3SecretKey       string
	S3Bucket          string
	S3Region          string
	Timezone          string
	Locale            string
}

func Load() *Config {
	return &Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgresql://zarishlog:zarishlog_dev_password@localhost:5432/zarishlog?sslmode=disable"),
		APIPort:           getEnv("API_PORT", "8080"),
		APIHost:           getEnv("API_HOST", "0.0.0.0"),
		OIDCIssuer:        getEnv("OIDC_ISSUER", "http://localhost:8080/realms/zarishlog"),
		OIDCClientID:      getEnv("OIDC_CLIENT_ID", "zarishlog-api"),
		OIDCClientSecret:  getEnv("OIDC_CLIENT_SECRET", "changeme"),
		JWTSecret:         getEnv("JWT_SECRET", "change-me-in-production"),
		RedisURL:          getEnv("REDIS_URL", "redis://localhost:6379"),
		MeilisearchURL:    getEnv("MEILISEARCH_URL", "http://localhost:7700"),
		MeilisearchAPIKey: getEnv("MEILISEARCH_API_KEY", "zarishlog_search_key"),
		S3Endpoint:        getEnv("S3_ENDPOINT", "http://localhost:9000"),
		S3AccessKey:       getEnv("S3_ACCESS_KEY", "zarishlog"),
		S3SecretKey:       getEnv("S3_SECRET_KEY", "zarishlog_dev_password"),
		S3Bucket:          getEnv("S3_BUCKET", "zarishlog-dev"),
		S3Region:          getEnv("S3_REGION", "us-east-1"),
		Timezone:          getEnv("DEFAULT_TIMEZONE", "Asia/Dhaka"),
		Locale:            getEnv("DEFAULT_LOCALE", "en-BD"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
