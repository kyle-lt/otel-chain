#!/usr/bin/env ruby
# frozen_string_literal: true

# Copyright The OpenTelemetry Authors
#
# SPDX-License-Identifier: Apache-2.0

require 'rubygems'
require 'bundler/setup'
require 'sinatra/base'
# Require otel-ruby
require 'opentelemetry/sdk'
require 'opentelemetry/exporter/otlp'

# Configure the SDK to use OTLP over HTTP
# export OTEL_RESOURCE_ATTRIBUTES=service.namespace=kjt-OTel-chain,service.name=ruby-chain
OpenTelemetry::SDK.configure do |c|
  c.add_span_processor(
    OpenTelemetry::SDK::Trace::Export::BatchSpanProcessor.new(
      OpenTelemetry::Exporter::OTLP::Exporter.new(
        endpoint: 'http://host.docker.internal:55681/v1/traces'
      )
    )
  )
end

# Rack middleware to extract span context, create child span, and add
# attributes/events to the span
class OpenTelemetryMiddleware
  def initialize(app)
    @app = app
    @tracer = OpenTelemetry.tracer_provider.tracer('sinatra', '1.0')
  end

  def call(env)
    # Extract context from request headers
    context = OpenTelemetry.propagation.extract(
      env,
      getter: OpenTelemetry::Context::Propagation.rack_env_getter
    )

    status, headers, response_body = 200, {}, ''

    # Span name SHOULD be set to route:
    span_name = env['PATH_INFO']

    # For attribute naming, see
    # https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/data-semantic-conventions.md#http-server

    # Span kind MUST be `:server` for a HTTP server span
    @tracer.in_span(
      span_name,
      attributes: {
        'component' => 'http',
        'http.method' => env['REQUEST_METHOD'],
        'http.route' => env['PATH_INFO'],
        'http.url' => env['REQUEST_URI'],
      },
      kind: :server,
      with_parent: context
    ) do |span|
      # Run application stack
      status, headers, response_body = @app.call(env)

      span.set_attribute('http.status_code', status)
    end

    [status, headers, response_body]
  end
end

class App < Sinatra::Base
  set :bind, '0.0.0.0'
  use OpenTelemetryMiddleware

  get '/node-chain' do
    'Hello World!'
  end

  run! if app_file == $0
end