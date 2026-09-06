-- +goose Up
CREATE TABLE transactions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind          TEXT NOT NULL CHECK (kind IN ('income','expense','transfer')),
    amount_minor  BIGINT NOT NULL CHECK (amount_minor > 0),
    account_id    UUID NOT NULL REFERENCES accounts(id),
    to_account_id UUID REFERENCES accounts(id),
    category_id   UUID REFERENCES categories(id),
    occurred_at   DATE NOT NULL,
    note          TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ,
    CONSTRAINT transactions_shape CHECK (
        (kind = 'transfer' AND to_account_id IS NOT NULL AND to_account_id <> account_id AND category_id IS NULL)
        OR
        (kind <> 'transfer' AND to_account_id IS NULL)
    )
);
CREATE INDEX transactions_occurred_idx ON transactions(occurred_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX transactions_account_idx  ON transactions(account_id)        WHERE deleted_at IS NULL;
CREATE INDEX transactions_category_idx ON transactions(category_id)       WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE transactions;
