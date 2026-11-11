CREATE TABLE IF NOT EXISTS links (
    id           BIGSERIAL PRIMARY KEY,
    code         TEXT NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_links_code ON links(code);

CREATE TABLE IF NOT EXISTS clicks (
    id         BIGSERIAL PRIMARY KEY,
    link_id    BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    user_agent TEXT
);

CREATE INDEX IF NOT EXISTS idx_clicks_link_id_clicked_at
    ON clicks(link_id, clicked_at DESC);