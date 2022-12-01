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
require 'opentelemetry/instrumentation/all'

# Export traces to console by default
ENV['OTEL_TRACES_EXPORTER'] ||= 'console'
# export OTEL_RESOURCE_ATTRIBUTES=service.namespace=kjt-OTel-chain,service.name=ruby-chain

# Configure the SDK to use OTLP exporter over HTTP via BatchSpanProcessor, and
# Console exporter via SimpleSpanProcessor
OpenTelemetry::SDK.configure do |c|
  #c.service_name = 'something_calculated_dynamically'
  c.add_span_processor(
    OpenTelemetry::SDK::Trace::Export::BatchSpanProcessor.new(
      OpenTelemetry::Exporter::OTLP::Exporter.new(
        endpoint: 'http://otel-collector:4318/v1/traces'
      )
    )
  )
  c.add_span_processor(
    OpenTelemetry::SDK::Trace::Export::SimpleSpanProcessor.new(
      OpenTelemetry::SDK::Trace::Export::ConsoleSpanExporter.new
    )
  )
  c.use_all() # enables all instrumentation!
end

##### BEGIN RACK MIDDLEWARE FOR MANUAL OTel Instrumentation #####

## Note, the middleware can be enabled/disabled in the "App" Class, at the bottom!

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
      #getter: OpenTelemetry::Context::Propagation.rack_env_getter
      getter: OpenTelemetry::Common::Propagation.rack_env_getter
    )

    status, headers, response_body = 200, {}, ['']

    # Span name SHOULD be set to route:
    span_name = env['PATH_INFO']

    # For attribute naming, see
    # https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/data-semantic-conventions.md#http-server

    ##### BEGIN OLD MANUAL INSTRUMENTATION CODE #####
    # Span kind MUST be `:server` for a HTTP server span
    # @tracer.start_span(
    #   span_name,
    #   attributes: {
    #     'component' => 'http',
    #     'http.method' => env['REQUEST_METHOD'],
    #     'http.route' => env['PATH_INFO'],
    #     'http.url' => env['REQUEST_URI'],
    #     'telemetry.sdk.language' => 'ruby',
    #   },
    #   kind: :server,
    #   with_parent: context
    # ) do |span|
    #   # Run application stack
    #   status, headers, response_body = @app.call(env)

    #   span.set_attribute('http.status_code', status)
    # end
    ##### END OLD MANUAL INSTRUMENTATION CODE #####

    #### BEGIN TESTING ######
    # create a span
    # @tracer.in_span('foo') do |span|
    #   # set an attribute
    #   span.set_attribute('platform', 'osx')
    #   # add an event
    #   span.add_event('event in bar')
    #   # create bar as child of foo
    #   @tracer.in_span('bar') do |child_span|
    #     # inspect the span
    #     p 'child_span created!'
    #     #pp child_span
    #   end
    # end
    #### END TESTING ######

    # Activate the extracted context
    OpenTelemetry::Context.with_current(context) do
      # Span kind MUST be `:server` for a HTTP server span
      @tracer.in_span(
        span_name,
        attributes: {
          'component' => 'http',
          'http.method' => env['REQUEST_METHOD'],
          'http.route' => env['PATH_INFO'],
          'http.url' => env['REQUEST_URI'],
        },
        kind: :server
      ) do |span|
        # Run application stack
        status, headers, response_body = @app.call(env)

        span.set_attribute('http.status_code', status)
      end
    end

    [status, headers, response_body]
  end
end

##### END RACK MIDDLEWARE FOR MANUAL OTel Instrumentation #####

class App < Sinatra::Base
  set :bind, '0.0.0.0'
  
  # Enable/Disable manual instrumentation by uncommenting/commenting the next line!
  #use OpenTelemetryMiddleware

  get '/node-chain' do
    'ruby-chain-works!'
  end

  run! if app_file == $0
end
