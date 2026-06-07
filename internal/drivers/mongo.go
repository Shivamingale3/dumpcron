package drivers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"dumpcron/internal/config"
)

type MongoDriver struct{}

func (d *MongoDriver) buildURI(job config.Job) string {
	if job.Username != "" && job.Password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%d",
			job.Username, job.Password, job.Host, job.Port)
	}
	return fmt.Sprintf("mongodb://%s:%d", job.Host, job.Port)
}

func (d *MongoDriver) Validate(job config.Job) error {
	uri := d.buildURI(job)

	pingCmd := exec.Command("mongosh",
		uri,
		"--quiet",
		"--eval", "db.adminCommand({ping:1})",
	)
	var stderr bytes.Buffer
	pingCmd.Stderr = &stderr

	out, err := pingCmd.Output()
	if err != nil {
		return fmt.Errorf("authentication failed: %s", strings.TrimSpace(stderr.String()))
	}

	if !strings.Contains(string(out), "ok") {
		return fmt.Errorf("authentication failed: ping returned unexpected response")
	}

	listCmd := exec.Command("mongosh",
		uri,
		"--quiet",
		"--eval", "JSON.stringify(db.adminCommand({listDatabases:1}).databases.map(d=>d.name))",
	)
	listOut, err := listCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list databases: %v", err)
	}

	var dbNames []string
	if err := json.Unmarshal(listOut, &dbNames); err != nil {
		return fmt.Errorf("failed to parse database list: %v", err)
	}

	existing := make(map[string]bool)
	for _, name := range dbNames {
		existing[name] = true
	}

	for _, dbName := range job.Databases {
		if !existing[dbName] {
			return fmt.Errorf("database %q does not exist", dbName)
		}
	}

	return nil
}

func (d *MongoDriver) Dump(ctx context.Context, job config.Job, dbName, outputPath string) error {
	uri := d.buildURI(job) + "/" + dbName

	cmd := exec.CommandContext(ctx, "mongodump",
		"--uri="+uri,
		"--archive",
	)

	zstd := exec.CommandContext(ctx, "zstd", "-o", outputPath)
	zstd.Stdin, _ = cmd.StdoutPipe()
	zstd.Stderr = os.Stderr

	if err := zstd.Start(); err != nil {
		return fmt.Errorf("zstd start: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("mongodump start: %w", err)
	}

	cmdErr := cmd.Wait()
	zstdErr := zstd.Wait()

	if cmdErr != nil {
		return fmt.Errorf("mongodump: %w", cmdErr)
	}
	if zstdErr != nil {
		return fmt.Errorf("zstd: %w", zstdErr)
	}

	return nil
}

func (d *MongoDriver) RequiredBinaries() []string {
	return []string{"mongodump"}
}

func (d *MongoDriver) Extension() string {
	return "json"
}


