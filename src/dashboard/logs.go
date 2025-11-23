package dashboard

import (
	"log/slog"
	"net/http"

	"github.com/lmittmann/tint"
	"go.chimbori.app/butterfly/conf"
	"go.chimbori.app/butterfly/db"
)

// GET /dashboard/logs
var logsHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	LogsTempl(conf.AppName).Render(req.Context(), w)
})

// GET /dashboard/logs/data
var logsDataHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	queries := db.New(db.Pool)
	logs, err := queries.GetRecentLogs(req.Context(), 1000)
	if err != nil {
		slog.Error("failed to fetch logs", tint.Err(err))
		ErrorTempl("Failed to fetch logs").Render(req.Context(), w)
		return
	}
	LogsTableTempl(logs).Render(req.Context(), w)
})
