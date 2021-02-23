use hyper::{Method, Body, Request, Response, Server, StatusCode, Client};
use opentelemetry::sdk;
use opentelemetry::trace::TraceError;
use opentelemetry::{
    global,
    sdk::{
        propagation::TraceContextPropagator,
        trace::{Config, Sampler},
    },
    trace::{TraceContextExt, Tracer},
    Context,
};
use opentelemetry_http::HeaderExtractor;
use opentelemetry_http::HeaderInjector;
use opentelemetry_semantic_conventions as semcov;
use std::{convert::Infallible, net::SocketAddr, sync::Arc};
use routerify::{Router, RouterService, RequestInfo};
use log::{info};

async fn node_chain_handler(req: Request<Body>) -> Result<Response<Body>, Infallible> {
    
    info!("Received call to node-chain");
    
    // Extract Propagation Context from incoming request headers
    let parent_cx = global::get_text_map_propagator(|propagator| {
        propagator.extract(&HeaderExtractor(req.headers()))
    });
    // Start Span using the extracted Context
    let parent_span = global::tracer("otel-rust").start_with_context("rust-chain", parent_cx);
    let current_cx = Context::current_with_span(parent_span);

    // Create new Hyper HTTP Client
    let client = Client::new();
    
    // Initialize Tracing
    let child_span = global::tracer("otel-rust").start_with_context("HTTP GET ruby-chain/node-chain", current_cx);
    let cx = Context::current_with_span(child_span);

    // Configure and send HTTP GET 
    let mut req = Request::builder()
        .method(Method::GET)
        .uri("http://host.docker.internal:44000/node-chain");
    
    // Inject Propagation Context into HTTP Headers
    global::get_text_map_propagator(|propagator| {
        propagator.inject_context(&cx, &mut HeaderInjector(&mut req.headers_mut().unwrap()))
    });
    
    // Send request, discard response
    let _res = client.request(req.body(Body::from("Hallo!")).expect("request builder")).await;

    Ok(Response::new("{\"otel\":\"rust\"}".into()))
}

fn init_tracer() -> Result<(sdk::trace::Tracer, opentelemetry_otlp::Uninstall), TraceError> {
    global::set_text_map_propagator(TraceContextPropagator::new());
    
    // Install the OTLP Exporter Pipeline
    opentelemetry_otlp::new_pipeline()
        .with_endpoint("http://host.docker.internal:4317")    
        .with_trace_config(Config {
            resource: Arc::new(sdk::Resource::new(vec![
                semcov::resource::SERVICE_NAME.string("rust-chain"),
                semcov::resource::SERVICE_NAMESPACE.string("kjt-OTel-chain"),
            ])),
            default_sampler: Box::new(Sampler::AlwaysOn),
            ..Default::default()
        })
        .install()
    // Might be able to simplify the Resource attributes like below...
    // .with_trace_config(
    //     trace::config()
    //          ...
    //          .with_resource(Resource::new(vec![KeyValue::new("key", "value")])),

}

// Define an error handler function which will accept the `routerify::Error`
// and the request information and generates an appropriate response.
async fn error_handler(err: routerify::Error, _: RequestInfo) -> Response<Body> {
    eprintln!("{}", err);
    Response::builder()
        .status(StatusCode::INTERNAL_SERVER_ERROR)
        .body(Body::from(format!("Something went wrong: {}", err)))
        .unwrap()
}

fn router() -> Router<Body, Infallible> {
    Router::builder()
        .get("/node-chain", node_chain_handler)
        .err_handler_with_info(error_handler)
        .build()
        .unwrap()
}

#[tokio::main]
async fn main() {

    // Instantiate Tracer
    let _guard = init_tracer();

    // Instantiate Router
    let router = router();

    // Create a Service from the router above to handle incoming requests.
    let service = RouterService::new(router).unwrap();

    // The address on which the server will be listening.
    let addr = SocketAddr::from(([0, 0, 0, 0], 43000));

    // Create a server by passing the created service to `.serve` method.
    let server = Server::bind(&addr).serve(service);    
    //let server = Server::bind(&addr).serve(make_svc);

    info!("Listening on {}", addr);
    if let Err(e) = server.await {
        eprintln!("server error: {}", e);
    }
}