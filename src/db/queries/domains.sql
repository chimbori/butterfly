-- name: ListDomains :many
SELECT * FROM domains
  ORDER BY domain;

-- name: UpsertDomain :one
INSERT INTO domains (domain, include_subdomains, authorized, updated_at)
  VALUES ($1, $2, $3, NOW())
  ON CONFLICT(domain)
  DO UPDATE SET
    include_subdomains = EXCLUDED.include_subdomains,
    authorized = EXCLUDED.authorized,
    updated_at = NOW()
  RETURNING *;

-- name: InsertUnauthorizedDomain :exec
INSERT INTO domains (domain, include_subdomains, authorized, updated_at)
  VALUES ($1, false, NULL, NOW())
  ON CONFLICT(domain) DO NOTHING;

-- name: DeleteDomain :exec
DELETE FROM domains
  WHERE domain = $1;

-- name: DeleteUnauthorizedStaleDomains :one
WITH deleted AS (
  DELETE FROM domains
    WHERE updated_at < NOW() - $1::interval
    AND authorized IS NULL OR authorized = FALSE
    RETURNING _id
  )
  SELECT COUNT(*) from deleted;

-- name: IsAuthorized :one
SELECT EXISTS (
  SELECT 1 FROM domains
  WHERE (domain ILIKE $1 OR (include_subdomains = true AND $1 ILIKE '%.' || domain))
  AND authorized IS TRUE
);
