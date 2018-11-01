# Start up builder image
FROM golang:alpine as builder
RUN apk update && apk add git

# Build app
ENV GO111MODULE=off
COPY . /go/src/github.com/lnsp/microlog
WORKDIR /go/src/github.com/lnsp/microlog/mail
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo -ldflags='-w -s' -o /go/bin/mail

# Copy into scratch image
FROM scratch
ADD ./mail/templates /app/templates
COPY --from=builder /go/bin/mail /app/mail

ENTRYPOINT ["/app/mail"]
