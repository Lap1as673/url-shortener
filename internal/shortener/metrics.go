package shortener

import "github.com/prometheus/client_golang/prometheus"

var ClicksTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "shortener",
		Name:      "clicks_total",
		Help:      "Total number of redirects served.",
	},
)

func init() {
	prometheus.MustRegister(ClicksTotal)
}
