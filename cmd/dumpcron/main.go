package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Shivamingale3/dumpcron/internal/config"
	"github.com/Shivamingale3/dumpcron/internal/core"
	"github.com/Shivamingale3/dumpcron/internal/drivers"
	"github.com/Shivamingale3/dumpcron/internal/events"
	"github.com/Shivamingale3/dumpcron/internal/scheduler"
	"github.com/Shivamingale3/dumpcron/internal/validate"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: dumpcron <command>")
		fmt.Fprintln(os.Stderr, "  validate   validate configuration")
		fmt.Fprintln(os.Stderr, "  run        start the backup scheduler")
		fmt.Fprintln(os.Stderr, "  version    print version")
		fmt.Fprintln(os.Stderr, "  uninstall  remove dumpcron from system")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "validate":
		cmdValidate()
	case "run":
		cmdRun()
	case "version":
		fmt.Println(core.Version)
	case "uninstall":
		cmdUninstall()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func cfgPath() string {
	if v := os.Getenv("DUMPCRON_CONFIG"); v != "" {
		return v
	}
	return "/etc/dumpcron/config.yaml"
}

func statePath() string {
	if v := os.Getenv("DUMPCRON_STATE"); v != "" {
		return v
	}
	return "/var/lib/dumpcron/last_run.json"
}

func makeDrivers() map[string]drivers.Driver {
	return map[string]drivers.Driver{
		"postgres": &drivers.PostgresDriver{},
		"mysql":    &drivers.MysqlDriver{},
		"mongo":    &drivers.MongoDriver{},
	}
}

func cmdValidate() {
	cfg, err := config.Load(cfgPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "config load: %v\n", err)
		os.Exit(1)
	}

	errs := validate.ValidateAll(cfg, makeDrivers())

	if len(errs) > 0 {
		fmt.Println()

		validationStep := ""

		for _, e := range errs {
			msg := e.Error()

			switch {
			case stepForValidation(msg) == "dependency":
				if validationStep != "dependency" {
					fmt.Println()
					validationStep = "dependency"
				}
				fmt.Printf("✗ %s\n", msg)

			case stepForValidation(msg) == "config":
				if validationStep != "config" {
					fmt.Println()
					validationStep = "config"
				}
				fmt.Printf("✗ %s\n", msg)

			case stepForValidation(msg) == "storage":
				if validationStep != "storage" {
					fmt.Println()
					validationStep = "storage"
				}
				fmt.Printf("✗ %s\n", msg)

			default:
				if validationStep != "database" {
					fmt.Println()
					validationStep = "database"
				}
				fmt.Printf("✗ %s\n", msg)
			}
		}

		fmt.Println()
		fmt.Println("Validation failed")
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✓ Configuration valid")
	fmt.Println("✓ Dependencies present")
	fmt.Println("✓ Storage valid")
	fmt.Println("✓ Database connectivity valid")
	fmt.Println()
	fmt.Println("Configuration OK")
}

func cmdRun() {
	cfg, err := config.Load(cfgPath())
	if err != nil {
		events.EmitConfigInvalid(fmt.Sprintf("failed to load config: %v", err))
		os.Exit(1)
	}

	drvs := makeDrivers()
	errs := validate.ValidateAll(cfg, drvs)

	if len(errs) > 0 {
		for _, e := range errs {
			msg := e.Error()
			switch {
			case strings.Contains(msg, "missing"):
				events.EmitDependencyMissing(msg)
			case strings.Contains(msg, "backup_root") || strings.Contains(msg, "writable"):
				events.EmitStorageInvalid(msg)
			default:
				events.EmitConfigInvalid(msg)
			}
		}
		os.Exit(1)
	}

	for _, dbType := range []string{"postgres", "mysql", "mongo"} {
		dir := filepath.Join(cfg.BackupRoot, dbType)
		if err := os.MkdirAll(dir, 0755); err != nil {
			events.EmitStorageInvalid(fmt.Sprintf("cannot create %s: %v", dir, err))
			os.Exit(1)
		}
	}

	events.EmitStarted()

	queue := make(chan scheduler.WorkItem, 100)
	worker := scheduler.NewWorker(queue, drvs)
	worker.Start()

	sched := scheduler.New(cfg, queue, statePath())
	if err := sched.LoadState(); err != nil {
		events.EmitConfigInvalid(fmt.Sprintf("failed to load state: %v", err))
		os.Exit(1)
	}
	sched.Start()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	<-sigCh

	sched.Stop()
	close(queue)
	worker.Wait()
	events.EmitStopped()
}

func stepForValidation(msg string) string {
	if strings.Contains(msg, "missing") {
		return "dependency"
	}
	if strings.Contains(msg, "backup_root") || strings.Contains(msg, "writable") {
		return "storage"
	}
	if strings.Contains(msg, "duplicate") || strings.Contains(msg, "required") || strings.Contains(msg, "invalid") || strings.Contains(msg, "positive") || strings.Contains(msg, "format") {
		return "config"
	}
	return "database"
}

func cmdUninstall() {
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "dumpcron uninstall must be run as root (try: sudo dumpcron uninstall)")
		os.Exit(1)
	}

	exec.Command("systemctl", "disable", "--now", "dumpcron").Run()

	os.Remove("/usr/local/bin/dumpcron")
	os.Remove("/etc/systemd/system/dumpcron.service")
	os.Remove("/etc/pidex/custom.d/dumpcron.conf")
	os.RemoveAll("/var/lib/dumpcron")

	var answer string
	fmt.Print("Remove /etc/dumpcron? [y/N]: ")
	fmt.Scanln(&answer)
	if answer == "y" || answer == "Y" {
		os.RemoveAll("/etc/dumpcron")
		fmt.Println("Removed /etc/dumpcron")
	}

	exec.Command("systemctl", "daemon-reload").Run()
	fmt.Println("Dumpcron uninstalled.")
}
