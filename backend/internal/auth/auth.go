package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/bookmarks-dashboard/backend/internal/config"
	"github.com/bookmarks-dashboard/backend/internal/models"
	"github.com/uptrace/bun"
)

type Claims struct {
	UserID     string            `json:"uid"`
	Username   string            `json:"username"`
	Role       models.Role       `json:"role"`
	AuthMethod models.AuthMethod `json:"auth_method"`
	jwt.RegisteredClaims
}

type Service struct {
	db  *bun.DB
	cfg *config.Config
}

var ErrOIDCRegistrationDisabled = errors.New("creating users through OIDC is disabled")

func NewService(db *bun.DB, cfg *config.Config) *Service {
	return &Service{db: db, cfg: cfg}
}

func (s *Service) HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

func (s *Service) CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (s *Service) CreateToken(user *models.User) (string, error) {
	claims := Claims{
		UserID:     user.ID,
		Username:   user.Username,
		Role:       user.Role,
		AuthMethod: user.AuthMethod,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *Service) ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (s *Service) Register(ctx context.Context, username, password string) (*models.User, error) {
	count, err := s.db.NewSelect().Model((*models.User)(nil)).Count(ctx)
	if err != nil {
		return nil, err
	}
	role := models.RoleUser
	if count == 0 {
		role = models.RoleAdmin
	}

	hash, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := &models.User{
		ID:           uuid.NewString(),
		Username:     username,
		PasswordHash: &hash,
		Role:         role,
	}
	if _, err := s.db.NewInsert().Model(user).Exec(ctx); err != nil {
		return nil, err
	}
	return user, nil
}

// BootstrapAdmin creates the first administrator from deployment credentials.
// Once any user exists, the supplied values are deliberately ignored.
func (s *Service) BootstrapAdmin(ctx context.Context, username, password string) (bool, error) {
	count, err := s.db.NewSelect().Model((*models.User)(nil)).Count(ctx)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}

	username = strings.TrimSpace(username)
	if username == "" && password == "" {
		return false, nil
	}
	if username == "" || password == "" {
		return false, errors.New("INITIAL_ADMIN_USERNAME and INITIAL_ADMIN_PASSWORD must both be set")
	}
	if len(username) > 64 {
		return false, errors.New("INITIAL_ADMIN_USERNAME must not exceed 64 characters")
	}
	if len(password) < 6 {
		return false, errors.New("INITIAL_ADMIN_PASSWORD must be at least 6 characters")
	}

	user, err := s.Register(ctx, username, password)
	if err != nil {
		return false, err
	}
	if user.Role != models.RoleAdmin {
		return false, errors.New("bootstrap user was not created as administrator")
	}
	return true, nil
}

func (s *Service) Login(ctx context.Context, username, password string) (*models.User, string, error) {
	user := new(models.User)
	err := s.db.NewSelect().Model(user).Where("username = ?", username).Scan(ctx)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}
	if user.PasswordHash == nil || !s.CheckPassword(*user.PasswordHash, password) {
		return nil, "", errors.New("invalid credentials")
	}
	user.AuthMethod = models.AuthMethodPassword
	token, err := s.CreateToken(user)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *Service) GetUser(ctx context.Context, id string) (*models.User, error) {
	user := new(models.User)
	err := s.db.NewSelect().Model(user).Where("id = ?", id).Scan(ctx)
	return user, err
}

func (s *Service) UpdateProfile(ctx context.Context, user *models.User, username, newPassword string) error {
	username = strings.TrimSpace(username)
	if username == "" || len(username) > 64 {
		return errors.New("username must be 1-64 characters")
	}

	var passwordHash *string
	if newPassword != "" {
		if len(newPassword) < 6 {
			return errors.New("new password must be at least 6 characters")
		}
		hash, err := s.HashPassword(newPassword)
		if err != nil {
			return err
		}
		passwordHash = &hash
	}

	query := s.db.NewUpdate().Model((*models.User)(nil)).Set("username = ?", username).Where("id = ?", user.ID)
	if passwordHash != nil {
		query = query.Set("password_hash = ?", *passwordHash)
	}
	if _, err := query.Exec(ctx); err != nil {
		return err
	}
	user.Username = username
	if passwordHash != nil {
		user.PasswordHash = passwordHash
	}
	return nil
}

func (s *Service) FindOrCreateOIDCUser(ctx context.Context, identity *OIDCIdentity, allowCreate bool) (*models.User, error) {
	user := new(models.User)
	err := s.db.NewSelect().Model(user).
		Where("oidc_issuer = ? AND oidc_subject = ?", identity.Issuer, identity.Subject).
		Scan(ctx)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	base := oidcUsername(identity)
	if !s.cfg.DisableOIDCUserMerge {
		existing := new(models.User)
		err := s.db.NewSelect().Model(existing).Where("username = ?", base).Scan(ctx)
		if err == nil && existing.OIDCIssuer == nil && existing.OIDCSubject == nil {
			res, err := s.db.NewUpdate().Model(existing).
				Set("oidc_issuer = ?", identity.Issuer).
				Set("oidc_subject = ?", identity.Subject).
				WherePK().
				Where("oidc_issuer IS NULL AND oidc_subject IS NULL").
				Exec(ctx)
			if err != nil {
				return nil, err
			}
			if affected, _ := res.RowsAffected(); affected == 1 {
				existing.OIDCIssuer = &identity.Issuer
				existing.OIDCSubject = &identity.Subject
				return existing, nil
			}
		} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}

	if !allowCreate {
		return nil, ErrOIDCRegistrationDisabled
	}

	username := base
	for suffix := 2; ; suffix++ {
		count, err := s.db.NewSelect().Model((*models.User)(nil)).Where("username = ?", username).Count(ctx)
		if err != nil {
			return nil, err
		}
		if count == 0 {
			break
		}
		username = fmt.Sprintf("%s-%d", base, suffix)
	}

	count, err := s.db.NewSelect().Model((*models.User)(nil)).Count(ctx)
	if err != nil {
		return nil, err
	}
	role := models.RoleUser
	if count == 0 {
		role = models.RoleAdmin
	}
	user = &models.User{
		ID:          uuid.NewString(),
		Username:    username,
		Role:        role,
		OIDCSubject: &identity.Subject,
		OIDCIssuer:  &identity.Issuer,
	}
	if _, err := s.db.NewInsert().Model(user).Exec(ctx); err != nil {
		return nil, err
	}
	return user, nil
}

type contextKey string

const UserContextKey contextKey = "user"

func Middleware(svc *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			var tokenStr string
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			} else if c, err := r.Cookie("token"); err == nil {
				tokenStr = c.Value
			}
			if tokenStr == "" {
				next.ServeHTTP(w, r)
				return
			}
			claims, err := svc.ParseToken(tokenStr)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			user, err := svc.GetUser(r.Context(), claims.UserID)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			if claims.AuthMethod != models.AuthMethodPassword && claims.AuthMethod != models.AuthMethodOIDC {
				next.ServeHTTP(w, r)
				return
			}
			user.AuthMethod = claims.AuthMethod
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromContext(ctx context.Context) *models.User {
	u, _ := ctx.Value(UserContextKey).(*models.User)
	return u
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if UserFromContext(r.Context()) == nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := UserFromContext(r.Context())
		if u == nil || u.Role != models.RoleAdmin {
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
