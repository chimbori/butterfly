package main

import (
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
	url, err := validateUrl(req.URL.Query().Get("url"))
	if err != nil {
		slog.Error("URL validation failed", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"status", http.StatusBadRequest)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	selector := req.URL.Query().Get("sel")
	if selector == "" {
		selector = "#link-preview"
	}

	cached, err := findCached(url, selector)
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

	if cached != nil {
		slog.Info("cached screenshot served",
			"method", req.Method,
			"path", req.URL.Path,
			"url", url)
		w.Header().Set("Content-Type", "image/png")
		w.Write(cached)
	} else {
		screenshot, err := takeScreenshot(req.Context(), url, selector)
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

		err = writeToCache(url, selector, screenshot)
		if err != nil {
			err = fmt.Errorf("error writing to cache: %s, %w", url, err)
			slog.Error("error taking screenshot", tint.Err(err),
				"method", req.Method,
				"path", req.URL.Path,
				"url", url,
				"status", http.StatusInternalServerError)
			// Still continue serving the image to clients even if caching failed.
		}

		slog.Info("new screenshot generated",
			"method", req.Method,
			"path", req.URL.Path,
			"url", url)
		w.Header().Set("Content-Type", "image/png")
		w.Write(screenshot)
	}
}

// Validates a URL provided by the user, and returns a formatted URL string.
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
		return "", errors.New("domain " + u.Host + " not authorized")
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
