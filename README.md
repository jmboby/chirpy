# chirpy

Twitter-like Go app. It serves a JSON API on port 8080 backed by PostgreSQL.

## Prerequisites

- Go 1.26 or newer
- Docker with Compose
- [goose](https://github.com/pressly/goose) for schema migrations

Install goose with Go.

```sh
go install github.com/pressly/goose/v3/cmd/goose@latest
```

If `goose --version` then reports command not found, add the Go bin directory to your PATH.

```sh
echo 'export PATH="$(go env GOPATH)/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

## First run

Copy the example env file and set a real password. Both Docker Compose and the app read `.env`.

```sh
cp .env.example .env
```

Start Postgres in the background.

```sh
docker compose up -d
```

Apply the schema. The first line exports every variable in `.env` into your shell so goose can read `DB_URL`.

```sh
set -a; source .env; set +a
goose -dir sql/schema postgres "$DB_URL" up
```

Download Go dependencies and start the app. The app loads `.env` itself.

```sh
go mod download
go run .
```

Check it is up.

```sh
./scripts/health-check.sh
```

## Day to day

Start the database and app.

```sh
docker compose up -d
go run .
```

Stop with Ctrl-C, then `docker compose down`. Add `-v` to wipe the database volume.

## Environment variables

| Variable | Used by | Purpose |
|---|---|---|
| `DB_URL` | app | PostgreSQL connection string |
| `PLATFORM` | app | `dev` enables `POST /admin/reset`. Any other value disables it |
| `POSTGRES_USER` | compose | Database superuser |
| `POSTGRES_PASSWORD` | compose | Superuser password. Must match `DB_URL` |
| `POSTGRES_DB` | compose | Database name. Must match `DB_URL` |

## Database changes

Migrations live in `sql/schema` as goose files. Queries live in `sql/queries`.

After changing either, regenerate the Go code with sqlc.

```sh
make sql-regenerate
```

## Testing

Unit tests run with `go test ./...`. The Makefile has curl targets against a running app.

```sh
make test
```
