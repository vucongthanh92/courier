package server

import (
	"github.com/vucongthanh92/go-base-utils/metrics"

	"github.com/vucongthanh92/go-base-utils/logger"
)

type HttpServerConfig struct {
	Port            string
	Development     bool
	ShutdownTimeout int

	Resources    []string
	RateLimiting *RateLimitingConfig
	Name         string
	AllowOrigins []string
	MetricConfig *metrics.MetricsConfig
}

type RateLimitingConfig struct {
	RateFormat string
}

type HttpServerOption func(*Server)

func WithLogger(log logger.Logger) HttpServerOption {
	return func(s *Server) {
		s.log = log
	}
}
