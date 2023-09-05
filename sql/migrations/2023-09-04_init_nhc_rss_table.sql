CREATE SCHEMA IF NOT EXISTS odin;
CREATE TABLE IF NOT EXISTS odin.nhc_rss (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    storm_id TEXT NOT NULL,
    adv_num INTEGER NOT NULL,
    parsed JSONB NOT NULL,
    processed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_odin_unique_storm ON odin.nhc_rss (storm_id, adv_num);