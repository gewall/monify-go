-- +goose Up
CREATE TABLE recurring_rules (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT NOT NULL,
    kind          TEXT NOT NULL CHECK (kind IN ('income','expense','transfer')),
    amount_minor  BIGINT NOT NULL CHECK (amount_minor > 0),
    account_id    UUID NOT NULL REFERENCES accounts(id),
    to_account_id UUID REFERENCES accounts(id),
    category_id   UUID REFERENCES categories(id),
    note          TEXT NOT NULL DEFAULT '',
    freq          TEXT NOT NULL CHECK (freq IN ('daily','weekly','monthly')),
    interval      INT NOT NULL DEFAULT 1 CHECK (interval > 0),
    day_of_month  INT CHECK (day_of_month BETWEEN 1 AND 31),
    next_run_at   DATE NOT NULL,
    end_at        DATE,
    active        BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);
CREATE INDEX recurring_due_idx ON recurring_rules(next_run_at)
    WHERE deleted_at IS NULL AND active;

-- +goose Down
DROP TABLE recurring_rules;
