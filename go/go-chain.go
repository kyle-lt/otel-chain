package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"

	"github.com/bmizerany/pat"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/propagation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
)

func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}

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
	if err != nil {
		handleErr(err, "failed to create exporter")
	} else {
		log.Printf("created new exporter!")
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String("go-chain"),
			// the service namespace used to display traces in backends
			semconv.ServiceNamespaceKey.String("kjt-OTel-chain"),
		),
	)
	if err != nil {
		handleErr(err, "failed to create resource")
	} else {
		log.Printf("created new resource!")
	}

	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	cont := controller.New(
		processor.New(
			simple.NewWithExactDistribution(),
			exp,
		),
		controller.WithExporter(exp),
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

	goPort := os.Getenv("GO_PORT")

	// Initialize OTLP & Trace Providers
	shutdown := initProvider()
	defer shutdown()

	// Set propagator
	propagator := propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})
	otel.SetTextMapPropagator(propagator)

	// Create a "pat" router
	r := pat.New()

	otelNodeChainHandler := otelhttp.NewHandler(http.HandlerFunc(nodeChainHandler), "GET /node-chain")
	otelGoChainHandler := otelhttp.NewHandler(http.HandlerFunc(goChainHandler), "GET /go-chain")
	
	r.Get("/node-chain", otelNodeChainHandler)
	r.Get("/go-chain", otelGoChainHandler)

	http.ListenAndServe(":" + goPort, r)
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

	/* Disabling the manual spans for now to keep this *relatively* simple
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
	*/

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
	var body []byte

	err := func(ctx context.Context) error {
		ctx, span := tracer.Start(ctx, "go-chain to python", trace.WithAttributes(semconv.PeerServiceKey.String("go-otel-api-client")))
		defer span.End()
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

		fmt.Printf("Sending request...\n")
		res, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		body, err = io.ReadAll(res.Body)
		_ = res.Body.Close()

		return err
	}(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Response Received!")
	_, _ = io.WriteString(w, "{\"Hello\":\"world\"}")
}