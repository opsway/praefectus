package metrics

import (
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	// Storages
	qStorage  *QueueStorage
	qmStorage *QueueMessageStorage
	wsStorage *WorkerStatStorage

	// Metrics
	workersInIdleState   prometheus.Gauge
	workersInBusyState   prometheus.Gauge
	messageFailedCount   QueryMessageGauge
	messageSucceedCount  QueryMessageGauge
	messageProcessedTime *prometheus.HistogramVec
	queueSize            *prometheus.GaugeVec
}

func NewMetrics(qStorage *QueueStorage, qmStorage *QueueMessageStorage, wsStorage *WorkerStatStorage) (*Metrics, error) {
	workersInIdleState := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "praefectus",
			Name:      "workers_in_idle_state",
			Help:      "Number of workers in IDLE state",
		},
	)
	if err := prometheus.Register(workersInIdleState); err != nil {
		return nil, err
	}

	workersInBusyState := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "praefectus",
			Name:      "workers_in_busy_state",
			Help:      "Number of workers in busy state",
		},
	)
	if err := prometheus.Register(workersInBusyState); err != nil {
		return nil, err
	}

	messageFailedCount := NewQueryMessageGauge(qmStorage, QueryMessageGaugeOpts{
		Namespace:    "praefectus",
		Name:         "message_failed_count",
		Help:         "Number of messages which processing was failed",
		MessageState: MessageStateFailed,
	})
	if err := prometheus.Register(messageFailedCount); err != nil {
		return nil, err
	}

	messageSucceedCount := NewQueryMessageGauge(qmStorage, QueryMessageGaugeOpts{
		Namespace:    "praefectus",
		Name:         "message_succeed_count",
		Help:         "Number of successfully processed messages",
		MessageState: MessageStateSucceed,
	})
	if err := prometheus.Register(messageSucceedCount); err != nil {
		return nil, err
	}

	messageProcessedTime := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "praefectus",
			Name:      "message_processed_time",
			Help:      "Duration of message processing",
			Buckets:   []float64{.5, 3, 30, 120, 600, 3600},
		},
		[]string{"name", "transport", "bus", "status"},
	)
	if err := prometheus.Register(messageProcessedTime); err != nil {
		return nil, err
	}

	queueSize := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "praefectus",
			Name:      "queue_size",
			Help:      "Message queue size",
		},
		[]string{"transport", "bus"},
	)
	if err := prometheus.Register(queueSize); err != nil {
		return nil, err
	}

	return &Metrics{
		qStorage:             qStorage,
		qmStorage:            qmStorage,
		wsStorage:            wsStorage,
		workersInIdleState:   workersInIdleState,
		workersInBusyState:   workersInBusyState,
		messageFailedCount:   messageFailedCount,
		messageSucceedCount:  messageSucceedCount,
		messageProcessedTime: messageProcessedTime,
		queueSize:            queueSize,
	}, nil
}

func (m *Metrics) Start() {
	lastGatherTime := time.Now()
	for {
		m.workersInIdleState.Set(m.wsStorage.CountByState(WorkerStateIdle))
		m.workersInBusyState.Set(m.wsStorage.CountByState(WorkerStateBusy))

		processedMessages := m.qmStorage.GetProcessedAfter(lastGatherTime)
		lastGatherTime = time.Now()
		for _, qm := range processedMessages {
			m.messageProcessedTime.
				WithLabelValues(qm.Name, qm.Transport, qm.Bus, strings.ToLower(qm.State.String())).
				Observe(qm.GetProcessedTime().Seconds())
		}
		m.updateQueueMetrics()

		time.Sleep(5 * time.Second)
	}
}

func (m *Metrics) updateQueueMetrics() {
	m.qStorage.mu.Lock()
	defer m.qStorage.mu.Unlock()

	for _, q := range m.qStorage.storage {
		m.queueSize.WithLabelValues(q.Transport, q.Bus).Set(float64(q.Size))
	}
}
