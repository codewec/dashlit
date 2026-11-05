package handlers

import (
	"net/http"
	"strings"
)

// MaxRequestBody caps every request body; icon uploads use a smaller limit.
const MaxRequestBody = 10 << 20

// SecurityHeaders disables MIME sniffing everywhere and sandboxes icon
// responses, so an uploaded or proxied SVG opened directly cannot run script.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if strings.HasPrefix(r.URL.Path, "/api/icons/") {
			w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; sandbox")
		}
		next.ServeHTTP(w, r)
	})
}

// LimitBody rejects request bodies larger than MaxRequestBody.
func LimitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > MaxRequestBody {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBody)
		next.ServeHTTP(w, r)
	})
}
