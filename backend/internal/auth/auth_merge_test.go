package auth

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/bookmarks-dashboard/backend/internal/config"
	appdb "github.com/bookmarks-dashboard/backend/internal/db"
	"github.com/bookmarks-dashboard/backend/internal/models"
)

func newMergeTestService(t *testing.T, disableMerge bool) *Service {
	t.Helper()
	cfg := &config.Config{
		DatabasePath:         filepath.Join(t.TempDir(), "auth.db"),
		JWTSecret:            "test-secret",
		DisableOIDCUserMerge: disableMerge,
	}
	database, err := appdb.Connect(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return NewService(database, cfg)
}

func TestFindOrCreateOIDCUserMergesPasswordUserByDefault(t *testing.T) {
	svc := newMergeTestService(t, false)
	// The first account becomes admin, which is never merged.
	if _, err := svc.Register(context.Background(), "admin", "password"); err != nil {
		t.Fatal(err)
	}
	passwordUser, err := svc.Register(context.Background(), "alice", "password")
	if err != nil {
		t.Fatal(err)
	}

	oidcUser, err := svc.FindOrCreateOIDCUser(context.Background(), &OIDCIdentity{
		Issuer: "https://id.example.com", Subject: "alice-subject", PreferredUsername: "alice", EmailVerified: true,
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	if oidcUser.ID != passwordUser.ID {
		t.Fatalf("OIDC login created a second user: got %q, want %q", oidcUser.ID, passwordUser.ID)
	}
	if oidcUser.PasswordHash == nil || oidcUser.OIDCIssuer == nil || oidcUser.OIDCSubject == nil {
		t.Fatal("merged user must retain password login and receive the OIDC identity")
	}
}

func TestFindOrCreateOIDCUserCanDisableMerge(t *testing.T) {
	svc := newMergeTestService(t, true)
	passwordUser, err := svc.Register(context.Background(), "alice", "password")
	if err != nil {
		t.Fatal(err)
	}

	oidcUser, err := svc.FindOrCreateOIDCUser(context.Background(), &OIDCIdentity{
		Issuer: "https://id.example.com", Subject: "alice-subject", PreferredUsername: "alice",
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	if oidcUser.ID == passwordUser.ID || oidcUser.Username != "alice-2" {
		t.Fatalf("expected separate alice-2 user, got id=%q username=%q", oidcUser.ID, oidcUser.Username)
	}
}

func TestUpdateProfileChangesPasswordWithoutCurrentPassword(t *testing.T) {
	svc := newMergeTestService(t, false)
	user, err := svc.Register(context.Background(), "alice", "password")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.UpdateProfile(context.Background(), user, "alice-renamed", "new-password"); err != nil {
		t.Fatal(err)
	}
	if user.Username != "alice-renamed" || user.PasswordHash == nil || !svc.CheckPassword(*user.PasswordHash, "new-password") {
		t.Fatal("profile was not updated")
	}
}

func TestUpdateProfileAllowsOIDCUserToSetFirstPassword(t *testing.T) {
	svc := newMergeTestService(t, false)
	user, err := svc.FindOrCreateOIDCUser(context.Background(), &OIDCIdentity{
		Issuer: "https://id.example.com", Subject: "alice-subject", PreferredUsername: "alice",
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.UpdateProfile(context.Background(), user, "alice", "new-password"); err != nil {
		t.Fatal(err)
	}
	if user.PasswordHash == nil || !svc.CheckPassword(*user.PasswordHash, "new-password") {
		t.Fatal("first password was not set")
	}
}

func TestUpdateThemePersistsCustomPalette(t *testing.T) {
	svc := newMergeTestService(t, false)
	user, err := svc.Register(context.Background(), "alice", "password")
	if err != nil {
		t.Fatal(err)
	}
	custom := `{"scheme":"dark","bg":"#111111","surface":"#222222","text":"#eeeeee","primary":"#89b4fa","accent":"#cba6f7","danger":"#f38ba8","success":"#a6e3a1","backgroundImage":""}`
	if err := svc.UpdateTheme(context.Background(), user, "custom", custom); err != nil {
		t.Fatal(err)
	}
	loaded, err := svc.GetUser(context.Background(), user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Theme != "custom" || loaded.CustomTheme == "" {
		t.Fatalf("theme was not persisted: theme=%q custom=%q", loaded.Theme, loaded.CustomTheme)
	}
	if err := svc.UpdateTheme(context.Background(), user, "black", ""); err != nil {
		t.Fatal(err)
	}
	if user.Theme != "black" || user.CustomTheme != "" {
		t.Fatalf("preset theme clear failed: theme=%q custom=%q", user.Theme, user.CustomTheme)
	}
	loaded, err = svc.GetUser(context.Background(), user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Theme != "black" || loaded.CustomTheme != "" {
		t.Fatalf("preset theme was not persisted: theme=%q custom=%q", loaded.Theme, loaded.CustomTheme)
	}
}

func TestFindOrCreateOIDCUserDoesNotMergeAdminsOrUnverifiedIdentities(t *testing.T) {
	svc := newMergeTestService(t, false)
	admin, err := svc.Register(context.Background(), "admin", "password")
	if err != nil {
		t.Fatal(err)
	}
	alice, err := svc.Register(context.Background(), "alice", "password")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		target *models.User
		id     *OIDCIdentity
	}{
		{admin, &OIDCIdentity{Issuer: "https://id.example.com", Subject: "s1", PreferredUsername: "admin", EmailVerified: true}},
		{alice, &OIDCIdentity{Issuer: "https://id.example.com", Subject: "s2", PreferredUsername: "alice"}},
	}
	for _, c := range cases {
		if _, err := svc.FindOrCreateOIDCUser(context.Background(), c.id, false); !errors.Is(err, ErrOIDCRegistrationDisabled) {
			t.Fatalf("%s: expected no merge, got err=%v", c.target.Username, err)
		}
		got, err := svc.FindOrCreateOIDCUser(context.Background(), c.id, true)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID == c.target.ID || got.Username != c.target.Username+"-2" {
			t.Fatalf("%s: OIDC identity was linked to the existing account (%q)", c.target.Username, got.Username)
		}
	}
}
