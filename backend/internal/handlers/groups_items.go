package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/bookmarks-dashboard/backend/internal/auth"
	"github.com/bookmarks-dashboard/backend/internal/models"
)

type GroupItemHandler struct {
	db *bun.DB
}

func NewGroupItemHandler(db *bun.DB) *GroupItemHandler {
	return &GroupItemHandler{db: db}
}

func (h *GroupItemHandler) canEditDashboard(r *http.Request, dashboardID string) (*models.Dashboard, bool) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		return nil, false
	}
	d := new(models.Dashboard)
	if err := h.db.NewSelect().Model(d).Where("id = ?", dashboardID).Scan(r.Context()); err != nil {
		return nil, false
	}
	if d.OwnerID != user.ID && user.Role != models.RoleAdmin {
		return nil, false
	}
	return d, true
}

type createGroupReq struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Icon        string          `json:"icon"`
	IconDark    string          `json:"iconDark"`
	ItemSize    models.ItemSize `json:"itemSize"`
	Position    int             `json:"position"`
}

func (h *GroupItemHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	dashboardID := chi.URLParam(r, "dashboardID")
	if _, ok := h.canEditDashboard(r, dashboardID); !ok {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var req createGroupReq
	if err := decodeJSON(r, &req); err != nil || req.Title == "" {
		writeError(w, http.StatusBadRequest, "title required")
		return
	}
	if req.ItemSize == "" {
		req.ItemSize = models.Size1x1
	}
	g := &models.Group{
		ID:          uuid.NewString(),
		DashboardID: dashboardID,
		Title:       req.Title,
		Description: req.Description,
		Icon:        req.Icon,
		IconDark:    req.IconDark,
		ItemSize:    req.ItemSize,
		Position:    req.Position,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if _, err := h.db.NewInsert().Model(g).Exec(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, g)
}

func (h *GroupItemHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	g := new(models.Group)
	if err := h.db.NewSelect().Model(g).Where("id = ?", id).Scan(r.Context()); err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if _, ok := h.canEditDashboard(r, g.DashboardID); !ok {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var req struct {
		Title       *string          `json:"title"`
		Description *string          `json:"description"`
		Icon        *string          `json:"icon"`
		IconDark    *string          `json:"iconDark"`
		ItemSize    *models.ItemSize `json:"itemSize"`
		Position    *int             `json:"position"`
		Collapsed   *bool            `json:"collapsed"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Title != nil {
		g.Title = *req.Title
	}
	if req.Description != nil {
		g.Description = *req.Description
	}
	if req.Icon != nil {
		g.Icon = *req.Icon
	}
	if req.IconDark != nil {
		g.IconDark = *req.IconDark
	}
	if req.ItemSize != nil {
		g.ItemSize = *req.ItemSize
	}
	if req.Position != nil {
		g.Position = *req.Position
	}
	if req.Collapsed != nil {
		g.Collapsed = *req.Collapsed
	}
	g.UpdatedAt = time.Now()
	if _, err := h.db.NewUpdate().Model(g).WherePK().Exec(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, g)
}

func (h *GroupItemHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	g := new(models.Group)
	if err := h.db.NewSelect().Model(g).Where("id = ?", id).Scan(r.Context()); err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if _, ok := h.canEditDashboard(r, g.DashboardID); !ok {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	if _, err := h.db.NewDelete().Model(g).WherePK().Exec(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type createItemReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Icon        string `json:"icon"`
	IconDark    string `json:"iconDark"`
	Position    int    `json:"position"`
}

func (h *GroupItemHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupID")
	g := new(models.Group)
	if err := h.db.NewSelect().Model(g).Where("id = ?", groupID).Scan(r.Context()); err != nil {
		writeError(w, http.StatusNotFound, "group not found")
		return
	}
	if _, ok := h.canEditDashboard(r, g.DashboardID); !ok {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var req createItemReq
	if err := decodeJSON(r, &req); err != nil || req.Title == "" || req.URL == "" {
		writeError(w, http.StatusBadRequest, "title and url required")
		return
	}
	if req.Icon == "" {
		req.Icon = "mdi:link"
	}
	item := &models.Item{
		ID:          uuid.NewString(),
		GroupID:     groupID,
		Title:       req.Title,
		Description: req.Description,
		URL:         req.URL,
		Icon:        req.Icon,
		IconDark:    req.IconDark,
		Position:    req.Position,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if _, err := h.db.NewInsert().Model(item).Exec(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *GroupItemHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item := new(models.Item)
	if err := h.db.NewSelect().Model(item).Where("id = ?", id).Scan(r.Context()); err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	g := new(models.Group)
	if err := h.db.NewSelect().Model(g).Where("id = ?", item.GroupID).Scan(r.Context()); err != nil {
		writeError(w, http.StatusNotFound, "group not found")
		return
	}
	if _, ok := h.canEditDashboard(r, g.DashboardID); !ok {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var req struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		URL         *string `json:"url"`
		Icon        *string `json:"icon"`
		IconDark    *string `json:"iconDark"`
		Position    *int    `json:"position"`
		GroupID     *string `json:"groupId"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Title != nil {
		item.Title = *req.Title
	}
	if req.Description != nil {
		item.Description = *req.Description
	}
	if req.URL != nil {
		item.URL = *req.URL
	}
	if req.Icon != nil {
		item.Icon = *req.Icon
	}
	if req.IconDark != nil {
		item.IconDark = *req.IconDark
	}
	if req.Position != nil {
		item.Position = *req.Position
	}
	if req.GroupID != nil {
		item.GroupID = *req.GroupID
	}
	item.UpdatedAt = time.Now()
	if _, err := h.db.NewUpdate().Model(item).WherePK().Exec(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *GroupItemHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item := new(models.Item)
	if err := h.db.NewSelect().Model(item).Where("id = ?", id).Scan(r.Context()); err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	g := new(models.Group)
	if err := h.db.NewSelect().Model(g).Where("id = ?", item.GroupID).Scan(r.Context()); err != nil {
		writeError(w, http.StatusNotFound, "group not found")
		return
	}
	if _, ok := h.canEditDashboard(r, g.DashboardID); !ok {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	if _, err := h.db.NewDelete().Model(item).WherePK().Exec(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Batch layout update after DnD
type layoutUpdateReq struct {
	Groups []struct {
		ID       string `json:"id"`
		Position int    `json:"position"`
	} `json:"groups"`
	Items []struct {
		ID       string `json:"id"`
		GroupID  string `json:"groupId"`
		Position int    `json:"position"`
	} `json:"items"`
}

func (h *GroupItemHandler) UpdateLayout(w http.ResponseWriter, r *http.Request) {
	dashboardID := chi.URLParam(r, "dashboardID")
	if _, ok := h.canEditDashboard(r, dashboardID); !ok {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var req layoutUpdateReq
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
	for _, g := range req.Groups {
		if _, err := tx.NewUpdate().Model((*models.Group)(nil)).
			Set("position = ?", g.Position).
			Set("updated_at = ?", time.Now()).
			Where("id = ? AND dashboard_id = ?", g.ID, dashboardID).
			Exec(ctx); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	for _, it := range req.Items {
		if _, err := tx.NewUpdate().Model((*models.Item)(nil)).
			Set("position = ?", it.Position).
			Set("group_id = ?", it.GroupID).
			Set("updated_at = ?", time.Now()).
			Where("id = ?", it.ID).
			Exec(ctx); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *GroupItemHandler) CloneGroup(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	g := new(models.Group)
	if err := h.db.NewSelect().Model(g).Where("id = ?", id).Relation("Items", func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.OrderExpr("position ASC")
	}).Scan(r.Context()); err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if _, ok := h.canEditDashboard(r, g.DashboardID); !ok {
		return
	}
	var maxPos int
	_ = h.db.NewSelect().Model((*models.Group)(nil)).ColumnExpr("COALESCE(MAX(position), -1)").Where("dashboard_id = ?", g.DashboardID).Scan(r.Context(), &maxPos)
	now := time.Now()
	ng := &models.Group{
		ID: uuid.NewString(), DashboardID: g.DashboardID,
		Title: g.Title + " (copy)", Description: g.Description, Icon: g.Icon, IconDark: g.IconDark,
		ItemSize: g.ItemSize, Position: maxPos + 1, CreatedAt: now, UpdatedAt: now,
	}
	if _, err := h.db.NewInsert().Model(ng).Exec(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for i, it := range g.Items {
		ni := &models.Item{
			ID: uuid.NewString(), GroupID: ng.ID,
			Title: it.Title, Description: it.Description, URL: it.URL,
			Icon: it.Icon, IconDark: it.IconDark, Position: i,
			CreatedAt: now, UpdatedAt: now,
		}
		if _, err := h.db.NewInsert().Model(ni).Exec(r.Context()); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	_ = h.db.NewSelect().Model(ng).WherePK().Relation("Items").Scan(r.Context())
	writeJSON(w, http.StatusCreated, ng)
}

func (h *GroupItemHandler) CloneItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item := new(models.Item)
	if err := h.db.NewSelect().Model(item).Where("id = ?", id).Scan(r.Context()); err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	g := new(models.Group)
	if err := h.db.NewSelect().Model(g).Where("id = ?", item.GroupID).Scan(r.Context()); err != nil {
		writeError(w, http.StatusNotFound, "group not found")
		return
	}
	if _, ok := h.canEditDashboard(r, g.DashboardID); !ok {
		return
	}
	var maxPos int
	_ = h.db.NewSelect().Model((*models.Item)(nil)).ColumnExpr("COALESCE(MAX(position), -1)").Where("group_id = ?", item.GroupID).Scan(r.Context(), &maxPos)
	now := time.Now()
	ni := &models.Item{
		ID: uuid.NewString(), GroupID: item.GroupID,
		Title: item.Title + " (copy)", Description: item.Description, URL: item.URL,
		Icon: item.Icon, IconDark: item.IconDark, Position: maxPos + 1,
		CreatedAt: now, UpdatedAt: now,
	}
	if _, err := h.db.NewInsert().Model(ni).Exec(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, ni)
}
