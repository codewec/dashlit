package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/bookmarks-dashboard/backend/internal/auth"
	"github.com/bookmarks-dashboard/backend/internal/models"
)

type DashboardHandler struct {
	db *bun.DB
}

func NewDashboardHandler(db *bun.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

func (h *DashboardHandler) List(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	ctx := r.Context()
	var list []*models.Dashboard
	q := h.db.NewSelect().Model(&list).Order("name ASC")
	if user == nil {
		q = q.Where("privacy = ?", models.PrivacyPublic)
	} else {
		q = q.Where("privacy = ? OR privacy = ? OR owner_id = ?", models.PrivacyPublic, models.PrivacyUsers, user.ID)
	}
	if err := q.Scan(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *DashboardHandler) Get(w http.ResponseWriter, r *http.Request) {
	idOrSlug := chi.URLParam(r, "id")
	user := auth.UserFromContext(r.Context())
	ctx := r.Context()

	d := new(models.Dashboard)
	q := h.db.NewSelect().Model(d).
		Relation("Groups", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("position ASC")
		}).
		Relation("Groups.Items", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("position ASC")
		}).
		Where("d.id = ? OR d.slug = ?", idOrSlug, idOrSlug)

	if err := q.Scan(ctx); err != nil {
		writeError(w, http.StatusNotFound, "dashboard not found")
		return
	}

	if !canView(user, d) {
		if user == nil {
			writeError(w, http.StatusUnauthorized, "login required")
		} else {
			writeError(w, http.StatusForbidden, "access denied")
		}
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (h *DashboardHandler) GetMain(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	ctx := r.Context()
	d := new(models.Dashboard)
	err := h.db.NewSelect().Model(d).
		Relation("Groups", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("position ASC")
		}).
		Relation("Groups.Items", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("position ASC")
		}).
		Where("is_main = ?", true).
		Scan(ctx)
	if err != nil {
		writeJSON(w, http.StatusOK, nil)
		return
	}
	if !canView(user, d) {
		if user == nil {
			writeError(w, http.StatusUnauthorized, "login required")
			return
		}
		writeError(w, http.StatusForbidden, "access denied")
		return
	}
	writeJSON(w, http.StatusOK, d)
}

type createDashboardReq struct {
	Name    string                 `json:"name"`
	Slug    string                 `json:"slug"`
	Layout  models.Layout          `json:"layout"`
	Width   models.Width           `json:"width"`
	Privacy models.Privacy         `json:"privacy"`
	Theme   *models.DashboardTheme `json:"theme"`
}

func (h *DashboardHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req createDashboardReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	req.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name required")
		return
	}
	if err := ValidateSlug(req.Slug); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Layout == "" {
		req.Layout = models.LayoutRows
	}
	if req.Width == "" {
		req.Width = models.WidthDefault
	}
	if req.Privacy == "" {
		req.Privacy = models.PrivacyPrivate
	}
	d := &models.Dashboard{
		ID:        uuid.NewString(),
		OwnerID:   user.ID,
		Name:      req.Name,
		Slug:      req.Slug,
		Layout:    req.Layout,
		Width:     req.Width,
		Privacy:   req.Privacy,
		Theme:     req.Theme,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if _, err := h.db.NewInsert().Model(d).Exec(r.Context()); err != nil {
		writeError(w, http.StatusConflict, "slug already exists")
		return
	}
	writeJSON(w, http.StatusCreated, d)
}

func (h *DashboardHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	d := new(models.Dashboard)
	if err := h.db.NewSelect().Model(d).Where("id = ?", id).Scan(r.Context()); err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if d.OwnerID != user.ID && user.Role != models.RoleAdmin {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var req createDashboardReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Name != "" {
		d.Name = req.Name
	}
	if req.Slug != "" {
		req.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
		if err := ValidateSlug(req.Slug); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		d.Slug = req.Slug
	}
	if req.Layout != "" {
		d.Layout = req.Layout
	}
	if req.Width != "" {
		d.Width = req.Width
	}
	if req.Privacy != "" {
		d.Privacy = req.Privacy
	}
	if req.Theme != nil {
		d.Theme = req.Theme
	}
	d.UpdatedAt = time.Now()
	if _, err := h.db.NewUpdate().Model(d).WherePK().Exec(r.Context()); err != nil {
		writeError(w, http.StatusConflict, "update failed")
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (h *DashboardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	d := new(models.Dashboard)
	if err := h.db.NewSelect().Model(d).Where("id = ?", id).Scan(r.Context()); err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if d.OwnerID != user.ID && user.Role != models.RoleAdmin {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	if _, err := h.db.NewDelete().Model(d).WherePK().Exec(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *DashboardHandler) SetMain(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil || user.Role != models.RoleAdmin {
		writeError(w, http.StatusForbidden, "admin only")
		return
	}
	id := chi.URLParam(r, "id")
	ctx := r.Context()
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()
	if _, err := tx.NewUpdate().Model((*models.Dashboard)(nil)).Set("is_main = ?", false).Where("is_main = ?", true).Exec(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	res, err := tx.NewUpdate().Model((*models.Dashboard)(nil)).Set("is_main = ?", true).Set("updated_at = ?", time.Now()).Where("id = ?", id).Exec(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func canView(user *models.User, d *models.Dashboard) bool {
	switch d.Privacy {
	case models.PrivacyPublic:
		return true
	case models.PrivacyUsers:
		return user != nil
	case models.PrivacyPrivate:
		return user != nil && (user.ID == d.OwnerID || user.Role == models.RoleAdmin)
	default:
		return false
	}
}
