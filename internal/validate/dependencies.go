package validate

import (
	"fmt"
	"os/exec"

	"github.com/Shivamingale3/dumpcron/internal/config"
)

func ValidateDependencies(cfg *config.Config) []error {
	var errs []error

	needed := map[string]bool{}
	hasAnyDB := false

	for i := range cfg.Jobs {
		hasAnyDB = true
		switch cfg.Jobs[i].Type {
		case "postgres":
			needed["pg_dump"] = true
		case "mysql":
			needed["mysqldump"] = true
		case "mongo":
			needed["mongodump"] = true
		}
	}

	if hasAnyDB {
		needed["zstd"] = true
	}

	for bin := range needed {
		if _, err := exec.LookPath(bin); err != nil {
			errs = append(errs, fmt.Errorf("%s missing", bin))
		}
	}

	return errs
}
