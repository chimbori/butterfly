package slogdb

import (
	"context"
	"log/slog"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.chimbori.app/butterfly/db"
)

// DBHandler is a slog.Handler that writes error-level logs to the PostgreSQL `errors` table.
// It wraps another handler to maintain normal console/file logging.
type DBHandler struct {
	parent slog.Handler
	pool   *pgxpool.Pool
	mu     sync.Mutex
}

// NewDBHandler creates a new database logging handler that wraps the parent handler.
// Only ERROR level logs are written to the database; all logs are passed to the parent.
func NewDBHandler(parent slog.Handler, pool *pgxpool.Pool) *DBHandler {
	return &DBHandler{
		parent: parent,
		pool:   pool,
	}
}

// Enabled reports whether the handler handles records at the given level. It delegates to the parent handler.
func (h *DBHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.parent.Enabled(ctx, level)
}

// Handle writes error-level logs to the database, then delegates to the parent handler.
func (h *DBHandler) Handle(ctx context.Context, r slog.Record) error {
	// Write to database if this is an error-level log
	if r.Level >= slog.LevelError {
		h.writeToDatabase(ctx, r)
	}
	// Always pass through to the parent handler for console/file logging
	return h.parent.Handle(ctx, r)
}

// WithAttrs returns a new handler with the given attributes added.
func (h *DBHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &DBHandler{
		parent: h.parent.WithAttrs(attrs),
		pool:   h.pool,
	}
}

// WithGroup returns a new handler with the given group added.
func (h *DBHandler) WithGroup(name string) slog.Handler {
	return &DBHandler{
		parent: h.parent.WithGroup(name),
		pool:   h.pool,
	}
}

// writeToDatabase extracts relevant information from the log record and writes it to the `logs` table.
func (h *DBHandler) writeToDatabase(ctx context.Context, r slog.Record) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Extract attributes from the log record
	var (
		slogErr       *string
		requestMethod *string
		requestPath   *string
		httpStatus    *int32
		url           *string
		hostname      *string
	)

	r.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "err":
			if s := a.Value.String(); s != "" {
				slogErr = &s
			}
		case "method":
			if s := a.Value.String(); s != "" {
				requestMethod = &s
			}
		case "path":
			if s := a.Value.String(); s != "" {
				requestPath = &s
			}
		case "url":
			if s := a.Value.String(); s != "" {
				url = &s
			}
		case "hostname":
			if s := a.Value.String(); s != "" {
				hostname = &s
			}
		case "status":
			if v := a.Value.Any(); v != nil {
				var i int32
				switch val := v.(type) {
				case int:
					i = int32(val)
				case int32:
					i = val
				case int64:
					i = int32(val)
				}
				httpStatus = &i
			}
		}
		return true
	})

	message := r.Message
	queries := db.New(h.pool)
	// Use context.Background() to avoid cancellation issues during shutdown
	err := queries.InsertLog(context.Background(), db.InsertLogParams{
		RequestMethod: requestMethod,
		RequestPath:   requestPath,
		HttpStatus:    httpStatus,
		Url:           url,
		Hostname:      hostname,
		Message:       &message,
		Err:           slogErr,
	})
	// If we fail to write to the database, log it to the parent handler,
	// but don’t propagate the error to avoid infinite loops.
	if err != nil {
		_ = h.parent.Handle(ctx, slog.NewRecord(r.Time, slog.LevelWarn, "Failed to write log to database", r.PC))
	}
}
