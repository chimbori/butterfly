-- +goose Up

CREATE TABLE qr_codes (
  _id               BIGSERIAL PRIMARY KEY,
  url               TEXT UNIQUE NOT NULL,
  generated_at      TIMESTAMPTZ DEFAULT NOW(),
  last_accessed_at  TIMESTAMPTZ DEFAULT NOW(),
  access_count      INTEGER DEFAULT 1
);

CREATE INDEX idx_qr_codes_url ON qr_codes(url);
CREATE INDEX idx_qr_codes_generated_at ON qr_codes(generated_at DESC);
CREATE INDEX idx_qr_codes_last_accessed_at ON qr_codes(last_accessed_at DESC);
CREATE INDEX idx_qr_codes_access_count ON qr_codes(access_count DESC);
