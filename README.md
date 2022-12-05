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

   > __Note:__  This project is configured to use the OpenTelemetry Collector to receive the spans - but the OTel Collector is not part of this repo!
   - This [local-monitoring-stack](https://github.com/kyle-lt/local-monitoring-stack) repo might help you to setup the OTel Collector, or check out the OTel Collector [Getting Started](https://opentelemetry.io/docs/collector/getting-started/) docs.

## How to Run

0. When running the first time, create docker network `monitoring`
```bash
docker create network monitoring
```

1. Start OpenTelemetry Collector (and, optionally, whichever desired backend).

   > __Note:__  As mentioned above, an OTel Collector is **required** but not included in this repo.

2. Build the services via `docker-compose`
```bash
docker-compose build
```

3. Start entire system via `docker-compose`
```bash
docker-compose up -d
```

4. Issue request to Node.js entrypoint
```bash
curl localhost:40000/node-start
```

5. Look at your backend to see the traces, or have a look at the OTel Collector's logs
```bash
docker logs -f otel-collector
```

## Individual Service Descriptions

### Node.js

The entry point service (root span) of the trace.

| Property | Value |
| ------ | ------- |
| Server | `Express` |
| Instrumentation | Manual |
| Default Port | `40000` | 

More detailed [README](/node/README.md).

### Go

The second service of the trace.

| Property | Value |
| ------ | ------- |
| Server | `net/http` |
| Instrumentation | Manual |
| Default Port | `41000` |

More detailed README.

### Python

The third service of the trace.

| Property | Value |
| ------ | ------- |
| Server | `flask` |
| Instrumentation | `flask`, `requests` Auto-Instrumentation |
| Default Port | `42000`|

More detailed README.

### Rust

The fourth service of the trace.

| Property | Value |
| ------ | ------- |
| Server | `hyper` |
| Instrumentation | Manual |
| Default Port | `43000` |

More detailed README.

### Java

The fifth service of the trace.

| Property | Value |
| ------ | ------- |
| Server | `spring-boot` |
| Instrumentation | Manual |
| Default Port | `44000` |

More detailed README.

### Dotnet

The sixth service of the trace.

| Property | Value |
| ------ | ------- |
| Server | `AspNetCore.Hosting` |
| Instrumentation | `OpenTelemetry.Extensions.Hosting`, `OpenTelemetry.Instrumentation.AspNetCore`, `OpenTelemetry.Instrumentation.Http` Auto-Instrumentation |
| Default Port | `45000` |

More detailed README.

### Ruby

The final service in the trace.

| Property | Value |
| ------ | ------- |
| Server | `Sinatra` |
| Instrumentation | `all` Auto-Instrumentation |
| Default Port | `47000` |


More detailed [README](/ruby/README.md)

## TODO

- rebuild java Dockerfile with multi-stage build
- load generator
- finish individual service `READMEs`