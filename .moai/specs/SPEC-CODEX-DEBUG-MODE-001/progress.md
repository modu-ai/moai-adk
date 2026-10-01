# SPEC-CODEX-DEBUG-MODE-001 — Progress

Status: draft (plan-phase artifacts authored 2026-10-01, card t1380). Premise table V1–V12 measured on tree `16424e9b4`; the RED-now observations in `acceptance.md` § Evidence Ledger were re-executed on tree `49a42c2fc` (the lane's artifact commit moved HEAD between the two measurement passes; none of the cited source files changed — anchors below re-verified at `49a42c2fc`).

## Premise Verification (plan-phase, this tree `16424e9b4`, branch `WT-codex-debug-mode`)

Verdict: **all four card premises HOLD; one external-support question closed as honestly unverifiable with a recorded fallback.** Evidence observed in this run, this worktree:

| # | Claim | Evidence (file:line + observed) |
|---|-------|-------------------------------|
| V1 | Codex launcher flag surface today is `--spawn` only | `internal/cli/codex_launcher.go:600` (`var codexCmd = &cobra.Command{`; the card's `:601` anchor names this command), `:641` (`DisableFlagParsing: true` — auditor-verified anchor), `:652` — the ONLY flag registration: `codexCmd.Flags().Bool("spawn", false, ...)` (re-verified at `49a42c2fc`); `:70-72` readout-refusal discipline precedent. Grep for `-d`/`--debug` over the file: zero hits, exit 1 (positive control `codexSpawnReadoutDiag` → `:70/:72/:721`, exit 0). |
| V2 | cc/glm have no `-d` handling — passthrough by structure | `grep '"-d"\|--debug\|debug'` over `internal/cli/cc.go` and `glm.go`: zero hits. Both are `DisableFlagParsing: true` (`cc.go:133`, `glm.go:137`); `runClaudeEntry` (`cc.go:146`) consumes only `--help/-h`, `--spawn` (`:164`), `-p` profile (`:168`), and the `-k/-f` entry shapes (`:175-179`) — leftover args flow into the child invocation, so `-d` reaches `claude` verbatim today. |
| V3 | claude CLI natively supports `-d` | `claude --help` (local, `/Users/goos/.local/bin/claude`): `-d, --debug [filter]  Enable debug mode with optional category` plus `--debug-file <path>`. |
| V4 | glm binary unavailable locally | `command -v glm` → no output. Child-native `-d` support UNVERIFIED for glm — deliberately non-load-bearing (observe-only design). |
| V5 | codex CLI rejects `-d`; no `--debug` flag; `debug` is a subcommand group | codex-cli 0.159.3 at `/Users/goos/.local/bin/codex`. `codex -d` → `error: unexpected argument '-d' found` (verbatim). `codex --help`: no `--debug` option; `debug` listed as a command ("Debugging tools"). `codex debug --help`: subcommands models / app-server / prompt-input only. `codex exec --help` grep for debug/log/RUST: no logging flag. |
| V6 | RUST_LOG: absent from the repo; undocumented for codex locally; literal present in the installed binary | `grep -rn RUST_LOG internal/ pkg/ cmd/` → exit 1 (zero hits). `strings /Users/goos/.local/bin/codex | grep -m3 RUST_LOG` → one concatenated literal hit (`=ErrorRUST_LOGcalled ...` — presence, not wiring proof). Upstream `docs/config.md` fetch → returned non-codex content (inconclusive). Resolution: `RUST_LOG=debug` injection recorded as a best-effort tracing-standard fallback; the launcher-side trace is the primary surface. |
| V7 | Launch-path phases to trace (slot claim, worktree, env injection) — anchors re-verified at `49a42c2fc` | Binary resolve `codex_launcher.go:1002`; timing gate (lane-only) `:1014-1015`; project root `:1023-1027`; init gate `:1033` (`codexInitOfferGate`); local instructions `:1034`; worktree resolve + writer check `:1054-1055` (`resolveCodexWorktreeDir` `:428`); factory entry: join gate `codex_factory.go:135-137` → `enterFactoryLaneRun`, lane claim `:148-151` → `kanban.ClaimFactoryLaneWithin`; env assembly `codexChildEnv` `codex_launcher.go:575-593` (drop-list + `CODEX_HOME` last-wins append) + `codexFactoryEnv` `codex_factory.go:206-227`; anchor lock + exec handoff `:1084` (`endHandoff`), threshold report sites `:1093/:1105`; direct door `codex_direct_posix.go:25-58` (`syscall.Exec` `:53`). |
| V8 | t1378 join reads recorded capacity in factory_slots.go | The card named no directory — the file is `internal/kanban/factory_slots.go` (NOT `internal/cli/`): `ClaimFactoryLaneWithin` `:159`, the recorded-capacity authority read inside the claim transaction `:170` (MX:ANCHOR), the no-recorded-capacity launcher-bound fallback `:280`. Launcher-side bound: `factoryJoinLaneBound` `internal/cli/codex_factory.go:180-186`. |
| V9 | Logging precedents | Level: `MOAI_LOG_LEVEL`/`MOAI_LOG_FORMAT` `internal/config/envkeys.go:40,43`, consumed by the slog default handler `internal/cli/logging.go:43-55` (default `warn`, `:22`). Destination: non-hook subcommands → stderr (`logging.go:78-83`); hook path → lazy-open `.moai/logs/` file sink (`:80`). `.moai/logs/` writer patterns exist for BACKGROUND processes: `internal/config/log.go:22-24` (`config.log`, best-effort append), `internal/codexadapter/diagnostics.go:16` (`.moai/logs/codex-adapter.jsonl`). Interactive launchers print diagnostics to `cmd.ErrOrStderr()` (`codex_launcher.go:1004, 1049, 1093, 1105`). |
| V10 | t1378 timing collector is the composition seam | `internal/cli/factory_launch_timing.go`: nil-safe (`:46-48`, "the cc/glm twins ... pass nil"), records every `begin()` step regardless of threshold (`:58-70`), threshold gates only `reportSlow` (`:87-101`); pre-seam discipline MX:NOTE `:85`; threshold default `internal/config/defaults.go:672` (2s), override key `internal/config/envkeys.go:297` (`MOAI_FACTORY_SLOW_LAUNCH_MS`). Wired lane-only at `codex_launcher.go:1012-1014`. |
| V11 | No dedicated debug env/flag exists anywhere in the launchers | `grep -rn RUST_LOG` (V6, zero) + V1/V2 greps: no `-d`/`--debug`/debug-env surface on any of the three launchers today. |
| V12 | Head/tail scoping precedent for the token scan | `runCodex` `--help` scan breaks at `--` (`codex_launcher.go:663-667`); `splitCodexDashDash` `:669`; `stripSpawnFlag` `:668` — the debug-token scan mirrors this pre-`--` scoping. |

SPEC ID self-check: `ID="SPEC-CODEX-DEBUG-MODE-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` → `PASS`. Uniqueness: `grep -rn SPEC-CODEX-DEBUG .moai/specs/` → zero hits (catalog scan observed this run).

## Plan-audit Iterations

- **Iteration 1 — FAIL 0.81 vs Tier M 0.80 (aggregate cleared the threshold; the MP-8 must-pass firewall forced FAIL on RED-now adoption)** (report: `.moai/reports/t1380/plan-audit.md`). Fixes applied 2026-10-01 on tree `49a42c2fc`, no commits: D1 (all 10 release-blocking ACs now carry RED-now cells with Command + fenced verbatim output + Exit + measured SHA — RED-1 invocation probe `go run ./cmd/moai codex -d status` at exit 1, RED-2/RED-3 greps with exit codes and positive controls per t1133, RED-4..RED-9 anchored test-absence selectors with vacuous-green red-reasons; single-invocation re-executability stated per cell, release-blocking retained), D2 (AC-015 added — pre-seam ordering on both doors, stub-seam-relative assertion, release-blocking; §C mapping REQ-009 → AC-015), D3 (citation convention fixed: `t1378 REQ-012` disambiguated from this SPEC's §B.4 REQ-012 across spec/plan/acceptance), D4 (uniform ledger carrier form: Command / Verbatim / Exit / Measured-at / Positive-control / Red-reason), D5 (chained Verify commands split into separately recordable parts — AC-013 parts 1-2, AC-014 split), D6 (V1/V7 line anchors re-verified and corrected at `49a42c2fc`: `:641` DisableFlagParsing, `:1033` init gate, `:1054-1055` worktree, `:1014-1015` timing wiring). No probe file cited by the ledger (all cells are invocation/grep/selector forms), so the t1378 committed-probe-file rule has no target here.

## §E.1 Plan-phase Audit-Ready Signal

_<pending plan-audit verdict — the audit loop fills plan_status / plan_audit_verdict / plan_audit_report / plan_audit_sha; no audit claims exist at authoring time>_

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
