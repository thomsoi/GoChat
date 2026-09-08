CREATE TABLE rooms (
    id BIGSERIAL PRIMARY KEY,
    room_name TEXT NOT NULL,
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);