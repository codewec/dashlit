package auth

import (
	"context"
	"testing"

	"github.com/bookmarks-dashboard/backend/internal/models"
)

func TestBootstrapAdminCreatesFirstUser(t *testing.T) {
	svc := newMergeTestService(t, false)
	created, err := svc.BootstrapAdmin(context.Background(), " admin ", "secret-password")
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("initial administrator was not created")
	}

	user := new(models.User)
	if err := svc.db.NewSelect().Model(user).Scan(context.Background()); err != nil {
		t.Fatal(err)
	}
	if user.Username != "admin" || user.Role != models.RoleAdmin || user.PasswordHash == nil || !svc.CheckPassword(*user.PasswordHash, "secret-password") {
		t.Fatalf("unexpected bootstrap user: %#v", user)
	}
}

func TestBootstrapAdminIgnoresCredentialsAfterFirstUser(t *testing.T) {
	svc := newMergeTestService(t, false)
	if _, err := svc.Register(context.Background(), "existing", "password"); err != nil {
		t.Fatal(err)
	}

	created, err := svc.BootstrapAdmin(context.Background(), "", "invalid")
	if err != nil || created {
		t.Fatalf("credentials were not ignored: created=%v, err=%v", created, err)
	}
	count, err := svc.db.NewSelect().Model((*models.User)(nil)).Count(context.Background())
	if err != nil || count != 1 {
		t.Fatalf("user count = %d, err=%v", count, err)
	}
}

func TestBootstrapAdminRequiresBothCredentialsForEmptyDatabase(t *testing.T) {
	svc := newMergeTestService(t, false)
	if _, err := svc.BootstrapAdmin(context.Background(), "admin", ""); err == nil {
		t.Fatal("missing password must fail on an empty database")
	}
}
