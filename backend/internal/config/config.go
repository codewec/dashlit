package config

import (
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Addr            string
	DataDir         string
	DatabasePath    string
	JWTSecret       string
	IconsDir        string
	IconCacheDir    string
	OIDCIssuer      string
	OIDCClientID    string
	OIDCClientSecret string
	OIDCRedirectURL string
	DevMode         bool
}

func Load() *Config {
	dataDir := env("DATA_DIR", "./data")
	_ = os.MkdirAll(dataDir, 0755)
	iconsDir := filepath.Join(dataDir, "icons")
	cacheDir := filepath.Join(dataDir, "icon-cache")
	_ = os.MkdirAll(iconsDir, 0755)
	_ = os.MkdirAll(cacheDir, 0755)

	return &Config{
		Addr:             env("ADDR", ":8080"),
		DataDir:          dataDir,
		DatabasePath:     env("DATABASE_PATH", filepath.Join(dataDir, "bookmarks.db")),
		JWTSecret:        env("JWT_SECRET", "change-me-in-production-please-use-long-secret"),
		IconsDir:         iconsDir,
		IconCacheDir:     cacheDir,
		OIDCIssuer:       os.Getenv("OIDC_ISSUER"),
		OIDCClientID:     os.Getenv("OIDC_CLIENT_ID"),
		OIDCClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
		OIDCRedirectURL:  env("OIDC_REDIRECT_URL", "http://localhost:8080/api/auth/oidc/callback"),
		DevMode:          envBool("DEV_MODE", false),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return fallback
}
