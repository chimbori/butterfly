package dashboard

import (
	"log/slog"
	"net/http"

	"chimbori.dev/butterfly/db"
	"chimbori.dev/butterfly/qrcode"
	"github.com/lmittmann/tint"
)

// GET /dashboard/qr-codes - List all QR Codes
func listQrCodesHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	queries := db.New(db.Pool)
	qrCodes, err := queries.ListQrCodes(ctx)
	if err != nil {
		slog.Error("failed to list QR Codes", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	QrCodesPageTempl(qrCodes).Render(ctx, w)
}

// DELETE /dashboard/qr-codes/url?url={url} - Delete a cached QR Code
func deleteQrCodeHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	queries := db.New(db.Pool)

	url := req.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "missing url parameter", http.StatusBadRequest)
		return
	}

	// Delete the cached file from disk
	if err := qrcode.DeleteCached(url); err != nil {
		slog.Warn("failed to delete cached QR Code file", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", url)
		// Continue anyway to remove from the database
	}

	// Delete the row from the database
	if err := queries.DeleteQrCode(ctx, url); err != nil {
		slog.Error("failed to delete QR Code", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", url,
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the updated list
	qrCodes, err := queries.ListQrCodes(ctx)
	if err != nil {
		slog.Error("failed to list QR Codes", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	QrCodesListTempl(qrCodes).Render(ctx, w)
}
