# lazycurl

A Go TUI that aims to be a Postman/Insomnia-like client while keeping vim-style
navigation. The product goal, coding guidelines and guardrails are in
[`.ai/system-prompt.md`](.ai/system-prompt.md); agent-specific context is in
[`AGENTS.md`](AGENTS.md).

## Current features

- Home screen with inline key hints and a scrollable `?` keybinds popup.
- Collections browser: navigate with `j`/`k`, fuzzy-filter with `/`, create with
  `n`, delete with `ctrl+d`, rename with `ctrl+e`, dismiss with `esc`.
- Create and rename prompts, delete/rename confirmations, and a status line that
  reports failed database operations.

Request editing, sending HTTP calls and viewing responses are **not implemented
yet**; only collection management works today.

## Keys

| Key | Action |
|---|---|
| `?` | keybinds popup |
| `c` | collections popup |
| `q` / `ctrl+c` | quit (home screen only) |
| `j` / `down`, `k` / `up` | move the collection cursor |
| `/` | filter collections (`enter` keeps the filter, `esc` clears it) |
| `n` | new collection |
| `ctrl+d` | delete selected collection |
| `ctrl+e` | rename selected collection |
| `esc` / `ctrl+c` | close the active popup or prompt |

## Develop

- `go build ./...` — build every package.
- `go test ./... -cover` — run tests with per-package coverage; the project
  target is at least 70%.
- `air` — rebuild and restart on source changes using `air.toml`; output goes to
  `bin/`.

The database is `tmp/lazycurl.db`, created and migrated on startup. Both `bin/`
and `tmp/` are gitignored.

## Layout

`main.go` starts the Bubble Tea program and delegates to `screen.Root` →
`screen.HomeScreen`. `screen` owns window sizing and closes popups; `chrome`
holds the popup manager and widgets; `keybind` defines bindings and help text;
`store` owns GORM/SQLite persistence; `ui` and `theme` hold shared drawing
helpers and colors.

Popups have one active model at a time and report results as messages rather
than acting directly. Store failures become `chrome.OperationError` messages
that render in the collections popup's status line.
