package timers

import (
	"github.com/sirupsen/logrus"
	"time"
)

type Liveness struct {
	timers *Timers
}

func newLiveness(timers *Timers) *Liveness {
	return &Liveness{
		timers: timers,
	}
}

func (s *Liveness) Check(request string, reply *bool) error {
	if s.timers.lastRun == nil {
		*reply = true
		logrus.WithField("result", true).Debug("Timers ipc check")
		return nil
	}
	checkResult := (time.Now().Unix() - s.timers.lastRun.Timestamp) <= int64(s.timers.config.Timer.Frequency)
	logrus.WithField("result", checkResult).Debug("Timers ipc check")
	*reply = checkResult

	return nil
}
