# Monify

Personal finance web app. Go + html/template + htmx + Alpine.js + PostgreSQL.

## Architecture

Layered / hexagonal-lite:
- `internal/domain` — pure entities + sentinel errors, zero dependencies
- `internal/service` — business logic; defines repo interfaces (consumer-side) in `ports.go`
- `internal/postgres` — repo implementations (pgx v5 + sqlc)
- `internal/http` — chi router, handlers return HTML fragments
- `internal/platform` — money type, session, validate helpers
- `web/` — templates + vendored static assets (no CDN)
- `migrations/` — goose SQL migrations

`cmd/monify/main.go` wires concrete implementations into services (manual DI).

## Rules

- Money is always `int64` minor units (cents). Never float.
- Repo interfaces live in `service`, not `postgres`.
- Handlers return HTML fragments; full page when `HX-Request` header absent.
- Alpine.js only for pure UI state (modals, toggles). Never for data fetching.
- All tables carry `id uuid`, `created_at`, `updated_at`, `deleted_at` (soft delete) for future sync.

## Commands

- `make up` — start Postgres + Adminer
- `make migrate` — run migrations (needs `goose`)
- `make seed-user` — create the first user
- `make run` / `make test` / `make lint`

## Plan

Full phased plan: `C:\Users\USER\.claude\plans\sequential-gliding-knuth.md`

Done:
- Fase 0 — server, config, pgxpool, /healthz, graceful shutdown
- Fase 1 — auth (bcrypt), HMAC signed-cookie session, login/logout, base layout, embedded templates, vendored htmx 2.0.4 + Alpine 3.14.8, SameOrigin CSRF guard
- Fase 2 — Accounts & Categories CRUD (the repo→service→handler→partial pattern all later resources copy; htmx inline-edit row swap)
- Fase 3 — Transactions: income/expense/transfer (transfer is one row touching two accounts, shape enforced by CHECK), filters + pagination, balances computed in SQL (never cached)
- Fase 4 — Dashboard (stat cards, top-expense bars, budget progress, avg monthly surplus) + Budgets per category/month (upsert)
- Fase 5 — Recurring rules + hourly in-process runner (pg advisory lock for idempotency) + "run now" + due-soon widget
- Fase 6 — Wishlist + affordability estimate (`domain.EstimateWishlist`): parallel vs sequential mode, "allocate savings", mark bought
- Fase 7 — CSV import (upload → preview with dup detection → commit selected) + CSV export + print-friendly /reports
- Fase 9 — Dockerfile (multi-stage, embeds web/), deploy/ (prod compose + Caddy + pg_dump backup.sh)

Deferred: Fase 8 (offline/PWA + sync) — user chose "online only dulu". Entities already carry uuid + updated_at + deleted_at so it can be added without migration churn.

## Test data

`testdata/sample-transactions.csv` — matches the import format; needs accounts "Bank BCA"/"Dompet Tunai" and categories "Gaji"/"Makan"/"Transport"/"Langganan" to exist first.

## Toolchain note

pgx v5 latest needs Go 1.25 but this machine has 1.23.5 and toolchain auto-download is broken.
Deps are pinned to 1.23-compatible versions and `GOTOOLCHAIN=local` is set in the Makefile + CI.
Pinned: `pgx/v5` v5.7.4, `x/term` v0.27.0, `goose/v3` v3.22.1.
Do NOT run `go mod tidy` — it upgrades goose/pgx past Go 1.23. Use `go get <pkg>@<ver>` + `go build`.

## Deploy

- `cmd/monify migrate` runs the embedded (`migrations/*.sql`) goose migrations; `cmd/monify seed-user` creates the first user.
- CI `.github/workflows/build.yml`: test job (vet/test/build) then image job pushes `ghcr.io/<owner>/monify` (`:latest`, `:sha-xxxxxxx`, `:x.y.z` on tags) on push to default branch / `v*` tags.
- VPS pulls via `deploy/docker-compose.prod.yml` (app + one-shot migrate + Postgres + Caddy). See `deploy/README.md`.
