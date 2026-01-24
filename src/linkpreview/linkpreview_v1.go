package linkpreview

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"

	"chimbori.dev/butterfly/conf"
	"chimbori.dev/butterfly/core"
	"chimbori.dev/butterfly/db"
	"chimbori.dev/butterfly/embedfs"
	"chimbori.dev/butterfly/validation"
	"github.com/lmittmann/tint"
)

var cache *core.DiskCache

func InitCache() {
	cache = core.NewDiskCache(filepath.Join(conf.Config.DataDir, "cache", "link-previews"))
}

func GetCache() *core.DiskCache {
	return cache
}

func SetupHandlers(mux *http.ServeMux) {
	mux.HandleFunc("GET /link-preview/v1", handleLinkPreview)
}

// GET /link-preview/v1?url={url}&sel={selector}
// Validates the URL, checks if it’s cached, generates screenshots, and serves them.
func handleLinkPreview(w http.ResponseWriter, req *http.Request) {
	slog.Debug("handleLinkPreview", "url", req.Method+" "+req.URL.String())

	reqUrl := req.URL.Query().Get("url")
	queries := db.New(db.Pool)

	url, hostname, err := validation.ValidateUrl(req.Context(), queries, reqUrl)
	if err != nil {
		slog.Error("URL validation failed", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", reqUrl,
			"hostname", hostname,
			"status", http.StatusUnauthorized)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	selector := req.URL.Query().Get("sel")
	if selector == "" {
		selector = "#link-preview"
	}

	var cached []byte

	// Only check cache if enabled
	if *conf.Config.LinkPreview.Cache.Enabled {
		var err error
		cached, err = cache.Find(url)
		if err != nil {
			err = fmt.Errorf("url: %s, %w", url, err)
			slog.Error("error during cache lookup", tint.Err(err),
				"method", req.Method,
				"path", req.URL.Path,
				"url", url,
				"hostname", hostname,
				"status", http.StatusInternalServerError)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if cached != nil {
		slog.Info("cached screenshot served",
			"method", req.Method,
			"path", req.URL.Path,
			"url", url,
			"hostname", hostname,
			"status", http.StatusOK)
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "max-age=31536000, immutable") // 1 year
		w.Write(cached)
		recordLinkPreviewAccessed(url)

	} else {
		ctx, cancel := context.WithTimeout(req.Context(), conf.Config.LinkPreview.Screenshot.Timeout)
		defer cancel()
		screenshot, err := takeScreenshot(ctx, url, selector)
		if err != nil {
			if !errors.Is(err, ErrMissingSelector) {
				err = fmt.Errorf("url: %s, %w", url, err)
				slog.Error("error taking screenshot", tint.Err(err),
					"method", req.Method,
					"path", req.URL.Path,
					"url", url,
					"hostname", hostname,
					"status", http.StatusInternalServerError)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			slog.Info("attempting with default template",
				"method", req.Method,
				"path", req.URL.Path,
				"url", url,
				"hostname", hostname,
				"status", http.StatusOK)
			title, description, fetchErr := fetchTitleAndDescription(ctx, url)
			if fetchErr != nil {
				err = fmt.Errorf("fetchTitleAndDescription failed: %w", fetchErr)
			} else {
				screenshot, err = takeScreenshotWithTemplate(ctx, embedfs.DefaultTemplate, url, title, description)
			}
			if err != nil {
				err = fmt.Errorf("url: %s, %w", url, err)
				slog.Error("error using default template", tint.Err(err),
					"method", req.Method,
					"path", req.URL.Path,
					"url", url,
					"hostname", hostname,
					"status", http.StatusInternalServerError)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Serve the screenshot immediately after generation, without waiting for compression.
		slog.Info("new screenshot generated",
			"method", req.Method,
			"path", req.URL.Path,
			"url", url,
			"hostname", hostname,
			"status", http.StatusOK)
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "max-age=31536000, immutable") // 1 year
		w.Write(screenshot)
		recordLinkPreviewCreated(url)

		// If cache is enabled, compress the generated screenshot and cache it, but without holding up the HTTP request
		go func() {
			if *conf.Config.LinkPreview.Cache.Enabled {
				dataToWrite := screenshot
				compressed, err := core.CompressPNG(screenshot)
				if err == nil {
					dataToWrite = compressed
					slog.Info("PNG compressed", "from", len(screenshot), "to", len(compressed), "%", (len(compressed) * 100 / len(screenshot)))
				} else {
					slog.Error("PNG compression failed", tint.Err(err), "url", url)
				}

				if err := cache.Write(url, dataToWrite); err != nil {
					err = fmt.Errorf("error writing to cache: %s, %w", url, err)
					slog.Error("error writing to cache", tint.Err(err),
						"method", req.Method,
						"path", req.URL.Path,
						"url", url,
						"hostname", hostname,
						"status", http.StatusInternalServerError)
				}
			}
		}()
	}
}

// Record when a link preview is created (for the first time)
func recordLinkPreviewCreated(url string) {
	queries := db.New(db.Pool)
	err := queries.RecordLinkPreviewCreated(context.Background(), url)
	if err != nil {
		slog.Error("failed to log link preview created", tint.Err(err))
	}
	// Don’t return an error to the caller; fulfill the request anyway.
}

// Record when a link preview is accessed from the cache
func recordLinkPreviewAccessed(url string) {
	queries := db.New(db.Pool)
	err := queries.RecordLinkPreviewAccessed(context.Background(), url)
	if err != nil {
		slog.Error("failed to log link preview created", tint.Err(err))
	}
	// Don’t return an error to the caller; fulfill the request anyway.
}

// DeleteCached removes a cached screenshot file from disk.
func DeleteCached(url string) error {
	return cache.Delete(url)
}
