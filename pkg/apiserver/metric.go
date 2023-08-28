package apiserver

import (
	"k8s.io/component-base/metrics"

	utilsmetrics "github.com/sunweiwe/horizon/pkg/utils/metrics"
)

var (
	RequestCounter = metrics.NewCounterVec(
		&metrics.CounterOpts{
			Name:           "hz_server_request_total",
			Help:           "Counter of hz_server requests broken out for each verb, group, version, resource and HTTP response code.",
			StabilityLevel: metrics.ALPHA,
		},
		[]string{"verb", "group", "version", "resource", "code"},
	)

	RequestLatencies = metrics.NewHistogramVec(
		&metrics.HistogramOpts{
			Name: "hz_server_request_duration_seconds",
			Help: "Response latency distribution in seconds for each verb, group, version, resource",
			Buckets: []float64{0.05, 0.1, 0.15, 0.2, 0.25, 0.3, 0.35, 0.4, 0.45, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0,
				1.25, 1.5, 1.75, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5, 6, 7, 8, 9, 10, 15, 20, 25, 30, 40, 50, 60},
			StabilityLevel: metrics.ALPHA,
		},
		[]string{"verb", "group", "version", "resource"},
	)

	metricsList = []metrics.Registerable{
		RequestCounter,
	}
)

func registerMetrics() {
	for _, m := range metricsList {
		utilsmetrics.MustRegister(m)
	}
}
