package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/bookmarks-dashboard/backend/internal/config"
)

type OIDCIdentity struct {
	Issuer            string
	Subject           string
	PreferredUsername string
	Email             string
	Name              string
}

type OIDCAuthenticator struct {
	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier
}

func NewOIDCAuthenticator(ctx context.Context, cfg *config.Config) (*OIDCAuthenticator, error) {
	if !cfg.OIDCEnabled() {
		return nil, nil
	}
	provider, err := oidc.NewProvider(ctx, strings.TrimRight(cfg.OIDCIssuer, "/"))
	if err != nil {
		return nil, fmt.Errorf("discover OIDC provider: %w", err)
	}
	return &OIDCAuthenticator{
		oauth2Config: oauth2.Config{
			ClientID:     cfg.OIDCClientID,
			ClientSecret: cfg.OIDCClientSecret,
			RedirectURL:  cfg.OIDCRedirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeEmail},
		},
		verifier: provider.Verifier(&oidc.Config{ClientID: cfg.OIDCClientID}),
	}, nil
}

func (a *OIDCAuthenticator) AuthorizationURL(state, nonce, verifier string) string {
	return a.oauth2Config.AuthCodeURL(state, oidc.Nonce(nonce), oauth2.S256ChallengeOption(verifier))
}

func (a *OIDCAuthenticator) Authenticate(ctx context.Context, code, expectedNonce, verifier string) (*OIDCIdentity, error) {
	token, err := a.oauth2Config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, fmt.Errorf("exchange authorization code: %w", err)
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, errors.New("OIDC response did not include an ID token")
	}
	idToken, err := a.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("verify ID token: %w", err)
	}
	if expectedNonce == "" || subtle.ConstantTimeCompare([]byte(idToken.Nonce), []byte(expectedNonce)) != 1 {
		return nil, errors.New("invalid OIDC nonce")
	}

	var claims struct {
		PreferredUsername string `json:"preferred_username"`
		Email             string `json:"email"`
		Name              string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("decode ID token claims: %w", err)
	}
	if idToken.Subject == "" {
		return nil, errors.New("OIDC subject is empty")
	}
	return &OIDCIdentity{
		Issuer:            idToken.Issuer,
		Subject:           idToken.Subject,
		PreferredUsername: claims.PreferredUsername,
		Email:             claims.Email,
		Name:              claims.Name,
	}, nil
}

var invalidUsernameChars = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func oidcUsername(identity *OIDCIdentity) string {
	candidates := []string{identity.PreferredUsername}
	if local, _, ok := strings.Cut(identity.Email, "@"); ok {
		candidates = append(candidates, local)
	}
	candidates = append(candidates, identity.Name)
	for _, candidate := range candidates {
		candidate = strings.Trim(invalidUsernameChars.ReplaceAllString(strings.TrimSpace(candidate), "-"), "-._")
		if candidate != "" {
			if len(candidate) > 64 {
				candidate = candidate[:64]
			}
			return candidate
		}
	}
	hash := sha256.Sum256([]byte(identity.Issuer + "\x00" + identity.Subject))
	return "oidc-" + hex.EncodeToString(hash[:6])
}
