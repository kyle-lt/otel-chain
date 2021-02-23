# otel-chain
## Overview

The purpose of this application is to create a distributed trace using OpenTelemetry libraries for many different languages.  Currently, the languages are:
- Node.js
- Go
- Python
- Rust
- Java
- Dotnet (core)
- Ruby

The current call flow is

Node.js --> Go --> Python --> Rust --> Java --> Dotnet //TODO --> Ruby

The project uses the OpenTelemetry Collector and Jaeger to collect the spans.

// TODO - the rest of the docs.