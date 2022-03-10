#!/bin/bash

sudo pip3 install \

# OTel API
#opentelemetry-api==0.17b0 --force-reinstall \
#opentelemetry-api==1.3.0 --force-reinstall \
opentelemetry-api==1.9.1 --force-reinstall \

# OTel SDK
#opentelemetry-sdk==0.17b0 --force-reinstall \
#opentelemetry-sdk==1.3.0 --force-reinstall \
opentelemetry-sdk==1.9.1 --force-reinstall \

# OTel Flask Auto-Instrumentation
#opentelemetry-instrumentation-flask==0.17b0 --force-reinstall \
#opentelemetry-instrumentation-flask==0.22b0 --force-reinstall \
opentelemetry-instrumentation-flask==0.28b1 --force-reinstall \

# OTel Requests Auto-Instrumentation
#opentelemetry-instrumentation-requests==0.17b0 --force-reinstall \
#opentelemetry-instrumentation-requests==0.22b0 --force-reinstall \
opentelemetry-instrumentation-requests==0.28b1 --force-reinstall \

# OTel OTLP Exporter
#opentelemetry-exporter-otlp==0.17b0 --force-reinstall
#opentelemetry-exporter-otlp==0.22b0 --force-reinstall
opentelemetry-exporter-otlp==1.9.1 --force-reinstall

exit 0