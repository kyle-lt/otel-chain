use actix_service::Service;
use actix_web::middleware::Logger;
use actix_web::{web, App, HttpServer};
use opentelemetry::trace::TraceError;
use opentelemetry::{global, sdk::trace as sdktrace};
use opentelemetry::{
    trace::{FutureExt, TraceContextExt, Tracer},
    Key,
    KeyValue,
};
use opentelemetry::sdk::{Resource};

fn init_tracer() -> Result<(sdktrace::Tracer, opentelemetry_otlp::Uninstall), TraceError> {
//fn init_tracer() -> Result<(sdktrace::Tracer, opentelemetry_jaeger::Uninstall), TraceError> {
    /*
    // Jaeger
    opentelemetry_jaeger::new_pipeline()
        .with_service_name("trace-http-demo")
        .install()
    */

    // OTLP
    opentelemetry_otlp::new_pipeline()
        .with_endpoint("http://host.docker.internal:4317")    
        //.with_trace_config(Config {
        //    resource: Arc::new(sdk::Resource::new(vec![
        //        semcov::resource::SERVICE_NAME.string("rust-chain"),
        //        semcov::resource::SERVICE_NAMESPACE.string("kjt-OTel-chain"),
        //    ])),
        //    default_sampler: Box::new(Sampler::AlwaysOn),
        //    ..Default::default()
        //})
        .with_trace_config(
            sdktrace::config()
                .with_resource(Resource::new(vec![KeyValue::new("service.name", "rust-chain-actix")])),
        )
        .install()

}

async fn index() -> &'static str {
    let tracer = global::tracer("request");
    tracer.in_span("index", |ctx| {
        ctx.span().set_attribute(Key::new("parameter").i64(10));
        "Index"
    })
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    std::env::set_var("RUST_LOG", "debug");
    env_logger::init();
    let (tracer, _uninstall) = init_tracer().expect("Failed to initialise tracer.");

    HttpServer::new(move || {
        let tracer = tracer.clone();
        App::new()
            .wrap(Logger::default())
            .wrap_fn(move |req, srv| {
                tracer.in_span("middleware", move |cx| {
                    cx.span()
                        .set_attribute(Key::new("path").string(req.path().to_string()));
                    srv.call(req).with_context(cx)
                })
            })
            .route("/", web::get().to(index))
    })
    .bind("127.0.0.1:8088")
    .unwrap()
    .run()
    .await
}