package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecurityHeadersSandboxIcons(t *testing.T) {
	h := SecurityHeaders(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/icons/abc", nil))
	if !strings.Contains(rec.Header().Get("Content-Security-Policy"), "sandbox") || rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("icon headers = %v", rec.Header())
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Header().Get("Content-Security-Policy") != "" || rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("page headers = %v", rec.Header())
	}
}

func TestLimitBodyRejectsOversizedBodies(t *testing.T) {
	h := LimitBody(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.ReadAll(r.Body); err != nil {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
		}
	}))
	big := strings.Repeat("a", MaxRequestBody+1)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/x", strings.NewReader(big)))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("declared length: code = %d", rec.Code)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/x", io.NopCloser(strings.NewReader(big)))
	req.ContentLength = -1
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("chunked: code = %d", rec.Code)
	}
}
