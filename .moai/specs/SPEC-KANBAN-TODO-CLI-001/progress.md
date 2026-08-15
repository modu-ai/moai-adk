---
id: SPEC-KANBAN-TODO-CLI-001
title: "Progress — moai todo CLI subcommand with lock-guarded backlog store"
version: "0.1.1"
status: in-progress
created: 2026-08-14
updated: 2026-08-14
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, cli, backlog, concurrency, progress"
tier: M
---

# progress.md — SPEC-KANBAN-TODO-CLI-001

## Authoring record

This is the **second delegation** of this plan phase. The first agent (kanban card t12 dispatch, 2026-08-14) delivered `spec.md` + `research.md` and then failed on autocompact thrashing before writing plan.md / acceptance.md / progress.md. This delegation verified the two surviving artifacts structurally (both complete, not truncated) and authored the remaining three. No SPEC body content from the first delegation was rewritten.

## Phase 1 — exploration SKIP rationale

Deep exploration was performed inline by the orchestrator **before delegation** and is fully recorded in `research.md` (12 evidence sections, every claim command-attributed: the missing CLI verb, the incident file, the lock substrate reads, the atomicfile exports, the twin diff, the CLI conventions, the dispatch-rule verb fit). Re-running an exploration pass here would duplicate verified evidence against a context budget that already consumed the previous agent — SKIP, evidence cited, not assumed.

## Decision Point 1 — satisfied by dispatch

The operator picked card t12 from the backlog queue and the kanban lead dispatched to plan per `.claude/rules/moai/workflow/kanban-dispatch.md` § The dispatch cycle. Entry into the board is the operator's act; this SPEC is the plan-phase product of that pick. The run session still faces **Implementation Kickoff Approval** before run-phase entry — nothing here pre-authorizes it.

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `research.md`, `progress.md` (Tier M set + research.md carried from the first delegation; `progress.md` at every tier).

- Requirements: 16 (`REQ-TODO-001` … `REQ-TODO-016`) — within the Tier M ceiling of 16 (post-v0.1.1 consolidation, spec.md §F History; originally authored at 19, see plan.md §G AP-PL-001).
- Acceptance criteria: 15 (`AC-TODO-001`, `AC-TODO-003` … `AC-TODO-016`; `AC-TODO-002` never defined — count labels corrected at v0.1.1) — within the Tier M ceiling; all 16 REQs covered (post-v0.1.1 consolidation, spec.md §F History) via acceptance.md §A matrix.
- Milestones: 4 (M1 store/lock/ids → M2 CLI verbs → M3 skill+mirror → M4 cross-platform verification), ordered by decision-reversibility per plan.md §F.
- Design decisions (a)–(h) from the dispatch contract: all stated and justified in plan.md §D; zero open clarification markers (each decision is bound to an authored REQ).
- Branch/worktree setup: deferred to the run session (Late-Branch policy; shared 5-session checkout; plan phase performs no git state changes) — plan.md §F Phase 13 note.

plan_status: audit-ready
plan_complete_at: 2026-08-14

### Plan-audit verdict (plan phase closed)

- Iteration 1 (2026-08-14): FAIL 0.90 — `.moai/reports/plan-audit/SPEC-KANBAN-TODO-CLI-001-review-1.md` (blocking: D1 AC-count misstatement, D2 REQ 19 > Tier M ceiling 16; optional: D3 negated marker token).
- Remediation: v0.1.1 — REQ consolidation 19→16 (spec.md §F History), count labels + coverage matrix rebuilt, token reworded, sibling frontmatter versions aligned.
- Iteration 2 (2026-08-14): **PASS 0.95** (Tier M threshold 0.80) — `.moai/reports/plan-audit/SPEC-KANBAN-TODO-CLI-001-review-2.md`; D1/D2/D3 RESOLVED with mechanical evidence; D4/D5 optional observations (D5 closed, D4 accepted as lint-surface artifact).
- Plan phase CLOSED. Next: Implementation Kickoff Approval (run-session human gate) → `/moai run SPEC-KANBAN-TODO-CLI-001`, M1 (store/lock/ids) first per plan.md §F ordering.

## §F Phase 4 Mode Selection

Input parameters: tier=M; scope ≈ 6 files (1 new store file + tests, CLI verb file + tests in M2, 2 skill twins + catalog.yaml in M3); domains = 2 (Go source, skill/template markdown); language mix = Go + markdown; concurrency benefit LOW (coding-heavy, sequential dependencies M1→M2→M3); Agent Teams prereqs N/A (Mode 3 retired).

| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | no | New concurrency-bearing store, not a typo fix |
| 2 background | no | Write-capable implementation work |
| 3 agent-team | no | RETIRED |
| 4 parallel | no | Coding-heavy per Anthropic's coding-task parallelism caveat |
| 5 sub-agent | YES | Sequential per-milestone delegation; M1 first (semi-autonomous progression — orchestrator stops after M1 for lead evidence review) |
| 6 workflow | no | Not ≥30-file mechanical transformation |

Decision: sub-agent
Justification: single-package Go implementation with tight sequential dependencies between milestones; one manager-develop delegation per milestone (Mode 5). Progression mode: semi-autonomous — the kanban lead directed an M1 stop-point for evidence review, so M2/M3/M4 await the lead's go decision. Implementation Kickoff Approval passed (operator, via kanban lead dispatch 2026-08-14); plan-audit PASS 0.95 ≥ Tier M threshold 0.80 with artifacts unchanged since (skip-eligible Phase 1 re-execution).

## §E.2 Run-phase Evidence

**Baseline-attribution**: all evidence below was captured by the orchestrator (run session, worktree `kanban-todo-cli`, branch `worktree-kanban-todo-cli`) against the tree `origin/main b73c81ab5` + the M1 working files (`internal/kanban/backlog_store.go` + 2 test files), before the M1 commit. Two manager-develop delegations died on autocompact context thrashing (delegation 1 after authoring the RED test file; delegation 2 after implementing to green, dying mid-coverage-measurement); the orchestrator then verified every claim directly — no §E row below is a delegation self-report.

All `go test` invocations use the shell-builtin env scrub `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED CLAUDE_CODE_STOP_HOOK_BLOCK_CAP && go test …` — the `env -u` wrapper form is rejected by the worktree guard hook in worktree-isolated sessions (verified; lesson recorded in memory).

### M1 — Backlog store + lock + ID issuance (COMPLETE)

Deliverables: `internal/kanban/backlog_store.go` (292 lines), `internal/kanban/backlog_store_test.go` (531 lines, 15 tests), `internal/kanban/backlog_store_errors_test.go` (151 lines).

**E8 — RED (TDD, verbatim pre-GREEN).** Inherited RED state was the test file against no implementation (compile failure), observed by the orchestrator:

```
$ unset MOAI_KANBAN … && go test ./internal/kanban/ -run TestNonexistent -count=1
internal/kanban/backlog_store_test.go:377:35: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/kanban [build failed]
FAIL
```

**E1 — M1 test matrix (store-level AC slice).**

```
$ unset MOAI_KANBAN … && go test ./internal/kanban/... -count=1 -cover
ok  	github.com/modu-ai/moai-adk/internal/kanban	24.727s	coverage: 87.3% of statements
```

All 15 tests in `backlog_store_test.go` pass (plus the error-path suite), covering: lock acquire/release incl. bounded-retry timeout naming `backlog.lock` (`TestBacklogLock_TimeoutNamesLockPath`); ID issuance under lock (`TestBacklogIDIssuedUnderLock_SequentialMutations`, `TestBacklogMutate_LoserSerializedNotFailed`); `last_seq` high-water surviving removal — done t20 then add → t21 (`TestBacklogHighWater_SurvivesRemoval`); derive-on-absent (`TestBacklogHighWater_DeriveOnAbsent`); hand-edited low `last_seq` guard (`TestBacklogHighWater_HandEditedLowSeq`); malformed file byte-identical (`TestBacklogMalformed_ReportedEverywhereFileUntouched`); missing file = empty queue (`TestBacklogLoad_MissingFileIsEmptyQueue`, `TestBacklogAdd_CreatesVersion1File`); round-trip field preservation (`TestBacklogRoundTrip_PreservesFieldsAndVersion`); no .tmp residue (`TestBacklogWrite_NoTmpResidue`); concurrent add unique IDs (`TestBacklogConcurrentAdd_UniqueIDs`); no lead-role guard (`TestBacklogStore_NoLeadRoleGuard`). CLI-process-level ACs (AC-TODO-001, 003–006) are **M2-pending**, not FAIL.

**E2 — Cross-platform.**

```
$ go build ./...                                        → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...              → exit 0
$ GOOS=windows GOARCH=amd64 go vet ./internal/kanban/... → exit 0 (echo ALL_BUILDS_VET_OK)
```

**E3 — Coverage.** `go test -cover ./internal/kanban/...` → `coverage: 87.3% of statements` (≥85% target; see E1 output — package-level).

**E4 — Full suite (attribution per failure, all baseline-verified by file-removal A/B):**

```
$ unset MOAI_KANBAN … && go test ./... -count=1
```

Failing packages: `internal/cli`, `internal/config`, `internal/hook` (+ `internal/spec` in one run, see below). Per-failure attribution, each verified by re-running with the M1 files moved aside (A/B):

- `internal/cli`: `--- FAIL: TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey` — the known pre-existing origin/main failure (lead-confirmed). Sole cli failure. Baseline.
- `internal/config`: `--- FAIL: TestAutonomyTierReader` — fails identically with M1 files removed → pre-existing baseline, not this SPEC.
- `internal/hook`: `--- FAIL: TestPreTool_AstGrepSkipReasonSurfaces` ("SystemMessage is empty: the ast-grep skip reason was dropped somewhere in the three-frame chain") — fails identically with M1 files removed → pre-existing baseline, not this SPEC.
- `internal/spec`: failed once under full-suite parallel load (85.9s), passed in two isolated re-runs with the SPEC dir present (`ok … 38.040s`, `ok … 45.995s`) → flaky under load, not deterministically broken by this SPEC's directory.

All other packages green. No failure attributable to M1.

**E5 — Lint.**

```
$ golangci-lint run --timeout=2m
0 issues.
```

Baseline was also `0 issues.` — zero NEW findings.

**E6 — Commit.** M1 commit lands on `worktree-kanban-todo-cli` with this progress update and the SPEC artifacts (frontmatter `status: draft → in-progress` on all four documents). Not pushed — PR/push awaits the kanban lead's M1 verdict per the semi-autonomous progression.

**Gaps.** (1) Per-file coverage of `backlog_store.go` alone was not measured (delegation 2 died mid-measurement; package-level 87.3% cited instead). (2) `internal/spec` flakiness under full-suite parallel load is observed, not root-caused — outside this SPEC's scope. (3) RED evidence is the compile-failure form; a behavioral (assertion-level) RED capture was lost with delegation 1's context.

**Residual-risk.** (1) The `internal/cli`/`internal/config`/`internal/hook` baseline failures were verified in THIS worktree; CI on the eventual PR runs on a clean runner where worktree-cwd-dependent behavior may differ. (2) M2's CLI verbs will exercise the store through real process concurrency (8-process `add`), a stronger test than the in-process concurrency test here.

### M2 — CLI verbs (COMPLETE; orchestrator-direct per lead instruction after M1 PASS)

Lead verdict on M1: **PASS** (2026-08-14, lead re-measured independently: the 3 evidence-test cluster green, coverage 87.3% reproduced, `GOOS=windows go vet` exit 0, and the source-level check that `rec.LastSeq++` exists only inside the `Mutate` callback — no issuance path outside the lock). M2 instruction: **implement directly, do not delegate** (2/2 manager-develop autocompact deaths vs. the surviving orchestrator-finished tail; M2 is low-uncertainty wiring over the finished store API).

Deliverables: `internal/cli/todo.go` (226 lines — `newTodoCmd` + `add`/`list --json`/`done <n>`/`next [<n>] [--spec]`, registered via `init()`; exit 0/1; stdout structured, stderr human), `internal/cli/todo_test.go` (13 tests + re-exec helper).

**E8 — RED (TDD, verbatim pre-GREEN).** Test file written first; captured against the no-implementation tree:

```
$ unset MOAI_KANBAN … && go vet ./internal/cli/
internal/cli/todo_test.go:15:2: "path/filepath" imported and not used
internal/cli/todo_test.go:29:9: undefined: newTodoCmd
internal/cli/todo_test.go:44:33: undefined: todoBacklogPath
…
```

One behavioral test fix mid-green: the 8-process helper initially reported ids via `t.Logf`, whose output the non-`-v` helper run drops — the id-parse assertion failed (`distinct issued ids printed = 0`); fixed by printing to process stdout. The helper then runs the FULL cobra verb (not `store.Add` directly), so AC-TODO-001's "concurrent `moai todo add` processes" is exercised on the binary's own code path.

**E1 — M2 test matrix.**

```
$ unset MOAI_KANBAN … && go test ./internal/cli/ -run 'TestTodo' -count=3
ok  	github.com/modu-ai/moai-adk/internal/cli	10.086s
```

13/13 pass ×3 consecutive runs: AC-TODO-001 (`TestTodoConcurrentAdd_8Processes` — 8 OS processes through the cobra `add` verb; `list --json` then reports 8 items with all texts, 8 distinct ids), AC-TODO-003 (`TestTodoList_LockFreeWhileForeignProcessHoldsLock` — foreign process parks inside `Mutate` holding `backlog.lock`; `list` succeeds; `TestTodoList_JSONStructured` — valid JSON, full record shape), AC-TODO-004 (`TestTodoDone_RemovesByID`, `TestTodoDone_MissReportedFileUntouched` — byte-identical on miss), AC-TODO-005 (`TestTodoNextBare_ReadOnlyOldestFirst` — oldest-first, file byte-identical), AC-TODO-006 (`TestTodoNextPick_OneLockedWrite` — exactly t2 `picked` + `spec_id: SPEC-X-001`, others untouched), AC-TODO-014 (`TestTodoCmd_NoAskUserQuestion` — static grep guard + negative control asserting the detector flags a synthetic violation), plus the acceptance §C edges: empty-text add rejected with no file write, out-of-range `done`/`next` miss reported + file byte-identical. M3/M4 ACs (015 skill/mirror, 016 full cross-platform sweep) remain pending.

**E2 — Cross-platform.**

```
$ go build ./...                                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...                → exit 0
$ GOOS=windows GOARCH=amd64 go vet ./internal/cli/... ./internal/kanban/... → exit 0 (BUILDS_VET_OK)
```

**E3 — Coverage (per AC §D "new store and CLI files").**

```
$ go test ./internal/cli/ ./internal/kanban/ -coverprofile=/tmp/t12-m2-cover.out
$ go tool cover -func=… | grep -E 'todo\.go:|backlog_store\.go:'
→ new-file avg: 94.1% over 20 funcs (todo.go: 75–100% per func incl. todoBacklogPath/done/normalize/init at 100%; backlog_store.go: Load/Add/acquireLock/normalize at 100%, Mutate 92.3%)
```

(internal/cli package-level shows 77.0% — the package aggregate over hundreds of pre-existing files; the AC's target is the new files, which the per-func profile above measures.)

**E4 — Suite.** `./internal/kanban/...` ok (87.3%); `./internal/cli` fails only on the known pre-existing `TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey` (M1-attributed baseline; lead notes PR #1527/t20 addresses it). No other failures.

**E5 — Lint.** `golangci-lint run --timeout=2m` → `0 issues.` (unchanged baseline).

**E6 — Commit.** M2 commit on `worktree-kanban-todo-cli` (this progress update rides it). Not pushed — awaiting lead M2 verdict per semi-autonomous progression.

**Gaps.** (1) No built-binary smoke run (`moai todo` via the compiled binary) — the cobra surface is exercised in-process and through re-exec helpers of the same test binary; the binary wiring is init()-registration verified by compile. (2) `internal/cli` package-level coverage (77.0%) is below 85% but that aggregate is dominated by pre-existing files outside this SPEC's scope.

**Residual-risk.** (1) The 8-process test's uniqueness assertion parses each helper's stdout (`t<n> <pos>`); a future output-format change to `add` needs the test's parser kept in step. (2) `resolveProjectDir()` (CLAUDE_PROJECT_DIR → cwd) is the path anchor; a caller invoking `moai todo` from outside the project without CLAUDE_PROJECT_DIR will target that outside directory's `.moai/state` — same behavior as the session/goal commands, noted not changed.

### M3 — Skill shrink + template mirror (COMPLETE; orchestrator-direct)

Lead verdict on M2: **PASS** (2026-08-14; lead filled the binary-smoke gap personally: built binary `moai todo add "첫 카드"` → `t1 1` exit 0; 12 OS processes concurrent add → 12/12 items, 12 unique gapless ids, `last_seq: 12`; high-water live check `t1 t2 t3` → `done 3` → add → `t1 t2 t4` — no t3 reuse). M3 instruction: same-direct treatment, with the explicit warning NOT to `cp` the template twin (neutralized-variant hazard — CI catches SPEC-IDs/dates but not internal-vocabulary leaks).

**Pre-edit state (measured before writing)**: `diff .claude/skills/moai/workflows/todo.md internal/template/templates/.claude/skills/moai/workflows/todo.md` → **byte-identical twins**. With no neutralized variant to preserve in THIS tree, applying the same rewrite to both files IS the same delta; the no-cp rule is honored in its substance (the variant check was performed, not assumed — research.md §8's claim re-verified on the actual M3 tree).

**Line-count correction (M4, lead-requested)**: the original M3 report's "105 → 119 lines" mixed a measured pre-edit value with an UNMEASURED post-edit estimate — the 119 was never observed and is wrong. The measured three-point facts (each attributed to its tree, `wc -l`): worktree base `b73c81ab5` todo.md = **105**; post-M3 twins = **107**; origin/main at M4 time = **112** (main advanced past this branch's base and touched todo.md, +7 vs base). Net: **112 → 107 (−5) against current origin/main** (the merge-relevant frame); 105 → 107 (+2) against this branch's base. The shrink direction holds — the earlier "grew to 119" claim is retracted. The upstream todo.md change also means the twins may conflict at PR-merge time; flagged to the lead at M4.

**The delta (REQ-TODO-015 / D-h)**: shrink to "run the command" form. Preserved: queue philosophy ("one line of intent", "deliberately thin"), the operator-act header + kanban-dispatch cross-reference, the record-shape explanation (now framed as reading `list --json` output, with `last_seq` documented), the [HARD] operator-selection rule verbatim in intent, Boundaries, Outside-Kanban-Mode, all three cross-references. Removed: the direct file-manipulation instructions (the verbs table's model-driven Read→Write semantics, "Write the file atomically (write a sibling temp file, then rename)…"), and the obsolete "any other argument form is treated as a description" rule (the real CLI takes explicit subcommands). Added: the mapping table of the five real CLI invocations and the "do not read or write the file directly" instruction (the whole point of M1+M2). Line counts: see the M4 correction above (112 → 107 against origin/main).

**AC-TODO-015 verification (each command + output)**:

```
$ diff .claude/skills/moai/workflows/todo.md internal/template/templates/.claude/skills/moai/workflows/todo.md && echo TWINS_STILL_IDENTICAL
TWINS_STILL_IDENTICAL                                    ← same delta on both twins

$ grep -n 'SPEC-KANBAN\|SPEC-TODO\|2026-' <both twins>
(no matches; exit 1)                                     ← no SPEC ids, no internal dates

$ grep -nE 'Write the file|temp file|then rename|Read.*backlog\.json.*Write' <both twins>
(no matches; exit 1)                                     ← removed instruction tokens absent (grep-for-removals clause)

$ make build → "catalog.yaml updated successfully (12407 bytes)", MAKE_EXIT=0
$ git status --porcelain internal/template/catalog.yaml
(no output)                                              ← regenerated catalog is byte-identical: the catalog's
                                                           hash subject set does not include workflow skills (grep 'todo'
                                                           in catalog.yaml → no entry). B5's hazard does not arise this
                                                           time; recorded as verified-absent, not assumed-absent.

$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	44.167s  ← neutrality gate (CI-equivalent strict mode) exit 0
```

Docs-vs-verbs cross-check (lead checklist item 5): cobra `Use:` strings in todo.go (`add <text>`, `list`, `done <n>`, `next [<n>]`) vs the skill's command table — all five documented invocations match the implemented surface (grep output captured above).

**Working-tree hygiene note**: the full-suite runs from M2 verification had rewritten `SPEC-HOOK-PRETOOL-PERF-001/{baseline,postchange}.md` in place (the known fixture-rewrite side effect); both were `git checkout --`-restored before the M3 commit — only the intended paths are in it.

**Gaps.** (1) The shrunk skill's behavioral effect (the model actually running `moai todo` instead of Read→Write) is verified by document inspection, not by an A/B model run — the skill-ab-testing methodology was out of M3's milestone scope (plan.md §F keeps M3 a documentation delta). (2) No 4-locale surface touched: the todo skill exists only in English (verified — no `todo.*.md` locale twins exist to mirror).

**Residual-risk.** The catalog-parity conclusion holds for THIS template state; if a future change adds workflow skills to the catalog's hash subject set, the B5 commit obligation revives for future edits of this file.

### M4 — Cross-platform + full-suite verification (COMPLETE; orchestrator-direct)

Lead verdict on M3: **PASS** (2026-08-14; lead re-measured: `cmp -s` twins identical at 107 lines, SPEC-ID regex `SPEC-[A-Z0-9-]+-[0-9]{3}` → 0 matches — the `<SPEC-ID>` placeholder in command syntax is not a violation —, catalog workflows-path entries 0, docs↔verbs match). One numeric correction requested and applied above (the unmeasured "119" line-count claim retracted; three-point measured values recorded with tree attribution).

**E2 — Cross-platform (exit codes cited verbatim; note the vet gate is FULL-REPO this time, stronger than M2's package-scoped run):**

```
$ go build ./...                          → BUILD_UNIX_EXIT=0
$ GOOS=windows GOARCH=amd64 go build ./... → BUILD_WIN_EXIT=0
$ GOOS=windows GOARCH=amd64 go vet ./...   → VET_WIN_EXIT=0
```

**E4 — Full suite** (`unset MOAI_KANBAN … && go test ./... -count=1`):

Exactly three failing tests, all previously A/B-attributed to baseline in M1 (files-removed reruns: each fails identically with the M1/M2 files absent):

- `TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey` (internal/cli) — the lead-declared known origin/main failure (t20; PR #1527 addresses it per lead).
- `TestAutonomyTierReader` (internal/config) — A/B'd in M1.
- `TestPreTool_AstGrepSkipReasonSurfaces` (internal/hook) — A/B'd in M1.

`internal/spec` (flaky in one M1 run) passed this run. **Zero NEW failures** — the lead's "A/B any failure beyond the known one" instruction did not trigger. All other packages green.

**E5 — Lint.**

```
$ golangci-lint run --timeout=5m → "0 issues." LINT_EXIT=0
```

Baseline (pre-M1, same tree family) was also "0 issues." — zero NEW findings across the entire SPEC's diff.

**E6 — Commit.** This M4 correction+evidence commit rides the same branch; push remains withheld for the lead's push/PR instruction.

**Gaps.** (1) The upstream todo.md change (origin/main advanced +7 lines past this branch's base) means a PR merge may conflict on the twins — resolution belongs to the merge step, flagged to the lead. (2) The M1 A/B attributions for the config/hook failures were performed against THIS worktree's base; the lead's note that #1527 addresses the cli one (and likely the config one, `TestAutonomyTierReader`, per the lead's M2 message) is their attribution on main — not re-measured here.

**Residual-risk.** CI on the eventual PR runs on a clean runner against current origin/main: the 3 known failures' behavior there is governed by main's state (post-#1527), not this branch's base.

### Post-M4 — rebase + push + draft PR (lead-instructed)

M4 verdict: **PASS**. Lead approved option (b): rebase onto origin/main, with the upstream +7-line todo.md change identified as `b689f492c` (#1526) — a batch-admission policy clause appended after the [HARD] block, which must SURVIVE the M3 shrink (it is queue policy, not file-manipulation guidance).

**Rebase (measured)**: `git rebase origin/main` (24e2bd402) completed with NO manual conflicts — git's 3-way merge produced the intended union: the M3 rewrite WITH the upstream batch-admission paragraph (lines 78–83) and its `kanban-dispatch.md` § Batch admission cross-reference intact on both twins. Neither side was wholesale-adopted; both survive. New chain: `c839a9a4e`(M1) `99dc73916`(M2) `4deaf395c`(M3) `1f591cbeb`(M4).

```
$ cmp -s <local twin> <template twin> && echo TWINS_IDENTICAL   → TWINS_IDENTICAL (114 lines = 107 + upstream 7)
$ go build ./...                                                 → BUILD_OK
$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...    → ok 24.176s
$ unset … && go test ./... -count=1 → --- FAIL ×2 (TestHandleCodexReviewGate…, TestPreTool_AstGrepSkipReasonSurfaces)
$ go build / GOOS=windows build / GOOS=windows go vet ./...      → 0 / 0 / 0
```

**TestAutonomyTierReader resolved by #1527** exactly as the lead predicted — known failures 3 → 2 at rebase. Zero NEW failures on the rebased tree.

**Push + PR**: branch `worktree-kanban-todo-cli` pushed; **draft PR #1529** opened (per the repo's draft-first discipline so auto-merge does not decide timing). PR body carries the milestone evidence summary + the known-failure table with per-failure attribution. Awaiting the lead's verification + undraft verdict. CodeRabbit review-limit caveat noted (row state ≠ verdict; read the comment body).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
