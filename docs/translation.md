# Gist.io — Go Translation

## Overview

This document describes the translation of Gist.io from a React/TypeScript
single-page application (SPA) to a server-side rendered Go web application.
The Go version produces identical HTML output to the original, is deployable as
a single self-contained binary, and has no CGO dependencies.

---

## Original Architecture (React/TypeScript)

| Concern | Implementation |
|---------|---------------|
| Routing | `react-router-dom` (client-side) |
| Markdown parsing | `marked` |
| Syntax highlighting | `highlight.js` (Paraíso Dark theme) |
| Rendering | React components, browser-side |
| Deployment | Static build (`react-scripts build`) uploaded to Vercel / Now |

**Flow:**
1. Browser fetches an empty HTML shell.
2. React bootstraps and evaluates the route.
3. For `/@user/id`, `useFetch` fetches the raw gist from GitHub, then
   `<Markdown>` converts it to HTML in the browser.

---

## Go Architecture

| Concern | Implementation |
|---------|---------------|
| Routing | `net/http` standard library mux |
| Markdown parsing | [`goldmark`](https://github.com/yuin/goldmark) |
| Syntax highlighting | [`goldmark-highlighting/v2`](https://github.com/yuin/goldmark-highlighting) + Chroma (Paraíso Dark theme) |
| Rendering | `html/template` server-side |
| Deployment | Single binary; all assets embedded via `go:embed` |

**Flow:**
1. Browser requests `/@user/id`.
2. Go server fetches the raw gist from `gist.githubusercontent.com`.
3. `goldmark` converts Markdown → HTML with inline syntax-highlighted code blocks.
4. The server renders the `gist.html` template and sends a complete page.

---

## File Map

```
main.go              — HTTP server, handlers, Markdown renderer
go.mod / go.sum      — module dependencies
templates/
  home.html          — home page (static, full HTML)
  gist.html          — gist page (uses .User, .ID, .Content)
static/
  style.css          — body/heading/code typography (from src/index.css)
  fonts.css          — @font-face declarations pointing to /fonts/…
src/fonts/           — ElenaWebBasic font files (embedded at /fonts/)
docs/
  translation.md     — this file
```

---

## Dependencies

| Package | Purpose | CGO? |
|---------|---------|------|
| `github.com/yuin/goldmark` | CommonMark-compliant Markdown → HTML | No |
| `github.com/yuin/goldmark-highlighting/v2` | Syntax-highlighted fenced code blocks | No |
| `github.com/alecthomas/chroma/v2` *(transitive)* | Language tokeniser / theme engine | No |

All external assets (Tachyons CSS, Google Fonts) are still loaded from CDN,
matching the original behaviour.

---

## Notable Differences from the React Version

| Feature | React | Go |
|---------|-------|-----|
| Rendering location | Browser | Server |
| Loading state | "Loading…" spinner | None (blocks until gist is fetched) |
| Page title (`<title>`) | `react-helmet` | Go template |
| Code highlighting | `highlight.js` CSS classes | Chroma inline styles |
| JavaScript required | Yes | No |
| Deployable artefact | Static file bundle | Single binary (~10 MB) |

The **Paraíso Dark** colour scheme is preserved — Chroma ships the same theme
under the name `paraiso-dark`.

The "Loading…" placeholder is intentionally removed: the server fetches the
gist before responding, so the browser always receives a fully-rendered page.

---

## Building

Requires Go 1.21 or later.

```bash
# Fetch dependencies
go mod download

# Compile
go build -o gistio .

# Run (defaults to port 8080)
./gistio

# Override port
PORT=3000 ./gistio
```

---

## Deployment

### Docker

```dockerfile
FROM golang:1.22-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o gistio .

FROM gcr.io/distroless/static
COPY --from=build /app/gistio /gistio
EXPOSE 8080
ENTRYPOINT ["/gistio"]
```

```bash
docker build -t gistio .
docker run -p 8080:8080 gistio
```

### Fly.io

```bash
fly launch          # accepts defaults; sets PORT automatically
fly deploy
```

### Railway / Render / Heroku

Set the start command to `./gistio` and ensure `PORT` is available as an
environment variable (all three platforms set this automatically).

### Vercel (Go runtime)

Create `api/index.go` wrapping the `http.Handler` and a `vercel.json` that
routes everything to it. The standard Go runtime on Vercel supports this
pattern directly.

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | TCP port the server listens on |
