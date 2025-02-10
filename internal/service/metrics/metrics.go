package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	RedirectCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "url_shortener_redirects_total",
			Help: "Number of redirects performed by the URL shortener",
		},
		[]string{"short_path"},
	)

	ReconcileErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "url_shortener_reconcile_errors_total",
			Help: "Number of reconciliation errors",
		},
	)
)

func init() {
	metrics.Registry.MustRegister(RedirectCount)
	metrics.Registry.MustRegister(ReconcileErrors)
}
