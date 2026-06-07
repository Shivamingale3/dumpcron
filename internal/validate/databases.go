package validate

import (
	"fmt"
	"net"
	"strconv"

	"github.com/Shivamingale3/dumpcron/internal/config"
	"github.com/Shivamingale3/dumpcron/internal/drivers"
)

func ValidateDatabases(cfg *config.Config, drvs map[string]drivers.Driver) []error {
	var errs []error

	for i := range cfg.Jobs {
		job := &cfg.Jobs[i]

		drv, ok := drvs[job.Type]
		if !ok {
			errs = append(errs, fmt.Errorf("job %q: no driver for type %q", job.Name, job.Type))
			continue
		}

		addr := net.JoinHostPort(job.Host, strconv.Itoa(job.Port))
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			errs = append(errs, fmt.Errorf("job %q: %s not reachable: %v", job.Name, addr, err))
			continue
		}
		conn.Close()

		if err := drv.Validate(*job); err != nil {
			errs = append(errs, fmt.Errorf("job %q: %v", job.Name, err))
		}
	}

	return errs
}
