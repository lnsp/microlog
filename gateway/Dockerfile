# Start up builder image
FROM golang:alpine as builder
RUN apk update && apk add git

# Build app
ENV GO111MODULE=off
COPY . /go/src/github.com/lnsp/microlog
WORKDIR /go/src/github.com/lnsp/microlog/gateway
RUN go get -d -v
RUN GO111MODULE=off CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo -ldflags='-w -s' -o /go/bin/gateway

# Copy into scratch image
FROM scratch
ADD ./gateway/web /app/web
COPY --from=builder /go/bin/gateway /app/gateway

ENTRYPOINT ["/app/gateway"]
