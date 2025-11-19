package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/lmittmann/tint"
	"go.chimbori.app/butterfly/conf"
)

// GET /link-preview/v1?url={url}&sel={selector}
// Validates the URL, checks if it’s cached, generates screenshots, and serves them.
func handleLinkPreview(w http.ResponseWriter, req *http.Request) {
	reqUrl := req.URL.Query().Get("url")
	url, err := validateUrl(reqUrl)
	if err != nil {
		slog.Error("URL validation failed", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", reqUrl,
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
		cached, err = findCached(url, selector)
		if err != nil {
			err = fmt.Errorf("url: %s, %w", url, err)
			slog.Error("error during cache lookup", tint.Err(err),
				"method", req.Method,
				"path", req.URL.Path,
				"url", url,
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
			"status", http.StatusOK)
		w.Header().Set("Content-Type", "image/png")
		w.Write(cached)
	} else {
		ctx, cancel := context.WithTimeout(req.Context(), conf.Config.LinkPreview.Screenshot.Timeout)
		defer cancel()
		screenshot, err := takeScreenshot(ctx, url, selector)
		if err != nil {
			err = fmt.Errorf("url: %s, %w", url, err)
			slog.Error("error taking screenshot", tint.Err(err),
				"method", req.Method,
				"path", req.URL.Path,
				"url", url,
				"status", http.StatusInternalServerError)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Only write to cache if enabled
		if *conf.Config.LinkPreview.Cache.Enabled {
			err = writeToCache(url, selector, screenshot)
			if err != nil {
				err = fmt.Errorf("error writing to cache: %s, %w", url, err)
				slog.Error("error writing to cache", tint.Err(err),
					"method", req.Method,
					"path", req.URL.Path,
					"url", url,
					"status", http.StatusInternalServerError)
				// Still continue serving the image to clients even if caching failed.
			}
		}

		slog.Info("new screenshot generated",
			"method", req.Method,
			"path", req.URL.Path,
			"url", url,
			"status", http.StatusOK)
		w.Header().Set("Content-Type", "image/png")
		w.Write(screenshot)
	}
}

// Validates a URL provided by the user, and returns a formatted URL as a string.
func validateUrl(userUrl string) (string, error) {
	if userUrl == "" {
		return "", errors.New("missing url")
	}

	if !strings.HasPrefix(userUrl, "https://") && !strings.HasPrefix(userUrl, "http://") {
		userUrl = "https://" + userUrl
	}

	u, err := url.Parse(userUrl)
	if err != nil {
		return "", errors.New("invalid url")
	}

	if !isAuthorized(u) {
		return "", errors.New("domain " + u.Hostname() + " not authorized")
	}

	return u.String(), nil
}

// isAuthorized returns true if the given URL’s domain is in the list of authorized domains.
func isAuthorized(u *url.URL) bool {
	hostname := u.Hostname()
	for _, domain := range conf.Config.LinkPreview.Domains {
		if hostname == domain || strings.HasSuffix(hostname, "."+domain) {
			return true
		}
	}
	return false
}
