package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"chimbori.dev/butterfly/conf"
	"chimbori.dev/butterfly/dashboard"
	"chimbori.dev/butterfly/db"
	"chimbori.dev/butterfly/github"
	"chimbori.dev/butterfly/linkpreviews"
	"chimbori.dev/butterfly/qrcode"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lmittmann/tint"
)

func performMaintenance() {
	ctx := context.Background()
	queries := db.New(db.Pool)

	dashboard.CleanupExpiredSessions()

	// Delete domains that haven’t been triaged in a while; they are blocked by default,
	// and do not need to be included on the dashboard.
	interval := pgtype.Interval{
		Microseconds: 7 * 24 * 60 * 60 * 1000000, // 7 days
		Valid:        true,
	}
	deletedDomains, err := queries.DeleteUnauthorizedStaleDomains(ctx, interval)
	if err != nil {
		slog.Error("failed to delete stale domains", tint.Err(err))
	} else {
		slog.Info(fmt.Sprintf("%d domains deleted", deletedDomains))
	}

	logRetentionInterval := pgtype.Interval{
		Microseconds: int64(conf.Config.Logs.Retention / time.Microsecond),
		Valid:        true,
	}
	deletedLogs, err := queries.DeleteOldLogs(ctx, logRetentionInterval)
	if err != nil {
		slog.Error("failed to delete old logs", tint.Err(err))
	} else {
		slog.Info(fmt.Sprintf("%d logs deleted", deletedLogs))
	}

	// Prune caches
	if linkpreviews.Cache != nil {
		if err := linkpreviews.Cache.Prune(); err != nil {
			slog.Error("failed to prune linkpreviews cache", tint.Err(err))
		}
	}
	if dashboard.ThumbnailCache != nil {
		if err := dashboard.ThumbnailCache.Prune(); err != nil {
			slog.Error("failed to prune linkpreview thumbnail cache", tint.Err(err))
		}
	}
	if qrcode.Cache != nil {
		if err := qrcode.Cache.Prune(); err != nil {
			slog.Error("failed to prune qrcode cache", tint.Err(err))
		}
	}
	if github.Cache != nil {
		if err := github.Cache.Prune(); err != nil {
			slog.Error("failed to prune github cache", tint.Err(err))
		}
	}
	slog.Info("Maintenance completed successfully")
}
