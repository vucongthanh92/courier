package tracing

import (
	"github.com/vucongthanh92/go-base-utils/logger"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func UseOpenTelemetry(config Config, logger ...logger.Logger) {
	tracerProvider, err := tracerProvider(config)
	if !config.Enable {
		return
	}
	if err != nil {
		panic(err)
	}
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	if len(logger) > 0 {
		otelzap.ReplaceGlobals(otelzap.New(logger[0].GetZapLogger()))
	}
}
