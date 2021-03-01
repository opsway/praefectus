package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type QueryMessageGauge interface {
	prometheus.Collector
}

// @see prometheus.Opts
type QueryMessageGaugeOpts struct {
	Namespace    string
	Subsystem    string
	Name         string
	Help         string
	ConstLabels  prometheus.Labels
	MessageState QueueMessageState
}

type queryMessageGaugeCollector struct {
	qmStorage      *QueueMessageStorage
	desc           *prometheus.Desc
	state          QueueMessageState
	lastGatherTime time.Time
}

func NewQueryMessageGauge(qmStorage *QueueMessageStorage, opts QueryMessageGaugeOpts) *queryMessageGaugeCollector {
	return &queryMessageGaugeCollector{
		qmStorage: qmStorage,
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
			opts.Help,
			nil,
			opts.ConstLabels,
		),
		state:          opts.MessageState,
		lastGatherTime: time.Now(),
	}
}

func (c *queryMessageGaugeCollector) Collect(ch chan<- prometheus.Metric) {
	value := c.qmStorage.CountByState(c.state, c.lastGatherTime)
	c.lastGatherTime = time.Now()

	ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, value)
}

func (c *queryMessageGaugeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.desc
}
