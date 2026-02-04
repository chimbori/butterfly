-- name: ListLinkPreviews :many
SELECT * FROM link_previews
  ORDER BY last_accessed_at DESC;

-- name: GetLinkPreview :one
SELECT * FROM link_previews
  WHERE url = $1;

-- name: DeleteLinkPreview :exec
DELETE FROM link_previews
  WHERE url = $1;

-- name: DeleteAllLinkPreviews :exec
DELETE FROM link_previews;

-- name: RecordLinkPreviewCreated :exec
INSERT INTO link_previews (url, generated_at, last_accessed_at, access_count)
  VALUES ($1, NOW(), NOW(), 1)
  ON CONFLICT(url)
  DO UPDATE SET
    generated_at = NOW(),
    last_accessed_at = NOW(),
    access_count = link_previews.access_count + 1
  RETURNING *;

-- name: RecordLinkPreviewAccessed :execrows
UPDATE link_previews
  SET last_accessed_at = NOW(),
    access_count = access_count + 1
  WHERE url = $1;
