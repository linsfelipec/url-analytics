CREATE TABLE urls (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(16) NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT urls_code_not_empty CHECK (code <> ''),
    CONSTRAINT urls_original_url_not_empty CHECK (original_url <> '')
);
