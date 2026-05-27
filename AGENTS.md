# AGENTS.md

## What This Is

Boule (`bo`) — a CLI message bus backed by SQLite for agent-to-agent and agent-to-human messaging. Part of the ttal toolchain's manager plane.

## Commands

```
make build    # → ./bo binary
make install  # → $GOPATH/bin/bo
make test     # go test ./...
make lint     # golangci-lint run ./...
make fmt      # gofmt -w .
```

### CLI

```
bo send --from <sender> <recipient>   # send message from stdin
bo inbox <recipient> [--json]         # unread messages
bo feed [--from] [--to] [--unread] [--json]  # all messages, optional filters
bo read <id> [--json]                 # mark message as read
bo entities [--json]                  # list all known sender/receiver identities
```

Default DB: `~/.boule/boule.db` (auto-created on first run, auto-migrated).
Override with `bo --db /path/to/db`.

## Architecture

```
main.go → cmd.Execute() → cobra command tree
cmd/                    # CLI command handlers (send, inbox, feed, read, entities)
internal/db/            # sqlc-generated — DO NOT EDIT BY HAND
internal/store/         # Open(), Migrate(), DefaultPath()
db/migrations/          # SQL schema (also DO NOT EDIT — it's the sqlc source of truth)
db/queries/             # sqlc query definitions (edit here, then regenerate)
```

Single-table design: `messages(id, sender, recipient, content, read, created_at)`. The `read` column is an integer boolean (0/1).

## sqlc Workflow

Generated code lives in `internal/db/`. To change queries:

1. Edit `db/queries/messages.sql` (query definitions) or `db/migrations/001_create_messages.sql` (schema)
2. Run `sqlc generate`
3. Generated Go appears in `internal/db/`

The `GetFeed` query uses sqlc's positional-parameter filtering pattern (`? = '' OR sender = ?`) which produces `Column1`/`Column3`/`Column5` fields in the params struct — these are the filter toggle values (empty string = no filter, 0 = no unread filter).

## Testing

Tests in `cmd/commands_test.go` rebuild each command independently via a `testEnv` struct (does not reuse the production root command). Each test gets a fresh temp SQLite database. Pattern:

```go
env := newTestEnv(t)
output, err := env.run("inbox", "bob")
// or with stdin:
output, err := env.runWithStdin("hello", "send", "--from", "alice", "bob")
```

`internal/store/store_test.go` tests migration idempotency and db open.

## Gotchas

- `internal/db/` is entirely sqlc-generated. Editing it by hand will be overwritten on next `sqlc generate`.
- `db/migrations/*.sql` is also a sqlc source — regenerate after editing, don't assume it's just migration files.
- `store.Open()` auto-runs migrations on every call. No separate migrate step needed.
- `store.Open()` sets `MaxOpenConns(1)` — SQLite is single-writer. Don't add concurrent write paths without changing this.
- No CGO — uses `modernc.org/sqlite` (pure Go). Works in minimal containers without gcc.
- `lefthook.yml` runs gofmt + goimports on pre-commit, golangci-lint on pre-push.
- The `--db` flag overrides the default path (`~/.boule/boule.db`). Useful for testing with `bo --db /tmp/test.db`.
- Commit convention: `feat(x):`, `fix(x):`, `refactor(x):`, `chore(x):` prefix required.
