FROM golang:1.24.5-alpine3.22
LABEL authors="blporter"

WORKDIR /usr/src/app

COPY go.mod ./
RUN go mod download
COPY webhook.go .
RUN go build -v -o /usr/local/bin/app ./...

CMD ["app"]
