-- +goose Up

CREATE TABLE link_previews (
  _id               BIGSERIAL PRIMARY KEY,
  url               TEXT UNIQUE NOT NULL,
  generated_at      TIMESTAMPTZ DEFAULT NOW(),
  last_accessed_at  TIMESTAMPTZ DEFAULT NOW(),
  access_count      INTEGER DEFAULT 1
);

CREATE INDEX idx_link_previews_url ON link_previews(url);
CREATE INDEX idx_link_previews_generated_at ON link_previews(generated_at DESC);
CREATE INDEX idx_link_previews_last_accessed_at ON link_previews(last_accessed_at DESC);
CREATE INDEX idx_link_previews_access_count ON link_previews(access_count DESC);
