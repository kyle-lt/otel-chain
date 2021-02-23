"use strict";

const { LogLevel } = require("@opentelemetry/core");
const { BasicTracerProvider, SimpleSpanProcessor } = require("@opentelemetry/tracing");
const { registerInstrumentations } = require('@opentelemetry/instrumentation');
const { CollectorTraceExporter } =  require('@opentelemetry/exporter-collector-grpc');

const collectorOptions = {
  serviceName: 'node-chain',
  url: 'host.docker.internal:4317',
  attributes: {
    'service.namespace' : 'kjt-OTel-chain'
  }
};

// Instantiate Tracer and Exporter
const provider = new BasicTracerProvider();
const exporter = new CollectorTraceExporter(collectorOptions);
provider.addSpanProcessor(new SimpleSpanProcessor(exporter));
provider.register();

// load old default plugins (?)
registerInstrumentations({
  tracerProvider: provider,
});

console.log("Node.js Tracing Initialized!");