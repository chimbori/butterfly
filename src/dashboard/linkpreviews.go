package dashboard

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"log/slog"
	"net/http"
	"runtime"
	"strconv"

	"butterfly.chimbori.dev/conf"
	"butterfly.chimbori.dev/core"
	"butterfly.chimbori.dev/db"
	"butterfly.chimbori.dev/linkpreviews"
	"butterfly.chimbori.dev/validation"
	nativewebp "github.com/HugoSmits86/nativewebp"
	"github.com/disintegration/imaging"
	"github.com/lmittmann/tint"
)

// compressionSem limits the number of concurrent image compression tasks.
var compressionSem chan struct{}

var ThumbnailCache *core.DiskCache

func init() {
	compressionSem = make(chan struct{}, runtime.NumCPU()*4)
}

// GET /dashboard/link-previews - Render the link previews page
func linkPreviewsPageHandler(w http.ResponseWriter, req *http.Request) {
	slog.Debug("linkPreviewsPageHandler", "url", req.Method+" "+req.URL.String())
	page := 1
	if pageStr := req.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	LinkPreviewsPageTempl(page).Render(req.Context(), w)
}

// GET /dashboard/link-previews/list - Get paginated link previews list
func linkPreviewsListHandler(w http.ResponseWriter, req *http.Request) {
	slog.Debug("linkPreviewsListHandler", "url", req.Method+" "+req.URL.String())

	ctx := req.Context()
	queries := db.New(db.Pool)

	page := 1
	if pageStr := req.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Fetch total count for pagination
	totalCount, err := queries.CountLinkPreviews(ctx)
	if err != nil {
		slog.Error("failed to count link previews", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", req.URL.String(),
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch paginated items
	offset := int32((page - 1) * conf.Config.Dashboard.Pagination.Limit)
	linkPreviews, err := queries.ListLinkPreviewsPaginated(ctx, db.ListLinkPreviewsPaginatedParams{
		Limit:  int32(conf.Config.Dashboard.Pagination.Limit),
		Offset: offset,
	})
	if err != nil {
		slog.Error("failed to list cached link previews", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", req.URL.String(),
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	LinkPreviewsListTempl(linkPreviews, page, totalCount).Render(ctx, w)
}

// DELETE /dashboard/link-previews/url?url=... - Delete a cached link preview
func deleteLinkPreviewHandler(w http.ResponseWriter, req *http.Request) {
	slog.Debug("deleteLinkPreviewHandler", "url", req.Method+" "+req.URL.String())

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

	// Return the updated list (go back to page 1)
	totalCount, err := queries.CountLinkPreviews(ctx)
	if err != nil {
		slog.Error("failed to count link previews", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", url,
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	linkPreviews, err := queries.ListLinkPreviewsPaginated(ctx, db.ListLinkPreviewsPaginatedParams{
		Limit:  int32(conf.Config.Dashboard.Pagination.Limit),
		Offset: 0,
	})
	if err != nil {
		slog.Error("failed to list cached link previews", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", url,
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Don’t go back to the first page.
	LinkPreviewsListTempl(linkPreviews, 1, totalCount).Render(ctx, w)
}

// GET /dashboard/link-previews/image?url={url}
// Serves a resized and compressed version of the cached link preview image.
func serveLinkPreviewHandler(w http.ResponseWriter, req *http.Request) {
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

	if ThumbnailCache != nil {
		if webp, err := ThumbnailCache.Find(url); err == nil && webp != nil {
			slog.Debug("serving from thumbnail cache", "url", url)
			w.Header().Set("Content-Type", "image/webp")
			w.Header().Set("Cache-Control", "public, max-age=31536000") // 1 year cache
			w.Write(webp)
			return
		}
	}

	if linkpreviews.Cache == nil {
		err := fmt.Errorf("preview unavailable for %s", url)
		slog.Error("cache disabled", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", url,
			"hostname", u.Hostname(),
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	png, err := linkpreviews.Cache.Find(url)
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
		err := fmt.Errorf("cached link preview not found")
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
	compressionSem <- struct{}{}
	defer func() { <-compressionSem }()

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

	// Use NearestNeighbor for simplicity & speed, since we’re only scaling down, never up.
	resized := imaging.Resize(img, 600, 0, imaging.NearestNeighbor)

	slog.Debug("image scaled successfully",
		"method", req.Method,
		"path", req.URL.Path,
		"url", url,
		"hostname", u.Hostname())

	// Encode as WebP.
	var webpBuf bytes.Buffer
	if err := nativewebp.Encode(&webpBuf, resized, &nativewebp.Options{}); err != nil {
		slog.Error("failed to encode WebP", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", url,
			"hostname", u.Hostname(),
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	webpData := webpBuf.Bytes()

	if ThumbnailCache != nil {
		go ThumbnailCache.Write(url, webpData)
	}

	w.Header().Set("Content-Type", "image/webp")
	w.Header().Set("Cache-Control", "public, max-age=31536000") // 1 year cache
	w.Write(webpData)

	slog.Debug("image converted to WebP & served successfully",
		"method", req.Method,
		"path", req.URL.Path,
		"url", url,
		"hostname", u.Hostname())
}

func calculateTotalPages(totalCount int64) int64 {
	limit := int64(conf.Config.Dashboard.Pagination.Limit)
	return (totalCount + (limit - 1)) / limit
}
