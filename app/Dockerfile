FROM node:20-alpine AS frontend
WORKDIR /build
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM node:20-alpine AS backend
WORKDIR /build
COPY backend/package*.json ./
RUN npm ci
COPY backend/ ./
RUN npm run build

FROM node:20-alpine
WORKDIR /app

COPY --from=backend  /build/dist         ./dist
COPY --from=backend  /build/node_modules ./node_modules
COPY --from=backend  /build/package.json ./
COPY --from=frontend /build/dist         ./public
COPY docs/   ./docs/
COPY config/ ./config/

ENV NODE_ENV=production \
    FRONTEND_PATH=./public \
    DOCS_PATH=./docs \
    CONFIG_PATH=./config

EXPOSE 3000 9090
USER node
CMD ["node", "dist/server.js"]