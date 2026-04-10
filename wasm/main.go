//go:build js && wasm

package main

import (
	"bytes"
	"syscall/js"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

var md goldmark.Markdown

func init() {
	md = goldmark.New(
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithStyle("paraiso-dark"),
				// Inline styles so no external CSS is required for syntax highlighting.
				highlighting.WithFormatOptions(
					chromahtml.WithClasses(false),
				),
			),
		),
	)
}

// renderMarkdown is exported to JavaScript as window.renderMarkdown(text) → html.
func renderMarkdown(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return ""
	}
	raw := args[0].String()
	var buf bytes.Buffer
	if err := md.Convert([]byte(raw), &buf); err != nil {
		return ""
	}
	return buf.String()
}

func main() {
	js.Global().Set("renderMarkdown", js.FuncOf(renderMarkdown))
	// Block forever so the WASM module (and its exported functions) stay alive.
	select {}
}
