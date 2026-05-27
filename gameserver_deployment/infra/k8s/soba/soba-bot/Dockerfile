FROM golang:1.21-alpine as builder
LABEL authors="blporter"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /bot


FROM alpine
ENV TOKEN=error
ENV GUILD_ID=error
ENV GAMES="game1=0000,game2=1111"
COPY --from=builder /bot /soba-bot
CMD ["/soba-bot"]