package drivers

import (
	"context"

	"dumpcron/internal/config"
)

type Driver interface {
	Validate(job config.Job) error
	Dump(ctx context.Context, job config.Job, dbName, outputPath string) error
	RequiredBinaries() []string
	Extension() string
}
