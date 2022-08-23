package rpc

import (
	"fmt"
	"github.com/opsway/praefectus/internal/config"
	"github.com/opsway/praefectus/internal/workers"
	"net/http"
)

type health struct {
	cfg *config.Config
}

func NewHealthHandler(cfg *config.Config) *health {
	return &health{
		cfg: cfg,
	}
}

func (h *health) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, command := range h.cfg.Workers {
		runningWorkers := workers.RunningProcesses(command)
		if len(runningWorkers) == 0 {
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("500 - No workers found for %s!", command)))
			return
		}
		if !workers.CheckWorkersIpcListener(runningWorkers...) {
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("500 - Ipc health-check failed for %s!", command)))
			return
		}
	}

	w.WriteHeader(200)
	w.Write([]byte("200 - All checks passed!"))
}
