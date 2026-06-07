package retention

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func Cleanup(backupRoot, dbType string, retentionDays int) error {
	dir := filepath.Join(backupRoot, dbType)

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read %s: %w", dir, err)
	}

	cutoff := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			if err := os.Remove(filePath); err != nil {
				fmt.Printf("retention: failed to delete %s: %v\n", filePath, err)
				continue
			}
			fmt.Printf("retention: deleted expired backup %s\n", filePath)
		}
	}

	return nil
}
