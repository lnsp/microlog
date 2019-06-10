# Start up builder image
FROM golang:1.12-alpine as builder
RUN apk update && apk add git

# Download grpc_health_probe
RUN GRPC_HEALTH_PROBE_VERSION=v0.3.0 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

# Install dependencies
RUN mkdir -p /src
WORKDIR /src
COPY go.mod /src
COPY go.sum /src
RUN go mod download

# Copy source code and build
COPY . .
WORKDIR /src/mail
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo -ldflags='-w -s' -o /bin/mail

# Copy into scratch image
FROM alpine:3.9
RUN apk update && apk add ca-certificates
WORKDIR /app
COPY --from=builder /src/mail/templates /app/templates
COPY --from=builder /bin/mail /bin/mail
COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe
ENTRYPOINT ["/bin/mail"]
