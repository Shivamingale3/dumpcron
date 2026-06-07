package validate

import (
	"fmt"
	"os"
)

func ValidateStorage(backupRoot string) []error {
	var errs []error

	info, err := os.Stat(backupRoot)
	if err != nil {
		if os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("backup_root %q does not exist", backupRoot))
		} else {
			errs = append(errs, fmt.Errorf("backup_root %q: %v", backupRoot, err))
		}
		return errs
	}

	if !info.IsDir() {
		errs = append(errs, fmt.Errorf("backup_root %q is not a directory", backupRoot))
		return errs
	}

	tmp, err := os.CreateTemp(backupRoot, ".dumpcron-write-test-*")
	if err != nil {
		errs = append(errs, fmt.Errorf("backup_root %q is not writable", backupRoot))
		return errs
	}
	tmp.Close()
	os.Remove(tmp.Name())

	return errs
}
