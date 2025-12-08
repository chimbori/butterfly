package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

var appWebManifestTemplate = `{
  "name": "%s",
  "start_url": "%s",
  "theme_color": "%s",
  "display": "standalone",
  "icons": [{
    "src": "/static/favicon.svg",
    "type": "image/svg+xml",
    "sizes": "144x144"
  }]
}`

func ServeWebManifest(mux *http.ServeMux, appName, url, themeColor string) {
	mux.HandleFunc("GET /app.webmanifest", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/manifest+json")
		var dst bytes.Buffer
		json.Compact(&dst, fmt.Appendf(nil, appWebManifestTemplate, appName, url, themeColor))
		w.Write(dst.Bytes())
	})
}
