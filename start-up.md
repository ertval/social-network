# Start-up

## Dependencies

- [Go](https://go.dev) (1.22+)
- [Docker](https://docker.com) & Docker Compose
- [OpenSSL](https://openssl.org) (for TLS cert generation)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI (`migrate`)

## Environment variables

Export all variables before running any target. The project expects `.env` files at the root, `db/postgres/.env`, and `db/redis/.env`:

```sh
set -a && source .env && set +a
set -a && source db/postgres/.env && set +a
set -a && source db/redis/.env && set +a
```

## Running

### Quick start (certs + Postgres + Redis + migrations + server)

```sh
make run-all
```

### Step by step

```sh
make certs          # generate TLS certificates
make postgres-up    # start Postgres container
make redis-up       # start Redis container
make migrate-up     # run DB migrations
make run            # start the Go server
```

## Useful commands

| Target | Description |
|---|---|
| `make certs` | Generate TLS certificates |
| `make postgres-up/down` | Start/stop Postgres container |
| `make redis-up/down` | Start/stop Redis container |
| `make containers-down` | Stop both containers |
| `make containers-delete` | Stop & remove volumes |
| `make migrate-up/down` | Apply/rollback migrations |
| `make migrate-version` | Show current migration version |
| `make run` | Start the Go server |
| `make t` | Fuzzy-find and run a target (requires `fzf` — see below) |

## Optional convenience feature

- **[fzf](https://github.com/junegunn/fzf)** — general-purpose fuzzy finder (probably already present on linux distros). Not required to run the app. If installed, `make t` pipes all available targets into an interactive `fzf` picker, letting you fuzzy-search and execute a target without typing its full name.
