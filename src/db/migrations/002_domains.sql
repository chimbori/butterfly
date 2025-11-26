-- +goose Up

CREATE TABLE domains (
  _id                BIGSERIAL PRIMARY KEY,
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  domain             TEXT UNIQUE NOT NULL,

  include_subdomains BOOLEAN DEFAULT FALSE,

  -- authorized is a tri-state:
  -- - NULL: no value set, implicitly unauthorized, but considered untriaged.
  -- - TRUE: explicitly authorized.
  -- - FALSE: explicitly unauthorized.
  authorized         BOOLEAN DEFAULT NULL
);
