FROM golang:1.24.5-alpine3.22

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY update-ip.go .
RUN go build -v -o /usr/local/bin/app ./...

CMD ["app"]
