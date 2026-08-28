package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/uptrace/bun"

	"github.com/bookmarks-dashboard/backend/internal/auth"
	"github.com/bookmarks-dashboard/backend/internal/config"
	"github.com/bookmarks-dashboard/backend/internal/models"
)

type AdminHandler struct {
	db      *bun.DB
	authSvc *auth.Service
	cfg     *config.Config
}

func NewAdminHandler(db *bun.DB, authSvc *auth.Service, cfg *config.Config) *AdminHandler {
	return &AdminHandler{db: db, authSvc: authSvc, cfg: cfg}
}

type adminUser struct {
	ID             string      `json:"id"`
	Username       string      `json:"username"`
	Role           models.Role `json:"role"`
	HasPassword    bool        `json:"hasPassword"`
	HasOIDC        bool        `json:"hasOIDC"`
	DashboardCount int         `json:"dashboardCount"`
}

func (h *AdminHandler) Overview(w http.ResponseWriter, r *http.Request) {
	users := make([]*models.User, 0)
	if err := h.db.NewSelect().Model(&users).Order("username ASC").Scan(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	dashboards := make([]*models.Dashboard, 0)
	if err := h.db.NewSelect().Model(&dashboards).Relation("Owner").Order("name ASC").Scan(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	counts := make(map[string]int, len(users))
	for _, dashboard := range dashboards {
		counts[dashboard.OwnerID]++
	}
	resultUsers := make([]adminUser, 0, len(users))
	for _, user := range users {
		resultUsers = append(resultUsers, adminUser{
			ID: user.ID, Username: user.Username, Role: user.Role,
			HasPassword: user.PasswordHash != nil, HasOIDC: user.OIDCIssuer != nil && user.OIDCSubject != nil,
			DashboardCount: counts[user.ID],
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"users":      resultUsers,
		"dashboards": dashboards,
		"flags": map[string]bool{
			"DEV_MODE":                      h.cfg.DevMode,
			"DISABLE_PASSWORD_REGISTRATION": h.cfg.DisablePasswordRegistration,
			"DISABLE_OIDC_REGISTRATION":     h.cfg.DisableOIDCRegistration,
			"DISABLE_PASSWORD_LOGIN":        h.cfg.DisablePasswordLogin,
			"DISABLE_OIDC_USER_MERGE":       h.cfg.DisableOIDCUserMerge,
		},
	})
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user := new(models.User)
	if err := h.db.NewSelect().Model(user).Where("id = ?", id).Scan(r.Context()); err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	var req struct {
		Username    string `json:"username"`
		NewPassword string `json:"newPassword"`
		ResetOIDC   bool   `json:"resetOIDC"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	username := strings.TrimSpace(req.Username)
	if username == "" || len(username) > 64 {
		writeError(w, http.StatusBadRequest, "username must be 1-64 characters")
		return
	}
	if req.NewPassword != "" && len(req.NewPassword) < 6 {
		writeError(w, http.StatusBadRequest, "new password must be at least 6 characters")
		return
	}
	if user.OIDCIssuer != nil && !req.ResetOIDC {
		writeError(w, http.StatusBadRequest, "OIDC reset confirmation is required")
		return
	}
	query := h.db.NewUpdate().Model((*models.User)(nil)).Set("username = ?", username).Where("id = ?", id)
	if req.NewPassword != "" {
		hash, err := h.authSvc.HashPassword(req.NewPassword)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not hash password")
			return
		}
		query = query.Set("password_hash = ?", hash)
	}
	if req.ResetOIDC {
		query = query.Set("oidc_issuer = NULL").Set("oidc_subject = NULL")
	}
	if _, err := query.Exec(r.Context()); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			writeError(w, http.StatusConflict, "username is already taken")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	current := auth.UserFromContext(r.Context())
	id := chi.URLParam(r, "id")
	if current != nil && current.ID == id {
		writeError(w, http.StatusBadRequest, "you cannot delete your own account")
		return
	}
	res, err := h.db.NewDelete().Model((*models.User)(nil)).Where("id = ?", id).Exec(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
