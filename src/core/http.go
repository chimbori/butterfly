package core

import (
	"net/http"
	"path/filepath"
)

// MaxAgeHandler wraps an HTTP handler to set cache control headers based on file extension.
// CSS files are cached for 1 day; all other static files for 1 year.
func MaxAgeHandler(h http.Handler) http.Handler {
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
