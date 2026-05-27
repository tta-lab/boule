# Handoff: Update AGENTS.md

## Context

AGENTS.md was created by the human but is mounted read-only in agent sessions (attachment). A coder with direct file access needs to apply two updates.

## Changes

### 1. Add CLI section after Commands

After the closing triple-backtick of the `## Commands` code block, insert:

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

### 2. Update Gotchas

In `## Gotchas`, find the bullet:

    `store.Open()` sets `MaxOpenConns(1)` -- SQLite is single-writer. Don't add concurrent write paths without changing this.

Insert a new bullet **before** it:

    `store.Open()` auto-runs migrations on every call. No separate migrate step needed.

## Verification

After editing, confirm the file has 8 sections (the original 6 + new `### CLI` subsection) and that `make test` still passes from the project root.
