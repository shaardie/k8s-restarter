package controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	histoBuckets = []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	opsRestarts  = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "k8s_restarter_restarts",
		Help: "The number of restarted apps in the last reconcilation",
	})
	opsExcluded = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "k8s_restarter_ignores",
		Help: "The number of ignored apps in the last reconcilation",
	})
	opsSkips = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "k8s_restarter_skips",
		Help: "The number of skipped apps in the last reconcilation",
	})
	opsRestartsHisto = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "k8s_restarter_restarts_histo",
		Help:    "The number of restarted apps",
		Buckets: histoBuckets,
	})
	opsExcludedHisto = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "k8s_restarter_ignores_histo",
		Help:    "The number of ignored apps",
		Buckets: histoBuckets,
	})
	opsSkipsHisto = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "k8s_restarter_skips_histo",
		Help:    "The number of skipped apps",
		Buckets: histoBuckets,
	})
)
