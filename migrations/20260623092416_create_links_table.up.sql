CREATE TABLE IF NOT EXISTS links (
    code         VARCHAR(6)   NOT NULL UNIQUE,
    original_url TEXT         NOT NULL UNIQUE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_links_code         ON links (code);
CREATE INDEX IF NOT EXISTS idx_links_original_url ON links (original_url);
