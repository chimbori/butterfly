-- +goose Up

CREATE TABLE logs (
  _id             BIGSERIAL PRIMARY KEY,
  logged_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  request_method  TEXT,
  request_path    TEXT,
  http_status     INTEGER,
  url             TEXT,
  hostname        TEXT,
  message         TEXT,
  err             TEXT
);

CREATE INDEX idx_logs_logged_at ON logs(logged_at DESC);

CREATE INDEX idx_logs_request_method ON logs(request_method);
CREATE INDEX idx_logs_request_path ON logs(request_path);
CREATE INDEX idx_logs_http_status ON logs(http_status);
CREATE INDEX idx_logs_hostname ON logs(hostname);
