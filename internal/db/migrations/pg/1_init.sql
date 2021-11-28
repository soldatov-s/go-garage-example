-- +goose Up
CREATE TABLE IF NOT EXISTS production.test (
    id SERIAL PRIMARY KEY,
    code text NOT NULL,
    meta jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

-- +goose Down
DROP TABLE IF EXISTS production.test;
