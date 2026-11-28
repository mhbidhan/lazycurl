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
- Explorer: pressing `enter` on a collection replaces the home screen with a
  full-screen tree browser. Navigate with vim motions, filter with `/`, create,
  rename, delete, mark to move or copy, paste and reorder nodes. Directories may
  contain other nodes; requests are leaves.

Request editing, sending HTTP calls and viewing responses are **not implemented
yet**. Request fields have persisted defaults, but only the tree can be edited
today.

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

### Explorer keys

| Key | Action |
|---|---|
| `enter` (collections popup) | open the selected collection |
| `j` / `down`, `k` / `up` | move the tree cursor |
| `h` / `left`, `l` / `right` | collapse or expand a directory |
| `enter` | toggle the selected directory |
| `g` / `G` | jump to top / bottom |
| `/` | fuzzy-filter nodes (`enter` keeps, `esc` clears) |
| `a` / `A` | add a request / directory (one name prompt) |
| `d` | delete the selected node and its descendants |
| `r` | rename the selected node |
| `x` / `c` | mark the selected node to move / copy it |
| `p` | paste the marked node into or after the target |
| `J` / `K` | move the node down / up among its siblings |
| `?` | keybinds popup |
| `esc` / `q` | return to the home screen |

`a` creates a request from the name you type; ending the name with `/` creates
a directory instead and the slash is not stored. `A` always creates a directory.
`x` marks a node to move; `c` marks it to copy. `p` pastes into a selected
directory as its last child, or after a selected request as the next sibling.
Copies get fresh IDs; copying a directory also copies its whole subtree and the
original stays untouched. The explorer expands a collapsed target directory so
the result stays visible. Cycles and moves into requests are rejected and shown
on the status line; a rejected paste keeps the mark so the target can be changed
without re-marking.

## Develop

- `go build ./...` — build every package.
- `go test ./... -cover` — run tests with per-package coverage; the project
  target is at least 70%.
- `air` — rebuild and restart on source changes using `air.toml`; output goes to
  `bin/`.

The database is `tmp/lazycurl.db`, created and migrated on startup. Both `bin/`
and `tmp/` are gitignored.

## Layout

`main.go` starts the Bubble Tea program and delegates to `screen.Root`, which
owns the active screen and switches between `screen.HomeScreen` and
`screen.ExplorerScreen` on `OpenCollectionMsg` / `BackToHomeMsg`. `screen` owns
window sizing and closes popups; `chrome` holds the popup manager and widgets;
`keybind` defines bindings and help text; `store` owns GORM/SQLite persistence
and tree operations; `ui` and `theme` hold shared drawing helpers and colors.

Popups have one active model at a time and report results as messages rather
than acting directly. Store failures become `chrome.OperationError` messages
that render in the collections popup's status line or the explorer's status
line.
