package auth

import (
	"context"
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
	UserID   string      `json:"uid"`
	Username string      `json:"username"`
	Role     models.Role `json:"role"`
	jwt.RegisteredClaims
}

type Service struct {
	db  *bun.DB
	cfg *config.Config
}

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
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
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
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if _, err := s.db.NewInsert().Model(user).Exec(ctx); err != nil {
		return nil, err
	}
	return user, nil
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
