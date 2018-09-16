FROM golang:latest
LABEL maintainer="lennart@espe.tech"
RUN mkdir -p /app
ADD . /app
WORKDIR /app
RUN go build -o microlog ./cmd/microlog
CMD [ "/app/microlog" ]