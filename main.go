package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

// Fonts are always present in the repository and are embedded in the binary.
// Static web assets (HTML, CSS, WASM) are served from the "static/" directory
// on disk so that the WASM binary can be built independently (run "make" first).

//go:embed src/fonts
var fontFiles embed.FS

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// Static web assets: HTML shell, CSS, wasm_exec.js, main.wasm.
	// Build the WASM first: run "make" or "make wasm".
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Embedded font files served from src/fonts/.
	fontsFS, err := fs.Sub(fontFiles, "src/fonts")
	if err != nil {
		log.Fatalf("fonts embed: %v", err)
	}
	mux.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.FS(fontsFS))))

	// Proxy: fetches external markdown on behalf of the browser to bypass CORS.
	mux.HandleFunc("/proxy", proxyHandler)

	// SPA fallback: serve index.html for all other paths so the client-side
	// router handles /@user/id and ?url=… routes.
	mux.HandleFunc("/", spaHandler)

	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

// spaHandler serves the single-page application shell for every route that is
// not handled by a more specific pattern (static assets, fonts, proxy).
func spaHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

// proxyHandler fetches a remote markdown file and streams it back to the
// browser, bypassing same-origin restrictions.  Only HTTPS URLs are allowed.
func proxyHandler(w http.ResponseWriter, r *http.Request) {
	rawURL := r.URL.Query().Get("url")
	if rawURL == "" {
		http.Error(w, "missing url parameter", http.StatusBadRequest)
		return
	}

	// Restrict to HTTPS to prevent requests to local or plain-text services.
	if !strings.HasPrefix(rawURL, "https://") {
		http.Error(w, "only https URLs are allowed", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(rawURL) //nolint:noctx
	if err != nil {
		http.Error(w, "failed to fetch URL", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("upstream returned %d", resp.StatusCode), resp.StatusCode)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("proxy copy error: %v", err)
	}
}
