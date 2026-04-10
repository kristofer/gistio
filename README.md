# GoRenderMand (a port of Gist.io)

A Go application that renders public Markdown files beautifully in the browser.
Markdown rendering runs entirely in a **Go WebAssembly (WASM)** module — no server-side rendering required.

## How it works

1. The browser loads a tiny HTML shell (`static/index.html`).
2. A Go WASM module (`static/main.wasm`) is loaded in the browser; it exposes `window.renderMarkdown(text) → html` using [goldmark](https://github.com/yuin/goldmark) with paraiso-dark syntax highlighting.
3. The Go HTTP server provides a `/proxy?url=<encoded-url>` endpoint that fetches any public HTTPS markdown file, bypassing browser CORS restrictions.
4. The JavaScript SPA shell fetches the markdown via the proxy and calls the WASM function to render it.

## Usage

### Render any public Markdown URL

Visit `/?url=https://raw.githubusercontent.com/user/repo/main/README.md`

Or paste the URL into the input form on the home page.

### GitHub Gist shortcut (backward-compatible)

Replace `gist.github.com/youruser/abc123` with `gist.io/@youruser/abc123`.

## Building

Requires [Go 1.21+](https://golang.org/dl/).

```sh
# Build the WASM module and the server binary
make

# Start the server (default port 8080, override with PORT env var)
./gistio
```

`make` does the following in order:

1. `GOARCH=wasm GOOS=js go build -o static/main.wasm ./wasm/` — compiles the Go WASM module.
2. Copies `wasm_exec.js` from your Go installation into `static/`.
3. `go build -o gistio .` — compiles the HTTP server.

### Individual targets

| Command | Description |
|---|---|
| `make wasm` | Build only the WASM module |
| `make server` | Build only the server binary (requires WASM built first) |
| `make run` | Build everything and start the server |
| `make clean` | Remove build artifacts |
| `make docker-build` | Build the Docker image (tagged `gistio`) |
| `make docker-run` | Run the Docker image on port 8080 |

## Docker

The repository ships with a multi-stage `Dockerfile` that handles the full
build — you don't need Go installed locally.

### Build and run with Docker

```sh
# Build the image (compiles WASM + server inside the container)
docker build -t gistio .

# Run on port 8080
docker run --rm -p 8080:8080 gistio

# Use a different host port (e.g. 3000 → 8080 inside container)
docker run --rm -p 3000:8080 gistio
```

Open <http://localhost:8080> in your browser.

### Build and run with Docker Compose

```sh
docker compose up --build
```

To run on a different host port set the `PORT` variable:

```sh
PORT=3000 docker compose up --build
```

### Environment variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | TCP port the server listens on inside the container |

## Project layout

```
wasm/main.go          Go WASM module — markdown renderer exported to JS
main.go               Go HTTP server — static files, font embedding, /proxy endpoint
static/index.html     SPA shell — routing, WASM loading, URL input form
static/style.css      Base styles
static/fonts.css      Custom font declarations
static/main.wasm      (build artifact) compiled Go WASM module
static/wasm_exec.js   (build artifact) Go WASM runtime helper
src/fonts/            Embedded Elena web fonts
Makefile              Build orchestration
```
