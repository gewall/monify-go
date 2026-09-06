# Deploy (VPS)

CI (`.github/workflows/build.yml`) builds and pushes `ghcr.io/<owner>/monify`
on every push to the default branch (`:latest` + `:sha-xxxxxxx`) and on `v*` tags
(`:1.2.3`).

## First time

```sh
git clone <repo> monify && cd monify/deploy
cp .env.example .env && edit .env          # set passwords + SESSION_SECRET

# If the GHCR package is private, log in once:
echo $GHCR_PAT | docker login ghcr.io -u <github-user> --password-stdin

# point deploy/Caddyfile at your real domain, then:
docker compose -f docker-compose.prod.yml up -d
```

The `migrate` service runs the embedded goose migrations before `app` starts.
Create the first user:

```sh
docker compose -f docker-compose.prod.yml run --rm app seed-user
```

## Update to the latest build

```sh
cd monify/deploy
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d      # re-runs migrate, restarts app
```

Pin a specific build instead of `latest`: set `MONIFY_IMAGE=ghcr.io/<owner>/monify:sha-abc1234` in `.env`.

## Backups

`backup.sh` — nightly `pg_dump` (see cron line in the file).
