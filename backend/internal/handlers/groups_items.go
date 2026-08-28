package handlers

import (
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/bookmarks-dashboard/backend/internal/auth"
	"github.com/bookmarks-dashboard/backend/internal/models"
)

type GroupItemHandler struct {
	db         *bun.DB
	httpClient *http.Client
}

func NewGroupItemHandler(db *bun.DB) *GroupItemHandler {
	return &GroupItemHandler{
		db: db,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
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
	Title        string `json:"title"`
	Description  string `json:"description"`
	URL          string `json:"url"`
	Icon         string `json:"icon"`
	IconDark     string `json:"iconDark"`
	PingEnabled  bool   `json:"pingEnabled"`
	PingOnlyDown bool   `json:"pingOnlyDown"`
	Position     int    `json:"position"`
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
		ID:           uuid.NewString(),
		GroupID:      groupID,
		Title:        req.Title,
		Description:  req.Description,
		URL:          req.URL,
		Icon:         req.Icon,
		IconDark:     req.IconDark,
		PingEnabled:  req.PingEnabled,
		PingOnlyDown: req.PingOnlyDown,
		Position:     req.Position,
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
		Title        *string `json:"title"`
		Description  *string `json:"description"`
		URL          *string `json:"url"`
		Icon         *string `json:"icon"`
		IconDark     *string `json:"iconDark"`
		PingEnabled  *bool   `json:"pingEnabled"`
		PingOnlyDown *bool   `json:"pingOnlyDown"`
		Position     *int    `json:"position"`
		GroupID      *string `json:"groupId"`
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
	if req.PingEnabled != nil {
		item.PingEnabled = *req.PingEnabled
	}
	if req.PingOnlyDown != nil {
		item.PingOnlyDown = *req.PingOnlyDown
	}
	if req.Position != nil {
		item.Position = *req.Position
	}
	if req.GroupID != nil {
		item.GroupID = *req.GroupID
	}
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
	ng := &models.Group{
		ID: uuid.NewString(), DashboardID: g.DashboardID,
		Title: g.Title + " (copy)", Description: g.Description, Icon: g.Icon, IconDark: g.IconDark,
		ItemSize: g.ItemSize, Position: maxPos + 1,
	}
	if _, err := h.db.NewInsert().Model(ng).Exec(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for i, it := range g.Items {
		ni := &models.Item{
			ID: uuid.NewString(), GroupID: ng.ID,
			Title: it.Title, Description: it.Description, URL: it.URL,
			Icon: it.Icon, IconDark: it.IconDark, PingEnabled: it.PingEnabled,
			PingOnlyDown: it.PingOnlyDown, Position: i,
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
	ni := &models.Item{
		ID: uuid.NewString(), GroupID: item.GroupID,
		Title: item.Title + " (copy)", Description: item.Description, URL: item.URL,
		Icon: item.Icon, IconDark: item.IconDark, PingEnabled: item.PingEnabled,
		PingOnlyDown: item.PingOnlyDown, Position: maxPos + 1,
	}
	if _, err := h.db.NewInsert().Model(ni).Exec(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, ni)
}

func (h *GroupItemHandler) PingItem(w http.ResponseWriter, r *http.Request) {
	item := new(models.Item)
	if err := h.db.NewSelect().Model(item).Where("id = ?", chi.URLParam(r, "id")).Scan(r.Context()); err != nil {
		writeError(w, http.StatusNotFound, "item not found")
		return
	}
	group := new(models.Group)
	if err := h.db.NewSelect().Model(group).Where("id = ?", item.GroupID).Scan(r.Context()); err != nil {
		writeError(w, http.StatusNotFound, "group not found")
		return
	}
	dashboard := new(models.Dashboard)
	if err := h.db.NewSelect().Model(dashboard).Where("id = ?", group.DashboardID).Scan(r.Context()); err != nil {
		writeError(w, http.StatusNotFound, "dashboard not found")
		return
	}
	if !canView(auth.UserFromContext(r.Context()), dashboard) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	if !item.PingEnabled {
		writeError(w, http.StatusBadRequest, "ping is disabled for this item")
		return
	}
	parsed, err := url.ParseRequestURI(item.URL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		writeJSON(w, http.StatusOK, map[string]bool{"reachable": false})
		return
	}

	reachable := false
	req, err := http.NewRequestWithContext(r.Context(), http.MethodHead, parsed.String(), nil)
	if err == nil {
		resp, requestErr := h.httpClient.Do(req)
		if requestErr == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusNotImplemented {
				getReq, getErr := http.NewRequestWithContext(r.Context(), http.MethodGet, parsed.String(), nil)
				if getErr == nil {
					getReq.Header.Set("Range", "bytes=0-0")
					getResp, getRequestErr := h.httpClient.Do(getReq)
					if getRequestErr == nil {
						getResp.Body.Close()
						reachable = getResp.StatusCode >= 200 && getResp.StatusCode < 400
					}
				}
			} else {
				reachable = resp.StatusCode >= 200 && resp.StatusCode < 400
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]bool{"reachable": reachable})
}
