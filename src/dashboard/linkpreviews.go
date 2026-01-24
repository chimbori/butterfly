package dashboard

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"log/slog"
	"net/http"

	"chimbori.dev/butterfly/db"
	"chimbori.dev/butterfly/linkpreview"
	"chimbori.dev/butterfly/validation"
	nativewebp "github.com/HugoSmits86/nativewebp"
	"github.com/lmittmann/tint"
	"golang.org/x/image/draw"
)

// GET /dashboard/link-previews - List all cached link previews
var linkPreviewsPageHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	slog.Debug("linkPreviewsPageHandler", "url", req.Method+" "+req.URL.String())

	ctx := req.Context()
	queries := db.New(db.Pool)
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
	LinkPreviewsPageTempl(linkPreviews).Render(ctx, w)
})

// DELETE /dashboard/link-previews/url?url=... - Delete a cached link preview
var deleteLinkPreviewHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	slog.Debug("deleteLinkPreviewHandler", "url", req.Method+" "+req.URL.String())

	ctx := req.Context()
	queries := db.New(db.Pool)

	url := req.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "missing url parameter", http.StatusBadRequest)
		return
	}

	// Delete the cached file from disk
	if err := linkpreview.DeleteCached(url); err != nil {
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
	LinkPreviewsListTempl(linkPreviews).Render(ctx, w)
})

// GET /dashboard/link-previews/image?url={url}
// Serves a resized and compressed version of the cached link preview image.
var serveLinkPreviewHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	slog.Debug("serveLinkPreviewHandler", "url", req.Method+" "+req.URL.String())

	reqUrl := req.URL.Query().Get("url")
	if reqUrl == "" {
		err := fmt.Errorf("missing URL parameter")
		slog.Error("missing URL parameter", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", reqUrl,
			"status", http.StatusBadRequest)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	u, err := validation.Canonicalize(reqUrl)
	if err != nil {
		slog.Error("URL validation failed", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", reqUrl,
			"status", http.StatusBadRequest)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	url := u.String()
	png, err := linkpreview.GetCache().Find(url)
	if err != nil {
		err = fmt.Errorf("url: %s, %w", url, err)
		slog.Error("error during cache lookup", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", url,
			"hostname", u.Hostname(),
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if png == nil {
		slog.Error("cached link preview not found", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", url,
			"hostname", u.Hostname(),
			"status", http.StatusNotFound)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Decode the PNG image from the cache & compress it to WebP on the fly.
	img, _, err := image.Decode(bytes.NewReader(png))
	if err != nil {
		slog.Error("Error decoding link preview image", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", url,
			"hostname", u.Hostname(),
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	origBounds := img.Bounds()
	newWidth := 600
	ratio := float64(newWidth) / float64(origBounds.Dx())
	newHeight := int(float64(origBounds.Dy()) * ratio)

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, origBounds, draw.Over, nil)

	slog.Debug("image scaled successfully",
		"method", req.Method,
		"path", req.URL.Path,
		"url", url,
		"hostname", u.Hostname())

	// Encode as WebP.
	w.Header().Set("Content-Type", "image/webp")
	w.Header().Set("Cache-Control", "public, max-age=31536000") // 1 year cache
	if err := nativewebp.Encode(w, dst, &nativewebp.Options{}); err != nil {
		slog.Error("failed to encode WebP", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", url,
			"hostname", u.Hostname(),
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	slog.Debug("image converted to WebP & served successfully",
		"method", req.Method,
		"path", req.URL.Path,
		"url", url,
		"hostname", u.Hostname())
})
