package dashboard

import (
	"context"
	"fmt"
	"log/slog"

	"chimbori.dev/butterfly/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lmittmann/tint"
)

func PerformMaintenance() {
	ctx := context.Background()
	queries := db.New(db.Pool)

	// Delete domains that haven’t been updated in 24 hours; they are blocked by default,
	// and do not need to be included on the dashboard.
	interval := pgtype.Interval{
		Microseconds: 24 * 60 * 60 * 1000000, // 24 hours in microseconds
		Valid:        true,
	}
	deletedDomains, err := queries.DeleteUnauthorizedStaleDomains(ctx, interval)
	if err != nil {
		slog.Error("failed to delete stale domains", tint.Err(err))
		return
	}

	slog.Info(fmt.Sprintf("Maintenance completed successfully; %d domains deleted", deletedDomains))
}
