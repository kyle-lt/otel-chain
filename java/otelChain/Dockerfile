FROM openjdk:11.0.2-jre-slim

RUN mkdir /app
WORKDIR /app

ADD target/otelChain-0.0.1-SNAPSHOT.jar .

ENTRYPOINT ["java", "-jar", "otelChain-0.0.1-SNAPSHOT.jar"]