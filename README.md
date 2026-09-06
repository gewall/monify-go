# Monify

Aplikasi web pengatur keuangan pribadi. Go + `html/template` + htmx + Alpine.js + PostgreSQL.

## Fitur

- Akun/dompet & kategori bertingkat
- Transaksi: pemasukan, pengeluaran, transfer antar akun
- Budget bulanan per kategori + dashboard (cashflow, pengeluaran terbesar, progress budget)
- Transaksi berulang (langganan/cicilan) + runner otomatis tiap jam + pengingat jatuh tempo
- Wishlist barang: estimasi kapan bisa dibeli dari rata-rata surplus bulanan (mode nabung paralel / beli berurutan)
- Impor & ekspor CSV, laporan bulanan siap cetak/PDF

## Menjalankan (lokal)

Prasyarat: Go 1.23+, Docker, [`goose`](https://github.com/pressly/goose) (`go install github.com/pressly/goose/v3/cmd/goose@latest`).

```sh
cp .env.example .env            # sesuaikan SESSION_SECRET
docker compose --profile local up -d   # Postgres + Adminer
make migrate                   # jalankan migrasi
make seed-user                 # buat user pertama (email + password)
make run                       # http://localhost:8080
```

`make test` — unit test · `make lint` — golangci-lint · `make build` — binary ke `bin/`.

## Arsitektur

```
cmd/monify           entrypoint + wiring + runner recurring
internal/domain      entity murni + estimasi wishlist + sentinel error (zero-dependency)
internal/service     business logic; mendefinisikan interface repo (consumer-side)
internal/postgres    implementasi repo (pgx v5, SQL manual)
internal/http        chi router, handler (balikin HTML fragment), middleware, render
internal/platform    money (int64 sen), session (cookie HMAC)
web/                 template + aset statik (di-embed ke binary)
migrations/          goose SQL
deploy/              compose produksi + Caddy + skrip backup
```

Detail rencana & progres: `CLAUDE.md`.
