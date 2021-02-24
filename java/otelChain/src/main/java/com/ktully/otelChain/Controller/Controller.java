package com.ktully.otelChain.Controller;

import java.util.Map;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestHeader;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.client.RestTemplate;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpMethod;
import org.springframework.http.ResponseEntity;

import com.ktully.otelChain.OtelTracerConfig;

import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.trace.Span;
import io.opentelemetry.api.trace.SpanKind;
import io.opentelemetry.api.trace.Tracer;
import io.opentelemetry.context.Context;
import io.opentelemetry.context.Scope;
import io.opentelemetry.context.propagation.TextMapGetter;
import io.opentelemetry.context.propagation.TextMapPropagator;
import io.opentelemetry.context.propagation.TextMapSetter;

@RestController
public class Controller {
	
	// Logger
	private static final Logger logger = LoggerFactory.getLogger(Controller.class);
	
	// OTel
	private static final OpenTelemetry openTelemetry = OtelTracerConfig.OpenTelemetryConfig();
	private static final Tracer tracer =
		      openTelemetry.getTracer("com.ktully.otelChain");
	
	/*
	 * Configuration for Context Propagation to be done via @RequestHeader
	 * extraction
	 */
	TextMapGetter<Map<String, String>> getter = new TextMapGetter<Map<String, String>>() {
		@Override
		public String get(Map<String, String> carrier, String key) {
			logger.debug("Key = " + key);
			logger.debug("Key found! " + key);
			logger.debug("Value = " + carrier.get(key));
			return carrier.get(key);
		}
		// 0.10.0 - didn't need this implementation for 0.8.0
		@Override
		public Iterable<String> keys(Map<String, String> carrier) {
			return carrier.keySet();
		}
	};

	/*
	 * Configuration for Context Propagation to be done via HttpHeaders injection
	 */
	private static final TextMapSetter<HttpHeaders> httpHeadersSetter = new TextMapSetter<HttpHeaders>() {
		@Override
		public void set(HttpHeaders carrier, String key, String value) {
			logger.debug("RestTemplate - Adding Header with Key = " + key);
			logger.debug("RestTemplate - Adding Header with Value = " + value);
			carrier.set(key, value);
		}
	};
	
	@GetMapping("/node-chain")
	private String distribute(@RequestHeader Map<String, String> headers) {
		
		Context extractedContext = null;
		try {
			logger.debug("Trying to extact Context Propagation Headers from ui to item-api/distribute.");
			extractedContext = openTelemetry.getPropagators().getTextMapPropagator()
					.extract(Context.current(), headers, getter);
			
			logger.debug(extractedContext.toString());
		} catch (Exception e) {
			logger.error("Exception caught while extracting Context Propagators", e);
		}
		
		Span serverSpan = tracer.spanBuilder("HTTP GET /node-chain").setParent(extractedContext).setSpanKind(SpanKind.SERVER).startSpan();
		try (Scope scope = serverSpan.makeCurrent()) {
			logger.debug("Trying to build Span and then make RestTemplate call downstream");
			// Add the attributes defined in the Semantic Conventions
			serverSpan.setAttribute("http.method", "GET");
			serverSpan.setAttribute("http.scheme", "http");
			serverSpan.setAttribute("http.host", "java-chain");
			serverSpan.setAttribute("http.target", "/node-chain");

			RestTemplate restTemplate = new RestTemplate();
			HttpHeaders propagationHeaders = new HttpHeaders();

			Span restTemplateSpan = tracer.spanBuilder("HTTP GET dotnet-chain/node-chain").setSpanKind(SpanKind.CLIENT).startSpan();
			try (Scope outgoingScope = restTemplateSpan.makeCurrent()) {
				// Add some important info to our Span
				restTemplateSpan.addEvent("Calling dotnet-chain/node-chain via RestTemplate"); // This ends up in "logs"
																							// section in
				// Add the attributes defined in the Semantic Conventions
				restTemplateSpan.setAttribute("http.method", "GET");
				restTemplateSpan.setAttribute("http.scheme", "http");
				restTemplateSpan.setAttribute("http.host", "dotnet-chain");
				restTemplateSpan.setAttribute("http.target", "/node-chain");

				// 0.14.1
				openTelemetry.getPropagators().getTextMapPropagator().inject(Context.current(), propagationHeaders, httpHeadersSetter);
				
				logger.debug("Injecting headers for call from java-chain to downstream API.");
				logger.debug("**** Here are the headers: " + headers.toString());
				HttpEntity<String> entity = new HttpEntity<String>("parameters", propagationHeaders);

				// Make outgoing call via RestTemplate
				ResponseEntity<String> response = restTemplate.exchange("http://host.docker.internal:45000/node-chain",
						HttpMethod.GET, entity, String.class);

				String responseString = response.getBody();
				logger.debug("Response from downstream: ");
				logger.debug(responseString);
			} catch (Exception e) {
				restTemplateSpan.addEvent("error");
				restTemplateSpan.addEvent(e.toString());
				restTemplateSpan.setAttribute("error", true);
				logger.error("Error during OT section, here it is!", e);
				return "error";
			} finally {
				restTemplateSpan.end();
			}

			return "{\"otel\":\"java\"}";
		} catch (Exception e) {
			logger.error("Exception caught attempting to create Span", e);
			return "Didn't work bro!";
		} finally {
			if (serverSpan != null) {
				serverSpan.end();
			}
		}	
	}
}
