# ruby-chain
## Overview

This service is a simple Ruby server (using `Sinatra`) that receives an HTTP GET request and sends a `String` response.

## Instrumentation

The OTel SDK is [instantiated](ruby-chain.rb#L22) to use 2 `SpanProcessors`:
- `BatchSpanProcessor` for `OTLP exporter`
- `SimpleSpanProcessor` for `console exporter`

The OTel [instrumentation](ruby-chain.rb#L36) is done using the `opentelemetry-instrumentation-all` [RubyGem](https://rubygems.org/search?query=opentelemetry-instrumentation-all).

Originally, [manual instrumentation](ruby-chain.rb#L45) was used via `Rack Middleware`:
- Incoming context was [extracted manually](ruby-chain.rb#L55)
- Spans were [created manually](ruby-chain.rb#L107)

## Build & Run

If desired, this service can be built and run separately.

0. When running the first time, create docker network `monitoring`
```bash
docker create network monitoring
```

1. Start OpenTelemetry Collector (and, optionally, whichever desired backend).

2. Build the Docker image via `docker-compose`
```bash
docker-compose build ruby-chain
```

3. Run the service via `docker-compose`
```bash
docker-compose up ruby-chain -d
```

4. Send requrest to the service
```bash
curl localhost:47000/node-chain
```

5. Check the `ruby-chain` logs
```bash
docker logs ruby-chain
```

and look for lines that start with something like:

```bash
#<struct OpenTelemetry::SDK::Trace::SpanData
```

