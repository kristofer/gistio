package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

//go:embed templates
var templateFiles embed.FS

//go:embed static
var staticFiles embed.FS

//go:embed src/fonts
var fontFiles embed.FS

var (
	templates *template.Template
	md        goldmark.Markdown
)

func init() {
	templates = template.Must(template.ParseFS(templateFiles, "templates/*.html"))

	// Markdown renderer with Paraiso-dark syntax highlighting (no CGO, pure Go).
	md = goldmark.New(
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithStyle("paraiso-dark"),
			),
		),
	)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// Embedded static files (CSS).
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("static embed: %v", err)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Embedded font files served from src/fonts/.
	fontsFS, err := fs.Sub(fontFiles, "src/fonts")
	if err != nil {
		log.Fatalf("fonts embed: %v", err)
	}
	mux.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.FS(fontsFS))))

	// Page routes — all unmatched paths go through dispatch.
	mux.HandleFunc("/", dispatch)

	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

// dispatch routes "/" to the home page and "/@user/id" to the gist page.
func dispatch(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/":
		homeHandler(w, r)
	case strings.HasPrefix(r.URL.Path, "/@"):
		gistHandler(w, r)
	default:
		http.NotFound(w, r)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	render(w, "home.html", map[string]string{
		"Title": "Gist.io • Writing for Hackers",
	})
}

func gistHandler(w http.ResponseWriter, r *http.Request) {
	// Parse /@user/id out of the URL path.
	parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/@"), "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		http.NotFound(w, r)
		return
	}
	user, id := parts[0], parts[1]

	// Fetch the raw Markdown from GitHub Gists.
	rawURL := fmt.Sprintf("https://gist.githubusercontent.com/%s/%s/raw", user, id)
	resp, err := http.Get(rawURL) //nolint:noctx
	if err != nil {
		http.Error(w, "failed to fetch gist", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "gist not found", http.StatusNotFound)
		return
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read gist", http.StatusInternalServerError)
		return
	}

	// Render Markdown → HTML.
	var buf bytes.Buffer
	if err := md.Convert(raw, &buf); err != nil {
		http.Error(w, "failed to render markdown", http.StatusInternalServerError)
		return
	}

	render(w, "gist.html", map[string]interface{}{
		"Title":   fmt.Sprintf("Gist.io • @%s/%s", user, id),
		"User":    user,
		"ID":      id,
		"Content": template.HTML(buf.String()),
	})
}

// render executes the named template and writes HTML to w.
func render(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
