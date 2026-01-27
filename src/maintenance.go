package main

import (
	"context"
	"fmt"
	"log/slog"

	"chimbori.dev/butterfly/db"
	"chimbori.dev/butterfly/github"
	"chimbori.dev/butterfly/linkpreview"
	"chimbori.dev/butterfly/qrcode"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lmittmann/tint"
)

func performMaintenance() {
	ctx := context.Background()
	queries := db.New(db.Pool)

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

	// Prune caches
	if linkpreview.Cache != nil {
		if err := linkpreview.Cache.Prune(); err != nil {
			slog.Error("failed to prune linkpreview cache", tint.Err(err))
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
