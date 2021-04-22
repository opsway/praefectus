package rpc

import (
	"net/rpc"

	"github.com/boodmo/praefectus/internal/metrics"
)

type PraefectusRPC struct {
	qStorage  *metrics.QueueStorage
	qmStorage *metrics.QueueMessageStorage
	wsStorage *metrics.WorkerStatStorage
}

func NewRPCHandler(qStorage *metrics.QueueStorage, qmStorage *metrics.QueueMessageStorage, wsStorage *metrics.WorkerStatStorage) *PraefectusRPC {
	return &PraefectusRPC{
		qStorage:  qStorage,
		qmStorage: qmStorage,
		wsStorage: wsStorage,
	}
}

func Register(h *PraefectusRPC) error {
	return rpc.Register(h)
}

func (h *PraefectusRPC) WorkerState(payload map[string]int, r *map[string]string) error {
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

	if ws := h.wsStorage.Get(pid); ws != nil {
		if err := h.wsStorage.ChangeState(ws, metrics.WorkerState(state)); err != nil {
			*r = map[string]string{"status": "Error", "msg": "State is required"}
			return nil // ToDo: or return error
		}

		*r = map[string]string{"status": "OK"}
		return nil
	}

	*r = map[string]string{"status": "Error", "msg": "Unknown process"}
	return nil // ToDo: or return error
}

func (h *PraefectusRPC) MessageState(payload map[string]interface{}, r *map[string]string) error {
	id := payload["id"]

	// ToDo: Map to struct
	qm := h.qmStorage.Get(id.(string))
	if qm != nil {
		state := payload["state"]
		if err := h.qmStorage.ChangeState(qm, metrics.QueueMessageState(state.(float64))); err != nil {
			*r = map[string]string{"status": "Error", "msg": "State is required"}
			return nil // ToDo: or return error
		}
	} else {
		name := payload["name"]
		transport := payload["transport"]
		bus := payload["bus"]
		h.qmStorage.Add(id.(string), name.(string), transport.(string), bus.(string))
	}

	return nil // ToDo: or return error
}

func (h *PraefectusRPC) QueueSize(payload map[string]interface{}, r *map[string]string) error {
	// ToDo: Map to struct
	transport := payload["transport"]
	bus := payload["bus"]
	size := payload["size"]

	q := h.qStorage.Get(transport.(string), bus.(string))
	if q == nil {
		q = h.qStorage.Add(transport.(string), bus.(string))
	}
	h.qStorage.ChangeSize(q, uint(size.(float64)))

	return nil // ToDo: or return error
}
