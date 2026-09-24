package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/bookmarks-dashboard/backend/internal/auth"
	"github.com/bookmarks-dashboard/backend/internal/config"
	"github.com/bookmarks-dashboard/backend/internal/models"
)

type IconHandler struct {
	db              *bun.DB
	cfg             *config.Config
	client          *http.Client
	selfhstMu       sync.RWMutex
	selfhstIcons    []selfhstIcon
	selfhstLoadedAt time.Time
}

func NewIconHandler(db *bun.DB, cfg *config.Config) *IconHandler {
	return &IconHandler{db: db, cfg: cfg, client: &http.Client{Timeout: 12 * time.Second}}
}

const (
	selfhstIndexURL = "https://cdn.jsdelivr.net/gh/selfhst/icons@main/index.json"
	selfhstCDNURL   = "https://cdn.jsdelivr.net/gh/selfhst/icons@main"
)

var iconNamePattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

type selfhstIcon struct {
	Name      string `json:"Name"`
	Reference string `json:"Reference"`
	SVG       string `json:"SVG"`
	PNG       string `json:"PNG"`
	Light     string `json:"Light"`
	Dark      string `json:"Dark"`
	Tags      string `json:"Tags"`
}

type iconSearchResult struct {
	Name     string `json:"name"`
	Icon     string `json:"icon"`
	IconDark string `json:"iconDark,omitempty"`
	Source   string `json:"source"`
	Variant  string `json:"variant,omitempty"`
}

func yes(value string) bool { return strings.EqualFold(value, "yes") }

func selfhstResults(icon selfhstIcon) []iconSearchResult {
	ext := ".png"
	if yes(icon.SVG) {
		ext = ".svg"
	}
	results := []iconSearchResult{{
		Name: icon.Name, Icon: "selfhst-icon:" + icon.Reference + ext,
		Source: "selfh.st", Variant: "color",
	}}
	if !yes(icon.Light) && !yes(icon.Dark) {
		return results
	}

	lightReference := icon.Reference
	darkReference := icon.Reference
	// selfh.st's dark-colored asset is intended for a light background and
	// its light-colored asset for a dark background.
	if yes(icon.Dark) {
		lightReference += "-dark"
	}
	if yes(icon.Light) {
		darkReference += "-light"
	}
	results = append(results, iconSearchResult{
		Name: icon.Name, Icon: "selfhst-icon:" + lightReference + ext,
		IconDark: "selfhst-icon:" + darkReference + ext, Source: "selfh.st", Variant: "monochrome",
	})
	return results
}

func (h *IconHandler) loadSelfhstIcons() ([]selfhstIcon, error) {
	h.selfhstMu.RLock()
	if len(h.selfhstIcons) > 0 && time.Since(h.selfhstLoadedAt) < 24*time.Hour {
		icons := h.selfhstIcons
		h.selfhstMu.RUnlock()
		return icons, nil
	}
	h.selfhstMu.RUnlock()

	h.selfhstMu.Lock()
	defer h.selfhstMu.Unlock()
	if len(h.selfhstIcons) > 0 && time.Since(h.selfhstLoadedAt) < 24*time.Hour {
		return h.selfhstIcons, nil
	}
	resp, err := h.client.Get(selfhstIndexURL)
	if err != nil {
		return h.selfhstIcons, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return h.selfhstIcons, fmt.Errorf("selfh.st index returned %s", resp.Status)
	}
	var icons []selfhstIcon
	if err := json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(&icons); err != nil {
		return h.selfhstIcons, err
	}
	h.selfhstIcons = icons
	h.selfhstLoadedAt = time.Now()
	return icons, nil
}

func searchSelfhst(icons []selfhstIcon, query string, limit int) []iconSearchResult {
	query = strings.ToLower(strings.TrimSpace(query))
	results := make([]iconSearchResult, 0, limit)
	for _, icon := range icons {
		haystack := strings.ToLower(icon.Name + " " + icon.Reference + " " + icon.Tags)
		if !strings.Contains(haystack, query) {
			continue
		}
		for _, result := range selfhstResults(icon) {
			results = append(results, result)
			if len(results) >= limit {
				return results
			}
		}
	}
	return results
}

func (h *IconHandler) SearchSelfhst(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeJSON(w, http.StatusOK, []iconSearchResult{})
		return
	}

	results := make([]iconSearchResult, 0, 32)
	if icons, err := h.loadSelfhstIcons(); err == nil || len(icons) > 0 {
		results = append(results, searchSelfhst(icons, query, 32)...)
	}
	writeJSON(w, http.StatusOK, results)
}

func (h *IconHandler) SearchIconify(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeJSON(w, http.StatusOK, []iconSearchResult{})
		return
	}
	results := make([]iconSearchResult, 0, 32)
	iconifyURL := "https://api.iconify.design/search?query=" + url.QueryEscape(query) + "&limit=32"
	if resp, err := h.client.Get(iconifyURL); err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var payload struct {
				Icons []string `json:"icons"`
			}
			if json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload) == nil {
				for _, icon := range payload.Icons {
					// Iconify republishes a selfhst collection under the same prefix
					// used by selfh.st. Direct results above have theme pairs and are
					// preferred, so omit these confusing duplicates.
					if strings.HasPrefix(icon, "selfhst:") {
						continue
					}
					name := icon
					if _, after, found := strings.Cut(icon, ":"); found {
						name = after
					}
					results = append(results, iconSearchResult{Name: name, Icon: icon, Source: "Iconify"})
				}
			}
		}
	}
	writeJSON(w, http.StatusOK, results)
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
	_, err = io.Copy(dst, file)
	if err != nil {
		_ = os.Remove(dstPath)
		writeError(w, http.StatusInternalServerError, "save failed")
		return
	}
	rec := &models.UploadedIcon{
		ID:       id,
		Filename: filename,
		Mime:     mime,
		OwnerID:  user.ID,
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

type iconifyIconData struct {
	Body   string `json:"body"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type iconifyAlias struct {
	Parent string `json:"parent"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	HFlip  bool   `json:"hFlip"`
	VFlip  bool   `json:"vFlip"`
}

type iconifySet struct {
	Prefix  string                     `json:"prefix"`
	Icons   map[string]iconifyIconData `json:"icons"`
	Aliases map[string]iconifyAlias    `json:"aliases"`
	Width   int                        `json:"width"`
	Height int                         `json:"height"`
}

func firstPositive(values ...int) int {
	for _, v := range values {
		if v > 0 {
			return v
		}
	}
	return 0
}

func (s *iconifySet) renderSVG(name string) (string, bool) {
	hFlip, vFlip := false, false
	widthOverride, heightOverride := 0, 0
	for hops := 0; ; hops++ {
		if icon, ok := s.Icons[name]; ok {
			width := firstPositive(widthOverride, icon.Width, s.Width, 16)
			height := firstPositive(heightOverride, icon.Height, s.Height, 16)
			body := icon.Body
			if hFlip || vFlip {
				var b strings.Builder
				b.WriteString(`<g transform="`)
				if hFlip {
					fmt.Fprintf(&b, "translate(%d 0) scale(-1 1) ", width)
				}
				if vFlip {
					fmt.Fprintf(&b, "translate(0 %d) scale(1 -1) ", height)
				}
				b.WriteString(`">`)
				b.WriteString(body)
				b.WriteString("</g>")
				body = b.String()
			}
			return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">%s</svg>`, width, height, width, height, body), true
		}
		if hops >= 8 {
			return "", false
		}
		alias, ok := s.Aliases[name]
		if !ok {
			return "", false
		}
		hFlip = hFlip != alias.HFlip
		vFlip = vFlip != alias.VFlip
		if alias.Width > 0 && widthOverride == 0 {
			widthOverride = alias.Width
		}
		if alias.Height > 0 && heightOverride == 0 {
			heightOverride = alias.Height
		}
		name = alias.Parent
	}
}

// ProxyIconify renders Iconify icons locally from the batch JSON API.
// The per-icon SVG endpoint is heavily rate-limited (HTTP 429), while the
// JSON endpoint is not, so icons are fetched as JSON and wrapped in an SVG.
func (h *IconHandler) ProxyIconify(w http.ResponseWriter, r *http.Request) {
	prefix := chi.URLParam(r, "prefix")
	name := chi.URLParam(r, "name")
	if !iconNamePattern.MatchString(prefix) || !iconNamePattern.MatchString(name) {
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
	apiURL := fmt.Sprintf("https://api.iconify.design/%s.json?icons=%s", prefix, name)
	resp, err := h.client.Get(apiURL)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		http.NotFound(w, r)
		return
	}
	var set iconifySet
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&set); err != nil {
		http.NotFound(w, r)
		return
	}
	svg, ok := set.renderSVG(name)
	if !ok {
		http.NotFound(w, r)
		return
	}
	data := []byte(svg)
	_ = os.MkdirAll(filepath.Dir(cachePath), 0755)
	_ = os.WriteFile(cachePath, data, 0644)
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(data)
}

func (h *IconHandler) ProxySelfhst(w http.ResponseWriter, r *http.Request) {
	filename := strings.Trim(strings.TrimSpace(chi.URLParam(r, "*")), "/")
	ext := strings.ToLower(filepath.Ext(filename))
	name := strings.TrimSuffix(filename, ext)
	if filepath.Base(filename) != filename || !iconNamePattern.MatchString(name) || (ext != ".svg" && ext != ".png" && ext != ".webp") {
		writeError(w, http.StatusBadRequest, "invalid selfh.st icon name")
		return
	}
	cachePath := filepath.Join(h.cfg.IconCacheDir, "selfhst", filename)
	contentType := map[string]string{".svg": "image/svg+xml", ".png": "image/png", ".webp": "image/webp"}[ext]
	if data, err := os.ReadFile(cachePath); err == nil {
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "public, max-age=86400")
		_, _ = w.Write(data)
		return
	}
	remoteURL := selfhstCDNURL + "/" + strings.TrimPrefix(ext, ".") + "/" + url.PathEscape(filename)
	resp, err := h.client.Get(remoteURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		http.NotFound(w, r)
		return
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_ = os.MkdirAll(filepath.Dir(cachePath), 0755)
	_ = os.WriteFile(cachePath, data, 0644)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(data)
}
