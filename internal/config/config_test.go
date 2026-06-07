package config

import (
	"os"
	"testing"
)

func TestLoadValid(t *testing.T) {
	yaml := `backup_root: /srv/backups
retention_days: 30

jobs:
  - name: postgres_main
    type: postgres
    host: localhost
    port: 5432
    username: backup_user
    password: secret
    databases:
      - app_db
      - auth_db
    time: "02:00"

  - name: mysql_main
    type: mysql
    host: localhost
    port: 3306
    username: backup_user
    password: secret
    databases:
      - users_db
    time: "03:00"

  - name: mongo_main
    type: mongo
    host: localhost
    port: 27017
    username: backup_user
    password: secret
    databases:
      - analytics
    time: "04:00"
`

	tmp := writeTemp(t, yaml)
	defer os.Remove(tmp)

	cfg, err := Load(tmp)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.BackupRoot != "/srv/backups" {
		t.Errorf("backup_root = %q, want /srv/backups", cfg.BackupRoot)
	}
	if cfg.RetentionDays != 30 {
		t.Errorf("retention_days = %d, want 30", cfg.RetentionDays)
	}
	if len(cfg.Jobs) != 3 {
		t.Fatalf("jobs count = %d, want 3", len(cfg.Jobs))
	}
	if cfg.Jobs[0].Name != "postgres_main" {
		t.Errorf("job[0].Name = %q", cfg.Jobs[0].Name)
	}
	if len(cfg.Jobs[0].Databases) != 2 {
		t.Errorf("job[0] databases = %d, want 2", len(cfg.Jobs[0].Databases))
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "dumpcron-test-*.yaml")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		t.Fatalf("write temp: %v", err)
	}
	f.Close()
	return f.Name()
}
