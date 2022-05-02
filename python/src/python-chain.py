import flask
import requests
import logging

from opentelemetry import trace
from opentelemetry.instrumentation.flask import FlaskInstrumentor
from opentelemetry.instrumentation.requests import RequestsInstrumentor
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import (
    ConsoleSpanExporter,
    SimpleSpanProcessor,
    BatchSpanProcessor
)
#from opentelemetry.exporter.otlp.trace_exporter import OTLPSpanExporter
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource

# Logging
logger = logging.getLogger('python-chain')
logger.setLevel(logging.INFO)
ch = logging.StreamHandler()
logger.addHandler(ch)

# Resource
# Not sure if namespace is working, but using env var, it did:
# export OTEL_RESOURCE_ATTRIBUTES=service.namespace=kjt-OTel-chain OR (?)
# export OTEL_RESOURCE_ATTRIBUTES=service.namespace:kjt-OTel-chain
resource = Resource(attributes={
    "service.name": "python-chain",
    "service.namespace": "kjt-OTel-chain",
    "telemetry.sdk.language": "python"
})

# Create OTLP Span Exporter
otlp_span_exporter = OTLPSpanExporter(endpoint="host.docker.internal:4317", insecure=True)

# Instantiate Tracer Provider, use resource defined above
trace.set_tracer_provider(TracerProvider(resource=resource))
# Add Console Exporter
trace.get_tracer_provider().add_span_processor(
    SimpleSpanProcessor(ConsoleSpanExporter())
)
# Add OTLP Exporter
trace.get_tracer_provider().add_span_processor(
    BatchSpanProcessor(otlp_span_exporter)
)

app = flask.Flask(__name__)
FlaskInstrumentor().instrument_app(app)
RequestsInstrumentor().instrument()

@app.route("/node-chain")
def hello():
    tracer = trace.get_tracer(__name__)
    with tracer.start_as_current_span("HTTP GET rust-chain/node-chain"):
        logger.info("made downstream request!")
        requests.get("http://host.docker.internal:43000/node-chain") # TODO: downstream call to rust will go here!
    return "{\"otel\":\"python\"}"

app.run(host="0.0.0.0", port=42000, debug=True)
