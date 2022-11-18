# otel-chain
## Overview

The purpose of this application is to create a distributed trace using OpenTelemetry libraries for many different languages.  So, the project essentially "chains" the services together in a single trace.

Currently, the languages are:
- Node.js
- Go
- Python
- Rust
- Java
- Dotnet (core)
- Ruby

The current call flow is

Node.js --> Go --> Python --> Rust --> Java --> Dotnet --> Ruby

The project is configured to use the OpenTelemetry Collector to receive the spans - but the OTel Collector is not part of this repo!

This [local-monitoring-stack](https://github.com/kyle-lt/local-monitoring-stack) repo might help to setup the OTel Collector.

## How to Run

1. Start on OpenTelemetry Collector and whichever desired backend.

2. Start entire system via `docker-compose`
```bash
docker-compose up -d
```

3. Issue request to Node.js entrypoint
```bash
curl localhost:40000/node-start
```

4. Look at your backend to see the traces, or have a look at the OTel Collector's logs
```bash
docker logs -f otel-collector
```

### Node.js

The entry point (root span) of the trace.

Port 40000

More detailed [README](/node/README.md).

### Go



### Python

### Rust

### Java

### Dotnet

### Ruby