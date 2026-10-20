FROM node:24-alpine AS frontend
WORKDIR /app/frontend
RUN npm install --global pnpm@11.25.0
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm test && pnpm build

FROM golang:1.26-alpine AS backend
WORKDIR /app/backend
COPY backend/ ./
RUN go test ./...
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /server ./cmd/server

FROM scratch
COPY --from=backend /server /server
COPY --from=frontend /app/frontend/dist /web
ENV ADDR=:8080 STATIC_DIR=/web
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/server"]
