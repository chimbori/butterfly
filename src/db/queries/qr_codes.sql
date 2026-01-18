-- name: ListQrCodes :many
SELECT * FROM qr_codes
  ORDER BY last_accessed_at DESC;

-- name: GetQrCode :one
SELECT * FROM qr_codes
  WHERE url = $1;

-- name: DeleteQrCode :exec
DELETE FROM qr_codes
  WHERE url = $1;

-- name: DeleteAllQrCodes :exec
DELETE FROM qr_codes;

-- name: RecordQrCodeCreated :exec
INSERT INTO qr_codes (url, generated_at, last_accessed_at, access_count)
  VALUES ($1, NOW(), NOW(), 1)
  ON CONFLICT(url)
  DO UPDATE SET
    generated_at = NOW(),
    last_accessed_at = NOW(),
    access_count = qr_codes.access_count + 1
  RETURNING *;

-- name: RecordQrCodeAccessed :exec
UPDATE qr_codes
  SET last_accessed_at = NOW(),
    access_count = access_count + 1
  WHERE url = $1;
