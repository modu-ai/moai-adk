---
id: SPEC-CLI-TUX-V3-005
title: "AC-TUX3-020 Printer Migration — fmt.Print* Ratchet Succession"
version: "0.1.0"
status: draft
created: 2026-07-14
updated: 2026-07-14
author: manager-spec
priority: P2
phase: "v3.0.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "cli, printer, ratchet, debt, tux-v3"
depends_on: [SPEC-CLI-TUX-V3-001]
tier: M
era: V3R6
---

# progress.md — SPEC-CLI-TUX-V3-005

> Plan-phase progress tracker. Run-phase evidence (§E.2-§E.4) is populated by manager-develop; sync-phase evidence (§E.4) by manager-docs. This file carries the §E section skeleton at plan-phase per the era-classification engine requirement.

---

## §A. Current State

| Field | Value |
|-------|-------|
| Phase | plan (artifacts authored, awaiting plan-audit) |
| Status | draft |
| plan_complete_at | _pending audit-ready verdict_ |
| Baseline (fmt.Print* count) | 38 (measured 2026-07-14) |

---

## §B. Phase Skip Rationale

### Phase 1 (Brain/analyze) — SKIP

**Rationale**: This SPEC is a debt split from SPEC-CLI-TUX-V3-003 (amendment re-close), not a new brain-proposal. The requirement (REQ-TUX3-020) was already analyzed and defined in SPEC-003 spec.md:98. The orchestrator's read-only investigation this session established all ground-truth context (38-call baseline distribution, existing Printer interface surface, gap-site classification). No new requirement analysis is needed — the migration target and scope are fully determined by the inherited REQ and the existing `printer.Printer` interface.

### Phase 3 (Context-First Discovery) — SKIP

**Rationale**: Intent clarity is 100%. The orchestrator's investigation (this session) fully established: (a) the 38-call baseline and per-file distribution, (b) the existing Printer interface method catalogue (printer.go:64-112), (c) the call-site classification (24 migratable, 13 gap, 1 dead code), (d) the dependency chain (SPEC-001 completed, SPEC-003 completed). The three clarification items (banner.go scope, branch_protection.go dead-code, state.go/migration.go human-format channel) were M1-gate decisions RESOLVED 2026-07-14 via user AskUserQuestion — see §D below. All three resolved with the default-recommended option (exclude / exclude / stdout Data() composed string). The domain (CLI output layer) is familiar; no Unknown-Unknowns are suspected.

---

## §C. Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-07-14 | SPEC ID: SPEC-CLI-TUX-V3-005 | Series continuity with TUX-V3-001/002/003/004. Alternative SPEC-CLI-PRINTER-001 considered but rejected — the debt originates from the TUX v3 series and the Printer was introduced by SPEC-001. |
| 2026-07-14 | Tier: M (standard) | Single domain (CLI output), interface already exists, 5 non-test files. Interface-design risk (Tier L signal) is retired. |
| 2026-07-14 | M1 architecture gate leads milestone ordering | banner.go + branch_protection.go scope decision determines whether Printer interface needs extension — highest reversibility. |

---

## §D. Resolved Decisions (formerly open clarification items; resolved 2026-07-14 via user AskUserQuestion)

These items MUST be resolved via orchestrator AskUserQuestion before run-phase entry:

1. **[DECISION 2026-07-14: uikit/banner.go scope — EXCLUDED as TUI render]** — The 12 `fmt.Print*` calls in banner.go (TUI render via lipgloss) are EXCLUDED; Printer interface not extended. (See research.md §C.6, plan.md §F M1.)

2. **[DECISION 2026-07-14: branch_protection.go ttyConfirmer — EXCLUDED as dead code]** — The 1 `fmt.Printf` prompt in `ttyConfirmer.Confirm()` (dead code, `nolint:unused`) is EXCLUDED; active path is yesConfirmer + --yes-branch-protection. (See research.md §C.5, plan.md §F M1.)

3. **[DECISION 2026-07-14: state.go/migration.go human-format — stdout Data() composed string]** — The multi-line human-readable state/migration display routes to stdout via `Data()` (composed string); scripted consumers unaffected. (See research.md §C.1/C.2, plan.md §F M2.)

---

## §E. Section Skeleton

> The following §E subsections are placeholder headings only at plan-phase. They are populated by manager-develop (§E.2, §E.3) and manager-docs (§E.4) during run/sync phases. The era-classification engine (`internal/spec/era.go` `hasAnyProgressMarker`) greps for the literal `§E.2`/`§E.3`/`§E.4` substrings to classify the SPEC era — emitting these placeholder headings at plan-phase prevents misclassification.

## §E.1 Plan-phase Audit-Ready Signal

_status: pending plan-auditor verdict_

Plan-phase artifacts (spec.md, plan.md, acceptance.md, research.md, progress.md) authored 2026-07-14. Awaiting plan-auditor independent audit. The §E.1 audit-ready signal will be populated upon receiving a PASS or PASS-WITH-DEBT verdict.

## §E.2 Run-phase Evidence

**M2 milestone complete** — 11 `fmt.Print*` calls in `internal/cli/state.go` migrated to `printer.Printer` methods (Data/Info).

Migration mapping (orchestrator-verified, working tree = state.go migrated + state_m2_test.go present):

- **Printer construction**: `printer.New(printer.WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr()))` injected into RunE closures at `state.go:49` and `state.go:68`. Function signatures `runStateDump` (:76), `printPhaseStateHuman` (:129), `runShowBlocker` (:152) accept `p printer.Printer` parameter.
- **JSON output → p.Data**: `state.go:120`, `state.go:206` (`p.Data(string(data))`). stdout preserved — scripted JSON consumers byte-identical.
- **Status messages → p.Info**: `state.go:107` ("No checkpoint found"), `state.go:168` ("No blockers found"), `state.go:195` ("No outstanding blockers found"). **Channel re-routed stdout→stderr** (p.Info writes stderr) — non-payload status text moved off the scripted-consumer stdout channel.
- **Human-format → p.Data composed string**: `state.go:148` (`p.Data(strings.Join(lines, "\n"))`). Per DECISION 2026-07-14, the multi-line `printPhaseStateHuman` output is composed into a single string then emitted via `p.Data`, preserving stdout. Byte-identical to the prior `fmt.Print*` path for scripted consumers.

Characterization tests (`state_m2_test.go`) pin both stdout and stderr paths and assert the channel split (JSON/human-format on stdout, status messages on stderr).

**AC mechanical verification** (orchestrator-verified verbatim — injected, NOT re-run by manager-develop):

```
$ go build ./internal/cli/...                                                        → exit 0
$ go test ./internal/cli/ -run 'M2|Checkpoint|Blocker|PhaseState|State' -count=1     → exit 0, "ok github.com/modu-ai/moai-adk/internal/cli 0.444s"
$ go test ./internal/cli/... -count=1                                                → exit 0 (17 packages all ok)
$ go vet ./internal/cli/...                                                          → exit 0
$ go test ./internal/cli/ -cover -count=1                                            → "ok ... 17.298s coverage: 73.8% of statements"
$ grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli/state.go | grep -v _test.go | wc -l
                                                                                     → 0   (AC-TUX3-021 PASS)
```

---

### M3 milestone — migration.go (8 fmt.Print* → Data/Info/Success)

**Scope**: `internal/cli/migration.go` (single-file DDD cycle) + new `internal/cli/migration_m3_test.go`. Follows the exact M2 pattern (state.go, committed 72e22b876).

**8-call mapping applied** (no behavior change; channel re-routing for status messages only):

| # | Line | Pre-migration | Post-migration | Channel |
|---|------|---------------|----------------|---------|
| 1 | 57   | `fmt.Println("실행할 pending 마이그레이션이 없습니다.")` | `p.Info("실행할 pending 마이그레이션이 없습니다.")` | stdout→stderr (status) |
| 2 | 61   | `fmt.Printf("성공: %d개 마이그레이션 적용됨 (버전: %v)\n", ...)` | `p.Success("성공: %d개 마이그레이션 적용됨 (버전: %v)", ...)` | stdout→stderr (status) |
| 3 | 103  | `fmt.Println(string(data))` | `_ = p.Data(string(data))` | stdout (JSON data) |
| 4 | 110  | `fmt.Printf("현재 버전: %d\n", current)` | composed-string `lines[0]` | stdout (data) |
| 5 | 112  | `fmt.Printf("Pending 마이그레이션 (%d개): %v\n", ...)` | composed-string `lines[1]` (pending>0) | stdout (data) |
| 6 | 114  | `fmt.Println("Pending 마이그레이션 없음 (최신 상태)")` | composed-string `lines[1]` (pending==0) | stdout (data) |
| 7 | 117  | `fmt.Printf("최근 적용: %s (버전 %d)\n", ...)` | composed-string `lines[2]` (lastApplied) | stdout (data) |
| 8 | 157  | `fmt.Printf("성공: 버전 %d로 롤백됨\n", v)` | `p.Success("성공: 버전 %d로 롤백됨", v)` | stdout→stderr (status) |

**DECISION 2026-07-14 (composed-string pattern)**: calls #4-#7 (human-format block) are composed into a single `[]string` and emitted via one `p.Data(strings.Join(lines, "\n"))`. This is byte-identical to the pre-migration sequential `fmt.Printf` block: `Data(non-JSON)` = `fmt.Fprintln(stdout, v)`, and `strings.Join(lines, "\n")` + Fprintln's trailing `\n` reproduces the sequential N-line output exactly. Golden-master tests (below) verify byte-identity empirically.

**Channel convention (per CLAUDE.md internal/cli output-stream)**: stdout = machine-readable data (JSON + human block), stderr = human status messages (Info/Success). 3 status calls (#1, #2, #8) move stdout→stderr; 5 data calls (#3, #4, #5, #6, #7) stay stdout. This is the documented convention, not a behavior regression — the data payloads (what scripts pipe) are unchanged.

**Korean strings preserved byte-for-byte (UTF-8)**: all 8 Korean strings carried verbatim into the new `p.Info` / `p.Success` / `p.Data` call arguments. No `\uXXXX` escapes introduced. `grep -rn 'fmt.Print' migration.go` = 0 (comment wording avoids the `fmt.Print*` substring to prevent the AC-TUX3-020 false-positive — comment says "sequential stdout writes").

**Characterization tests** (`migration_m3_test.go`, 6 tests, all PASS):

| Test | Branch covered | Assertion |
|------|----------------|-----------|
| `TestM3_Status_JSON_NoLastApplied` | line 103 (JSON Data) | stdout byte-identical 73-byte golden master; stderr empty |
| `TestM3_Status_Human_NoPending_NoLastApplied` | lines 110+114 (human Data) | stdout byte-identical 72-byte golden master; stderr empty |
| `TestM3_Status_Human_WithLastApplied` | lines 110+114+117 (human Data + lastApplied) | stdout byte-identical 120-byte golden master; stderr empty |
| `TestM3_Status_JSON_WithLastApplied` | line 103 (JSON Data + nested lastApplied) | stdout byte-identical 172-byte golden master; stderr empty |
| `TestM3_Run_NoPending` | line 57 (Info, channel change) | stdout empty; stderr contains Korean text |
| `TestM3_PrinterFormats_UnreachableBranches` | lines 61, 112, 157 (byte-pin at Printer level) | apply-success, pending>0, rollback-success format strings byte-pinned |

**Registry caveat**: the cli test binary does NOT import `internal/migration/migrations`, so m001/m002 never `Register()` and the registry is empty. This makes `Pending()` always empty and `Apply()` always return `[]`, so the apply-success (61), pending>0 (112), and rollback-success (157) branches cannot be reached via the Runner. Those three call sites are byte-pinned at the Printer-call level in `TestM3_PrinterFormats_UnreachableBranches`, using the identical Printer methods whose byte-identity the reachable command-level tests prove.

**Verification (bounded verbatim outputs, redirected to `/tmp/m3-verify/`)**:
```
$ go build ./internal/cli/...                                                       → exit 0
$ go vet ./internal/cli/...                                                         → exit 0
$ go test ./internal/cli/ -run 'M3' -count=1 -v > /tmp/m3-verify/m3-tests.log       → exit 0 (6/6 PASS)
$ go test ./internal/cli/... -count=1 > /tmp/m3-verify/cli-full.log                 → exit 0 (17 packages all ok)
$ grep -rn 'fmt\.Print' internal/cli --include='*.go' | grep -v _test.go | wc -l
                                                                                     → 19  (was 27 after M2; −8 this milestone; AC-TUX3-020 PARTIAL — M4 tmux 5 calls still pending, target 14)
$ grep -c 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli/migration.go | tr -d ' '
                                                                                     → 0   (AC-TUX3-021 PASS for migration.go)
$ grep -cn 'p\.Data\|p\.Info\|p\.Success' internal/cli/migration.go | tr -d ' '
                                                                                     → 6   (reachability: Data×2 + Info×1 + Success×2 + printer.New×3 construction sites)
```

### M4 milestone — tmux_integration.go (5 fmt.Printf → Info; stdout→stderr)

Migrates the 5 consecutive `fmt.Printf` calls in the "Print log output" block of
`CreateTmuxSession` to `p.Info(...)`. Simplest milestone: all 5 calls map to the
SAME method (`Info`), no composed strings, no Korean. Channel change stdout→stderr;
these are status lines with no scripted consumer, so Reversibility risk is LOW.

Mechanism: injectable `tmuxSessionPrinterFactory` (package-level function var),
mirroring the existing `tmuxSessionFactory` / `isTmuxAvailableFunc` convention —
no signature change, preserves all 4 existing sec_harden callers + the `new.go`
production caller (no `new.go` touch).

Characterization test `tmux_integration_m4_test.go` pins the 5 lines byte-for-byte
on stderr in ModePlain (fixed "·" info glyph, no ANSI). The "· " prefix is emitted
exclusively by the Printer status path, so observing it on stderr is the
reachability proof that `p.Info` was actually called (AC-TUX3-021).

```
$ go build ./internal/cli/...                                                       → exit 0
$ go vet ./internal/cli/...                                                         → exit 0
$ go test ./internal/cli/worktree/ -run 'TestCreateTmuxSession_ResultBlock_OnStderrViaPrinter' -count=1 -v
                                                                                     → exit 0 (PASS — 1/1)
$ go test ./internal/cli/worktree/... -count=1                                      → exit 0
$ grep -rn 'fmt\.Print' internal/cli --include='*.go' | grep -v _test.go | wc -l
                                                                                     → 14  (was 19 after M3; −5 this milestone; AC-TUX3-020 migratable-set COMPLETE — banner 12 TUI + branch_protection 1 dead EXCLUDED per M1 arch gate)
$ grep -c 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli/worktree/tmux_integration.go | tr -d ' '
                                                                                     → 0   (AC-TUX3-021 PASS for tmux_integration.go)
$ grep -n 'p\.Info(' internal/cli/worktree/tmux_integration.go | grep -v '^[0-9]*:[ \t]*//' | wc -l
                                                                                     → 5   (reachability: 5 actual p.Info calls lines 115-119, comment-excluded)
```

## §E.3 Run-phase Audit-Ready Signal

**M2-local readiness statement** (NOT full run-phase completion — M3/M4/M5 still pending):

- **AC-TUX3-021 (method-usage) — PASS**: 0 `fmt.Print*` calls in `state.go` (grep verbatim above); all 11 migrated sites reach Printer methods via the construction+signature chain documented in §E.2.
- **AC-TUX3-022 (behavior preservation via characterization tests) — PASS**: `state_m2_test.go` pins stdout/stderr byte output for JSON, human-format, and status-message paths; full `./internal/cli/...` suite green (exit 0).
- **AC-TUX3-020 (ratchet) — PARTIAL (M2-local)**: global `fmt.Print*` count 38→27 after M2 (state.go 11 removed). Final 38→14 target requires M3 (migration.go, 8 calls) + M4 (tmux_integration.go, 5 calls). Not yet at run-phase exit.
- **AC-TUX3-023 (coverage) — SHOULD-PASS**: 73.8% (≥ pre-state after `state_m2_test.go` added; precise pre/post delta is a manager-docs sync-phase concern).

Remaining milestones: **M3** migration.go (8 calls), **M4** tmux_integration.go (5 calls), **M5** ratchet verify (38→14 final). M2 is an intermediate run-phase milestone; this §E.3 records M2-local readiness, not the terminal run-phase signal.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

**Input parameters**:
- tier: M
- scope: 5 non-test files (state.go, migration.go, tmux_integration.go, banner.go, branch_protection.go); 24 migratable call sites
- domain count: 1 (CLI output layer)
- file language mix: Go (100%)
- concurrency benefit: LOW (coding-heavy migration per Anthropic coding-task parallelism caveat)

**Mode evaluation**:
- Mode 1 (trivial): NO — multi-file migration with characterization tests
- Mode 2 (background): NO — write-capable implementation work
- Mode 3 (agent-team): RETIRED
- Mode 4 (parallel): NO — coding-heavy, single domain, few parallelizable tasks
- Mode 5 (sub-agent): SELECTED — sequential per-milestone DDD cycles
- Mode 6 (workflow): NO — semantic per-file migration (not a single uniform mechanical transform); <30 files

**Decision**: sub-agent (Mode 5)

**Justification**: Tier M single-domain coding-heavy migration. Each milestone (M2 state.go 11 calls, M3 migration.go 8 calls, M4 tmux_integration.go 5 calls) is an independent DDD ANALYZE-PRESERVE-IMPROVE cycle on one file — sequential sub-agent per milestone matches Anthropic's coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research"). M1 architecture gate already RESOLVED (3 decisions recorded 2026-07-14: banner.go excluded, ttyConfirmer excluded, human-format stdout Data()). M5 is ratchet verification. Per-milestone dispatch also avoids the autocompact-thrash hazard of a single large-scope spawn.

**Implementation Kickoff Approval**: OBTAINED 2026-07-14 (user AskUserQuestion — run-phase entry approved, score-independent mandatory gate cleared).

**Plan Audit Gate (Phase 1)**: PASS 0.95 (plan-auditor iter-2, 2026-07-14). skip-eligible (score ≥0.90 + artifact hash stable post-acd7d8046 + within 24h) — Phase 1 re-execution skipped; the iter-2 audit IS the verdict of record.
