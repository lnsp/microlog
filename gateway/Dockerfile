# Start up builder image
FROM golang:1.12-alpine as builder
RUN apk update && apk add git

# Install dependencies
RUN mkdir -p /src
WORKDIR /src
COPY go.mod /src
COPY go.sum /src
RUN go mod download

# Copy source code and build
COPY . .
WORKDIR /src/gateway
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo -ldflags='-w -s' -o /bin/gateway

# Copy into scratch image
FROM scratch
WORKDIR /app
COPY --from=builder /src/CHANGELOG.md /app/CHANGELOG.md
COPY --from=builder /src/gateway/web /app/web
COPY --from=builder /bin/gateway /gateway
ENTRYPOINT ["/gateway"]
