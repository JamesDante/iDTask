package monitor

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	workerTasksExecuted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "worker_tasks_executed_total",
		Help: "Total number of tasks executed by worker",
	})

	workerTasksFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "worker_tasks_failed_total",
		Help: "Total number of failed tasks",
	})

	workerTaskExecDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "worker_task_execution_duration_seconds",
		Help:    "Histogram of task execution duration",
		Buckets: prometheus.DefBuckets,
	})
)

// Scheduler metrics
var (
	schedulerTasksScheduled = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "scheduler_tasks_scheduled_total",
		Help: "Total number of tasks scheduled by the scheduler",
	})

	schedulerTasksFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "scheduler_tasks_failed_total",
		Help: "Total number of failed task scheduling attempts",
	})
)

var (
	apiRequestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "api_requests_total",
		Help: "Total number of API requests received",
	})
)

// Export accessors
func WorkerTasksExecuted() prometheus.Counter {
	return workerTasksExecuted
}

func WorkerTasksFailed() prometheus.Counter {
	return workerTasksFailed
}

func WorkerTaskExecDuration() prometheus.Histogram {
	return workerTaskExecDuration
}

func SchedulerTasksScheduled() prometheus.Counter {
	return schedulerTasksScheduled
}

func SchedulerTasksFailed() prometheus.Counter {
	return schedulerTasksFailed
}

func InitWorkerMetrics() {
	prometheus.MustRegister(workerTasksExecuted, workerTasksFailed, workerTaskExecDuration)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Println("📈 Prometheus metrics exposed at :8083/metrics")
		if err := http.ListenAndServe(":8083", nil); err != nil {
			log.Fatalf("metrics server error: %v", err)
		}
	}()
}

func InitSchedulerMetrics() {
	prometheus.MustRegister(schedulerTasksScheduled, schedulerTasksFailed)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Println("📈 Prometheus metrics exposed at :8082/metrics")
		if err := http.ListenAndServe(":8082", nil); err != nil {
			log.Fatalf("metrics server error: %v", err)
		}
	}()
}

func InitApiMetrics() {
	prometheus.MustRegister(apiRequestsTotal)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Println("📈 API Prometheus metrics exposed at :8081/metrics")
		if err := http.ListenAndServe(":8081", mux); err != nil {
			log.Fatalf("api metrics server error: %v", err)
		}
	}()
}

func ApiRequestsTotal() prometheus.Counter {
	return apiRequestsTotal
}
