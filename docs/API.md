# Monify External API (v1)

A token-authenticated JSON REST API for integrating external applications
with Monify, separate from the cookie/htmx-based web UI.

Base URL: `https://<your-host>/api/v1`

> The web UI (`/`, `/accounts`, `/transactions`, ...) is a different surface:
> HTML fragments, session cookies, same-origin only. It is not part of this
> API and isn't meant for external clients.

## Authentication

Every request must carry a bearer token:

```
Authorization: Bearer mnfy_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

Tokens are minted from the server (there is no self-service signup endpoint,
matching the app's single-operator `seed-user` model):

```
make api-key EMAIL=you@example.com NAME="my integration"
# or: go run ./cmd/monify apikey create you@example.com "my integration"
```

The plaintext token is printed **once** and cannot be recovered afterwards —
only its hash is stored. Store it securely (e.g. a secrets manager). There is
currently no endpoint to list or revoke keys from the API itself; do that via
direct database access to the `api_keys` table (set `revoked_at`) if a key
needs to be revoked.

A missing or invalid token returns:

```json
HTTP/1.1 401 Unauthorized
{"error": "missing bearer token"}
```
or
```json
{"error": "invalid or revoked api key"}
```

### CORS

`/api/v1/*` sends permissive CORS headers (`Access-Control-Allow-Origin: *`),
so it can be called directly from browser-based external apps. This is safe
because auth here is an explicit bearer token, never an implicit cookie.

## Conventions

- **Money** is always an integer in minor units (cents / *sen*), never a
  float. `amount_minor: 150000` means Rp 1.500,00. Field names ending in
  `_minor` follow this rule everywhere in this API.
- **Dates** are `"YYYY-MM-DD"` strings (no time component). **Months** (used
  for budgets/reports) are `"YYYY-MM"`.
- **Timestamps** (`created_at`, `updated_at`, ...) are RFC 3339.
- **IDs** are UUID strings.
- Request bodies are JSON (`Content-Type: application/json`); unknown fields
  in a request body are rejected.
- Successful responses: `200 OK` (read/update), `201 Created` (create),
  `204 No Content` (delete).
- Errors are always `{"error": "<message>"}` with one of these statuses:

  | Status | Meaning |
  |---|---|
  | 400 | validation error (message explains what's wrong) |
  | 401 | missing/invalid bearer token |
  | 404 | resource not found |
  | 409 | conflict |
  | 500 | internal error |

- Deletes are soft deletes; deleted resources simply stop appearing in list/get.

---

## Accounts

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/accounts` | List all accounts, with computed balance |
| GET | `/api/v1/accounts/{id}` | Get one account |
| POST | `/api/v1/accounts` | Create an account |
| PUT | `/api/v1/accounts/{id}` | Replace an account |
| DELETE | `/api/v1/accounts/{id}` | Soft-delete an account |

`type` is one of `cash`, `bank`, `ewallet`, `credit`.

Request body (POST/PUT):

```json
{
  "name": "Bank BCA",
  "type": "bank",
  "initial_balance_minor": 500000,
  "archived": false
}
```

Response (account object):

```json
{
  "ID": "b1e...", "Name": "Bank BCA", "Type": "bank",
  "InitialBalance": 500000, "Currency": "IDR", "Archived": false,
  "CreatedAt": "2026-09-01T00:00:00Z", "UpdatedAt": "2026-09-01T00:00:00Z",
  "Balance": 1250000
}
```

`Balance` = `InitialBalance` plus the net effect of every posted transaction
touching this account; it's only meaningful on read, ignored on write.

---

## Categories

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/categories` | List all categories |
| GET | `/api/v1/categories/{id}` | Get one category |
| POST | `/api/v1/categories` | Create a category |
| PUT | `/api/v1/categories/{id}` | Replace a category |
| DELETE | `/api/v1/categories/{id}` | Soft-delete a category |

`kind` is `income` or `expense`. `parent_id` is `""`/omitted for a top-level
category, or an existing category's id to nest under it (a category cannot
be its own parent).

Request body:

```json
{ "name": "Makan", "kind": "expense", "parent_id": "" }
```

---

## Transactions

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/transactions` | List transactions (filtered, paginated) |
| GET | `/api/v1/transactions/{id}` | Get one transaction |
| POST | `/api/v1/transactions` | Create a transaction |
| PUT | `/api/v1/transactions/{id}` | Replace a transaction |
| DELETE | `/api/v1/transactions/{id}` | Soft-delete a transaction |

`kind` is `income`, `expense`, or `transfer`.

### List query params

| Param | Meaning |
|---|---|
| `from`, `to` | `YYYY-MM-DD`, inclusive date range |
| `account_id` | filter to one account (either side of a transfer) |
| `category_id` | filter to one category |
| `kind` | `income` \| `expense` \| `transfer` |
| `limit` | page size, default 25, max 200 |
| `offset` | default 0 |

Response:

```json
{
  "transactions": [ { "ID": "...", "Kind": "expense", "AmountMinor": 50000,
                       "AccountID": "...", "CategoryID": "...",
                       "OccurredAt": "2026-09-10T00:00:00Z", "Note": "lunch",
                       "AccountName": "Dompet Tunai", "CategoryName": "Makan" } ],
  "total": 137,
  "limit": 25,
  "offset": 0
}
```

### Create / update body

```json
{
  "kind": "expense",
  "amount_minor": 50000,
  "account_id": "b1e...",
  "category_id": "c2f...",
  "occurred_at": "2026-09-10",
  "note": "lunch"
}
```

For `kind: "transfer"`: omit `category_id`, include `to_account_id` (must
differ from `account_id`). `amount_minor` must be > 0.

---

## Budgets

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/budgets?month=YYYY-MM` | Budget vs. actual spend for every category in a month |
| POST | `/api/v1/budgets` | Set (upsert) one category's budget for a month |

`month` defaults to the current month if omitted.

GET response:

```json
{
  "month": "2026-09",
  "lines": [
    { "CategoryID": "...", "CategoryName": "Makan", "AmountMinor": 2000000, "SpentMinor": 850000 }
  ]
}
```

POST body:

```json
{ "category_id": "c2f...", "month": "2026-09", "amount_minor": 2000000 }
```

---

## Wishlist

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/wishlist?mode=parallel\|sequential` | List items with affordability estimates |
| GET | `/api/v1/wishlist/{id}` | Get one item |
| POST | `/api/v1/wishlist` | Create an item |
| PUT | `/api/v1/wishlist/{id}` | Replace an item |
| POST | `/api/v1/wishlist/{id}/save` | Add (or, with a negative value, remove) saved amount |
| POST | `/api/v1/wishlist/{id}/status` | Change status |
| DELETE | `/api/v1/wishlist/{id}` | Soft-delete an item |

`status` is `active`, `bought`, or `archived`. `mode=sequential` estimates
items being funded one after another (in priority order) from one shared
monthly surplus; the default `parallel` estimates each item independently.

Create/update body:

```json
{
  "name": "Kamera Mirrorless",
  "target_minor": 15000000,
  "saved_minor": 2000000,
  "priority": 100,
  "url": "https://example.com/product",
  "status": "active",
  "target_date": "2027-01-01"
}
```

`POST .../save` body: `{ "amount_minor": 500000 }`
`POST .../status` body: `{ "status": "bought" }`

List response wraps each item in an estimate:

```json
{
  "Estimates": [
    {
      "Item": { "ID": "...", "Name": "Kamera Mirrorless", "TargetAmountMinor": 15000000,
                "SavedAmountMinor": 2000000, "Status": "active", ... },
      "Estimable": true,
      "AlreadyFunded": false,
      "MonthsNeeded": 9,
      "ReadyDate": "2027-06-01T00:00:00Z",
      "CumulativeRemaining": 13000000
    }
  ],
  "AvgSurplus": 1500000,
  "Sequential": false
}
```

`Estimable: false` means the rolling 3-month average surplus is ≤ 0, so no
completion date can be projected.

---

## Reports / Dashboard

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/dashboard?month=YYYY-MM` | Summary figures for one month |

`month` defaults to the current month.

```json
{
  "summary": {
    "Month": "2026-09-01T00:00:00Z",
    "TotalBalance": 12500000,
    "MonthIncome": 8000000,
    "MonthExpense": 5200000,
    "TopExpenses": [ { "CategoryID": "...", "CategoryName": "Makan", "AmountMinor": 2100000 } ],
    "Budgets": [ { "CategoryID": "...", "CategoryName": "Makan", "AmountMinor": 2000000, "SpentMinor": 2100000 } ]
  },
  "avg_monthly_surplus_minor": 1500000
}
```

`TotalBalance` is across all accounts, as of now (not scoped to `month`).
`avg_monthly_surplus_minor` is a rolling 3-month average of income minus
expense.

---

## Not yet exposed

Recurring rules and CSV import/export are only available through the
htmx web UI today, not this API. They can be added the same way as the
resources above (`internal/service` already has the business logic;
`internal/http/apihandler` just needs a thin JSON wrapper) if external
integration needs them — ask if that's required.

## Example

```bash
curl -s https://your-host/api/v1/transactions \
  -H "Authorization: Bearer mnfy_..." \
  -H "Content-Type: application/json" \
  -d '{"kind":"expense","amount_minor":50000,"account_id":"<account-uuid>",
       "category_id":"<category-uuid>","occurred_at":"2026-09-16","note":"lunch"}'
```

## Implementation reference

- Routes: `internal/http/router.go` (`/api/v1` group)
- Handlers: `internal/http/apihandler/*.go`
- Auth middleware: `internal/http/middleware/auth.go` (`RequireAPIKey`, `CORS`)
- Token generation/hashing: `internal/platform/apikey/apikey.go`
- Key issuance: `cmd/monify apikey create` (`cmd/monify/main.go`)
