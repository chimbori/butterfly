package dashboard

import (
	"log/slog"
	"net/http"
	"strconv"

	"butterfly.chimbori.dev/conf"
	"butterfly.chimbori.dev/db"
	"github.com/lmittmann/tint"
)

// GET /dashboard/logs
func logsHandler(w http.ResponseWriter, req *http.Request) {
	page := 1
	if pageStr := req.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	LogsTempl(conf.AppName, page).Render(req.Context(), w)
}

// GET /dashboard/logs/data
func logsDataHandler(w http.ResponseWriter, req *http.Request) {
	queries := db.New(db.Pool)

	page := 1
	if pageStr := req.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Fetch total count for pagination
	totalCount, err := queries.CountLogs(req.Context())
	if err != nil {
		slog.Error("failed to count logs", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch paginated items
	limit := int32(conf.Config.Logs.Pagination.Limit)
	offset := int32((page - 1) * int(limit))
	logs, err := queries.GetRecentLogsPaginated(req.Context(), db.GetRecentLogsPaginatedParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		slog.Error("failed to fetch logs", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	LogsListTempl(logs, page, totalCount).Render(req.Context(), w)
}
