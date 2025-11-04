package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Settings struct {
	Port            int
	DBURL           string
	APIKey          string
	APISecret       string
	SecretKey       string
	HLBaseURL       string
	HLWSURL         string
	JWTSecret       string
	DeepseekAPIKey  string
	DeepseekBaseURL string
	DeepseekModel   string
	FeeRate         float64
}

func Load() (*Settings, error) {
	// Try to load .env if present; ignore error inside Docker where envs are injected
	_ = godotenv.Load()

	port := getInt("PORT", 8080)
	cfg := &Settings{
		Port:            port,
		DBURL:           getStr("DB_URL", "postgres://postgres:postgres@db:5432/deepseek_trader?sslmode=disable"),
		APIKey:          getStr("API_KEY", ""),
		APISecret:       getStr("API_SECRET", ""),
		SecretKey:       getStr("SECRET_KEY", "0123456789abcdef0123456789abcdef"),
		HLBaseURL:       getStr("HL_BASE_URL", "https://api.hyperliquid.xyz"),
		HLWSURL:         getStr("HL_WS_URL", "wss://api.hyperliquid.xyz/ws"),
		JWTSecret:       getStr("JWT_SECRET", "change-me-super-secret"),
		DeepseekAPIKey:  getStr("DEEPSEEK_API_KEY", ""),
		DeepseekBaseURL: getStr("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
		DeepseekModel:   getStr("DEEPSEEK_MODEL", "deepseek-chat"),
		FeeRate:         getFloat("FEE_RATE", 0.0005),
	}
	return cfg, nil
}

func getStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return def
}

func getFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}
