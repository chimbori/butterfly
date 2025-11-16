package embedfs

import (
	"embed"
	"net/http"

	"go.chimbori.app/butterfly/core"
)

//go:embed static
var staticFiles embed.FS

func ServeStaticFS(mux *http.ServeMux) {
	mux.Handle("GET /static/", core.MaxAgeHandler(http.FileServer(http.FS(staticFiles))))
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFileFS(w, req, staticFiles, "static/favicon.svg")
	})
}
