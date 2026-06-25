CREATE TABLE IF NOT EXISTS links (
    code         VARCHAR(6)   NOT NULL UNIQUE,
    original_url TEXT         NOT NULL UNIQUE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);