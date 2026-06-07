package retention

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCleanup(t *testing.T) {
	dir := t.TempDir()
	dbDir := filepath.Join(dir, "postgres")
	os.MkdirAll(dbDir, 0755)

	oldFile := filepath.Join(dbDir, "old.zst")
	newFile := filepath.Join(dbDir, "new.zst")

	os.WriteFile(oldFile, []byte("old"), 0644)
	oldTime := time.Now().Add(-31 * 24 * time.Hour)
	os.Chtimes(oldFile, oldTime, oldTime)

	os.WriteFile(newFile, []byte("new"), 0644)

	if err := Cleanup(dir, "postgres", 30); err != nil {
		t.Fatalf("cleanup error: %v", err)
	}

	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("old file should have been deleted")
	}
	if _, err := os.Stat(newFile); err != nil {
		t.Error("new file should still exist")
	}
}

func TestCleanupMissingDir(t *testing.T) {
	err := Cleanup("/nonexistent", "postgres", 30)
	if err != nil {
		t.Fatalf("expected no error for missing dir, got: %v", err)
	}
}

func TestCleanupEmptyDir(t *testing.T) {
	dir := t.TempDir()
	dbDir := filepath.Join(dir, "mysql")
	os.MkdirAll(dbDir, 0755)

	if err := Cleanup(dir, "mysql", 30); err != nil {
		t.Fatalf("cleanup error: %v", err)
	}
}

func TestCleanupSkipsDirectories(t *testing.T) {
	dir := t.TempDir()
	dbDir := filepath.Join(dir, "mongo")
	subDir := filepath.Join(dbDir, "subdir")
	os.MkdirAll(subDir, 0755)

	if err := Cleanup(dir, "mongo", 30); err != nil {
		t.Fatalf("cleanup error: %v", err)
	}

	if _, err := os.Stat(subDir); err != nil {
		t.Error("subdirectory should not be deleted")
	}
}
