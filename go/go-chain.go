package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/bmizerany/pat"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var serviceName = semconv.ServiceNameKey.String("go-chain")
var serviceNameSpace = semconv.ServiceNamespaceKey.String("kjt-OTel-chain")

// Initialize a gRPC connection to be used bythe tracer providers
func initConn() (*grpc.ClientConn, error) {
	// It connects the OpenTelemetry Collector through local gRPC connection.
	// You may replace `localhost:4317` with your endpoint.
	conn, err := grpc.NewClient("host.docker.internal:4317",
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	return conn, err
}

// Initializes an OTLP exporter, and configures the corresponding trace provider.
func initTracerProvider(ctx context.Context, res *resource.Resource, conn *grpc.ClientConn) (func(context.Context) error, error) {
	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider.Shutdown, nil
}

func main() {

	log.Println("Starting main!")
	log.Printf("Waiting for connection...")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	conn, err := initConn()
	if err != nil {
		log.Fatal(err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// The service name used to display traces in backends
			serviceName,
			serviceNameSpace,
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	shutdownTracerProvider, err := initTracerProvider(ctx, res, conn)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdownTracerProvider(ctx); err != nil {
			log.Fatalf("failed to shutdown TracerProvider: %s", err)
		}
	}()

	goPort := os.Getenv("GO_PORT")

	// Set propagator
	//propagator := propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})
	//otel.SetTextMapPropagator(propagator)

	// Create a "pat" router
	r := pat.New()

	otelNodeChainHandler := otelhttp.NewHandler(http.HandlerFunc(nodeChainHandler), "GET /node-chain")
	otelGoChainHandler := otelhttp.NewHandler(http.HandlerFunc(goChainHandler), "GET /go-chain")

	r.Get("/node-chain", otelNodeChainHandler)
	r.Get("/go-chain", otelGoChainHandler)

	http.ListenAndServe(":"+goPort, r)
}

func nodeChainHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Incoming node-chain Request Received!")

	// Grab URLs and Ports
	pythonUrl := os.Getenv("PYTHON_URL")
	pythonPort := os.Getenv("PYTHON_PORT")
	// Create target URL for downstream call
	pythonTarget := "http://" + pythonUrl + ":" + pythonPort + "/node-chain"
	url := pythonTarget
	//url := flag.String("server", pythonTarget, "server url")
	//flag.Parse()

	// Create a named tracer
	tracer := otel.Tracer("go-otel-api")

	// Grab incoming context from request
	ctx := req.Context()
	// Start span from incoming context
	reqSpan := trace.SpanFromContext(ctx)
	defer reqSpan.End()

	// Disabling the manual spans for now to keep this *relatively* simple
	// Create a parent span called "operation"
	func(ctx context.Context) {
		var span trace.Span
		ctx, span = tracer.Start(ctx, "node-chain operation")
		defer span.End()

		// Create a child span called "sub-operation"
		func(ctx context.Context) {
			var span trace.Span
			ctx, span = tracer.Start(ctx, "node-chain sub-operation")
			defer span.End()

			// Wait between 0 and 3 seconds
			//rand.Seed(time.Now().UnixNano())
			//n := rand.Intn(3) // n will be between 0 and 3
			//log.Printf("Sleeping %d seconds...\n", n)
			//time.Sleep(time.Duration(n) * time.Second)
		}(ctx)
	}(ctx)

	// Let's create another span to make a downstream call

	// Create HTTP Client and enable Otel Propagation
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	var body []byte

	err := func(ctx context.Context) error {
		ctx, span := tracer.Start(ctx, "node-chain to python", trace.WithAttributes(semconv.PeerServiceKey.String("go-otel-api-client")))
		defer span.End()
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

		log.Printf("Sending request to " + pythonTarget)
		res, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		body, err = io.ReadAll(res.Body)
		_ = res.Body.Close()

		return err
	}(ctx)
	if err != nil {
		log.Fatalf("Response NOT received!")
		panic(err)
	}
	log.Printf("Response Received!")
	log.Printf(string(body))
	_, _ = io.WriteString(w, "{\"otel\":\"go\"}")
}

func goChainHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Incoming go-chain Request Received!")

	// Grab URLs and Ports
	pythonUrl := os.Getenv("PYTHON_URL")
	pythonPort := os.Getenv("PYTHON_PORT")
	// Create target URL for downstream call
	pythonTarget := "http://" + pythonUrl + ":" + pythonPort
	url := pythonTarget
	//url := flag.String("server", pythonTarget, "server url")
	//flag.Parse()

	// Create a named tracer
	tracer := otel.Tracer("go-otel-api")

	// Grab incoming context from request
	ctx := req.Context()
	// Start span from incoming context
	reqSpan := trace.SpanFromContext(ctx)
	defer reqSpan.End()

	// Create a parent span called "operation"
	func(ctx context.Context) {
		var span trace.Span
		ctx, span = tracer.Start(ctx, "go-chain operation")
		defer span.End()

		// Create a child span called "sub-operation"
		func(ctx context.Context) {
			var span trace.Span
			ctx, span = tracer.Start(ctx, "go-chain sub-operation")
			defer span.End()

			// Wait between 0 and 3 seconds
			rand.Seed(time.Now().UnixNano())
			n := rand.Intn(3) // n will be between 0 and 3
			fmt.Printf("Sleeping %d seconds...\n", n)
			time.Sleep(time.Duration(n) * time.Second)
		}(ctx)
	}(ctx)

	// Let's create another span to make a downstream call

	// Create HTTP Client and enable Otel Propagation
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	err := func(ctx context.Context) error {

		ctx, span := tracer.Start(ctx, "go-chain to python", trace.WithAttributes(semconv.PeerServiceKey.String("go-otel-api-client")))
		defer span.End()
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

		fmt.Printf("Sending request...\n")
		res, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		//body, err = io.ReadAll(res.Body)
		_ = res.Body.Close()

		return err
	}(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Response Received!")
	_, _ = io.WriteString(w, "{\"Hello\":\"world\"}")
}
