package com.ktully.otelChain;

import java.util.concurrent.TimeUnit;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

// 0.13.1
//import io.opentelemetry.exporter.jaeger.JaegerGrpcSpanExporter;
import io.opentelemetry.exporter.logging.LoggingSpanExporter;
import io.opentelemetry.api.trace.propagation.W3CTraceContextPropagator;
import io.opentelemetry.context.propagation.ContextPropagators;
import io.opentelemetry.sdk.OpenTelemetrySdk;
import io.opentelemetry.sdk.resources.Resource;
import io.opentelemetry.sdk.trace.samplers.Sampler;
import io.opentelemetry.semconv.resource.attributes.ResourceAttributes;
import io.opentelemetry.sdk.trace.export.BatchSpanProcessor;
import io.opentelemetry.sdk.trace.export.SimpleSpanProcessor;

// 0.14.1
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.common.Attributes;
import io.opentelemetry.sdk.trace.SdkTracerProvider;

// OTLP Exporter
import io.opentelemetry.exporter.otlp.trace.OtlpGrpcSpanExporter;

@Configuration
public class OtelTracerConfig {

	@Bean
	//public Tracer OtelTracer() throws Exception {
	public static OpenTelemetry OpenTelemetryConfig() {
		
		// ** General Comments **
		// Re-writing a lot of this code for changes put into effect for 0.14.1
		// The original class returned a Tracer, but this time, we return the OpenTelemetrySDK
		
		// ** Create Jaeger Exporter **
		// COMMENTING OUT THE JAEGER EXPORTER, PUTTING IN PLACE THE OTLP COLLECTOR
		//JaegerGrpcSpanExporter jaegerExporter = JaegerGrpcSpanExporter.builder().setServiceName("garagesale-ui")
		//		.setChannel(ManagedChannelBuilder.forAddress("jaeger", 14250).usePlaintext().build()).build();
		
	    // ** Create OTLP gRPC Span Exporter & BatchSpanProcessor **
		OtlpGrpcSpanExporter spanExporter =
	            OtlpGrpcSpanExporter.builder()
	            	.setEndpoint("http://host.docker.internal:4317")
	            	.setTimeout(2, TimeUnit.SECONDS).build();
	        BatchSpanProcessor spanProcessor =
	            BatchSpanProcessor.builder(spanExporter)
	                .setScheduleDelay(100, TimeUnit.MILLISECONDS)
	                .build();	   
	        
		// This was working using OTEL_RESOURCE_ATTRIBUTES env var, but apparently not anymore with 0.15.0
	    Resource serviceNameResource =
	                Resource.create(Attributes.of(ResourceAttributes.SERVICE_NAME, "java-chain",
	                		ResourceAttributes.SERVICE_NAMESPACE, "kjt-OTel-chain"));
		
		// ** Create OpenTelemetry SdkTracerProvider
		// Use OTLP & Logging Exporters
	    // Use Service Name Resource (and attributes) defined above
		// Use AlwaysOn TraceConfig
        SdkTracerProvider sdkTracerProvider =
                SdkTracerProvider.builder()
                    .addSpanProcessor(spanProcessor) // OTLP
                    .addSpanProcessor(SimpleSpanProcessor.create(new LoggingSpanExporter()))
                    .setResource(Resource.getDefault().merge(serviceNameResource))
                    .setSampler(Sampler.alwaysOn())
                    .build();
	        
	    // ** Create OpenTelemetry SDK **
		// Use W3C Trace Context Propagation
        // Use the SdkTracerProvider instantiated above
	    OpenTelemetrySdk openTelemetrySdk =
	            OpenTelemetrySdk.builder()
                	.setTracerProvider(sdkTracerProvider)
	            	.setPropagators(ContextPropagators.create(W3CTraceContextPropagator.getInstance()))
	                .build();
	    			//.buildAndRegisterGlobal();  // can/should I use this?  Maybe later...
	    
	    
	    //  ** Create Shutdown Hook **
	    Runtime.getRuntime().addShutdownHook(new Thread(sdkTracerProvider::shutdown));
	    
	    return openTelemetrySdk;

	}

}
