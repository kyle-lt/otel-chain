FROM golang:1.16-alpine AS base
COPY . /src/
WORKDIR /src

FROM base AS example-http-client
RUN go install ./go-chain.go

CMD ["/go/bin/go-chain"]
