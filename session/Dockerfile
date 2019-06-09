# Start up builder image
FROM golang:alpine as builder
RUN apk update && apk add git

# Build app
ENV GO111MODULE=off
COPY . /go/src/github.com/lnsp/microlog
WORKDIR /go/src/github.com/lnsp/microlog/session
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo -ldflags='-w -s' -o /go/bin/session

# Copy into scratch image
FROM alpine:3.9
COPY --from=builder /go/bin/session /app/session

# Install grpc_health_probe
RUN GRPC_HEALTH_PROBE_VERSION=v0.3.0 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

# Run as non-root
RUN adduser -S microlog
USER microlog

ENTRYPOINT ["/app/session"]
