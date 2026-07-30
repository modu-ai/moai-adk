# SPEC-CLIFIX-HYGIENE-001 M1 — Threshold & Timeout Literals Inventory (Frozen Fixture)

> Frozen by manager-develop M1 on 2026-07-30 against worktree HEAD `359e887b9`.
> This is the baseline M3 will collapse to a single `defaults.go` source per
> spec.md REQ-HYG-001-004 / plan.md §B row 4.

## Tier-threshold literal `[]int{1, 3, 5, 10}` — THREE sites (M3 collapses to ONE)

| Site | Line | Context |
|---|---|---|
| `internal/cli/harness.go:150` | `thresholds = []int{1, 3, 5, 10}` | default branch (computed when config omits `tier_thresholds`) |
| `internal/cli/harness.go:480` | `TierThresholds: []int{1, 3, 5, 10},` | struct-default initializer |
| `internal/cli/hook.go:1013` | `var defaultTierThresholds = []int{1, 3, 5, 10}` | package-level `defaultTierThresholds` var |

**Plan correction**: the original audit said "duplicated ×2" — this is an
undercount. The current tree has **THREE** sites, all carrying the identical
literal. M3 reduces to ONE `defaults.go` constant referenced by all three.

## Dispatcher timeouts — TWO `30*time.Second` sites in hook.go (M3 extracts to `defaults.go`)

| Site | Line | Context |
|---|---|---|
| `internal/cli/hook.go:237` | `ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)` | dispatcher call site A |
| `internal/cli/hook.go:361` | `ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)` | dispatcher call site B |

M3 introduces a `defaults.go` constant (e.g. `DefaultHookDispatcherTimeout`) and
both sites reference it.

## Size caps / retry / circuit literals

Deferred to M3 inline extraction — the audit named these but did not anchor
them with stable line numbers. M3 re-derives them via grep at run time and
adds them to `defaults.go`.

## M3 review note

Per CLAUDE.local.md §14, all thresholds MUST live in `config/defaults.go` as a
single source. The three-site tier-threshold duplication is the load-bearing
collapse target; the two dispatcher timeouts are the secondary collapse target.
