package config

import (
	"os"
	"testing"
)

func TestLoadReadsDotEnvWithoutOverridingEnvironment(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile(".env", []byte("OIDC_BUTTON_TITLE=From file\nDISABLE_PASSWORD_REGISTRATION=true\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	oldTitle, hadTitle := os.LookupEnv("OIDC_BUTTON_TITLE")
	oldDisable, hadDisable := os.LookupEnv("DISABLE_PASSWORD_REGISTRATION")
	_ = os.Unsetenv("OIDC_BUTTON_TITLE")
	_ = os.Unsetenv("DISABLE_PASSWORD_REGISTRATION")
	t.Cleanup(func() {
		if hadTitle {
			_ = os.Setenv("OIDC_BUTTON_TITLE", oldTitle)
		} else {
			_ = os.Unsetenv("OIDC_BUTTON_TITLE")
		}
		if hadDisable {
			_ = os.Setenv("DISABLE_PASSWORD_REGISTRATION", oldDisable)
		} else {
			_ = os.Unsetenv("DISABLE_PASSWORD_REGISTRATION")
		}
	})

	cfg := Load()
	if cfg.OIDCButtonTitle != "From file" || !cfg.DisablePasswordRegistration {
		t.Fatalf(".env was not loaded: %+v", cfg)
	}
	if cfg.DataDir != "./data" {
		t.Fatalf("unexpected data directory: %q", cfg.DataDir)
	}

	if err := os.Setenv("OIDC_BUTTON_TITLE", "From environment"); err != nil {
		t.Fatal(err)
	}
	if got := Load().OIDCButtonTitle; got != "From environment" {
		t.Fatalf("process environment must win, got %q", got)
	}
}

func TestPasswordLoginCanOnlyBeDisabledWithOIDC(t *testing.T) {
	cfg := &Config{DisablePasswordLogin: true}
	if !cfg.PasswordLoginEnabled() {
		t.Fatal("password login must remain available without configured OIDC")
	}
	cfg.OIDCIssuer = "https://id.example.com"
	cfg.OIDCClientID = "dashlit"
	if cfg.PasswordLoginEnabled() {
		t.Fatal("password login should be disabled when OIDC is configured")
	}
}

func TestPasswordRegistrationRequiresPasswordLogin(t *testing.T) {
	cfg := &Config{
		DisablePasswordLogin: true,
		OIDCIssuer:           "https://id.example.com",
		OIDCClientID:         "dashlit",
	}
	if cfg.PasswordRegistrationEnabled() {
		t.Fatal("password registration must be disabled when password login is disabled")
	}

	cfg.DisablePasswordLogin = false
	if !cfg.PasswordRegistrationEnabled() {
		t.Fatal("password registration should be enabled when neither policy disables it")
	}

	cfg.DisablePasswordRegistration = true
	if cfg.PasswordRegistrationEnabled() {
		t.Fatal("explicit password registration policy must still be respected")
	}
}

func TestOIDCUserMergeIsEnabledByDefault(t *testing.T) {
	t.Setenv("DISABLE_OIDC_USER_MERGE", "")
	t.Chdir(t.TempDir())
	if cfg := Load(); cfg.DisableOIDCUserMerge {
		t.Fatal("OIDC user merge must be enabled by default")
	}
}
