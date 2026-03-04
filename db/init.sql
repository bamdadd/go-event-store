CREATE TABLE IF NOT EXISTS events (
    sequence_number BIGSERIAL PRIMARY KEY,
    id              TEXT NOT NULL,
    name            TEXT NOT NULL,
    stream          TEXT NOT NULL,
    category        TEXT NOT NULL,
    position        INTEGER NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}',
    observed_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (category, stream, position)
);

CREATE INDEX IF NOT EXISTS idx_events_category ON events (category);
CREATE INDEX IF NOT EXISTS idx_events_stream ON events (category, stream);
