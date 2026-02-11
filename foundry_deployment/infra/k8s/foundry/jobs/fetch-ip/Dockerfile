FROM golang:1.24.5-alpine3.22
LABEL authors="blporter"

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download
COPY update-ip.go .
RUN go build -v -o /usr/local/bin/app ./...

CMD ["app"]
