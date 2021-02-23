package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"google.golang.org/grpc"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/propagation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
)

// Initializes an OTLP exporter, and configures the corresponding trace provider
func initProvider() func() {
	ctx := context.Background()

	// If the OpenTelemetry Collector is running on a local cluster (minikube or
	// microk8s), it should be accessible through the NodePort service at the
	// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
	// endpoint of your cluster. If you run the app inside k8s, then you can
	// probably connect directly to the service through dns
	driver := otlpgrpc.NewDriver(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint("host.docker.internal:4317"),
		otlpgrpc.WithDialOption(grpc.WithBlock()), // useful for testing
	)
	exp, err := otlp.NewExporter(ctx, driver)
	handleErr(err, "failed to create exporter")

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String("go-otel-api-server"),
			// the service namespace used to display traces in backends
			semconv.ServiceNamespaceKey.String("kjt-OTel-go-otel-api"),
		),
	)
	handleErr(err, "failed to create resource")

	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	cont := controller.New(
		processor.New(
			simple.NewWithExactDistribution(),
			exp,
		),
		controller.WithPusher(exp),
		controller.WithCollectPeriod(2*time.Second),
	)

	// set global propagator to tracecontext (the default is no-op)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	//otel.SetMeterProvider(cont.MeterProvider())
	//otel.SetTextMapPropagator(propagation.TraceContext{})
	handleErr(cont.Start(context.Background()), "failed to start controller")

	return func() {
		// Shutdown will flush any remaining spans.
		handleErr(tracerProvider.Shutdown(ctx), "failed to shutdown TracerProvider")

		// Push any last metric events to the exporter.
		handleErr(cont.Stop(context.Background()), "failed to stop controller")
	}
}

func main() {

	// Initialize OTLP & Trace Providers
	shutdown := initProvider()
	defer shutdown()

	// Create a named tracer
	tracer := otel.Tracer("go-otel-api")
	//ctx := context.Background()
	
	// labels represent additional key-value descriptors that can be bound to a
	// metric observer or recorder.
	/*
	operationLabels := []label.KeyValue{
		label.String("author", "ktully"),
		label.String("org", "appd"),
	}
	*/

	//ctx = baggage.ContextWithValues(ctx, fooKey.String("foo1"), barKey.String("bar1"))

	// Set propagator
	propagator := propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})
	otel.SetTextMapPropagator(propagator)

	helloHandler := func(w http.ResponseWriter, req *http.Request) {

		// Grab incoming context from request
		ctx := req.Context()
		// Start span from incoming context
		reqSpan := trace.SpanFromContext(ctx)
		defer reqSpan.End()

		// Create a parent span called "operation"
		func(ctx context.Context) {
			var span trace.Span
			ctx, span = tracer.Start(ctx, "operation")
			defer span.End()

			span.AddEvent("Nice operation!", trace.WithAttributes(label.Int("bogons", 100)))
			anotherKey := label.Key("go-otel-api/derp")
			span.SetAttributes(anotherKey.String("derpOperation"))

			// Create a child span called "sub-operation"
			func(ctx context.Context) {
				var span trace.Span
				ctx, span = tracer.Start(ctx, "sub-operation")
				defer span.End()

				span.SetAttributes(anotherKey.String("derpSubOperation"))
				span.AddEvent("Sub span event")

				// Wait between 0 and 3 seconds
				rand.Seed(time.Now().UnixNano())
				n := rand.Intn(3) // n will be between 0 and 3
				fmt.Printf("Sleeping %d seconds...\n", n)
				time.Sleep(time.Duration(n)*time.Second)
			}(ctx)
		}(ctx)

		_, _ = io.WriteString(w, "{\"Hello\":\"world\"}")
	}

	otelHandler := otelhttp.NewHandler(http.HandlerFunc(helloHandler), "Hello")
	
	// Run the HTTP server and handle request to /
	http.Handle("/", otelHandler)
	err := http.ListenAndServe(":9000", nil)
	if err != nil {
		panic(err)
	}
}

func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}