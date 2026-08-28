package auth

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/bookmarks-dashboard/backend/internal/config"
	appdb "github.com/bookmarks-dashboard/backend/internal/db"
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
	passwordUser, err := svc.Register(context.Background(), "alice", "password")
	if err != nil {
		t.Fatal(err)
	}

	oidcUser, err := svc.FindOrCreateOIDCUser(context.Background(), &OIDCIdentity{
		Issuer: "https://id.example.com", Subject: "alice-subject", PreferredUsername: "alice",
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
