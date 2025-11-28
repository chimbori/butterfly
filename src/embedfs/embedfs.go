package embedfs

import (
	"embed"
	"net/http"
	"path/filepath"
)

//go:embed static
var staticFiles embed.FS

//go:embed "default-template.html"
var DefaultTemplate string

func ServeStaticFS(mux *http.ServeMux) {
	mux.Handle("GET /static/", maxAgeHandler(http.FileServer(http.FS(staticFiles))))
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFileFS(w, req, staticFiles, "static/favicon.svg")
	})
}

// maxAgeHandler wraps an HTTP handler to set cache control headers based on file extension.
// CSS files are cached for 1 day; all other static files for 1 year.
func maxAgeHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ext := filepath.Ext(req.URL.Path)
		if ext == ".css" {
			w.Header().Set("Cache-Control", "max-age=86400") // 1 day
		} else {
			w.Header().Set("Cache-Control", "max-age=31536000, immutable") // 1 year
		}
		h.ServeHTTP(w, req)
	})
}
