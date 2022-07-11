package workers

import (
	"github.com/rs/xid"
	"sync"
)

type CommandState uint8

const (
	Fresh CommandState = iota
	MarkRemove
	Remove
)

type WorkerStorage struct {
	storage map[xid.ID]*WorkerCommand
	mu      sync.Mutex
}

type ProcessId struct {
	id int
}

type WorkerCommand struct {
	id        xid.ID
	state     CommandState
	command   string
	processId *ProcessId
	stop      chan struct{}
}

func NewWorkerCommand(command string) *WorkerCommand {
	return &WorkerCommand{
		id:      xid.New(),
		state:   Fresh,
		command: command,
		stop:    make(chan struct{}),
	}
}
