# Watchlist Contract

## Purpose

Define the normative machine-local storage contract for the first easyharness
watchlist. This contract exists so later watchlist registration, read-model,
and dashboard work can build on one clear local storage shape instead of
reopening basic questions about what is watched, where it is stored, and which
facts belong in persisted data versus derived UI state.

This spec is intentionally narrow. It defines the minimal persisted watchlist
record for local use. It does not define write triggers, backend shape, or UI
behavior beyond the persisted-versus-derived boundary needed to keep later
slices coherent.

## Watched Unit

The watched unit is a `git-backed workspace`.

A watched workspace is a local filesystem checkout that:

- is backed by a Git working tree
- is intended to run easyharness in that checkout

This definition includes both:

- a repository's primary checkout
- a linked checkout created through `git worktree`

The watchlist contract must not assume that every watched workspace is a
linked worktree. Direct work in a repository's primary checkout is a
first-class case.

This first contract does not support non-git directories as watched items.

## Storage Location

The watchlist file lives at:

- `~/.easyharness/watchlist.json`

This location is machine-local and user-private. It is not a repository-shared
artifact and must not be written into the repository itself.

The parent directory may later hold other machine-local easyharness files, but
this contract defines only the watchlist file above.

## File Shape

The watchlist file is UTF-8 JSON with one top-level object:

```json
{
  "version": 1,
  "workspaces": [
    {
      "workspace_path": "/absolute/path/to/workspace"
    }
  ]
}
```

Contract:

- `version` is required and identifies the watchlist file format version
- `workspaces` is required and contains zero or more watched workspace records
- each `workspace_path` is required
- each `workspace_path` must be an absolute canonical local filesystem path
- duplicate `workspace_path` values are invalid

The minimal persisted workspace record is intentionally small:

- `workspace_path`

This contract does not require any additional persisted per-workspace fields in
the first slice.

## Identity Model

For this first machine-local contract, watched-workspace identity is the
canonical absolute `workspace_path`.

This choice is intentionally local and path-oriented:

- it keeps the persisted record small enough for an XS foundation slice
- it supports both primary checkouts and linked worktrees with one model
- it avoids introducing synthetic IDs before later read-model or write-path
  work proves they are necessary

If a workspace moves to a different path, that is a different watched
workspace under this initial contract. The first contract does not attempt to
preserve identity across path moves.

## Derived Repository Grouping

The watchlist persists watched workspaces, not repository groups.

Repository-family grouping is derived at read time from Git metadata. The
intended local grouping model is:

- a repository's primary checkout and any linked git worktrees belong to the
  same local repository family
- separate local clones remain separate families even when they point to the
  same remote

The exact Git probe is an implementation detail, but the grouping contract
must be consistent with local repository-family identity rather than
remote-URL-based project identity.

This lets the UI treat `workspace` as the base watched unit while still
grouping related local checkouts together.

## Persisted Versus Derived Fields

Persist only the minimal watched-workspace set in the watchlist file.

Persisted:

- `version`
- `workspaces[].workspace_path`

Derived at read time:

- repository root or top-level path
- local repository-family grouping key
- branch name
- whether the workspace is the repository's primary checkout or a linked
  worktree
- live harness status or dashboard summary fields

The contract prefers deriving these facts from the current filesystem and Git
state instead of persisting copies that can drift.

## Deferred View State

Dashboard-only view state is out of scope for this minimal watchlist contract.

In particular, fields such as:

- `hidden`
- completion filtering
- manual dismissal or archive-like dashboard preferences

must not be folded into the minimal persisted watchlist record defined here.

Those concerns belong to a later local view-model or lifecycle contract once
the dashboard behavior is ready to define.

## Non-Goals

This spec does not:

- define when or how workspaces are added to the watchlist
- define daemon versus on-demand backend architecture
- define a dashboard read model beyond the persisted-versus-derived boundary
- merge separate local clones into one project because they share a remote
- support non-git watched directories in the first slice
