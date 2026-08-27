package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
)

var reservedSlugs = map[string]bool{
	"login": true, "logout": true, "settings": true, "admin": true,
	"api": true, "assets": true, "_app": true, "oidc": true, "auth": true,
	"icons": true, "static": true, "favicon.ico": true,
}

var slugRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,62}[a-z0-9])?$`)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func ValidateSlug(slug string) error {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if slug == "" {
		return errInvalid("slug is required")
	}
	if reservedSlugs[slug] {
		return errInvalid("slug is reserved")
	}
	if !slugRegex.MatchString(slug) {
		return errInvalid("slug must be 1-64 chars, lowercase alphanumeric and hyphens")
	}
	return nil
}

type apiError struct {
	msg string
}

func (e *apiError) Error() string { return e.msg }

func errInvalid(msg string) error { return &apiError{msg: msg} }
