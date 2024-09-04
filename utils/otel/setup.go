package otel

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitProvider(ctx context.Context) func() {
	res := newResource(ctx)
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Use insecure or secure credentials based on your needs
	}
	conn, err := grpc.NewClient(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"), opts...)
	if err != nil {
		panic(err)
	}
	tracerProvider, traceExporter := newTracerProvider(ctx, res, conn)
	loggerProvider := newLoggerProvider(ctx, res, conn)
	setPropagator()

	global.SetLoggerProvider(loggerProvider)
	otel.SetTracerProvider(tracerProvider)

	return func() {
		shutdownProviders(ctx, traceExporter, loggerProvider)
	}
}

func newResource(ctx context.Context) *resource.Resource {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("sample-app-otel-collector"),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		log.Fatalf("Failed to create resource: %v", err)
	}
	return res
}

func newTracerProvider(ctx context.Context, res *resource.Resource, conn *grpc.ClientConn) (*sdktrace.TracerProvider, *otlptrace.Exporter) {
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithGRPCConn(conn),
	)

	traceExporter, err := otlptrace.New(ctx, client)
	if err != nil {
		log.Fatalf("Failed to create trace exporter: %v", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	return tracerProvider, traceExporter
}

func newLoggerProvider(ctx context.Context, res *resource.Resource, conn *grpc.ClientConn) *sdklog.LoggerProvider {
	exporter, err := otlploggrpc.New(ctx, otlploggrpc.WithGRPCConn(conn))
	if err != nil {
		return nil
	}
	processor := sdklog.NewBatchProcessor(exporter)
	provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(processor),
	)
	return provider
}

func setPropagator() {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
}

func shutdownProviders(ctx context.Context, traceExporter *otlptrace.Exporter, loggerProvider *sdklog.LoggerProvider) {
	if err := traceExporter.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down trace exporter: %v", err)
	}
	if err := loggerProvider.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down logger provider: %v", err)
	}
}
