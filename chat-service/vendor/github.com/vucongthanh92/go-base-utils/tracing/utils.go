package tracing

import (
	"context"

	"github.com/vucongthanh92/go-base-utils/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc/metadata"
)

func StartHttpServerTracerSpan(c *gin.Context, operationName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
	ctx, serverSpan := tracer.Start(ctx, operationName, append(opts, trace.WithSpanKind(trace.SpanKindServer))...)
	return ctx, serverSpan
}

func StartSpanFromContext(ctx context.Context, operationName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tracer.Start(ctx, operationName, append(opts, trace.WithSpanKind(trace.SpanKindServer))...)
}

func GetTextMapCarrierFromMetaData(ctx context.Context) propagation.MapCarrier {
	metadataMap := make(propagation.MapCarrier)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for key := range md.Copy() {
			metadataMap.Set(key, md.Get(key)[0])
		}
	}
	return metadataMap
}

func StartGrpcServerTracerSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	textMapCarrierFromMetaData := GetTextMapCarrierFromMetaData(ctx)
	ctx = otel.GetTextMapPropagator().Extract(ctx, textMapCarrierFromMetaData)
	ctx, serverSpan := tracer.Start(ctx, operationName, trace.WithSpanKind(trace.SpanKindServer))
	return ctx, serverSpan
}

func StartKafkaConsumerTracerSpan(ctx context.Context, headers []kafka.Header, operationName string) (context.Context, trace.Span) {
	carrierFromKafkaHeaders := TextMapCarrierFromKafkaMessageHeaders(headers)
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrierFromKafkaHeaders)
	ctx, serverSpan := tracer.Start(ctx, operationName, trace.WithSpanKind(trace.SpanKindServer))
	return ctx, serverSpan
}

func TextMapCarrierToKafkaMessageHeaders(textMap propagation.MapCarrier) []kafka.Header {
	headers := make([]kafka.Header, 0, len(textMap))
	for k, v := range textMap {
		headers = append(headers, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}
	return headers
}

func TextMapCarrierFromKafkaMessageHeaders(headers []kafka.Header) propagation.MapCarrier {
	textMap := make(map[string]string, len(headers))
	for _, header := range headers {
		textMap[header.Key] = string(header.Value)
	}
	return textMap
}

func GetKafkaTracingHeadersFromCtx(ctx context.Context) []kafka.Header {
	textMap := make(propagation.MapCarrier)
	otel.GetTextMapPropagator().Inject(ctx, textMap)
	kafkaMessageHeaders := TextMapCarrierToKafkaMessageHeaders(textMap)
	return kafkaMessageHeaders
}

func InjectTextMapCarrierToGrpcMetaData(ctx context.Context) context.Context {
	textMap := make(propagation.MapCarrier)
	otel.GetTextMapPropagator().Inject(ctx, textMap)
	md := metadata.New(textMap)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx
}

func RecordSpan(ctx context.Context, message string, err error, span trace.Span, log ...logger.Logger) {
	if err != nil {
		logger.WithTrace(ctx, log...).Error(message, zap.Error(err))
		span.SetStatus(codes.Error, err.Error())
	}
	span.RecordError(err)
	span.End()
}
