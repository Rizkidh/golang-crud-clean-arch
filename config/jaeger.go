package config

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel/exporters/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func InitTracerWithProvider(serviceName string) (func(), *sdktrace.TracerProvider) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(
		jaeger.WithEndpoint("http://jaeger:14268/api/traces"),
	))
	if err != nil {
		log.Fatalf("❌ Failed to initialize Jaeger exporter: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
	)

	fmt.Println("✅ Jaeger tracer initialized")
	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("⚠️ Error shutting down tracer provider: %v", err)
		}
	}, tp
}
