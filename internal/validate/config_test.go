package validate

import (
	"testing"

	"github.com/Shivamingale3/dumpcron/internal/config"
)

func TestValidateConfigValid(t *testing.T) {
	cfg := &config.Config{
		BackupRoot:    "/srv/backups",
		RetentionDays: 30,
		Jobs: []config.Job{
			{
				Name:      "postgres_main",
				Type:      "postgres",
				Host:      "localhost",
				Port:      5432,
				Username:  "backup",
				Password:  "secret",
				Databases: []string{"app_db"},
				Time:      "02:00",
			},
		},
	}

	errs := ValidateConfig(cfg)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateConfigEmptyBackupRoot(t *testing.T) {
	cfg := &config.Config{
		RetentionDays: 30,
		Jobs: []config.Job{
			{
				Name:      "pg",
				Type:      "postgres",
				Host:      "localhost",
				Port:      5432,
				Username:  "u",
				Databases: []string{"db"},
				Time:      "02:00",
			},
		},
	}

	errs := ValidateConfig(cfg)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if errs[0].Error() != "backup_root is required" {
		t.Errorf("got %q", errs[0].Error())
	}
}

func TestValidateConfigDuplicateNames(t *testing.T) {
	cfg := &config.Config{
		BackupRoot:    "/srv/backups",
		RetentionDays: 30,
		Jobs: []config.Job{
			{Name: "same", Type: "postgres", Host: "localhost", Port: 5432, Username: "u", Databases: []string{"db"}, Time: "02:00"},
			{Name: "same", Type: "mysql", Host: "localhost", Port: 3306, Username: "u", Databases: []string{"db"}, Time: "03:00"},
		},
	}

	errs := ValidateConfig(cfg)
	found := false
	for _, e := range errs {
		if e.Error() == `job "same": duplicate name` {
			found = true
		}
	}
	if !found {
		t.Errorf("expected duplicate name error, got: %v", errs)
	}
}

func TestValidateConfigInvalidType(t *testing.T) {
	cfg := &config.Config{
		BackupRoot:    "/srv/backups",
		RetentionDays: 30,
		Jobs: []config.Job{
			{Name: "pg", Type: "oracle", Host: "localhost", Port: 5432, Username: "u", Databases: []string{"db"}, Time: "02:00"},
		},
	}

	errs := ValidateConfig(cfg)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
}

func TestValidateConfigInvalidTimeFormat(t *testing.T) {
	tests := []string{"25:00", "12:60", "12-00", "1200", "abc", "12:00:00", "12:00pm"}
	for _, tm := range tests {
		cfg := &config.Config{
			BackupRoot:    "/srv/backups",
			RetentionDays: 30,
			Jobs: []config.Job{
				{Name: "pg", Type: "postgres", Host: "localhost", Port: 5432, Username: "u", Databases: []string{"db"}, Time: tm},
			},
		}
		errs := ValidateConfig(cfg)
		found := false
		for _, e := range errs {
			if e.Error() != "" {
				found = true
			}
		}
		if !found {
			t.Errorf("time %q should have failed validation", tm)
		}
	}
}

func TestValidateConfigValidTimes(t *testing.T) {
	tests := []string{"00:00", "01:30", "12:00", "23:59", "02:00", "2:00", "9:05"}
	for _, tm := range tests {
		cfg := &config.Config{
			BackupRoot:    "/srv/backups",
			RetentionDays: 30,
			Jobs: []config.Job{
				{Name: "pg", Type: "postgres", Host: "localhost", Port: 5432, Username: "u", Databases: []string{"db"}, Time: tm},
			},
		}
		errs := ValidateConfig(cfg)
		if len(errs) != 0 {
			t.Errorf("time %q should be valid, got: %v", tm, errs)
		}
	}
}
