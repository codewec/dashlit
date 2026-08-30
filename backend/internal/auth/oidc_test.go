package auth

import (
	"net/http"
	"testing"
)

func TestNewOIDCHTTPClient(t *testing.T) {
	if client := newOIDCHTTPClient(false); client != nil {
		t.Fatal("secure OIDC configuration returned a custom HTTP client")
	}

	client := newOIDCHTTPClient(true)
	transport, ok := client.Transport.(*http.Transport)
	if !ok || transport.TLSClientConfig == nil || !transport.TLSClientConfig.InsecureSkipVerify {
		t.Fatal("insecure OIDC HTTP client does not skip TLS verification")
	}
}

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
