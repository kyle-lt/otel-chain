"use strict";

const opentelemetry = require('@opentelemetry/api');

// Not functionally required but gives some insight what happens behind the scenes
const { diag, DiagConsoleLogger, DiagLogLevel } = opentelemetry;
diag.setLogger(new DiagConsoleLogger(), DiagLogLevel.INFO);

const { SimpleSpanProcessor } = require("@opentelemetry/tracing");
const { NodeTracerProvider } = require('@opentelemetry/node');
const { registerInstrumentations } = require('@opentelemetry/instrumentation');
const { CollectorTraceExporter } =  require('@opentelemetry/exporter-collector-grpc');
const { HttpInstrumentation } = require('@opentelemetry/instrumentation-http');
const { Resource } = require('@opentelemetry/resources');
const { ResourceAttributes } = require('@opentelemetry/semantic-conventions');

// Originally, adding serviceName and service.namespace worked here,
// But it doesn't seem to work anymore.  Instead, added via Resource (below)
const collectorOptions = {
  serviceName: 'node-chain',
  url: 'http://host.docker.internal:4317',
  attributes: {
    'service.namespace' : 'kjt-OTel-chain'
  }
};

// Instantiate Tracer and Exporter
const provider = new NodeTracerProvider({
  resource: new Resource({
    [ResourceAttributes.SERVICE_NAME]: 'node-chain',
    [ResourceAttributes.SERVICE_NAMESPACE]: 'kjt-OTel-chain',
  }),
});
const exporter = new CollectorTraceExporter(collectorOptions);
provider.addSpanProcessor(new SimpleSpanProcessor(exporter));
provider.register();

registerInstrumentations({
  tracerProvider: provider,
  instrumentations: [
    //HttpInstrumentation: {
    new HttpInstrumentation({
      requestHook: (span, request) => {
        const spanAttrs = span.attributes;
        var spanName = "";
        for (const property in spanAttrs) {
          //console.log(`${property}: ${spanAttrs[property]}`);
          if (property === "http.method") {
            //console.log("Found HTTP method! It equals " + spanAttrs[property]);
            spanName += spanAttrs[property] + " ";
          } else if (property === "http.route") {
            //console.log("Found HTTP route! It equals " + spanAttrs[property]);
            spanName += spanAttrs[property];
          }
        }
        console.log(`Creating Span with derived name: \"${spanName}\"`);
        span.updateName(spanName);
      },
    }),
  ],
});

/*
          if (property === "http.method") {

            method = spanAttrs[property]
          } else if (property === "http.route") {
            url = spanAttrs[property]
          }
*/

// register and load old default plugins (?)
// Disabling express plugin to test - the result is much cleaner, prob won't turn it back on
/*
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
            } else {
              // hard-code the span name
              span.updateName("GET /node-start");
            }
          },
        }
      },
    }
  ]
});
*/

console.log("Node.js Tracing Initialized!");