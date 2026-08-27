package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr                        string
	DataDir                     string
	DatabasePath                string
	JWTSecret                   string
	IconsDir                    string
	IconCacheDir                string
	OIDCIssuer                  string
	OIDCClientID                string
	OIDCClientSecret            string
	OIDCRedirectURL             string
	OIDCButtonTitle             string
	DisablePasswordRegistration bool
	DisableOIDCRegistration     bool
	DisablePasswordLogin        bool
	DevMode                     bool
}

func Load() *Config {
	// Existing process variables take precedence over values from .env.
	_ = godotenv.Load()

	dataDir := env("DATA_DIR", "./data")
	_ = os.MkdirAll(dataDir, 0755)
	iconsDir := filepath.Join(dataDir, "icons")
	cacheDir := filepath.Join(dataDir, "icon-cache")
	_ = os.MkdirAll(iconsDir, 0755)
	_ = os.MkdirAll(cacheDir, 0755)

	return &Config{
		Addr:                        env("ADDR", ":8080"),
		DataDir:                     dataDir,
		DatabasePath:                env("DATABASE_PATH", filepath.Join(dataDir, "bookmarks.db")),
		JWTSecret:                   env("JWT_SECRET", "change-me-in-production-please-use-long-secret"),
		IconsDir:                    iconsDir,
		IconCacheDir:                cacheDir,
		OIDCIssuer:                  os.Getenv("OIDC_ISSUER"),
		OIDCClientID:                os.Getenv("OIDC_CLIENT_ID"),
		OIDCClientSecret:            os.Getenv("OIDC_CLIENT_SECRET"),
		OIDCRedirectURL:             env("OIDC_REDIRECT_URL", "http://localhost:8080/api/auth/oidc/callback"),
		OIDCButtonTitle:             env("OIDC_BUTTON_TITLE", "Sign in with OIDC"),
		DisablePasswordRegistration: envBool("DISABLE_PASSWORD_REGISTRATION", false),
		DisableOIDCRegistration:     envBool("DISABLE_OIDC_REGISTRATION", false),
		DisablePasswordLogin:        envBool("DISABLE_PASSWORD_LOGIN", false),
		DevMode:                     envBool("DEV_MODE", false),
	}
}

func (c *Config) OIDCEnabled() bool {
	return strings.TrimSpace(c.OIDCIssuer) != "" && strings.TrimSpace(c.OIDCClientID) != ""
}

// Password login is always retained as a recovery path until OIDC is fully
// configured, even when DISABLE_PASSWORD_LOGIN is set.
func (c *Config) PasswordLoginEnabled() bool {
	return !c.DisablePasswordLogin || !c.OIDCEnabled()
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
