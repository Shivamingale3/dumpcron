package validate

import (
	"github.com/Shivamingale3/dumpcron/internal/config"
	"github.com/Shivamingale3/dumpcron/internal/drivers"
)

func ValidateAll(cfg *config.Config, drvs map[string]drivers.Driver) []error {
	var errs []error

	if e := ValidateConfig(cfg); len(e) > 0 {
		return e
	}

	if e := ValidateDependencies(cfg); len(e) > 0 {
		return e
	}

	if e := ValidateStorage(cfg.BackupRoot); len(e) > 0 {
		return e
	}

	if e := ValidateDatabases(cfg, drvs); len(e) > 0 {
		return e
	}

	return errs
}
