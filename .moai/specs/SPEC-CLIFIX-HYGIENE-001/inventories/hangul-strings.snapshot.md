# SPEC-CLIFIX-HYGIENE-001 M1 — Hangul Strings Inventory (Frozen Fixture)

> Frozen by manager-develop M1 on 2026-07-30 against worktree HEAD `359e887b9`.
> This is the baseline M4 will diff against. The 6 in-scope files are per
> spec.md REQ-HYG-001-005 / plan.md §B row 5: doctor.go, migration.go, clean.go,
> web_port*.go (web_port.go + web_port_posix.go + web_port_windows.go).
>
> Counts obtained via: `grep -c '[가-힣]' <file>` (BSD grep over Hangul ranges is
> unreliable for `-l`; per-`-c` is stable for non-zero counts; rg cross-check
> pending at M4 — the AC uses `rg -l`).

## Per-file Hangul counts (M1 baseline)

| File | Hangul line count | Status for M4 |
|---|---:|---|
| `internal/cli/doctor.go` | **0** | **Already English — no M4 work needed.** Drift finding vs plan.md §B row 5 (which lists doctor.go as in-scope); recorded here, NOT a defect — the file was cleaned up between the audit and this SPEC. |
| `internal/cli/migration.go` | 16 | M4 target — Korean user-facing strings (e.g. `:45` cwd error, `:52` migrate-fail, `:57` no-pending, `:61` success count). |
| `internal/cli/clean.go` | 1 | M4 target — `:38` `--force` flag help text. |
| `internal/cli/web_port.go` | 29 | M4 target — port-recovery docstrings + user-facing errors (`:81` non-moai holder, `:84` reuse notice, `:86` kill-fail, `:95` post-SIGTERM still-held). |
| `internal/cli/web_port_posix.go` | 17 | M4 target — lsof/ps error wrappers (`:29,33,37,42`). |
| `internal/cli/web_port_windows.go` | 10 | M4 target — file-header docstring + unsupported stubs (`:18-19,24-26`). |

## Total in-scope Hangul lines (M1 baseline): **73** across 5 files (doctor.go excluded — already clean)

## Out-of-scope (~24 additional Hangul-bearing production files)

Per spec.md §C `Out of Scope — Broader i18n sweep`, deferred to a follow-up
SPEC. Listed for traceability; M4 does NOT touch these:

`fang.go`, `glm_tools.go`, `harness_clusters.go`, `loop.go`,
`preference/{correction,decay,freshness,gate,proficiency,toggle}.go`,
`schema_bridge.go`, `specid/specid.go`, `state.go`, `web.go`,
`wizard/questions.go`, `wizard/translations.go`, `worktree/new.go`,
`harness/execute.go`, `pr/watch.go`, `agentlint/agent_lint.go`,
`profile_setup*.go`, `glamour_style.go`.

## M4 review note

Per plan.md §G anti-pattern: each message must be reviewed so diagnostics stay
accurate — NO blind sed. The migration.go success/pending counts (`:57,:61`)
and web_port error messages with PID interpolation (`:81,:84,:86,:95`) carry
semantic content (counts, PIDs, port numbers) that must survive the
English rewrite.
