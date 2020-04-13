package metric

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/ipfs-force-community/gosf/logger"
	"github.com/ipfs-force-community/gosf/proc"
)

const (
	envForcePrometheusGateway = "FORCE_PROMETHEUS_GATEWAY"

	defaultGateway = "127.0.0.1:9091"
)

var (
	pusherOnce sync.Once
	collectors []prometheus.Collector
)

func gatewayFromEnv() string {
	if g, _ := os.LookupEnv(envForcePrometheusGateway); g != "" {
		return g
	}

	return defaultGateway
}

// DefaultConfig default push config
var DefaultConfig = PushConfig{
	Timeout:  3 * time.Second,
	Interval: 15 * time.Second,
}

// PushConfig push config
type PushConfig struct {
	Gateway  string
	Timeout  time.Duration
	Interval time.Duration
}

// Collect register collectors
func Collect(c ...prometheus.Collector) {
	collectors = append(collectors, c...)
}

// Run start a pusher
func Run(ctx context.Context, cfg PushConfig) {
	pusherOnce.Do(func() {
		run(ctx, cfg)
	})
}

func run(ctx context.Context, cfg PushConfig) {
	if len(collectors) == 0 {
		logger.LS().Info("no prometheus collector applied")
		return
	}

	gateway := cfg.Gateway
	if gateway == "" {
		gateway = gatewayFromEnv()
		logger.LS().Infof("prometheus gateway not set in config, use %s", gateway)
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultConfig.Timeout
	}

	if cfg.Interval <= 0 {
		cfg.Interval = DefaultConfig.Interval
	}

	pusher := push.New(gateway, "backend").Grouping("host", proc.Hostname())
	for i := range collectors {
		pusher = pusher.Collector(collectors[i])
	}

	pushTicker := time.NewTicker(cfg.Interval)

	for {
		select {
		case <-ctx.Done():
			return

		case <-pushTicker.C:
			if err := pusher.Push(); err != nil {
				logger.LS().Debugf("unable to push to prometheus gateway, err=%v", err)
			}
		}
	}
}
