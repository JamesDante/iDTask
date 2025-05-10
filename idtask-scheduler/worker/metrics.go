package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	tasksExecuted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "worker_tasks_executed_total",
		Help: "Total number of tasks executed by worker",
	})

	tasksFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "worker_tasks_failed_total",
		Help: "Total number of failed tasks",
	})

	taskExecDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "worker_task_execution_duration_seconds",
		Help:    "Histogram of task execution duration",
		Buckets: prometheus.DefBuckets,
	})
)

func initWorkerMetrics() {
	prometheus.MustRegister(tasksExecuted, tasksFailed, taskExecDuration)

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Println("📈 Prometheus metrics exposed at :8083/metrics")
		if err := http.ListenAndServe(":8083", nil); err != nil {
			log.Fatalf("metrics server error: %v", err)
		}
	}()
}
