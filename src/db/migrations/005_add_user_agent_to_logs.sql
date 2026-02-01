-- +goose Up

ALTER TABLE logs ADD COLUMN user_agent TEXT;

CREATE INDEX idx_logs_user_agent ON logs(user_agent);
