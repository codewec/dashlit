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
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	svc  *auth.Service
	cfg  *config.Config
	oidc *auth.OIDCAuthenticator
}

func NewAuthHandler(svc *auth.Service, cfg *config.Config, oidcAuthenticator *auth.OIDCAuthenticator) *AuthHandler {
	return &AuthHandler{svc: svc, cfg: cfg, oidc: oidcAuthenticator}
}

type authConfigResponse struct {
	PasswordLoginEnabled        bool   `json:"passwordLoginEnabled"`
	PasswordRegistrationEnabled bool   `json:"passwordRegistrationEnabled"`
	OIDCEnabled                 bool   `json:"oidcEnabled"`
	OIDCButtonTitle             string `json:"oidcButtonTitle"`
}

func (h *AuthHandler) Configuration(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, authConfigResponse{
		PasswordLoginEnabled:        h.cfg.PasswordLoginEnabled(),
		PasswordRegistrationEnabled: !h.cfg.DisablePasswordRegistration,
		OIDCEnabled:                 h.oidc != nil,
		OIDCButtonTitle:             h.cfg.OIDCButtonTitle,
	})
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
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
	if h.cfg.DisablePasswordRegistration {
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
	user, err := h.svc.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		writeError(w, http.StatusConflict, "username taken or error")
		return
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

func (h *AuthHandler) oidcFailure(w http.ResponseWriter, r *http.Request, message string) {
	target := "/#/login?oidc_error=" + url.QueryEscape(message)
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
	for _, name := range []string{oidcStateCookie, oidcNonceCookie, oidcVerifierCookie, oidcReturnCookie} {
		h.setOIDCCookie(w, name, "", -1)
	}
	state := r.URL.Query().Get("state")
	if stateErr != nil || nonceErr != nil || verifierErr != nil || state == "" || subtle.ConstantTimeCompare([]byte(state), []byte(stateCookie.Value)) != 1 {
		h.oidcFailure(w, r, "OIDC session expired or is invalid")
		return
	}
	if providerError := r.URL.Query().Get("error"); providerError != "" {
		h.oidcFailure(w, r, "OIDC provider rejected the login")
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		h.oidcFailure(w, r, "OIDC response did not include a code")
		return
	}
	identity, err := h.oidc.Authenticate(r.Context(), code, nonceCookie.Value, verifierCookie.Value)
	if err != nil {
		h.oidcFailure(w, r, "OIDC token validation failed")
		return
	}
	user, err := h.svc.FindOrCreateOIDCUser(r.Context(), identity, !h.cfg.DisableOIDCRegistration)
	if errors.Is(err, auth.ErrOIDCRegistrationDisabled) {
		h.oidcFailure(w, r, "No linked account exists and OIDC registration is disabled")
		return
	}
	if err != nil {
		h.oidcFailure(w, r, "Could not load the OIDC account")
		return
	}
	token, err := h.svc.CreateToken(user)
	if err != nil {
		h.oidcFailure(w, r, "Could not create an application session")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: "token", Value: token, Path: "/", HttpOnly: true,
		Secure:   strings.HasPrefix(strings.ToLower(h.cfg.OIDCRedirectURL), "https://"),
		SameSite: http.SameSiteLaxMode, MaxAge: 7 * 24 * 3600,
	})
	returnTo := "/"
	if returnCookie != nil {
		if decoded, err := base64.RawURLEncoding.DecodeString(returnCookie.Value); err == nil {
			if validated := validOIDCReturnURL(string(decoded), r.Host); validated != "" {
				returnTo = validated
			}
		}
	}
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
