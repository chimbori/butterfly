package dashboard

import (
	"log/slog"
	"net/http"

	"chimbori.dev/butterfly/db"
	"github.com/lmittmann/tint"
)

// GET /dashboard/link-previews - List all domains and cached link previews
var linkPreviewsHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	queries := db.New(db.Pool)
	domains, err := queries.ListDomains(ctx)
	if err != nil {
		slog.Error("failed to list domains", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", req.URL.String(),
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	linkPreviews, err := queries.ListLinkPreviews(ctx)
	if err != nil {
		slog.Error("failed to list cached link previews", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", req.URL.String(),
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	LinkPreviewsTempl(domains, linkPreviews).Render(ctx, w)
})
