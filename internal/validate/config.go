package validate

import (
	"fmt"
	"regexp"

	"dumpcron/internal/config"
)

var timeRegex = regexp.MustCompile(`^([01]?\d|2[0-3]):([0-5]\d)$`)

var validTypes = map[string]bool{
	"postgres": true,
	"mysql":    true,
	"mongo":    true,
}

func ValidateConfig(cfg *config.Config) []error {
	var errs []error

	if cfg.BackupRoot == "" {
		errs = append(errs, fmt.Errorf("backup_root is required"))
	}

	if cfg.RetentionDays <= 0 {
		errs = append(errs, fmt.Errorf("retention_days must be positive"))
	}

	if len(cfg.Jobs) == 0 {
		errs = append(errs, fmt.Errorf("at least one job is required"))
		return errs
	}

	names := make(map[string]bool)

	for i := range cfg.Jobs {
		job := &cfg.Jobs[i]

		if job.Name == "" {
			errs = append(errs, fmt.Errorf("job[%d]: name is required", i))
		} else if names[job.Name] {
			errs = append(errs, fmt.Errorf("job %q: duplicate name", job.Name))
		} else {
			names[job.Name] = true
		}

		if !validTypes[job.Type] {
			errs = append(errs, fmt.Errorf("job %q: invalid type %q, must be postgres, mysql, or mongo", job.Name, job.Type))
		}

		if job.Host == "" {
			errs = append(errs, fmt.Errorf("job %q: host is required", job.Name))
		}

		if job.Port <= 0 {
			errs = append(errs, fmt.Errorf("job %q: port must be positive", job.Name))
		}

		if job.Username == "" {
			errs = append(errs, fmt.Errorf("job %q: username is required", job.Name))
		}

		if len(job.Databases) == 0 {
			errs = append(errs, fmt.Errorf("job %q: at least one database is required", job.Name))
		}

		if !timeRegex.MatchString(job.Time) {
			errs = append(errs, fmt.Errorf("job %q: invalid time %q, must be HH:MM 24-hour format", job.Name, job.Time))
		}
	}

	return errs
}
