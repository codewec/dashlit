package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ValidateStorage verifies all directories that DashLit writes to before the
// database and HTTP server are started. This produces an actionable error for
// bind mounts instead of letting the process enter a crash/restart loop.
func (c *Config) ValidateStorage() error {
	directories := []string{
		c.DataDir,
		filepath.Dir(c.DatabasePath),
		c.IconsDir,
		c.IconCacheDir,
	}
	checked := make(map[string]struct{}, len(directories))

	for _, directory := range directories {
		clean := filepath.Clean(directory)
		if _, ok := checked[clean]; ok {
			continue
		}
		checked[clean] = struct{}{}

		if err := os.MkdirAll(clean, 0755); err != nil {
			return fmt.Errorf("create %q: %w", clean, err)
		}

		probe, err := os.CreateTemp(clean, ".dashlit-write-test-*")
		if err != nil {
			return fmt.Errorf("write to %q: %w", clean, err)
		}
		probePath := probe.Name()
		if err := probe.Close(); err != nil {
			_ = os.Remove(probePath)
			return fmt.Errorf("close write test in %q: %w", clean, err)
		}
		if err := os.Remove(probePath); err != nil {
			return fmt.Errorf("remove write test from %q: %w", clean, err)
		}
	}

	if info, err := os.Stat(c.DatabasePath); err == nil && !info.IsDir() {
		database, err := os.OpenFile(c.DatabasePath, os.O_WRONLY|os.O_APPEND, 0)
		if err != nil {
			return fmt.Errorf("write to database %q: %w", c.DatabasePath, err)
		}
		if err := database.Close(); err != nil {
			return fmt.Errorf("close database write test %q: %w", c.DatabasePath, err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("inspect database %q: %w", c.DatabasePath, err)
	}

	return nil
}
