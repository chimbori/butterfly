-- name: InsertLog :exec
INSERT INTO logs (
  request_method, request_path, http_status,
  url, hostname, user_agent,
  message, err
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetRecentLogs :many
SELECT * FROM logs
  ORDER BY logged_at DESC
  LIMIT $1;

-- name: DeleteOldLogs :execrows
DELETE FROM logs
  WHERE logged_at < NOW() - $1;
