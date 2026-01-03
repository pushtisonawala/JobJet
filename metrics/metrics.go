package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	JobsRecieved = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "jobs_recieved_total",
			Help: "Total number of jobs recieved",
		},
	)
	JobsProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "jobs_processed_total",
			Help: "Total number of jobs processed",
		},
	)
	JobsFailed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "jobs_failed_total",
			Help: "Total number of jobs failed",
		},
	)
	JobsRetried = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "jobs_retried_total",
			Help: "Total number of jobs retried",
		},
	)
	JobsDLQ = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "jobs_dlq_total",
			Help: "Total number of jobs moved to dlq",
		},
	)
	JobQueueLength = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "jobs_queue_length",
			Help: "Current job queue length",
		},
	)
	ProcessingQueueLength = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "processing_queue_length",
			Help: "Current processing queue length",
		},
	)
	RetryQueueLength = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "retry_queue_length",
			Help: "current retry queue length",
		},
	)
	DLQLength = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "dlq_queue_length",
			Help: "current dlq length",
		},
	)
	WorkerRestarts = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "worker_restarts_total",
			Help: "Total number of worker process starts",
		},
	)
)

func Init() {
	prometheus.MustRegister(
		JobsRecieved, JobsProcessed, JobsFailed, JobsRetried, JobsDLQ, JobQueueLength, ProcessingQueueLength, RetryQueueLength, DLQLength, WorkerRestarts,
	)
}
