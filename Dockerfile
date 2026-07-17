FROM node:20-alpine AS frontend
WORKDIR /build
COPY app/frontend/package*.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci --prefer-offline
COPY app/frontend/ ./
RUN npm run build

FROM node:20-alpine AS backend
WORKDIR /build
COPY app/backend/package*.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci --prefer-offline
COPY app/backend/ ./
RUN npm run build && npm prune --omit=dev

FROM node:20-alpine
WORKDIR /app

COPY --from=backend  /build/dist         ./dist
COPY --from=backend  /build/node_modules ./node_modules
COPY --from=backend  /build/package.json ./
COPY --from=frontend /build/dist         ./public
COPY docs   ./docs/

ENV NODE_ENV=production \
    FRONTEND_PATH=./public \
    DOCS_PATH=./docs

EXPOSE 3000 9090
USER node
CMD ["node", "dist/server.js"]