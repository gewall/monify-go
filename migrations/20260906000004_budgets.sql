-- +goose Up
CREATE TABLE budgets (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id  UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    period       DATE NOT NULL, -- first day of the budgeted month
    amount_minor BIGINT NOT NULL CHECK (amount_minor >= 0),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,
    UNIQUE (category_id, period)
);

-- +goose Down
DROP TABLE budgets;
