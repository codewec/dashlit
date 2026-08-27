package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/bookmarks-dashboard/backend/internal/auth"
	"github.com/bookmarks-dashboard/backend/internal/config"
	"github.com/bookmarks-dashboard/backend/internal/models"
)

type IconHandler struct {
	db  *bun.DB
	cfg *config.Config
}

func NewIconHandler(db *bun.DB, cfg *config.Config) *IconHandler {
	return &IconHandler{db: db, cfg: cfg}
}

func (h *IconHandler) Upload(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "file too large or invalid")
		return
	}
	file, header, err := r.FormFile("icon")
	if err != nil {
		writeError(w, http.StatusBadRequest, "icon file required")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]string{".png": "image/png", ".svg": "image/svg+xml", ".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".webp": "image/webp", ".ico": "image/x-icon"}
	mime, ok := allowed[ext]
	if !ok {
		writeError(w, http.StatusBadRequest, "unsupported file type")
		return
	}
	id := uuid.NewString()
	filename := id + ext
	dstPath := filepath.Join(h.cfg.IconsDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "save failed")
		return
	}
	defer dst.Close()
	size, err := io.Copy(dst, file)
	if err != nil {
		_ = os.Remove(dstPath)
		writeError(w, http.StatusInternalServerError, "save failed")
		return
	}
	rec := &models.UploadedIcon{
		ID:           id,
		Filename:     filename,
		OriginalName: header.Filename,
		Mime:         mime,
		Size:         size,
		OwnerID:      user.ID,
	}
	if _, err := h.db.NewInsert().Model(rec).Exec(r.Context()); err != nil {
		_ = os.Remove(dstPath)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":   id,
		"icon": "local:" + id,
		"url":  "/api/icons/" + id,
	})
}

func (h *IconHandler) Serve(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// strip extension if present
	id = strings.TrimSuffix(id, filepath.Ext(id))
	rec := new(models.UploadedIcon)
	if err := h.db.NewSelect().Model(rec).Where("id = ?", id).Scan(r.Context()); err != nil {
		http.NotFound(w, r)
		return
	}
	path := filepath.Join(h.cfg.IconsDir, rec.Filename)
	w.Header().Set("Content-Type", rec.Mime)
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	http.ServeFile(w, r, path)
}

// Proxy Iconify SVG and cache to disk
func (h *IconHandler) ProxyIconify(w http.ResponseWriter, r *http.Request) {
	prefix := chi.URLParam(r, "prefix")
	name := chi.URLParam(r, "name")
	if prefix == "" || name == "" {
		writeError(w, http.StatusBadRequest, "prefix and name required")
		return
	}
	cachePath := filepath.Join(h.cfg.IconCacheDir, prefix, name+".svg")
	if data, err := os.ReadFile(cachePath); err == nil {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		_, _ = w.Write(data)
		return
	}
	url := fmt.Sprintf("https://api.iconify.design/%s/%s.svg", prefix, name)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		http.NotFound(w, r)
		return
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_ = os.MkdirAll(filepath.Dir(cachePath), 0755)
	_ = os.WriteFile(cachePath, data, 0644)
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(data)
}
