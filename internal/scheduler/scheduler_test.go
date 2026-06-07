package scheduler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"dumpcron/internal/config"
)

func TestIsDueNoLastRunPast(t *testing.T) {
	yesterday := time.Now().Add(-1 * time.Hour)
	// schedule at one hour ago
	hh := yesterday.Format("15")
	mm := yesterday.Format("04")

	job := config.Job{
		Name: "test",
		Time: hh + ":" + mm,
	}

	s := &Scheduler{
		lastRun: make(map[string]time.Time),
	}

	if !s.isDue(job) {
		t.Error("job should be due (no last run, scheduled time has passed)")
	}
}

func TestIsDueNoLastRunFuture(t *testing.T) {
	tomorrow := time.Now().Add(1 * time.Hour)
	hh := tomorrow.Format("15")
	mm := tomorrow.Format("04")

	job := config.Job{
		Name: "test",
		Time: hh + ":" + mm,
	}

	s := &Scheduler{
		lastRun: make(map[string]time.Time),
	}

	if s.isDue(job) {
		t.Error("job should not be due (scheduled in future)")
	}
}

func TestIsDueAlreadyRunToday(t *testing.T) {
	now := time.Now()
	// schedule at start of this hour
	hh := now.Format("15")
	mm := now.Format("04")

	job := config.Job{
		Name: "test",
		Time: hh + ":" + mm,
	}

	s := &Scheduler{
		lastRun: map[string]time.Time{
			"test": now,
		},
	}

	if s.isDue(job) {
		t.Error("job should not be due (already ran after scheduled time)")
	}
}

func TestIsDueRanYesterday(t *testing.T) {
	now := time.Now()
	// schedule 5 minutes ago
	sched := now.Add(-5 * time.Minute)
	hh := sched.Format("15")
	mm := sched.Format("04")

	job := config.Job{
		Name: "test",
		Time: hh + ":" + mm,
	}

	s := &Scheduler{
		lastRun: map[string]time.Time{
			"test": now.Add(-25 * time.Hour),
		},
	}

	if !s.isDue(job) {
		t.Error("job should be due (last ran yesterday, schedule passed today)")
	}
}

func TestMarkRunSavesState(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "last_run.json")

	cfg := &config.Config{Jobs: []config.Job{}}
	queue := make(chan WorkItem, 10)

	s := New(cfg, queue, statePath)
	s.markRun("test_job")

	if err := s.saveState(); err != nil {
		t.Fatalf("saveState: %v", err)
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read state: %v", err)
	}

	var state map[string]time.Time
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}

	if _, ok := state["test_job"]; !ok {
		t.Error("test_job not found in saved state")
	}
}

func TestLoadStateEmpty(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "nonexistent.json")

	cfg := &config.Config{Jobs: []config.Job{}}
	queue := make(chan WorkItem, 10)

	s := New(cfg, queue, statePath)
	if err := s.LoadState(); err != nil {
		t.Fatalf("LoadState on missing file: %v", err)
	}
}

func TestLoadStateWithData(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "last_run.json")

	state := map[string]time.Time{
		"pg_main": time.Date(2026, 6, 7, 2, 0, 0, 0, time.UTC),
	}
	data, _ := json.Marshal(state)
	os.WriteFile(statePath, data, 0644)

	cfg := &config.Config{Jobs: []config.Job{}}
	queue := make(chan WorkItem, 10)

	s := New(cfg, queue, statePath)
	if err := s.LoadState(); err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	s.mu.Lock()
	loaded := s.lastRun["pg_main"]
	s.mu.Unlock()

	if loaded.IsZero() {
		t.Error("pg_main not loaded from state")
	}
}
