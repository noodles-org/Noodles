FROM golang:1.26.4-alpine3.24
LABEL authors="blporter"

WORKDIR /usr/src/app

COPY go.mod ./
COPY webhook.go .
COPY models.go .
RUN go get ./...
RUN go build -v -o /usr/local/bin/app ./...

CMD ["app"]
