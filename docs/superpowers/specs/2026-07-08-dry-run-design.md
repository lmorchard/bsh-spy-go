# `run --dry-run` — observational mode

Date: 2026-07-08

## Problem

`run` (with or without `--once`) always performs the real work: it adds the
current song to the Spotify playlist and writes to the local cache. There is no
way to see what the bot *would* do — which song it scraped, whether it matched
on Spotify, and whether it would add or skip it — without mutating the playlist
and cache.

## Goal

Add a `--dry-run` flag to `run` that reports the per-poll decision but performs
no mutations, so an operator can preview behavior safely.

## Design

### Surface

A boolean `--dry-run` flag on the existing `run` command (default false =
current behavior). Composes with `--once`: `run --once --dry-run` previews a
single poll; `run --dry-run` previews continuously on the ticker.

### Behavior

The poll cycle runs normally through the read-only steps — scrape, Spotify
search, and the cache membership check — then reports the decision and returns
*before* any mutation. In dry-run, NOTHING is mutated: no playlist POST, no
`tracks` cache write, no `mystery_songs` write.

Per current song, it logs one of:
- found & not in cache → `would add: <artist> - <title>` (with track id + uri)
- found & already in cache → `already in playlist — would skip`
- no Spotify match → `no match — would record mystery`

### Implementation

Two files:

- `internal/runner/runner.go`: add a `DryRun bool` field to `Deps`. In
  `ProcessSong`, run `SearchTrack` and (when matched) `Store.Has` as today, but
  when `DryRun` is true, log the "would…" outcome and `return nil` before
  `AddToPlaylist`, `Store.Add`, or `RecordMystery`. The non-dry-run path is
  unchanged.
- `cmd/run.go`: add the `--dry-run` flag and set `deps.DryRun` from it.

### Testing

Extend `internal/runner/runner_test.go` with dry-run cases using the existing
fakes (which count `addCalls`, `added`, `mysteries`):
- dry-run + new track → `addCalls == 0 && len(added) == 0` (returns nil, logs "would add").
- dry-run + no match → `mysteries == 0 && addCalls == 0` (returns nil).
- (sanity) known-track skip is already a no-op for mutations regardless of dry-run.

Then a live `run --once --dry-run` against real credentials to observe the
report — safe, since it performs no writes.

## Non-goals

- No change to the non-dry-run behavior.
- No separate `preview`/`check` command (rejected — duplicates run wiring for a
  one-flag feature).
- No dry-run reporting format beyond the log lines above (no JSON/table output).
