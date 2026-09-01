package handlers

import (
	"net/http"

	"github.com/bookmarks-dashboard/backend/internal/auth"
	"github.com/bookmarks-dashboard/backend/internal/models"
	"github.com/bookmarks-dashboard/backend/internal/updatecheck"
)

type SystemHandler struct {
	updates *updatecheck.Checker
}

func NewSystemHandler(updates *updatecheck.Checker) *SystemHandler {
	return &SystemHandler{updates: updates}
}

func (h *SystemHandler) Version(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil || user.Role != models.RoleAdmin {
		writeJSON(w, http.StatusOK, h.updates.Current())
		return
	}
	writeJSON(w, http.StatusOK, h.updates.Info(r.Context()))
}
