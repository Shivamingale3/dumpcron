package drivers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"dumpcron/internal/config"
)

type MysqlDriver struct{}

func (d *MysqlDriver) Validate(job config.Job) error {
	env := append(os.Environ(), "MYSQL_PWD="+job.Password)

	cmd := exec.Command("mysqladmin",
		"ping",
		"-h", job.Host,
		"-P", strconv.Itoa(job.Port),
		"-u", job.Username,
	)
	cmd.Env = env
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("authentication failed: %s", strings.TrimSpace(stderr.String()))
	}

	listCmd := exec.Command("mysqlshow",
		"-h", job.Host,
		"-P", strconv.Itoa(job.Port),
		"-u", job.Username,
	)
	listCmd.Env = env
	out, err := listCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list databases: %v", err)
	}

	existing := parseMysqlDbList(string(out))

	for _, dbName := range job.Databases {
		if !existing[dbName] {
			return fmt.Errorf("database %q does not exist", dbName)
		}
	}

	return nil
}

func (d *MysqlDriver) Dump(ctx context.Context, job config.Job, dbName, outputPath string) error {
	cmd := exec.CommandContext(ctx, "mysqldump",
		"-h", job.Host,
		"-P", strconv.Itoa(job.Port),
		"-u", job.Username,
		dbName,
	)
	cmd.Env = append(os.Environ(), "MYSQL_PWD="+job.Password)

	zstd := exec.CommandContext(ctx, "zstd", "-o", outputPath)
	zstd.Stdin, _ = cmd.StdoutPipe()
	zstd.Stderr = os.Stderr

	if err := zstd.Start(); err != nil {
		return fmt.Errorf("zstd start: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("mysqldump start: %w", err)
	}

	cmdErr := cmd.Wait()
	zstdErr := zstd.Wait()

	if cmdErr != nil {
		return fmt.Errorf("mysqldump: %w", cmdErr)
	}
	if zstdErr != nil {
		return fmt.Errorf("zstd: %w", zstdErr)
	}

	return nil
}

func (d *MysqlDriver) RequiredBinaries() []string {
	return []string{"mysqldump"}
}

func (d *MysqlDriver) Extension() string {
	return "sql"
}

func parseMysqlDbList(output string) map[string]bool {
	dbs := make(map[string]bool)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "Databases") || strings.Contains(line, "+-") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			name := parts[0]
			if name != "" {
				dbs[name] = true
			}
		}
	}
	return dbs
}
