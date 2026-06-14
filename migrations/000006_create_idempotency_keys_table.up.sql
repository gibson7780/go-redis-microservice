CREATE TABLE idempotency_keys (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key         VARCHAR(255) NOT NULL UNIQUE,  -- idem key，加 unique constraint
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status      TEXT NOT NULL DEFAULT 'in_flight',  -- in_flight, complete, failed
    response    JSONB,  -- 儲存成功的 response
    lease       TIMESTAMPTZ,  -- now() + 30s
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);