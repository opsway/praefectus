package rpc

import (
	"fmt"
	"net/rpc"

	"github.com/boodmo/praefectus/internal/storage"
)

type WorkerRPCHandler struct {
	stateStorage *storage.ProcStorage
}

func NewRPCHandler(ps *storage.ProcStorage) *WorkerRPCHandler {
	return &WorkerRPCHandler{
		stateStorage: ps,
	}
}

func Register(h *WorkerRPCHandler) error {
	return rpc.Register(h)
}

func (h *WorkerRPCHandler) ChangeState(payload map[string]int, r *map[string]string) error {
	pid, found := payload["pid"]
	if !found {
		*r = map[string]string{"status": "Error", "msg": "PID is required"}
		return nil // ToDo: or return error
	}

	state, found := payload["state"]
	if !found {
		*r = map[string]string{"status": "Error", "msg": "State is required"}
		return nil // ToDo: or return error
	}

	if ws := h.stateStorage.Get(pid); ws != nil {
		if err := h.stateStorage.ChangeState(ws, state); err != nil {
			*r = map[string]string{"status": "Error", "msg": "State is required"}
			return nil // ToDo: or return error
		}

		fmt.Printf("State changed to <%d>\n", state)
		*r = map[string]string{"status": "OK"}
		return nil
	}

	*r = map[string]string{"status": "Error", "msg": "Unknown process"}
	return nil // ToDo: or return error
}
