package updatecheck

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultReleaseAPI = "https://api.github.com/repos/codewec/dashlit/releases/latest"
	releasesURL       = "https://github.com/codewec/dashlit/releases"
)

var versionPattern = regexp.MustCompile(`^v([0-9]+)\.([0-9]+)\.([0-9]+)$`)

type Info struct {
	Version         string `json:"version"`
	Commit          string `json:"commit"`
	LatestVersion   string `json:"latestVersion,omitempty"`
	UpdateAvailable bool   `json:"updateAvailable,omitempty"`
	ReleaseURL      string `json:"releaseUrl,omitempty"`
}

type Checker struct {
	version        string
	commit         string
	enabled        bool
	latestOverride string
	apiURL         string
	client         *http.Client
	cacheDuration  time.Duration

	mu        sync.Mutex
	checkedAt time.Time
	latest    string
	release   string
}

func New(version, commit string, enabled bool, latestOverride string) *Checker {
	return &Checker{
		version: strings.TrimSpace(version), commit: strings.TrimSpace(commit), enabled: enabled,
		latestOverride: strings.TrimSpace(latestOverride), apiURL: defaultReleaseAPI,
		client: &http.Client{Timeout: 5 * time.Second}, cacheDuration: 12 * time.Hour,
	}
}

func (c *Checker) Current() Info {
	return Info{Version: fallback(c.version, "dev"), Commit: fallback(c.commit, "unknown")}
}

func (c *Checker) Info(ctx context.Context) Info {
	info := c.Current()
	if !c.enabled || !validVersion(info.Version) {
		return info
	}

	latest, release := c.latestRelease(ctx)
	info.LatestVersion = latest
	info.ReleaseURL = release
	info.UpdateAvailable = newer(latest, info.Version)
	return info
}

func (c *Checker) latestRelease(ctx context.Context) (string, string) {
	if validVersion(c.latestOverride) {
		return c.latestOverride, releasesURL + "/tag/" + c.latestOverride
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.checkedAt.IsZero() && time.Since(c.checkedAt) < c.cacheDuration {
		return c.latest, c.release
	}
	c.checkedAt = time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL, nil)
	if err != nil {
		return c.latest, c.release
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "DashLit/"+c.version)
	resp, err := c.client.Do(req)
	if err != nil {
		return c.latest, c.release
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return c.latest, c.release
	}
	var payload struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload) != nil || !validVersion(payload.TagName) {
		return c.latest, c.release
	}
	c.latest = payload.TagName
	c.release = payload.HTMLURL
	if c.release == "" {
		c.release = releasesURL + "/tag/" + c.latest
	}
	return c.latest, c.release
}

func validVersion(value string) bool { return versionPattern.MatchString(value) }

func newer(candidate, current string) bool {
	a := versionParts(candidate)
	b := versionParts(current)
	if a == nil || b == nil {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}

func versionParts(value string) []uint64 {
	match := versionPattern.FindStringSubmatch(value)
	if match == nil {
		return nil
	}
	parts := make([]uint64, 3)
	for i := range parts {
		part, err := strconv.ParseUint(match[i+1], 10, 64)
		if err != nil {
			return nil
		}
		parts[i] = part
	}
	return parts
}

func fallback(value, fallbackValue string) string {
	if value == "" {
		return fallbackValue
	}
	return value
}
