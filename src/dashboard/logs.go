package dashboard

import (
	"log/slog"
	"net/http"

	"chimbori.dev/butterfly/conf"
	"chimbori.dev/butterfly/db"
	"github.com/lmittmann/tint"
)

// GET /dashboard/logs
func logsHandler(w http.ResponseWriter, req *http.Request) {
	LogsTempl(conf.AppName).Render(req.Context(), w)
}

// GET /dashboard/logs/data
func logsDataHandler(w http.ResponseWriter, req *http.Request) {
	queries := db.New(db.Pool)
	logs, err := queries.GetRecentLogs(req.Context(), 1000)
	if err != nil {
		slog.Error("failed to fetch logs", tint.Err(err))
		ErrorTempl("Failed to fetch logs").Render(req.Context(), w)
		return
	}
	LogsTableTempl(logs).Render(req.Context(), w)
}
