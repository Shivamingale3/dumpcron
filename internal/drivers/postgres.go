package drivers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Shivamingale3/dumpcron/internal/config"
)

type PostgresDriver struct{}

func (d *PostgresDriver) Validate(job config.Job) error {
	env := append(os.Environ(), "PGPASSWORD="+job.Password)

	cmd := exec.Command("psql",
		"-h", job.Host,
		"-p", strconv.Itoa(job.Port),
		"-U", job.Username,
		"-c", "SELECT 1",
	)
	cmd.Env = env
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("authentication failed: %s", strings.TrimSpace(stderr.String()))
	}

	listCmd := exec.Command("psql",
		"-h", job.Host,
		"-p", strconv.Itoa(job.Port),
		"-U", job.Username,
		"-l", "-t", "-A", "-F|",
	)
	listCmd.Env = env
	out, err := listCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list databases: %v", err)
	}

	existing := parsePsqlDbList(string(out))

	for _, dbName := range job.Databases {
		if !existing[dbName] {
			return fmt.Errorf("database %q does not exist", dbName)
		}
	}

	return nil
}

func (d *PostgresDriver) Dump(ctx context.Context, job config.Job, dbName, outputPath string) error {
	cmd := exec.CommandContext(ctx, "pg_dump",
		"-h", job.Host,
		"-p", strconv.Itoa(job.Port),
		"-U", job.Username,
		dbName,
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+job.Password)

	zstd := exec.CommandContext(ctx, "zstd", "-o", outputPath)
	zstd.Stdin, _ = cmd.StdoutPipe()
	zstd.Stderr = os.Stderr

	if err := zstd.Start(); err != nil {
		return fmt.Errorf("zstd start: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("pg_dump start: %w", err)
	}

	cmdErr := cmd.Wait()
	zstdErr := zstd.Wait()

	if cmdErr != nil {
		return fmt.Errorf("pg_dump: %w", cmdErr)
	}
	if zstdErr != nil {
		return fmt.Errorf("zstd: %w", zstdErr)
	}

	return nil
}

func (d *PostgresDriver) RequiredBinaries() []string {
	return []string{"pg_dump"}
}

func (d *PostgresDriver) Extension() string {
	return "sql"
}

func parsePsqlDbList(output string) map[string]bool {
	dbs := make(map[string]bool)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) > 0 {
			name := strings.TrimSpace(parts[0])
			if name != "" {
				dbs[name] = true
			}
		}
	}
	return dbs
}
