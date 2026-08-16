# t56 — kanban companion naming: bare roles, numeric collision bump

Card t56 (run tjv7iy, round 4). Operator decision 2026-08-17: "한 머신 한 런"
premise ACCEPTED — remove the run-id suffix from kanban companion names; plain
role names (plan / run / review / sync) with numeric bump on collision
(plan-1, plan-2 …).

## Files in this directory

- `red-kanban.txt`, `red-cli.txt`, `red-hook.txt` — RED (TDD phase 1). New
  contract fails to compile against the old grammar: `SplitCompanionLabel`
  bare-role table rows, `CompanionLabel` single-arg signature,
  `CompanionNumberLabel` / `resolveCompanionName` / `companionRegistryPath`
  undefined. Assertion-level RED for the announcement shape is inside the same
  test files (regex + contains flips).
- `green-kanban-config-final.txt` — `go test ./internal/kanban/ ./internal/config/`
  (final tree) — ok / ok, incl. `TestAlwaysLoadedTokenBudget` PASS.
- `green-hook-final.txt` — `go test ./internal/hook/` (final tree) — ok.
- `green-cli-final.txt` — `go test ./internal/cli/` (final tree, full package) — ok.
- `green-template-final.txt` — `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/`
  (final tree) — ok (mirror parity + neutrality).
- `lint-final.txt` — `golangci-lint run` over the five touched packages — 0 issues.
- `make-build.txt`, `make-build-final.txt` — template re-embed builds, exit 0.

## Decisions

1. **Legacy run-id form is MIGRATED, not rejected.** `SplitCompanionLabel`
   keeps admitting `<role>-<lowercase-alnum>`; only the bare-role fast path is
   added. Rejecting the suffix would fail the shape check, and
   `-k --name plan-abc123` (stale muscle memory) would then reroute down the
   LEAD branch of the launcher truth table — a `-k` launch whose name is not
   companion-shaped seeds a whole second chain. A silent misroute is worse
   than joining under a stale suffix.
2. **The lead keeps `lead-<run-id>`.** The run id remains lead-owned state
   (its session name survives /clear; its leader socket path). Companion
   surfaces carry no run id at all.
3. **Companion bump mirrors the t68 factory bump.** Liveness-checked pid
   registry at `.moai/state/kanban/companions.json`, dead claims pruned and
   reclaimed, bumped candidates are `<role>-<n>` (a held legacy
   `plan-abc123` bumps to `plan-1`, never a second hyphen), final label
   reaches the backend argv via `replaceNamedLabel`, and the SessionStart
   companion notice names the final label (the reliable surface for a bumped
   name — the stderr note is gone by the time the TUI takes the screen).
4. **`enterKanbanCompanionMode` no longer sets MOAI_KANBAN_ID.** Nothing on
   the companion path reads it; the announcement and the label are the whole
   membership.

## t21 — incident class removed (absorbed)

t21's shape: the lead's `MOAI_KANBAN_ID` disagreed with a live companion's
name suffix, so the SessionStart notice announced companion commands
(`moai cc -k --name <role>-<run-id>`) composed from a run id no live session
carried — a ghost run announcement. Two structural removals land here:

- the announced commands are now `moai cc -k --name <role>` — no run id
  appears in any copyable companion command, so a disagreeing id has nothing
  to disagree about (`internal/hook/session_start_kanban.go`);
- the companion branch no longer derives/publishes `MOAI_KANBAN_ID` from its
  label at all (`internal/cli/kanban.go enterKanbanCompanionMode`) — the
  suffix is a collision number now, and publishing it as a run id would
  recreate the mismatch surface.

With no companion surface carrying a run id, the lead-env-vs-companion-name
mismatch class has no remaining producer.

## Verification commands (final tree, all observed)

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR \
      MOAI_KANBAN_SETTINGS_INJECTED && <cmd>
```

- `go build ./...` — exit 0
- `GOOS=windows go build ./... && GOOS=windows go vet <5 pkgs>` — exit 0
- `gofmt -l <my files>` — empty after `gofmt -w` on two files
- `go test ./internal/kanban/ ./internal/config/ -count=1` — ok
- `go test ./internal/hook/ -count=1` — ok
- `go test ./internal/cli/ -count=1 -timeout 540s` — ok
- `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -count=1` — ok
- `golangci-lint run ./internal/cli/... ./internal/kanban/... ./internal/hook/... ./internal/config/... ./internal/template/...` — 0 issues
- `make build` — exit 0 (templates re-embedded)

Gaps: no live multi-terminal launch was exercised (the `moai cc` launch path
mutates real settings files; §13 forbids running it in the dev project). The
bump, the announcement, and the env publication are covered by unit tests
with the liveness seam overridden. First real launch happens on the next
kanban run.

Residual risk: `TestAlwaysLoadedTokenBudget` now passes with ≤3 tokens of
headroom (76,000 budget) — the next always-loaded rule edit will trip it
again; kanban-dispatch.md remains the top diet candidate (known from the
session-cost analysis).
