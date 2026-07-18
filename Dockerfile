FROM --platform=$BUILDPLATFORM node:20-alpine AS frontend
WORKDIR /build
COPY app/frontend/package*.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci --prefer-offline
COPY app/frontend/ ./
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS backend
ARG TARGETOS TARGETARCH
WORKDIR /build
COPY app/backend/go.mod app/backend/go.sum ./
RUN go mod download
COPY app/backend/ ./
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o server ./cmd/server

FROM alpine:3.21
WORKDIR /app

COPY --from=backend  /build/server       ./server
COPY --from=frontend /build/dist         ./public
COPY docs   ./docs/

ENV NODE_ENV=production \
    FRONTEND_PATH=./public \
    DOCS_PATH=./docs

EXPOSE 3000 9090
USER nobody
CMD ["./server"]