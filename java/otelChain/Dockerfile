FROM maven:3-openjdk-18 as builder

RUN mkdir /build
WORKDIR /build

ADD src src/
ADD pom.xml .
RUN mvn clean package

FROM openjdk:18.0-slim

RUN mkdir /app
WORKDIR /app

COPY --from=builder /build/target/otelChain-0.0.1-SNAPSHOT.jar .

ENTRYPOINT ["java", "-jar", "otelChain-0.0.1-SNAPSHOT.jar"]