package auth

import "testing"

func TestOIDCUsernameFallbacks(t *testing.T) {
	tests := []struct {
		name     string
		identity OIDCIdentity
		want     string
	}{
		{name: "preferred username", identity: OIDCIdentity{PreferredUsername: "john.doe"}, want: "john.doe"},
		{name: "email", identity: OIDCIdentity{Email: "john@example.com"}, want: "john"},
		{name: "display name sanitized", identity: OIDCIdentity{Name: "John Doe"}, want: "John-Doe"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := oidcUsername(&tt.identity); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
