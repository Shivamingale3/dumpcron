package events

import (
	"fmt"
)

func EmitStarted() {
	fmt.Println("dumpcron DUMPCRON_STARTED: service started")
}

func EmitStopped() {
	fmt.Println("dumpcron DUMPCRON_STOPPED: service stopped")
}

func EmitConfigInvalid(reason string) {
	fmt.Printf("dumpcron DUMPCRON_CONFIG_INVALID: %s\n", reason)
}

func EmitDependencyMissing(binary string) {
	fmt.Printf("dumpcron DUMPCRON_DEPENDENCY_MISSING: %s not found\n", binary)
}

func EmitStorageInvalid(reason string) {
	fmt.Printf("dumpcron DUMPCRON_STORAGE_INVALID: %s\n", reason)
}

func EmitJobStarted(jobName string) {
	fmt.Printf("dumpcron BACKUP_JOB_STARTED: %s\n", jobName)
}

func EmitJobCompleted(jobName string) {
	fmt.Printf("dumpcron BACKUP_JOB_COMPLETED: %s\n", jobName)
}

func EmitJobFailed(jobName string) {
	fmt.Printf("dumpcron BACKUP_JOB_FAILED: %s\n", jobName)
}

func EmitDatabaseBackupFailed(jobName, dbName, reason string) {
	fmt.Printf("dumpcron DATABASE_BACKUP_FAILED: %s/%s: %s\n", jobName, dbName, reason)
}
