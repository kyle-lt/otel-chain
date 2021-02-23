#!/bin/bash

export OTEL_RESOURCE_ATTRIBUTES=service.name=java-chain,service.namespace=kjt-OTel-chain

java -jar target/otelChain-0.0.1-SNAPSHOT.jar

exit 0
