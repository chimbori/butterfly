package dashboard

import (
	"log/slog"
	"net/http"

	"chimbori.dev/butterfly/db"
	"chimbori.dev/butterfly/linkpreviews"
	"github.com/lmittmann/tint"
)

// GET /dashboard/link-previews/list - List all cached link previews
var getCachedLinkPreviewsHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	queries := db.New(db.Pool)
	linkPreviews, err := queries.ListLinkPreviews(ctx)
	if err != nil {
		slog.Error("failed to list cached link previews", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	CachedLinkPreviewsTempl(linkPreviews).Render(ctx, w)
})

// DELETE /dashboard/link-previews/url?url=... - Delete a cached link preview
var deleteCachedLinkPreviewHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	queries := db.New(db.Pool)

	url := req.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "missing url parameter", http.StatusBadRequest)
		return
	}

	// Delete the cached file from disk
	if err := linkpreviews.DeleteCached(url); err != nil {
		slog.Warn("failed to delete cached file", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", url,
			"status", http.StatusInternalServerError)
		// Continue anyway to remove from the database
	}

	// Delete the row from the database
	err := queries.DeleteLinkPreview(ctx, url)
	if err != nil {
		slog.Error("failed to delete cached link preview", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", url,
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the updated list
	linkPreviews, err := queries.ListLinkPreviews(ctx)
	if err != nil {
		slog.Error("failed to list cached link previews", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", url,
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	CachedLinkPreviewsTempl(linkPreviews).Render(ctx, w)
})

// POST /dashboard/link-previews/regenerate?url=... - Regenerate a link preview
var regenerateCachedLinkPreviewHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	queries := db.New(db.Pool)

	err := req.ParseForm()
	if err != nil {
		slog.Error("failed to parse form", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"status", http.StatusBadRequest)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	url := req.FormValue("url")
	if url == "" {
		slog.Error("missing url parameter", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"status", http.StatusBadRequest)
		http.Error(w, "missing url parameter", http.StatusBadRequest)
		return
	}

	// Delete the cached link preview from disk
	if err := linkpreviews.DeleteCached(url); err != nil {
		slog.Warn("failed to delete cached file for regeneration", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", url,
			"status", http.StatusBadRequest)
		// Continue anyway
	}

	// Delete the database record so it gets recreated on next access
	err = queries.DeleteLinkPreview(ctx, url)
	if err != nil {
		slog.Error("failed to delete cached link preview record for regeneration", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", url,
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.Info("link preview marked for regeneration",
		"method", req.Method,
		"path", req.URL.Path,
		"url", url,
		"status", http.StatusInternalServerError)

	// Return the updated list
	linkPreviews, err := queries.ListLinkPreviews(ctx)
	if err != nil {
		slog.Error("failed to list cached link previews", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", url,
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	CachedLinkPreviewsTempl(linkPreviews).Render(ctx, w)
})
