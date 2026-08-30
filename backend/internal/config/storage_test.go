package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateStorageCreatesWritableDirectories(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	cfg := &Config{
		DataDir:      dataDir,
		DatabasePath: filepath.Join(dataDir, "bookmarks.db"),
		IconsDir:     filepath.Join(dataDir, "icons"),
		IconCacheDir: filepath.Join(dataDir, "icon-cache"),
	}

	if err := cfg.ValidateStorage(); err != nil {
		t.Fatalf("ValidateStorage() error = %v", err)
	}
	for _, directory := range []string{cfg.DataDir, cfg.IconsDir, cfg.IconCacheDir} {
		if info, err := os.Stat(directory); err != nil || !info.IsDir() {
			t.Fatalf("directory %q was not created", directory)
		}
	}
}

func TestValidateStorageRejectsNonDirectoryDataPath(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	if err := os.WriteFile(dataDir, []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := &Config{
		DataDir:      dataDir,
		DatabasePath: filepath.Join(dataDir, "bookmarks.db"),
		IconsDir:     filepath.Join(dataDir, "icons"),
		IconCacheDir: filepath.Join(dataDir, "icon-cache"),
	}

	if err := cfg.ValidateStorage(); err == nil {
		t.Fatal("ValidateStorage() error = nil, want an invalid storage error")
	}
}
