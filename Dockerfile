FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY backend/go.mod backend/go.sum ./backend/
WORKDIR /src/backend
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/primordia-engine ./cmd/primordia

FROM node:22-alpine AS frontend-builder

WORKDIR /src/frontend

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

FROM alpine:3.21

RUN adduser -D -u 10001 primordia

WORKDIR /app
COPY --from=builder /out/primordia-engine /app/primordia-engine
COPY --from=frontend-builder /src/frontend/dist /app/web

USER primordia
EXPOSE 8080

ENTRYPOINT ["/app/primordia-engine"]