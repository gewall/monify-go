# Deploy (VPS)

CI (`.github/workflows/build.yml`) builds and pushes `ghcr.io/gewall/monify-go`
on every push to the default branch (`:latest` + `:sha-xxxxxxx`) and on `v*` tags
(`:1.2.3`).

## Host layout on the VPS

- **Postgres**: shared stack at `~/infra/postgres` (container `shared-postgres-postgres-1`,
  external docker network `shared-db`). The `monify` database + role were created with
  `~/infra/postgres/scripts/create-project-db.sh monify`.
- **TLS + routing**: host `nginx` + `certbot` (same as other sites, e.g. readmeplz).
- **This app**: `~/apps/monify/` holds `docker-compose.prod.yml` + `.env`. The `app`
  container publishes `127.0.0.1:8090` only; nginx terminates TLS and proxies to it.

## First time

```sh
mkdir -p ~/apps/monify && cd ~/apps/monify
# copy docker-compose.prod.yml + .env.example here (scp or from a checkout)
cp .env.example .env && vim .env     # set DATABASE_URL password + SESSION_SECRET

# GHCR package is private — log in once (PAT needs read:packages):
echo "$GHCR_PAT" | docker login ghcr.io -u gewall --password-stdin

docker compose -f docker-compose.prod.yml up -d
```

`migrate` runs the embedded goose migrations against the shared DB before `app` starts.

Create the first user (interactive, needs a TTY):

```sh
docker compose -f docker-compose.prod.yml run --rm app seed-user
```

### nginx + TLS

```sh
sudo cp <checkout>/deploy/nginx/monify.gewall.my.id.conf \
        /etc/nginx/sites-available/monify.gewall.my.id
sudo ln -s /etc/nginx/sites-available/monify.gewall.my.id /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
sudo certbot --nginx -d monify.gewall.my.id
```

Verify: `curl https://monify.gewall.my.id/healthz` → `{"status":"ok"}`.

## Update to the latest build

```sh
cd ~/apps/monify
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d      # re-runs migrate, restarts app
```

Pin a specific build: set `MONIFY_IMAGE=ghcr.io/gewall/monify-go:sha-abc1234` in `.env`.

### Auto-deploy

`auto-pull.sh` polls GHCR and redeploys only when the `:latest` image actually
changed (silent no-op otherwise). Installed via cron on the VPS, every 5 min:

```sh
crontab -e
# add:
*/5 * * * * /home/gewall/apps/monify/auto-pull.sh >> /var/log/monify-autopull.log 2>&1
```

## Backups

The shared Postgres stack already runs a nightly `pg_dumpall`. `backup.sh` is an
optional extra monify-only `pg_dump -Fc` (see the cron line in the file).
