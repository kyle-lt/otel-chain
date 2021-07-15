#!/bin/bash

sudo pip3 install \
#opentelemetry-api==0.17b0 --force-reinstall \
opentelemetry-api==1.3.0 --force-reinstall \
#opentelemetry-sdk==0.17b0 --force-reinstall \
opentelemetry-sdk==1.3.0 --force-reinstall \
#opentelemetry-instrumentation-flask==0.17b0 --force-reinstall \
opentelemetry-instrumentation-flask==0.22b0 --force-reinstall \
#opentelemetry-instrumentation-requests==0.17b0 --force-reinstall \
opentelemetry-instrumentation-requests==0.22b0 --force-reinstall \
#opentelemetry-exporter-otlp==0.17b0 --force-reinstall
opentelemetry-exporter-otlp==0.22b0 --force-reinstall

exit 0