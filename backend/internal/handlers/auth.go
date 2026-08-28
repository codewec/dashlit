package handlers

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bookmarks-dashboard/backend/internal/auth"
	"github.com/bookmarks-dashboard/backend/internal/config"
	"github.com/bookmarks-dashboard/backend/internal/legacy"
	"github.com/bookmarks-dashboard/backend/internal/models"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	svc    *auth.Service
	cfg    *config.Config
	oidc   *auth.OIDCAuthenticator
	legacy *legacy.Migrator
}

func NewAuthHandler(svc *auth.Service, cfg *config.Config, oidcAuthenticator *auth.OIDCAuthenticator, legacyMigrator *legacy.Migrator) *AuthHandler {
	return &AuthHandler{svc: svc, cfg: cfg, oidc: oidcAuthenticator, legacy: legacyMigrator}
}

type authConfigResponse struct {
	PasswordLoginEnabled        bool   `json:"passwordLoginEnabled"`
	PasswordRegistrationEnabled bool   `json:"passwordRegistrationEnabled"`
	OIDCEnabled                 bool   `json:"oidcEnabled"`
	OIDCButtonTitle             string `json:"oidcButtonTitle"`
	LegacyDashboardAvailable    bool   `json:"legacyDashboardAvailable"`
}

func (h *AuthHandler) Configuration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, authConfigResponse{
		PasswordLoginEnabled:        h.cfg.PasswordLoginEnabled(),
		PasswordRegistrationEnabled: h.cfg.PasswordRegistrationEnabled(),
		OIDCEnabled:                 h.oidc != nil,
		OIDCButtonTitle:             h.cfg.OIDCButtonTitle,
		LegacyDashboardAvailable:    h.legacy.Available(r.Context()),
	})
}

type loginReq struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	ImportLegacy bool   `json:"importLegacy"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.PasswordLoginEnabled() {
		writeError(w, http.StatusForbidden, "password login is disabled")
		return
	}
	var req loginReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password required")
		return
	}
	user, token, err := h.svc.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 3600,
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  user,
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.PasswordRegistrationEnabled() {
		writeError(w, http.StatusForbidden, "password registration is disabled")
		return
	}
	var req loginReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Username == "" || len(req.Password) < 6 {
		writeError(w, http.StatusBadRequest, "username required, password min 6 chars")
		return
	}
	// Capture this before Register inserts the first user. Available deliberately
	// becomes false as soon as any user exists.
	importLegacy := req.ImportLegacy && h.legacy.Available(r.Context())
	user, err := h.svc.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		writeError(w, http.StatusConflict, "username taken or error")
		return
	}
	user.AuthMethod = models.AuthMethodPassword
	if importLegacy {
		if err := h.legacy.ImportForFirstUser(r.Context(), user); err != nil {
			writeError(w, http.StatusInternalServerError, "account created, but legacy dashboard import failed")
			return
		}
	}
	token, err := h.svc.CreateToken(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token error")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 3600,
	})
	writeJSON(w, http.StatusCreated, map[string]any{
		"token": token,
		"user":  user,
	})
}

const (
	oidcStateCookie    = "oidc_state"
	oidcNonceCookie    = "oidc_nonce"
	oidcVerifierCookie = "oidc_verifier"
	oidcReturnCookie   = "oidc_return_to"
	oidcLegacyCookie   = "oidc_import_legacy"
)

func randomURLSafeString() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func (h *AuthHandler) setOIDCCookie(w http.ResponseWriter, name, value string, maxAge int) {
	secure := strings.HasPrefix(strings.ToLower(h.cfg.OIDCRedirectURL), "https://")
	http.SetCookie(w, &http.Cookie{
		Name: name, Value: value, Path: "/api/auth/oidc", HttpOnly: true,
		Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: maxAge,
	})
}

func (h *AuthHandler) OIDCLogin(w http.ResponseWriter, r *http.Request) {
	if h.oidc == nil {
		writeError(w, http.StatusNotFound, "OIDC is not configured")
		return
	}
	state, err := randomURLSafeString()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not start OIDC login")
		return
	}
	nonce, err := randomURLSafeString()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not start OIDC login")
		return
	}
	verifier := oauth2.GenerateVerifier()
	returnTo := validOIDCReturnURL(r.URL.Query().Get("return_to"), r.Host)
	h.setOIDCCookie(w, oidcStateCookie, state, 600)
	h.setOIDCCookie(w, oidcNonceCookie, nonce, 600)
	h.setOIDCCookie(w, oidcVerifierCookie, verifier, 600)
	if returnTo != "" {
		h.setOIDCCookie(w, oidcReturnCookie, base64.RawURLEncoding.EncodeToString([]byte(returnTo)), 600)
	}
	if r.URL.Query().Get("import_legacy") == "1" && h.legacy.Available(r.Context()) {
		h.setOIDCCookie(w, oidcLegacyCookie, "1", 600)
	}
	http.Redirect(w, r, h.oidc.AuthorizationURL(state, nonce, verifier), http.StatusFound)
}

func validOIDCReturnURL(raw, requestHost string) string {
	u, err := url.Parse(raw)
	if err != nil || !u.IsAbs() || u.User != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return ""
	}
	requestURL, err := url.Parse("http://" + requestHost)
	if err != nil || !strings.EqualFold(u.Hostname(), requestURL.Hostname()) {
		return ""
	}
	return u.Scheme + "://" + u.Host + "/"
}

func oidcReturnURL(returnCookie *http.Cookie, requestHost string) string {
	if returnCookie != nil {
		if decoded, err := base64.RawURLEncoding.DecodeString(returnCookie.Value); err == nil {
			if validated := validOIDCReturnURL(string(decoded), requestHost); validated != "" {
				return validated
			}
		}
	}
	return "/"
}

func (h *AuthHandler) oidcFailure(w http.ResponseWriter, r *http.Request, returnCookie *http.Cookie, message string) {
	target := oidcReturnURL(returnCookie, r.Host) + "#/login?oidc_error=" + url.QueryEscape(message)
	http.Redirect(w, r, target, http.StatusFound)
}

func (h *AuthHandler) OIDCCallback(w http.ResponseWriter, r *http.Request) {
	if h.oidc == nil {
		writeError(w, http.StatusNotFound, "OIDC is not configured")
		return
	}
	stateCookie, stateErr := r.Cookie(oidcStateCookie)
	nonceCookie, nonceErr := r.Cookie(oidcNonceCookie)
	verifierCookie, verifierErr := r.Cookie(oidcVerifierCookie)
	returnCookie, _ := r.Cookie(oidcReturnCookie)
	legacyCookie, _ := r.Cookie(oidcLegacyCookie)
	for _, name := range []string{oidcStateCookie, oidcNonceCookie, oidcVerifierCookie, oidcReturnCookie, oidcLegacyCookie} {
		h.setOIDCCookie(w, name, "", -1)
	}
	state := r.URL.Query().Get("state")
	if stateErr != nil || nonceErr != nil || verifierErr != nil || state == "" || subtle.ConstantTimeCompare([]byte(state), []byte(stateCookie.Value)) != 1 {
		h.oidcFailure(w, r, returnCookie, "OIDC session expired or is invalid")
		return
	}
	if providerError := r.URL.Query().Get("error"); providerError != "" {
		h.oidcFailure(w, r, returnCookie, "OIDC provider rejected the login")
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		h.oidcFailure(w, r, returnCookie, "OIDC response did not include a code")
		return
	}
	identity, err := h.oidc.Authenticate(r.Context(), code, nonceCookie.Value, verifierCookie.Value)
	if err != nil {
		h.oidcFailure(w, r, returnCookie, "OIDC token validation failed")
		return
	}
	user, err := h.svc.FindOrCreateOIDCUser(r.Context(), identity, !h.cfg.DisableOIDCRegistration)
	if errors.Is(err, auth.ErrOIDCRegistrationDisabled) {
		h.oidcFailure(w, r, returnCookie, "Registration through OIDC is disabled")
		return
	}
	if err != nil {
		h.oidcFailure(w, r, returnCookie, "Could not load the OIDC account")
		return
	}
	user.AuthMethod = models.AuthMethodOIDC
	if legacyCookie != nil && legacyCookie.Value == "1" {
		if err := h.legacy.ImportForFirstUser(r.Context(), user); err != nil {
			h.oidcFailure(w, r, returnCookie, "Account created, but legacy dashboard import failed")
			return
		}
	}
	token, err := h.svc.CreateToken(user)
	if err != nil {
		h.oidcFailure(w, r, returnCookie, "Could not create an application session")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: "token", Value: token, Path: "/", HttpOnly: true,
		Secure:   strings.HasPrefix(strings.ToLower(h.cfg.OIDCRedirectURL), "https://"),
		SameSite: http.SameSiteLaxMode, MaxAge: 7 * 24 * 3600,
	})
	returnTo := oidcReturnURL(returnCookie, r.Host)
	separator := "?"
	if strings.Contains(returnTo, "?") {
		separator = "&"
	}
	http.Redirect(w, r, returnTo+separator+"oidc=1", http.StatusFound)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Username    string `json:"username"`
		NewPassword string `json:"newPassword"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if user.AuthMethod == models.AuthMethodOIDC && strings.TrimSpace(req.Username) != user.Username {
		writeError(w, http.StatusForbidden, "username cannot be changed during an OIDC session")
		return
	}
	if req.NewPassword != "" && !h.cfg.PasswordLoginEnabled() {
		writeError(w, http.StatusForbidden, "password login is disabled")
		return
	}
	if err := h.svc.UpdateProfile(r.Context(), user, req.Username, req.NewPassword); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			writeError(w, http.StatusConflict, "username is already taken")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, user)
}
