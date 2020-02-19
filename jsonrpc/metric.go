package jsonrpc

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.forceup.in/dev-go/gosf/metric"
	"gitlab.forceup.in/dev-go/gosf/proc"
)

func init() {
	metric.Collect(rpcResponseMetric)
}

var rpcResponseMetric = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "service",
		Subsystem: proc.AppName(),
		Name:      "rpc",
		// 50ms, 100ms, 200ms, 500ms, 1s, 2s, 5s, 10s
		Buckets: []float64{0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10},
	},
	[]string{"url", "code"},
)

func rpcMetricAdd(u string, code int, dur time.Duration) {
	rpcResponseMetric.With(prometheus.Labels{
		"url":  u,
		"code": httpCodeRange(code),
	}).Observe(dur.Seconds())
}

func httpCodeRange(code int) string {
	if code <= 0 {
		return "unknown"
	}

	return fmt.Sprintf("%dxx", code/100)
}
