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
FROM golang:alpine
COPY --from=builder /go/bin/session /app/session
EXPOSE 8080
ENTRYPOINT ["/app/session"]
