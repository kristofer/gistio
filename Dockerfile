# ── Build stage ──────────────────────────────────────────────────────────────
# Uses the full Go toolchain to compile the WebAssembly module and the server.
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Download dependencies first for better layer caching.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source tree.
COPY . .

# 1. Compile the Go WebAssembly module (runs in the browser).
# 2. Copy the matching wasm_exec.js runtime helper from the Go installation.
# 3. Compile the HTTP server binary as a fully-static Linux binary.
RUN GOARCH=wasm GOOS=js go build -o static/main.wasm ./wasm/ && \
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" static/wasm_exec.js && \
    CGO_ENABLED=0 GOOS=linux go build -o gistio .

# ── Runtime stage ─────────────────────────────────────────────────────────────
# distroless/static includes CA certificates (required for the HTTPS proxy)
# but nothing else — no shell, no package manager.
FROM gcr.io/distroless/static-debian12

WORKDIR /app

# Copy the server binary and the pre-built static assets.
COPY --from=builder /app/gistio   ./gistio
COPY --from=builder /app/static   ./static

EXPOSE 8080

ENV PORT=8080

ENTRYPOINT ["/app/gistio"]
