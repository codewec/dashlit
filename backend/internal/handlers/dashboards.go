package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
	list := make([]*models.Dashboard, 0)
	q := h.db.NewSelect().Model(&list).Relation("Owner").Order("name ASC")
	if user == nil {
		q = q.Where("privacy = ?", models.PrivacyPublic)
	} else if user.Role != models.RoleAdmin {
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
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
	Icon        string         `json:"icon"`
	IconDark    string         `json:"iconDark"`
	Layout      models.Layout  `json:"layout"`
	Width       models.Width   `json:"width"`
	Privacy     models.Privacy `json:"privacy"`
	CleanMode   bool           `json:"cleanMode"`
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
		ID:          uuid.NewString(),
		OwnerID:     user.ID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Icon:        req.Icon,
		IconDark:    req.IconDark,
		Layout:      req.Layout,
		Width:       req.Width,
		Privacy:     req.Privacy,
	}
	d.CleanMode = req.CleanMode
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
	d.Description = req.Description
	d.Icon = req.Icon
	d.IconDark = req.IconDark
	if req.Layout != "" {
		d.Layout = req.Layout
	}
	if req.Width != "" {
		d.Width = req.Width
	}
	if req.Privacy != "" {
		if d.IsMain && req.Privacy == models.PrivacyPrivate {
			writeError(w, http.StatusBadRequest, "system default dashboard cannot be private")
			return
		}
		d.Privacy = req.Privacy
	}
	d.CleanMode = req.CleanMode
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
	d := new(models.Dashboard)
	if err := h.db.NewSelect().Model(d).Where("id = ?", id).Scan(r.Context()); err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if d.Privacy == models.PrivacyPrivate {
		writeError(w, http.StatusBadRequest, "system default dashboard cannot be private")
		return
	}
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
	res, err := tx.NewUpdate().Model((*models.Dashboard)(nil)).Set("is_main = ?", true).Where("id = ?", id).Exec(ctx)
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

// SetDefault selects a dashboard as the current user's personal default.
// Personal defaults are restricted to owned dashboards and are independent
// from the administrator-managed system main dashboard.
func (h *DashboardHandler) SetDefault(w http.ResponseWriter, r *http.Request) {
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
	if d.OwnerID != user.ID {
		writeError(w, http.StatusForbidden, "only an owned dashboard can be the personal default")
		return
	}

	var req struct {
		IsDefault bool `json:"isDefault"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	ctx := r.Context()
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	if req.IsDefault {
		if _, err := tx.NewUpdate().Model((*models.Dashboard)(nil)).
			Set("is_default = ?", false).
			Where("owner_id = ? AND is_default = ?", user.ID, true).
			Exec(ctx); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	res, err := tx.NewUpdate().Model((*models.Dashboard)(nil)).
		Set("is_default = ?", req.IsDefault).
		Where("id = ? AND owner_id = ?", id, user.ID).
		Exec(ctx)
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

// ExportFormatVersion is bumped when the export JSON schema changes.
const ExportFormatVersion = 1

type exportPayload struct {
	Version   int             `json:"version"`
	Dashboard exportDashboard `json:"dashboard"`
}

type exportDashboard struct {
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
	Icon        string         `json:"icon"`
	IconDark    string         `json:"iconDark"`
	Layout      models.Layout  `json:"layout"`
	Width       models.Width   `json:"width"`
	Privacy     models.Privacy `json:"privacy"`
	CleanMode   bool           `json:"cleanMode"`
	Groups      []exportGroup  `json:"groups"`
}

type exportGroup struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Icon        string          `json:"icon"`
	IconDark    string          `json:"iconDark"`
	ItemSize    models.ItemSize `json:"itemSize"`
	Position    int             `json:"position"`
	Items       []exportItem    `json:"items"`
}

type exportItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Icon        string `json:"icon"`
	IconDark    string `json:"iconDark"`
	Position    int    `json:"position"`
}

func (h *DashboardHandler) loadFull(ctx context.Context, idOrSlug string) (*models.Dashboard, error) {
	d := new(models.Dashboard)
	err := h.db.NewSelect().Model(d).
		Where("id = ? OR slug = ?", idOrSlug, idOrSlug).
		Relation("Groups", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.OrderExpr("position ASC")
		}).
		Relation("Groups.Items", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.OrderExpr("position ASC")
		}).
		Scan(ctx)
	return d, err
}

func dashboardToExport(d *models.Dashboard) exportPayload {
	eg := make([]exportGroup, 0, len(d.Groups))
	for _, g := range d.Groups {
		ei := make([]exportItem, 0, len(g.Items))
		for _, it := range g.Items {
			ei = append(ei, exportItem{
				Title: it.Title, Description: it.Description, URL: it.URL,
				Icon: it.Icon, IconDark: it.IconDark, Position: it.Position,
			})
		}
		eg = append(eg, exportGroup{
			Title: g.Title, Description: g.Description, Icon: g.Icon, IconDark: g.IconDark,
			ItemSize: g.ItemSize, Position: g.Position, Items: ei,
		})
	}
	return exportPayload{
		Version: ExportFormatVersion,
		Dashboard: exportDashboard{
			Name: d.Name, Slug: d.Slug, Description: d.Description,
			Icon: d.Icon, IconDark: d.IconDark, Layout: d.Layout, Width: d.Width,
			Privacy: d.Privacy, CleanMode: d.CleanMode, Groups: eg,
		},
	}
}

func (h *DashboardHandler) Export(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	d, err := h.loadFull(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if !canView(user, d) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	payload := dashboardToExport(d)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+d.Slug+".dashlit.json\"")
	_ = json.NewEncoder(w).Encode(payload)
}

func uniqueSlug(ctx context.Context, db *bun.DB, base string) string {
	base = strings.ToLower(strings.TrimSpace(base))
	if base == "" {
		base = "dashboard"
	}
	slug := base
	for i := 0; i < 50; i++ {
		exists, _ := db.NewSelect().Model((*models.Dashboard)(nil)).Where("slug = ?", slug).Exists(ctx)
		if !exists {
			return slug
		}
		slug = fmt.Sprintf("%s-%d", base, i+2)
	}
	return fmt.Sprintf("%s-%s", base, uuid.NewString()[:8])
}

func (h *DashboardHandler) importPayload(ctx context.Context, user *models.User, payload exportPayload, namePrefix string) (*models.Dashboard, error) {
	if payload.Version < 1 || payload.Version > ExportFormatVersion {
		return nil, fmt.Errorf("unsupported export version %d (supported: 1–%d)", payload.Version, ExportFormatVersion)
	}
	src := payload.Dashboard
	name := src.Name
	if namePrefix != "" {
		name = namePrefix + name
	}
	slugBase := src.Slug
	if slugBase == "" {
		slugBase = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	}
	slug := uniqueSlug(ctx, h.db, slugBase)
	layout := src.Layout
	if layout == "" {
		layout = models.LayoutRows
	}
	width := src.Width
	if width == "" {
		width = models.WidthDefault
	}
	privacy := src.Privacy
	if privacy == "" {
		privacy = models.PrivacyPrivate
	}
	d := &models.Dashboard{
		ID: uuid.NewString(), OwnerID: user.ID, Name: name, Slug: slug,
		Description: src.Description, Icon: src.Icon, IconDark: src.IconDark,
		Layout: layout, Width: width, Privacy: privacy, CleanMode: src.CleanMode,
	}
	if _, err := h.db.NewInsert().Model(d).Exec(ctx); err != nil {
		return nil, err
	}
	for gi, gs := range src.Groups {
		g := &models.Group{
			ID: uuid.NewString(), DashboardID: d.ID,
			Title: gs.Title, Description: gs.Description, Icon: gs.Icon, IconDark: gs.IconDark,
			ItemSize: gs.ItemSize, Position: gs.Position,
		}
		if g.ItemSize == "" {
			g.ItemSize = models.Size1x1
		}
		if g.Position == 0 && gi > 0 {
			g.Position = gi
		}
		if _, err := h.db.NewInsert().Model(g).Exec(ctx); err != nil {
			return nil, err
		}
		for ii, is := range gs.Items {
			it := &models.Item{
				ID: uuid.NewString(), GroupID: g.ID,
				Title: is.Title, Description: is.Description, URL: is.URL,
				Icon: is.Icon, IconDark: is.IconDark, Position: is.Position,
			}
			if it.Icon == "" {
				it.Icon = "mdi:link"
			}
			if it.Position == 0 && ii > 0 {
				it.Position = ii
			}
			if _, err := h.db.NewInsert().Model(it).Exec(ctx); err != nil {
				return nil, err
			}
		}
	}
	return h.loadFull(ctx, d.ID)
}

func (h *DashboardHandler) Import(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var payload exportPayload
	if err := decodeJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	d, err := h.importPayload(r.Context(), user, payload, "")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, d)
}

func (h *DashboardHandler) Clone(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	src, err := h.loadFull(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if !canView(user, src) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	payload := dashboardToExport(src)
	d, err := h.importPayload(r.Context(), user, payload, "Copy of ")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, d)
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
