-- +goose Up
CREATE TABLE wishlist_items (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                TEXT NOT NULL,
    target_amount_minor BIGINT NOT NULL CHECK (target_amount_minor > 0),
    saved_amount_minor  BIGINT NOT NULL DEFAULT 0 CHECK (saved_amount_minor >= 0),
    priority            INT NOT NULL DEFAULT 100,
    target_date         DATE,
    url                 TEXT NOT NULL DEFAULT '',
    status              TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','bought','archived')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at          TIMESTAMPTZ
);
CREATE INDEX wishlist_active_idx ON wishlist_items(priority, target_date)
    WHERE deleted_at IS NULL AND status = 'active';

-- +goose Down
DROP TABLE wishlist_items;
