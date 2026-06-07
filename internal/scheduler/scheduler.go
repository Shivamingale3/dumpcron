package scheduler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"dumpcron/internal/config"
)

type WorkItem struct {
	Job    config.Job
	Config *config.Config
}

type Scheduler struct {
	cfg       *config.Config
	queue     chan<- WorkItem
	statePath string
	lastRun   map[string]time.Time
	mu        sync.Mutex
	running   bool
	stopCh    chan struct{}
}

func New(cfg *config.Config, queue chan<- WorkItem, statePath string) *Scheduler {
	return &Scheduler{
		cfg:       cfg,
		queue:     queue,
		statePath: statePath,
		lastRun:   make(map[string]time.Time),
		stopCh:    make(chan struct{}),
	}
}

func (s *Scheduler) LoadState() error {
	data, err := os.ReadFile(s.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return json.Unmarshal(data, &s.lastRun)
}

func (s *Scheduler) saveState() error {
	dir := filepath.Dir(s.statePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	s.mu.Lock()
	data, err := json.Marshal(s.lastRun)
	s.mu.Unlock()

	if err != nil {
		return err
	}

	return os.WriteFile(s.statePath, data, 0644)
}

func (s *Scheduler) isDue(job config.Job) bool {
	s.mu.Lock()
	lastRun := s.lastRun[job.Name]
	s.mu.Unlock()

	now := time.Now()

	parts := [2]int{}
	if _, err := fmt.Sscanf(job.Time, "%d:%d", &parts[0], &parts[1]); err != nil {
		return false
	}

	scheduled := time.Date(now.Year(), now.Month(), now.Day(),
		parts[0], parts[1], 0, 0, now.Location())

	if now.Before(scheduled) {
		return false
	}

	if lastRun.IsZero() {
		return true
	}

	return lastRun.Before(scheduled)
}

func (s *Scheduler) markRun(jobName string) {
	s.mu.Lock()
	s.lastRun[jobName] = time.Now()
	s.mu.Unlock()
}

func (s *Scheduler) Start() {
	s.running = true
	ticker := time.NewTicker(1 * time.Minute)

	go func() {
		s.tick()

		for {
			select {
			case <-ticker.C:
				s.tick()
			case <-s.stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	s.stopCh <- struct{}{}
}

func (s *Scheduler) tick() {
	if !s.running {
		return
	}

	for i := range s.cfg.Jobs {
		job := &s.cfg.Jobs[i]
		if s.isDue(*job) {
			s.markRun(job.Name)
			s.saveState()

			s.queue <- WorkItem{
				Job:    *job,
				Config: s.cfg,
			}
		}
	}
}
