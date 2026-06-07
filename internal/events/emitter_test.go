package events

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestEmitStarted(t *testing.T) {
	out := captureOutput(EmitStarted)
	if !strings.Contains(out, "dumpcron DUMPCRON_STARTED: service started") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestEmitStopped(t *testing.T) {
	out := captureOutput(EmitStopped)
	if !strings.Contains(out, "dumpcron DUMPCRON_STOPPED: service stopped") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestEmitConfigInvalid(t *testing.T) {
	out := captureOutput(func() { EmitConfigInvalid("bad field") })
	if !strings.Contains(out, "dumpcron DUMPCRON_CONFIG_INVALID: bad field") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestEmitDependencyMissing(t *testing.T) {
	out := captureOutput(func() { EmitDependencyMissing("pg_dump") })
	if !strings.Contains(out, "dumpcron DUMPCRON_DEPENDENCY_MISSING: pg_dump not found") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestEmitStorageInvalid(t *testing.T) {
	out := captureOutput(func() { EmitStorageInvalid("not writable") })
	if !strings.Contains(out, "dumpcron DUMPCRON_STORAGE_INVALID: not writable") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestEmitJobStarted(t *testing.T) {
	out := captureOutput(func() { EmitJobStarted("postgres_main") })
	if !strings.Contains(out, "dumpcron BACKUP_JOB_STARTED: postgres_main") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestEmitJobCompleted(t *testing.T) {
	out := captureOutput(func() { EmitJobCompleted("postgres_main") })
	if !strings.Contains(out, "dumpcron BACKUP_JOB_COMPLETED: postgres_main") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestEmitJobFailed(t *testing.T) {
	out := captureOutput(func() { EmitJobFailed("mysql_main") })
	if !strings.Contains(out, "dumpcron BACKUP_JOB_FAILED: mysql_main") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestEmitDatabaseBackupFailed(t *testing.T) {
	out := captureOutput(func() { EmitDatabaseBackupFailed("pg_main", "app_db", "connection refused") })
	if !strings.Contains(out, "dumpcron DATABASE_BACKUP_FAILED: pg_main/app_db: connection refused") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestAllEventsUnique(t *testing.T) {
	events := []string{
		"dumpcron DUMPCRON_STARTED:",
		"dumpcron DUMPCRON_STOPPED:",
		"dumpcron DUMPCRON_CONFIG_INVALID:",
		"dumpcron DUMPCRON_DEPENDENCY_MISSING:",
		"dumpcron DUMPCRON_STORAGE_INVALID:",
		"dumpcron BACKUP_JOB_STARTED:",
		"dumpcron BACKUP_JOB_COMPLETED:",
		"dumpcron BACKUP_JOB_FAILED:",
		"dumpcron DATABASE_BACKUP_FAILED:",
	}

	seen := make(map[string]bool)
	for _, e := range events {
		if seen[e] {
			t.Errorf("duplicate event prefix: %q", e)
		}
		seen[e] = true
	}
}

func captureOutput(fn func()) string {
	r, w, _ := os.Pipe()
	stdout := os.Stdout
	os.Stdout = w

	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	w.Close()
	os.Stdout = stdout

	return <-done
}
