package storage

const (
	StateUnknown = iota
	StateStarting
	StateStared
	StateIdle
	StateBusy
	StateStopping
	StateStopped
)

func StateName(state int) string {
	stateName := "Unknown"
	switch state {
	case StateStarting:
		stateName = "Starting"
	case StateStared:
		stateName = "Started"
	case StateIdle:
		stateName = "IDLE"
	case StateBusy:
		stateName = "Busy"
	case StateStopping:
		stateName = "Stopping"
	case StateStopped:
		stateName = "Stopped"
	}

	return stateName
}
