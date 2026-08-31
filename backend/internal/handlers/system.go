package handlers

import (
	"net/http"

	"github.com/bookmarks-dashboard/backend/internal/updatecheck"
)

type SystemHandler struct {
	updates *updatecheck.Checker
}

func NewSystemHandler(updates *updatecheck.Checker) *SystemHandler {
	return &SystemHandler{updates: updates}
}

func (h *SystemHandler) Version(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.updates.Info(r.Context()))
}
