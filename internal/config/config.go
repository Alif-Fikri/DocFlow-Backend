package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Config struct {
	Port           string
	APIKey         string
	SofficeBin     string
	WorkDir        string
	MaxUploadBytes int64
	ConvertTimeout time.Duration
}

func Load() Config {
	return Config{
		Port:           env("PORT", "8080"),
		APIKey:         env("API_KEY", ""),
		SofficeBin:     env("SOFFICE_BIN", "soffice"),
		WorkDir:        env("WORK_DIR", filepath.Join(os.TempDir(), "docflow")),
		MaxUploadBytes: int64(envInt("MAX_UPLOAD_MB", 50)) * 1024 * 1024,
		ConvertTimeout: time.Duration(envInt("CONVERT_TIMEOUT_SEC", 120)) * time.Second,
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
