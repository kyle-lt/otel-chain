"use strict";

const { LogLevel } = require("@opentelemetry/core");
const { BasicTracerProvider, SimpleSpanProcessor } = require("@opentelemetry/tracing");
const { NodeTracerProvider } = require('@opentelemetry/node');
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
// Trying out NodeTracerProvider
//const provider = new BasicTracerProvider();
const provider = new NodeTracerProvider();
const exporter = new CollectorTraceExporter(collectorOptions);
provider.addSpanProcessor(new SimpleSpanProcessor(exporter));
provider.register();

// register and load old default plugins (?)
// Disabling express plugin to test - the result is much cleaner, prob won't turn it back on
registerInstrumentations({
  tracerProvider: provider,
  instrumentations: [
    {
      plugins: {
        express: {
          enabled: false,
        },
        http: {
          requestHook: (span, request) => {
            //span.setAttribute("custom request hook attribute", "request");
            // Name the Span using the URL path, if possible
            if (!typeof request.path === 'undefined') {
              span.updateName(request.method + " " + request.path);
            }
          },
        }
      },
    }
  ]
});

console.log("Node.js Tracing Initialized!");