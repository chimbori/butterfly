-- +goose Up

ALTER TABLE link_previews ADD COLUMN canonical_user_agent TEXT;

CREATE INDEX idx_link_previews_canonical_user_agent ON link_previews(canonical_user_agent);
