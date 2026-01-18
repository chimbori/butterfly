package linkpreviews

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"chimbori.dev/butterfly/conf"
	"chimbori.dev/butterfly/db"
	"chimbori.dev/butterfly/embedfs"
	"github.com/lmittmann/tint"
)

func SetupHandlers(mux *http.ServeMux) {
	mux.HandleFunc("GET /link-preview/v1", handleLinkPreview)
}

// GET /link-preview/v1?url={url}&sel={selector}
// Validates the URL, checks if it’s cached, generates screenshots, and serves them.
func handleLinkPreview(w http.ResponseWriter, req *http.Request) {
	reqUrl := req.URL.Query().Get("url")
	queries := db.New(db.Pool)
	url, hostname, err := validateUrl(req.Context(), queries, reqUrl)
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
		cached, err = findCached(url)
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
		w.Write(cached)
		recordLinkPreviewCreated(url)

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
				screenshot, err = takeScreenshotWithTemplate(ctx, url, embedfs.DefaultTemplate, title, description)
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

		// Only write to cache if enabled
		if *conf.Config.LinkPreview.Cache.Enabled {
			err = writeToCache(url, screenshot)
			if err != nil {
				err = fmt.Errorf("error writing to cache: %s, %w", url, err)
				slog.Error("error writing to cache", tint.Err(err),
					"method", req.Method,
					"path", req.URL.Path,
					"url", url,
					"hostname", hostname,
					"status", http.StatusInternalServerError)
				// Still continue serving the image to clients even if caching failed.
			}
		}

		slog.Info("new screenshot generated",
			"method", req.Method,
			"path", req.URL.Path,
			"url", url,
			"hostname", hostname,
			"status", http.StatusOK)
		w.Header().Set("Content-Type", "image/png")
		w.Write(screenshot)
		recordLinkPreviewCreated(url)
	}
}

// Validates a URL provided by the user, and returns a formatted URL as a string.
func validateUrl(ctx context.Context, q *db.Queries, userUrl string) (validatedUrl string, hostname string, err error) {
	if userUrl == "" {
		return "", "", errors.New("missing url")
	}

	if !strings.HasPrefix(userUrl, "https://") && !strings.HasPrefix(userUrl, "http://") && !strings.Contains(userUrl, "://") {
		userUrl = "https://" + userUrl
	}

	u, err := url.Parse(userUrl)
	if err != nil {
		return "", "", errors.New("invalid url")
	}

	authorized, err := isAuthorized(ctx, q, u)
	if err != nil {
		return "", u.Hostname(), err
	}

	if !authorized {
		return "", u.Hostname(), errors.New("domain " + u.Hostname() + " not authorized")
	}

	return u.String(), u.Hostname(), nil
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

// isAuthorized returns true if the given URL's domain is in the list of authorized domains.
// As a side effect, if the domain is not authorized and doesn’t exist in the database,
// it will be added (default blocked) for future triage.
func isAuthorized(ctx context.Context, q *db.Queries, u *url.URL) (bool, error) {
	hostname := u.Hostname()
	authorized, err := q.IsAuthorized(ctx, hostname)
	if err != nil {
		return false, err
	}

	// If not authorized, add it to the database for future triage.
	if !authorized {
		err = q.InsertUnauthorizedDomain(ctx, hostname)
		if err != nil {
			// Log the error but don’t fail the authorization check.
			slog.Error("failed to insert unauthorized domain", tint.Err(err), "domain", hostname)
		}
	}

	return authorized, nil
}
