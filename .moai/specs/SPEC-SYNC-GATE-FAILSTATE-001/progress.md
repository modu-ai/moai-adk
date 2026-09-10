# Progress — SPEC-SYNC-GATE-FAILSTATE-001

Card: t624 · Branch: `WT-sync-gate-failstate` · Plan-phase tree: `fa96fe644`

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts authored 2026-09-10: `spec.md`, `plan.md`, `acceptance.md`, this file.
  Status `draft`. Tier M (no `design.md` / `research.md`).
- Tree check: `git rev-parse --show-toplevel` →
  `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t624`; `git branch --show-current` →
  `WT-sync-gate-failstate`; `git rev-parse HEAD` → `fa96fe644fcff8a15ac336833a4e816fc0a46fe3`.
- SPEC ID regex check executed:
  `[[ "SPEC-SYNC-GATE-FAILSTATE-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`.
- Defect evidence: `.moai/reports/t624/h01-repro-develop.md` (tree `d5dc42959`). Target files
  are identical at `fa96fe644` (acceptance.md §D.0 L-01).
- Plan-phase decisions B1-B6 were resolved by lead ruling on 2026-09-10 and are recorded in
  `plan.md §B`. No clarification markers remain.
- `plan_status: audit-ready`

### Plan-audit and kickoff provenance (recorded 2026-09-10)

| Round | Verdict | Score | Blocking | Source |
|---|---|---|---|---|
| r1 | FAIL | 0.60 | 11 | `.moai/reports/t624/plan-audit.md` |
| r2 | FAIL | 0.86 | 4 | `.moai/reports/t624/plan-audit-r2.md` |
| r3 | FAIL | 0.86 | 1 (NEW-1) | `.moai/reports/t624/plan-audit-r3.md` |

- 2026-09-10 — plan-audit round 3 FAIL (score 0.86, 1 blocking NEW-1) — `.moai/reports/t624/plan-audit-r3.md`; operator decision relayed by the factory lead: PASS-with-debt, Implementation Kickoff approved on condition that NEW-1 is resolved before M1; debt paid in the first run-phase commit (this change set). Standing condition: if at M1 an observed RED reason again differs from its stated reason, the run stops and reports without editing the criterion.
- Earlier verdicts: `.moai/reports/t624/plan-audit.md` (round 1, FAIL 0.60) and `.moai/reports/t624/plan-audit-r2.md` (round 2, FAIL 0.86).

## §E.2 Run-phase Evidence

### M1 — RED tests (test-only commit), recorded 2026-09-10 by manager-develop

Every command below ran in this run, in `.claude/worktrees/t624`. Long outputs are persisted under
`.moai/reports/t624/` and cited by path; the excerpts quoted here are verbatim lines from those files.

#### Pre-flight (plan.md §E), measured on `1a3b12ce5` before any change

| # | Command | Verbatim stdout | Exit |
|---|---|---|---|
| P1 | `git rev-parse --show-toplevel` | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t624` | 0 |
| P2 | `git branch --show-current` | `WT-sync-gate-failstate` | 0 |
| P3 | `git rev-parse --short HEAD` | `1a3b12ce5` | 0 |
| P4 | `cmp .claude/hooks/moai/sync-phase-quality-gate.sh internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` | *(empty)* | 0 |
| P5 | `git diff --stat fa96fe644 -- <2 hook copies> <2 quality-gates-quality.md copies>` | *(empty)* | 0 |
| P6 | `unset MOAI_SYNC_GATE_BLOCKING MOAI_AUTONOMY_TIER && go test ./internal/hook/ -run 'TestAC004_SyncGateAdvisoryAtFullyAutonomous\|TestAC002_NonSyncHeadSkipsVetBuild\|TestHookWrapperCopiesStayIdentical' -count=1 -v` | full output `.moai/reports/t624/m1-preflight-run.txt`; last line `ok  	github.com/modu-ai/moai-adk/internal/hook	8.597s` | 0 |

(In P6 the `\|` is table escaping; the executed selector carries a plain `|`.)

#### Test commit (ordering witness)

- Commit: `f611a7060106e13875510217f13d07c2a659c5a1` — `test(t624): M1 RED tests for sync gate failure-state record`.
  It contains exactly `internal/hook/sync_gate_failstate_test.go` (+820), `internal/hook/sync_gate_failstate_unix_test.go`
  (+246), and the `spec.md` frontmatter `status: draft` → `status: in-progress` (1 line; `updated:` was already
  `2026-09-10`). No hook and no document changed.
- `git log -1 --format='%(trailers:key=Authored-By-Agent,valueonly)'` → `manager-develop`.
- Static checks on the test files (before the commit; the files are unchanged since):
  `gofmt -l internal/hook/` → *(empty)*, exit 0; `go vet ./internal/hook/` → *(empty)*, exit 0
  (`m1-vet-native.txt`); `GOOS=windows GOARCH=amd64 go vet ./internal/hook/` → *(empty)*, exit 0 (`m1-vet-windows.txt`).
- Harness fixes before the commit: **none.** A dry run of the same selector (scratch output, not evidence) showed
  every row executing, with no guard message, panic, or `[no tests to run]`.

#### RED run (release-blocking cells, lead ruling B4)

- Tree: commit `f611a7060106e13875510217f13d07c2a659c5a1`.
- Command: `unset MOAI_SYNC_GATE_BLOCKING MOAI_AUTONOMY_TIER && go test ./internal/hook/ -run '^TestSyncGateFailState' -count=1 -v`
- Exit code: **1**. Full stdout: `.moai/reports/t624/m1-red-run.txt`; every assertion and log line:
  `.moai/reports/t624/m1-red-excerpt.txt` (86 lines).
- Swept set: `grep -c '^=== RUN' m1-red-run.txt` → `69`; `grep -c 'no tests to run' m1-red-run.txt` → `0`.
- Top-level tests that ran (13): `TestSyncGateFailState_AC001_FailureRedeliveredOnSameHead` (FAIL),
  `_AC002_NewFailingHeadBlocksWithNewResult` (PASS), `_AC003_PassThenSameHeadStaysSilent` (PASS),
  `_AC004_StopHookActiveDefersRedelivery` (FAIL), `_AC005_UnknownAndLegacyRecordsRegate` (FAIL),
  `_AC005_TornWriteNeverSilentPass` (FAIL), `_AC006_RunningRecordStaleWindow` (FAIL),
  `_AC007_NoPathLooserThanToday` (PASS), `_AC008_RedeliveryFollowsModeResolution` (FAIL),
  `_AC013_RetryByDeletionNoStaleAuxState` (PASS), `_AC014_StaleWindowEqualsRegisteredTimeout` (FAIL),
  `_AC006c_RetryBoundAndNotice` (FAIL), `_AC015_NoticesNeverBlockOrConsumeCap` (FAIL). Last line: `FAIL	github.com/modu-ai/moai-adk/internal/hook	94.627s`.
- `grep '\[regression-guard\]' m1-red-excerpt.txt | grep -v 'call1='` → *(empty)*: no assertion tagged
  regression-guard failed (the only regression-guard-tagged lines are AC-008 `t.Logf` lines carrying `call1=`).

Verbatim excerpt (from `m1-red-excerpt.txt`):

```
sync_gate_failstate_test.go:332: AC-001 [release-blocking] call1 stdout (237 bytes)="{\"hookSpecificOutput\":{\"hookEventName\":\"Stop\",\"decision\":\"block\",\"reason\":\"go vet failed\"},\"systemMessage\":\"sync-phase quality gate BLOCKED: go vet failed (go vet=1 go build=0 deps_modified=0). Detail: .moai/logs/sync-quality-gate.log\"}\n" stub=2; call2 stdout (0 bytes)="" stub=2
sync_gate_failstate_test.go:337: AC-001 [release-blocking]: call 2 stdout is not byte-identical to call 1 (call1 237 bytes, call2 0 bytes); call2=""
sync_gate_failstate_test.go:456: AC-005 L1 run-and-block [release-blocking]: stub invoked 0 time(s) on this call; want >= 1 (checks must run)
sync_gate_failstate_test.go:467: AC-005 U4 record-format [release-blocking]: record = "58f144313005495b771a5498be3f1fc50f2971db"; want "58f144313005495b771a5498be3f1fc50f2971db fail"
sync_gate_failstate_test.go:530: AC-005 TA1 [release-blocking]: stub invoked 2 time(s) on this call; want 0 (checks must not run)
sync_gate_failstate_test.go:564: AC-006 a50 [release-blocking]: stub invoked 2 time(s) on this call; want 0 (checks must not run)
sync_gate_failstate_test.go:693: AC-008 A4 [release-blocking]: call 2 stdout is not byte-identical to call 1's block (call1 241 bytes, call2 0 bytes); call2=""
sync_gate_failstate_test.go:771: AC-014 [release-blocking] stale-window assignments found: []
sync_gate_failstate_test.go:807: AC-014 [release-blocking] settings timeouts for sync-phase-quality-gate.sh entries: [60]
sync_gate_failstate_unix_test.go:114: AC-006c run 1: marker observed for this run; process group killed
sync_gate_failstate_unix_test.go:114: AC-006c run 2: marker not observed; hook exited on its own; output=""
sync_gate_failstate_unix_test.go:119: AC-006c [release-blocking] assertion 2: run 2 did not observe its own marker — the one allowed stale re-run did not start
sync_gate_failstate_unix_test.go:128: AC-006c [release-blocking] run 3 stdout (0 bytes)="" stub delta=0 exit=0
sync_gate_failstate_unix_test.go:136: AC-006c [release-blocking] assertion 5: run 3 stdout is not a systemMessage saying the gate run has not completed and naming deletion of sync-quality-gate.last; stdout=""
sync_gate_failstate_unix_test.go:211: AC-015 N1 [release-blocking] summary: nonzero-exit=0 missing-systemMessage=9 decision-count=0 stub-delta=0
sync_gate_failstate_unix_test.go:224: AC-015 N2a [release-blocking] summary: nonzero-exit=0 missing-systemMessage=8 decision-count=1 stub-delta=2
sync_gate_failstate_unix_test.go:226: AC-015 N2a [release-blocking]: record after run 9 = "a570028487af8b103d10c98dd93bb67e6e8cf960"; want "a570028487af8b103d10c98dd93bb67e6e8cf960 running"
sync_gate_failstate_unix_test.go:239: AC-015 N2b [release-blocking] summary: nonzero-exit=0 missing-systemMessage=0 decision-count=9 stub-delta=18
sync_gate_failstate_unix_test.go:242: AC-015 swept invocations: 27 (want 27)
```

#### Row-by-row comparison (operator condition)

Each check run on a Go fixture invokes the stub twice (`go vet`, then `go build`).

| Row | Class | Stated RED reason / "must pass" | Observed on `f611a7060` | Match |
|---|---|---|---|---|
| AC-001 | release-blocking | call 2 writes 0 bytes; stored failure not re-delivered | call1 237-byte block, call2 0 bytes; stub 2 → 2; both exit 0; only the byte-identity assertion failed | yes |
| AC-002 | regression-guard | must pass | PASS | yes |
| AC-003 | regression-guard | must pass | PASS | yes |
| AC-004 F1-F4 (a) | release-blocking | passes vacuously today (nothing re-delivered) | no (a) assertion failed in any form | yes |
| AC-004 F1-F4 (b) | release-blocking | not re-delivered on the unflagged call (0 bytes) | each form: 0 bytes vs call1 237 bytes | yes |
| AC-004 (c) | release-blocking | not re-delivered (0 bytes) | 0 bytes vs 237 bytes | yes |
| AC-005 L1 | release-blocking | checks never run, stdout empty (bare SHA = HEAD short-circuits) | stub 0; stdout `""`; record = bare SHA | yes |
| AC-005 U1-U5 run-and-block | regression-guard | must pass | no regression-guard assertion failed in any U row | yes |
| AC-005 U1-U5 record-format | release-blocking | red for the REQ-001 reason (today rewrites a bare SHA) | each U row: record = bare 40-hex SHA, want `<HEAD> fail` (U4 ran: non-root session, read step applied) | yes |
| AC-005 TA1 | release-blocking | notice absent; `<HEAD> running` read as foreign, checks run, block | stub 2; stdout is the block JSON (carries `"decision"`) | yes |
| AC-005 TA2, TA3, TA5-bare, TA5-pass | regression-guard | must pass | PASS (4 subtests) | yes |
| AC-005 TA4 | release-blocking | checks never run, stdout empty | stub 0; stdout `""` | yes |
| AC-005 TB1 | release-blocking | same as TA1 | stub 2; block JSON | yes |
| AC-005 TB2 | regression-guard | must pass | PASS | yes |
| AC-006 a0, a50 | release-blocking | notice absent; checks run (count > 0) and block | each: stub 2; block JSON | yes |
| AC-006 b61, b70, b120 | regression-guard | must pass | PASS (3 subtests) | yes |
| AC-006c | release-blocking | assertion 2 fails (run 2 never invokes the stub) and 5 fails (run 3 empty); 1, 3, 4 pass; no harness wait | run 1 marker observed + group killed; run 2 marker not observed, hook exited, output `""`; run 3 0 bytes, stub delta 0, exit 0; errors only for assertions 2 and 5 | yes |
| AC-007 R1-R17 (R10a, R10b) | regression-guard | must pass | PASS (18 subtests) | yes |
| AC-008 A1-A3, A5-A8 | regression-guard | must pass | PASS (7 subtests); call 2 `""`, stub 2 → 2 in each | yes |
| AC-008 A4 | release-blocking | call 2 writes 0 bytes | call1 241-byte block, call2 0 bytes | yes |
| AC-008 A9 | release-blocking | call 2 writes 0 bytes | call1 237-byte block, call2 0 bytes | yes |
| AC-013 D1, S1 | regression-guard | must pass | PASS (2 subtests) | yes |
| AC-014 | release-blocking | named stale-window variable absent; settings entry reads 60 | assignments found `[]`; settings timeouts `[60]` | yes |
| AC-015 N1 | release-blocking | setup assertion 2 fails; all 9 stdouts empty (`systemMessage` fails); decision-count and stub-count pass | setup assertion 2 failed; missing-systemMessage=9, decision-count=0, stub-delta=0; notice text absent in 9 of 9 | yes |
| AC-015 N2a | release-blocking | decision count 1; stub +1 check run (2 invocations); 8 empty stdouts; final record bare SHA | decision-count=1, stub-delta=2, missing-systemMessage=8, final record bare SHA | yes |
| AC-015 N2b | release-blocking | decision count 9; stub +9 check runs (18 invocations); `systemMessage` assertion passes | decision-count=9, stub-delta=18, missing-systemMessage=0 | yes |
| AC-015 swept set | — | 27 invocations (9 × 3) | 27 | yes |

**Result: every release-blocking row failed for its named RED reason, and every regression-guard row passed.**

#### Test-side contracts M2 must satisfy (choices the criteria leave open)

- AC-014 finds the stale window as exactly one line-start assignment `NAME=<integer>` whose name contains `STALE`
  (case-insensitive), in the **template** hook.
- AC-006c assertion 5 and the AC-015 N1 notice check: the lowercased stdout contains `not completed`,
  `sync-quality-gate.last`, and `delet`.
- Platform split: AC-006c and all of AC-015 (N1, N2a, N2b) are in `sync_gate_failstate_unix_test.go`
  (`//go:build !windows`). N2a and N2b need no process group; they are there so the 27-invocation sweep stays in one
  test. U4 skips on Windows and as root.
- AC-005 TA5 is exercised in both named forms (`TA5-bare`, `TA5-pass`). The interrupted-run harness ages the record
  after every interrupted run (AC-006c wording). It carries a 90 s guard against a hook that neither writes the marker
  nor exits; the guard never fired (no guard line in `m1-red-run.txt`).

#### acceptance.md §D.0 ledger gaps, measured at M1 (tree `f611a7060`)

| Gap | Command | Verbatim stdout | Exit |
|---|---|---|---|
| L-10 phrase, local doc | `/usr/bin/grep -c "Audit ALL of the following manifest files present at project root" .claude/skills/moai/workflows/sync/quality-gates-quality.md` | `1` | 0 |
| `manifest audit`, local hook | `/usr/bin/grep -c "manifest audit" .claude/hooks/moai/sync-phase-quality-gate.sh` | `2` | 0 |
| `SPEC-`, local hook | `/usr/bin/grep -c "SPEC-" .claude/hooks/moai/sync-phase-quality-gate.sh` | `0` | 1 |
| L-19 pattern, local copies | `/usr/bin/grep -cE '(^\|[^A-Za-z0-9_])t[0-9]{2,4}([^A-Za-z0-9_]\|$)\|20[0-9]{2}-[0-9]{2}-[0-9]{2}' .claude/hooks/moai/sync-phase-quality-gate.sh .claude/skills/moai/workflows/sync/quality-gates-quality.md` | `.claude/hooks/moai/sync-phase-quality-gate.sh:0` and `.claude/skills/moai/workflows/sync/quality-gates-quality.md:0` | 1 |
| `SPEC-`, template doc (AC-009, not in the ledger) | `/usr/bin/grep -c "SPEC-" internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md` | `0` | 1 |

(`\|` in the L-19 cell is table escaping; the executed regex carries a plain `|`.) The local values equal the
template values recorded at L-10, L-12, L-14, and L-19, so the "L-13/L-16 imply equal values" inference is now
measured.

**L-21 (commit-SHA coverage of the delegated template-leak guards).**
- Command, run verbatim from the fenced block:
  `go test ./internal/template/ -run 'TestTemplateNoInternalContentLeak|TestLeakClassNoDateShaInDefaultTier|TestC7PackageRestriction' -count=1 -v`
  → exit 0; full output `.moai/reports/t624/m1-l21-run.txt`. The names were confirmed before the result was read:
  `=== RUN   TestTemplateNoInternalContentLeak`, `=== RUN   TestLeakClassNoDateShaInDefaultTier`,
  `=== RUN   TestC7PackageRestriction`, all three `--- PASS`; `no tests to run` count `0`; last line
  `ok  	github.com/modu-ai/moai-adk/internal/template	1.921s`.
- Tier read (`internal/template/internal_content_leak_test.go`):
  - Both template paths are scanned: `.sh` and `.md` are in `leakTextExtensions`, and neither path appears in
    `skipPaths` (grep exit 1).
  - The default tier (`leakClasses` C1, C2, C3, C4, C5, C1b, C6, C7, S3, C1c, C2b, C2c, C8) carries **no**
    commit-SHA pattern. `TestLeakClassNoDateShaInDefaultTier` asserts that no default-tier class matches the
    short-SHA probe `a1b2c3d `.
  - The only SHA class, `S2-short-sha-sentence-final` (`\b[0-9a-f]{7,8}([\s\.,;:!?]|$)`), sits in
    `strictLeakClasses`. That tier runs only with `MOAI_TEMPLATE_LEAK_STRICT=1`, which the L-21 command does not
    set, and it matches 7-8 hex digits, not 7-40.
  - **Verdict: neither path is in a tier whose class checks 7-40-hex SHA tokens.** Both are uncovered.
- Contingency applied (acceptance.md §D.0, L-21 contingency), one invocation per uncovered file:
  - `/usr/bin/grep -nE '(^|[^0-9A-Za-z])[0-9a-f]{7,40}([^0-9A-Za-z]|$)' internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` → *(empty)*, exit 1
  - `/usr/bin/grep -nE '(^|[^0-9A-Za-z])[0-9a-f]{7,40}([^0-9A-Za-z]|$)' internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md` → *(empty)*, exit 1
  - Zero hits, so there is nothing to classify and no undecidable hit; REQ-012's commit-SHA clause is not an explicit
    gap at this tree. The check is a one-time measurement, not a standing guard, so it has to be re-run after the
    M2 and M3 edits to these files.

#### Gaps (not observed in M1)

- AC-009, AC-010, AC-011, AC-012 are not M1 rows (plan.md §F M1 scope); their checks run after M2/M3.
- AC-013's reviewer read (the hook's leading comment names deletion of `.moai/state/sync-quality-gate.last` as the
  retry) is not automated; it applies after M2 rewrites the header.
- `golangci-lint` was not run on the new test files (not in the M1 deliverable list); `gofmt` and `go vet`
  (native and `GOOS=windows`) were.
- The Windows execution of the portable rows was not observed: only `GOOS=windows go vet` compilation was.
- Coverage is not measured: M1 adds tests only.

#### Residual risk

- The notice-text and stale-window name contracts above are test-side readings of REQ-009 and AC-014 wording. An M2
  implementation that satisfies the requirement with different wording or a different variable name would fail these
  tests for a reason the SPEC does not state.
- The AC-006 a50 and b61 rows rely on the hook reading the clock within 10 s (a50) and 1 s of slack being harmless (b61
  grows toward stale). Under heavy load an M2 hook slower than 10 s could flip a50; the boundary is not at risk on
  today's hook, which ignores age.
- The harness measures today's hook; a stub that ignores `$1` or a PATH where the stub is shadowed would change the
  counts, but the observed counts (2 per check run, 0 on short-circuit) match the named reasons, which argues against
  that.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Recorded 2026-09-10 by the lane orchestrator (lane-10) before the first run-phase `Agent()` spawn.
Implementation Kickoff Approval: granted by the operator (relayed by the factory lead) as PASS-with-debt after plan-audit r3; the debt (NEW-1) was paid in `1a3b12ce5` before this spawn.

Input parameters:
- tier: M
- scope: 2 byte-identical hook copies (`internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh`, `.claude/hooks/moai/sync-phase-quality-gate.sh`), 2 doc copies (`quality-gates-quality.md`, template + local), new Go tests under `internal/hook/`
- domain count: 2 (shell hook behavior + Go test harness; doc wording is a separate milestone)
- file language mix: bash + Go + markdown
- concurrency benefit: LOW — coding-heavy, milestones depend on each other (M1 tests pin the behavior M2 implements)

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | multi-file behavioral change with release-blocking criteria, not a trivial edit |
| serial | **yes** | coding-heavy work with sequential milestone dependencies; one write-capable agent at a time |
| fanout | no | not multi-domain research; concurrent write agents on one worktree are prohibited |
| sweep | no | not a high-volume uniform mechanical transform |

Decision: serial

Justification: the run phase is coding-heavy (a stateful shell hook plus a Go fixture harness), and each milestone consumes the previous one's output — M1's test-only commit fixes the RED reasons that M2's implementation must turn green. Anthropic's coding-task parallelism caveat applies, and the one-writer-per-worktree rule rules out concurrent write agents. `manager-develop` is spawned once per milestone.
