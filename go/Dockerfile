FROM golang:1.23-alpine AS base
COPY . /src/
WORKDIR /src

FROM base AS example-http-client
RUN go install ./go-chain.go

CMD ["/go/bin/go-chain"]
