package qrcode

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"path/filepath"

	"chimbori.dev/butterfly/conf"
	"chimbori.dev/butterfly/core"
	"chimbori.dev/butterfly/db"
	"chimbori.dev/butterfly/validation"
	"github.com/lmittmann/tint"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

var cache *core.Cache

func InitCache() {
	cache = core.NewCache(filepath.Join(conf.Config.DataDir, "cache", "qr-codes"))
}

func SetupHandlers(mux *http.ServeMux) {
	mux.HandleFunc("GET /qrcode/v1", handleQrCode)
}

// GET /qrcode/v1?url={url}
// Validates the URL, checks if it’s cached, generates QR Code, and serves it.
func handleQrCode(w http.ResponseWriter, req *http.Request) {
	reqUrl := req.URL.Query().Get("url")
	queries := db.New(db.Pool)

	validatedUrl, hostname, err := validation.ValidateUrl(req.Context(), queries, reqUrl)
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

	var cached []byte

	// Only check cache if enabled
	if *conf.Config.QrCode.Cache.Enabled {
		cached, err = cache.Find(validatedUrl)
		if err != nil {
			slog.Error("error during cache lookup", tint.Err(err),
				"method", req.Method,
				"path", req.URL.Path,
				"url", validatedUrl,
				"hostname", hostname,
				"status", http.StatusInternalServerError)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if cached != nil {
		slog.Info("cached QR Code served",
			"method", req.Method,
			"path", req.URL.Path,
			"url", validatedUrl,
			"hostname", hostname,
			"status", http.StatusOK)
		w.Header().Set("Content-Type", "image/png")
		w.Write(cached)
		recordQrCodeAccessed(validatedUrl)
		return
	}

	// Generate new QR Code
	png, err := generateQrCode(validatedUrl)
	if err != nil {
		slog.Error("error generating QR Code", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", validatedUrl,
			"hostname", hostname,
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write to cache if enabled
	if *conf.Config.QrCode.Cache.Enabled {
		if err := cache.Write(validatedUrl, png); err != nil {
			slog.Error("error writing to cache", tint.Err(err),
				"method", req.Method,
				"path", req.URL.Path,
				"url", validatedUrl,
				"hostname", hostname)
			// Continue serving even if caching failed
		}
	}

	slog.Info("new QR Code generated",
		"method", req.Method,
		"path", req.URL.Path,
		"url", validatedUrl,
		"hostname", hostname,
		"status", http.StatusOK)
	w.Header().Set("Content-Type", "image/png")
	w.Write(png)
	recordQrCodeCreated(validatedUrl)
}

// writeCloser wraps an io.Writer and adds a no-op Close method
type writeCloser struct {
	*bytes.Buffer
}

func (wc *writeCloser) Close() error {
	return nil
}

// generateQrCode creates a QR Code PNG for the given URL
func generateQrCode(url string) ([]byte, error) {
	qrc, err := qrcode.New(url)
	if err != nil {
		return nil, err
	}

	// Create a buffer to write the QR code to
	var buf bytes.Buffer
	wc := &writeCloser{&buf}
	writer := standard.NewWithWriter(wc, standard.WithBuiltinImageEncoder(standard.PNG_FORMAT))

	if err := qrc.Save(writer); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// recordQrCodeCreated records when a QR Code is created (for the first time)
func recordQrCodeCreated(url string) {
	queries := db.New(db.Pool)
	err := queries.RecordQrCodeCreated(context.Background(), url)
	if err != nil {
		slog.Error("failed to log QR Code created", tint.Err(err))
	}
}

// recordQrCodeAccessed records when a QR Code is accessed from the cache
func recordQrCodeAccessed(url string) {
	queries := db.New(db.Pool)
	err := queries.RecordQrCodeAccessed(context.Background(), url)
	if err != nil {
		slog.Error("failed to log QR Code accessed", tint.Err(err))
	}
}

// DeleteCached removes a cached QR Code file from disk.
func DeleteCached(url string) error {
	return cache.Delete(url)
}
