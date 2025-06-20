package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Email metrics
	EmailsQueued = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gomailer_emails_queued_total",
		Help: "The total number of emails queued",
	})

	EmailsSent = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gomailer_emails_sent_total",
		Help: "The total number of emails sent",
	})

	EmailErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gomailer_email_errors_total",
		Help: "The total number of email sending errors",
	})

	// Queue metrics
	QueueSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gomailer_queue_size",
		Help: "Current number of emails in the queue",
	})

	QueueLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "gomailer_queue_latency_seconds",
		Help:    "Time taken for an email to be processed from queue",
		Buckets: prometheus.DefBuckets,
	})

	// TCP metrics
	TCPConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gomailer_tcp_connections_current",
		Help: "Current number of active TCP connections",
	})

	TCPErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gomailer_tcp_errors_total",
		Help: "Total number of TCP connection errors",
	})

	TCPAuthSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gomailer_tcp_auth_success_total",
		Help: "Total number of successful TCP authentications",
	})

	TCPAuthErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gomailer_tcp_auth_errors_total",
		Help: "Total number of failed TCP authentications",
	})
) 