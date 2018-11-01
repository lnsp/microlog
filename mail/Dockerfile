FROM golang:latest
LABEL maintainer="lennart@espe.tech"
RUN mkdir -p /app
ADD . /app
WORKDIR /app
RUN go build -o mail ./
CMD [ "/app/mail" ]
