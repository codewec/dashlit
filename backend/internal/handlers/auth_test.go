package handlers

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidOIDCReturnURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		host string
		want string
	}{
		{name: "dev frontend port", raw: "http://localhost:5173/", host: "localhost:8080", want: "http://localhost:5173/"},
		{name: "same production host", raw: "https://dash.example.com/", host: "dash.example.com", want: "https://dash.example.com/"},
		{name: "external host rejected", raw: "https://evil.example/", host: "dash.example.com", want: ""},
		{name: "relative URL rejected", raw: "//evil.example/", host: "dash.example.com", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validOIDCReturnURL(tt.raw, tt.host); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOIDCFailureRedirectsToFrontendReturnURL(t *testing.T) {
	h := &AuthHandler{}
	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/api/auth/oidc/callback", nil)
	returnCookie := &http.Cookie{
		Name:  oidcReturnCookie,
		Value: base64.RawURLEncoding.EncodeToString([]byte("http://localhost:5173/")),
	}
	recorder := httptest.NewRecorder()

	h.oidcFailure(recorder, req, returnCookie, "Registration through OIDC is disabled")

	if recorder.Code != http.StatusFound {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusFound)
	}
	want := "http://localhost:5173/#/login?oidc_error=Registration+through+OIDC+is+disabled"
	if got := recorder.Header().Get("Location"); got != want {
		t.Fatalf("got redirect %q, want %q", got, want)
	}
}
