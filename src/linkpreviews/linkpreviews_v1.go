package linkpreviews

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"regexp"

	"butterfly.chimbori.dev/conf"
	"butterfly.chimbori.dev/core"
	"butterfly.chimbori.dev/db"
	"butterfly.chimbori.dev/embedfs"
	"butterfly.chimbori.dev/validation"
	"github.com/lmittmann/tint"
)

var Cache *core.DiskCache

var selectorRegex = regexp.MustCompile(`^[#.][a-zA-Z0-9_-]+$`)

func Init(mux *http.ServeMux) {
	if *conf.Config.LinkPreviews.Cache.Enabled {
		Cache = core.NewDiskCache(
			filepath.Join(conf.Config.DataDir, "cache", "link-previews"),
			core.WithTTL(conf.Config.LinkPreviews.Cache.TTL),
			core.WithMaxSize(conf.Config.LinkPreviews.Cache.MaxSizeBytes),
		)
	} // else cache will be nil

	mux.HandleFunc("GET /link-previews/v1", handleLinkPreview)
}

// GET /link-previews/v1?url={url}&sel={selector}
// Validates the URL, checks if it’s cached, generates screenshots, and serves them.
func handleLinkPreview(w http.ResponseWriter, req *http.Request) {
	slog.Debug("handleLinkPreview", "url", req.Method+" "+req.URL.String())

	reqUrl := req.URL.Query().Get("url")
	userAgent := req.Header.Get("User-Agent")
	queries := db.New(db.Pool)

	url, hostname, err := validation.ValidateUrl(req.Context(), queries, reqUrl)
	if err != nil {
		slog.Error("URL validation failed", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", reqUrl,
			"hostname", hostname,
			"user-agent", userAgent,
			"status", http.StatusUnauthorized)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	selector := req.URL.Query().Get("sel")
	if selector == "" {
		selector = "#link-preview"
	} else if !selectorRegex.MatchString(selector) {
		err := fmt.Errorf("invalid selector")
		slog.Error("selector validation failed", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", reqUrl,
			"hostname", hostname,
			"user-agent", userAgent,
			"status", http.StatusBadRequest)
		http.Error(w, "invalid selector", http.StatusBadRequest)
		return
	}

	var cached []byte

	// Only check cache if enabled
	if *conf.Config.LinkPreviews.Cache.Enabled {
		var err error
		cached, err = Cache.Find(url)
		if err != nil {
			err = fmt.Errorf("url: %s, %w", url, err)
			slog.Error("error during cache lookup", tint.Err(err),
				"method", req.Method,
				"path", req.URL.Path,
				"url", url,
				"hostname", hostname,
				"user-agent", userAgent,
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
			"user-agent", userAgent,
			"status", http.StatusOK)
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "max-age=31536000, immutable") // 1 year
		w.Write(cached)
		recordLinkPreviewAccessed(url)

	} else {
		ctx, cancel := context.WithTimeout(req.Context(), conf.Config.LinkPreviews.Screenshot.Timeout)
		defer cancel()
		screenshot, err := core.TakeScreenshot(ctx, url, selector)
		if err != nil {
			if !errors.Is(err, core.ErrMissingSelector) {
				err = fmt.Errorf("url: %s, %w", url, err)
				slog.Error("error taking screenshot", tint.Err(err),
					"method", req.Method,
					"path", req.URL.Path,
					"url", url,
					"hostname", hostname,
					"user-agent", userAgent,
					"status", http.StatusInternalServerError)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			slog.Info("attempting with default template",
				"method", req.Method,
				"path", req.URL.Path,
				"url", url,
				"hostname", hostname,
				"user-agent", userAgent,
				"status", http.StatusOK)
			title, description, fetchErr := core.FetchTitleAndDescription(ctx, url)
			if fetchErr != nil {
				err = fmt.Errorf("fetchTitleAndDescription failed: %w", fetchErr)
			} else {
				screenshot, err = core.TakeScreenshotWithTemplate(ctx, embedfs.DefaultTemplate, url, "#link-preview", title, description)
			}
			if err != nil {
				err = fmt.Errorf("url: %s, %w", url, err)
				slog.Error("error using default template", tint.Err(err),
					"method", req.Method,
					"path", req.URL.Path,
					"url", url,
					"hostname", hostname,
					"user-agent", userAgent,
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
			"user-agent", userAgent,
			"status", http.StatusOK)
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "max-age=31536000, immutable") // 1 year
		w.Write(screenshot)
		recordLinkPreviewCreated(url)

		// If cache is enabled, compress the generated screenshot and cache it, but without holding up the HTTP request
		go func() {
			if *conf.Config.LinkPreviews.Cache.Enabled {
				dataToWrite := screenshot
				compressed, err := core.CompressPNG(screenshot)
				if err == nil {
					dataToWrite = compressed
					slog.Info("PNG compressed", "from", len(screenshot), "to", len(compressed), "%", (len(compressed) * 100 / len(screenshot)))
				} else {
					slog.Error("PNG compression failed", tint.Err(err), "url", url)
				}

				if err := Cache.Write(url, dataToWrite); err != nil {
					err = fmt.Errorf("error writing to cache: %s, %w", url, err)
					slog.Error("error writing to cache", tint.Err(err),
						"method", req.Method,
						"path", req.URL.Path,
						"url", url,
						"hostname", hostname,
						"user-agent", userAgent,
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
	rowsUpdated, err := queries.RecordLinkPreviewAccessed(context.Background(), url)
	if err != nil {
		slog.Error("failed to log link preview created", tint.Err(err))
	}
	if rowsUpdated == 0 { // If not already in the database, add it now.
		recordLinkPreviewCreated(url)
	}
	// Don’t return an error to the caller; fulfill the request anyway.
}

// DeleteCached removes a cached screenshot file from disk.
func DeleteCached(url string) error {
	return Cache.Delete(url)
}
