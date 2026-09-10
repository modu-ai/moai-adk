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

### AC-013 amendment (recorded 2026-09-11 by manager-spec)

- **Trigger:** in the run-phase M4 mutant probes (HEAD `989ef144b`), mutant M18 survived.
  Replacing hook line 372 (`rm -f "$PAYLOAD_FILE" 2>/dev/null || true`) with `:` left
  `TestSyncGateFailState_AC013_RetryByDeletionNoStaleAuxState` at exit 0, D1 and S1 PASS.
  Evidence: `.moai/reports/t624/m4-mutant-M18.txt`.
  - **Why S1 missed it:** its failing call 2 rewrites the payload before call 3 reads it.
- **Decision:** lead decision A. AC-013 gains rows that kill M18 directly. Approved without a
  new plan-audit round under the kickoff debt rule. SPEC status is unchanged (`in-progress`).
- **Rows added:**
  - **S2:** the payload file is absent at every stub invocation of a check-running call that
    follows a stored block.
  - **S3:** a partial write failure (payload `mv` refused by a PATH shim, `fail` record
    written) re-gates on the next call instead of re-delivering the stale block.
- **Knock-on edits:**
  - §D.16: M18 now names S2 and S3; S1 is recorded as not an M18 target.
  - plan.md §H: the stale-payload risk row now cites S2 and S3.
- **For the run phase:** the M18 probe re-run on the tree where S2/S3 land is their RED
  evidence. No RED-now cell exists (acceptance.md §D.13 class note).

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

### M2 — outcome record and auxiliary state in the hook (GREEN), recorded 2026-09-11 by manager-develop

**Interruption note.** The first M2 verification batch ran on 2026-09-10 and was never reported. An
API rate limit ended the session first. The coordinator's instruction on resumption was to treat
those `m2-*.txt` files as unattributed. Every verification below was therefore **re-run on
2026-09-11 in the resumed run**, and each file under `.moai/reports/t624/m2-*.txt` was overwritten
by that re-run. One fact from the interrupted session is carried here only because it explains the
committed code, not as evidence:

- The first selector run against the new hook failed `TestSyncGateFailState_AC014_StaleWindowEqualsRegisteredTimeout`
  with `want exactly one stale-window assignment; got [SYNC_GATE_STALE_WINDOW STALE_RERUN]`.
- Cause: an internal flag named `STALE_RERUN=0` also matched the test's `STALE` contract.
- Fix: the flag was renamed `RERUN_OF_RUNNING` in the hook. The test was not changed.
- The re-run below is the evidence that this fix holds.

#### Pre-flight (resumed run, before any commit)

| # | Command | Verbatim stdout | Exit |
|---|---|---|---|
| P1 | `git rev-parse --show-toplevel` | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t624` | 0 |
| P2 | `git branch --show-current` | `WT-sync-gate-failstate` | 0 |
| P3 | `git rev-parse --short HEAD` | `2232e1a5f` | 0 |
| P4 | `git --no-optional-locks status --short` | ` M .claude/hooks/moai/sync-phase-quality-gate.sh` / ` M internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` / five `?? .moai/reports/t624/m2-*.txt` | 0 |
| P5 | `git diff --stat` | both hook copies `284 +++++++++++++++++----`; `2 files changed, 466 insertions(+), 102 deletions(-)` | 0 |
| P6 | `cmp .claude/hooks/moai/sync-phase-quality-gate.sh internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` | *(empty)* | 0 |

Neither M1 test file appears in P4 or P5, so no test file changed.

#### Hook commit

- Commit `6f098a45e8ec098b11e9cbbd6323d1c6dd6c60f7`:
  `feat(t624): M2 record gate outcome and re-deliver stored failure in sync quality gate`.
  `git show --stat` lists exactly the two hook copies (`2 files changed, 466 insertions(+), 102 deletions(-)`).
- Trailer: `git log -1 --format='%(trailers:key=Authored-By-Agent,valueonly)'` → `manager-develop`.
- Attribution: the verification batch below ran on the uncommitted tree at HEAD `2232e1a5f`.
  `git diff --stat 6f098a45e -- <both hook copies>` printed nothing, so the measured hook bytes equal
  the committed ones.

New header block (template hook, lines 28-55, verbatim):

```
# Outcome record: for the HEAD it gates, the hook keeps one line
# "<head-sha> <outcome>" in .moai/state/sync-quality-gate.last. <outcome> is
# exactly one of running, pass, fail. The record reads "running" before any check
# starts, then "fail" if a check failed (whether the mode blocked or only advised)
# or "pass" otherwise. On a later turn with the same HEAD:
#   - pass: no checks, empty stdout.
#   - fail: no checks. The failing run's exact stdout, its kind (block or
#     advisory), and the failed-check exit codes are kept in
#     .moai/state/sync-quality-gate.payload. A stored block is re-delivered
#     byte-identical while the mode resolved on that turn is blocking; a stored
#     block under an advisory resolution, or a stored advisory message, stays
#     silent (the advisory warning is written once, by the run that checked).
#   - stop_hook_active: when stdin carries "stop_hook_active": true, a stored
#     block is not re-delivered on that turn and no state changes, so the next
#     turn without the flag re-delivers it. The flag never suppresses the output
#     of a run that executes the checks.
#   - running: a run did not finish. While the record is at most
#     SYNC_GATE_STALE_WINDOW seconds old, no checks run and a non-blocking notice
#     is emitted. An older record gets ONE re-run for that HEAD, recorded in
#     .moai/state/sync-quality-gate.retry; once that re-run is used, later turns
#     emit a non-blocking notice instead of re-running.
#   - no record, a record for another HEAD, or an empty, unreadable, legacy
#     (bare SHA), or malformed record: the checks run.
# Every state write goes through a temporary file renamed into place, and a
# failing run writes its payload before its "fail" record.
#
# Forcing a re-gate: delete .moai/state/sync-quality-gate.last (or .moai/state as
# a whole). There is no flag or environment variable for retrying.
```

#### Verification (resumed run)

| Check | Command | Exit | Verbatim tail | Full output |
|---|---|---|---|---|
| M1 selector | `unset MOAI_SYNC_GATE_BLOCKING MOAI_AUTONOMY_TIER && go test ./internal/hook/ -run '^TestSyncGateFailState' -count=1 -v` | 0 | `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/hook	148.147s` | `m2-green-run.txt` |
| Six existing guards | `unset MOAI_SYNC_GATE_BLOCKING MOAI_AUTONOMY_TIER && go test ./internal/hook/ ./internal/template/ -run '^(TestHookWrapperCopiesStayIdentical\|TestAC004_SyncGateAdvisoryAtFullyAutonomous\|TestAC002_NonSyncHeadSkipsVetBuild\|TestHookOfficialCompliance_AC002_SyncGateStopHookSpecificOutput\|TestTemplateNoInternalContentLeak\|TestTemplateNeutralityAudit)$' -count=1 -v` | 0 | `ok  	github.com/modu-ai/moai-adk/internal/hook	11.201s` / `ok  	github.com/modu-ai/moai-adk/internal/template	2.724s` | `m2-guards-run.txt` |
| Both packages | `unset MOAI_SYNC_GATE_BLOCKING MOAI_AUTONOMY_TIER && go test ./internal/hook/ ./internal/template/ -count=1` | 0 | `ok  	github.com/modu-ai/moai-adk/internal/hook	165.423s` / `ok  	github.com/modu-ai/moai-adk/internal/template	42.197s` | `m2-packages-run.txt` |
| Lint | `golangci-lint run ./internal/hook/...` | 0 | `0 issues.` | `m2-lint.txt` |
| Windows cross-vet | `GOOS=windows GOARCH=amd64 go vet ./internal/hook/` | 0 | *(empty; file size 0 bytes)* | `m2-vet-windows.txt` |
| Format | `gofmt -l internal/hook/` | 0 | *(empty)* | — |
| Shell syntax, local copy | `bash -n .claude/hooks/moai/sync-phase-quality-gate.sh` | 0 | *(empty)* | — |
| Shell syntax, template copy | `bash -n internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` | 0 | *(empty)* | — |

(`\|` in the guard selector is table escaping; the executed selector carries a plain `|`.)

M1 selector swept set: `grep -c '^=== RUN' m2-green-run.txt` → `69`; `grep -c 'no tests to run'` → `0`;
subtests `--- PASS` → `56`, `--- FAIL|SKIP` → `0`. All 13 top-level tests PASS:

1. `TestSyncGateFailState_AC001_FailureRedeliveredOnSameHead`
2. `TestSyncGateFailState_AC002_NewFailingHeadBlocksWithNewResult`
3. `TestSyncGateFailState_AC003_PassThenSameHeadStaysSilent`
4. `TestSyncGateFailState_AC004_StopHookActiveDefersRedelivery`
5. `TestSyncGateFailState_AC005_UnknownAndLegacyRecordsRegate`
6. `TestSyncGateFailState_AC005_TornWriteNeverSilentPass`
7. `TestSyncGateFailState_AC006_RunningRecordStaleWindow`
8. `TestSyncGateFailState_AC007_NoPathLooserThanToday`
9. `TestSyncGateFailState_AC008_RedeliveryFollowsModeResolution`
10. `TestSyncGateFailState_AC013_RetryByDeletionNoStaleAuxState`
11. `TestSyncGateFailState_AC014_StaleWindowEqualsRegisteredTimeout`
12. `TestSyncGateFailState_AC006c_RetryBoundAndNotice`
13. `TestSyncGateFailState_AC015_NoticesNeverBlockOrConsumeCap`

Guards, each reported `--- PASS` by name in `m2-guards-run.txt`: `TestAC004_SyncGateAdvisoryAtFullyAutonomous`,
`TestAC002_NonSyncHeadSkipsVetBuild`, `TestHookWrapperCopiesStayIdentical`,
`TestTemplateNoInternalContentLeak`, `TestHookOfficialCompliance_AC002_SyncGateStopHookSpecificOutput`,
`TestTemplateNeutralityAudit`; `grep -c 'no tests to run'` → `0`.

Neutrality positive control: `grep -c -- '--- PASS: TestTemplateNeutralityAudit ' m2-guards-run.txt` →
`1`. The neutrality test this evidence relies on ran by name and passed. It was not selected away.

#### Test edits

None. The M1 test files are byte-unchanged: they appear in neither P4 nor P5. The only fix needed
during M2 was in the hook (the `STALE_RERUN` rename above).

#### Gaps (not observed in M2)

- AC-009's L-18 jq regex, the L-19 card-id/date regex, and the SHA-token scan were run on the
  template hook in the interrupted session, with 0 hits each, and were not re-run in the resumed
  run. The template-leak and neutrality guards above did run by name and passed; the standalone greps
  belong to AC-009, which is checked after M3.
- AC-010, AC-011, and AC-012 are M3. AC-013's reviewer read of the new header is not recorded as a
  verdict here. Mutant probes are M4.
- The mtime-unreadable fallback, nested `stop_hook_active` keys, and the exact 60 s point remain the
  explicit unverified gaps of acceptance.md §D.0.
- No Windows execution of the hook. No live Claude Code Stop event: the real stdin shape and the
  runtime block cap were not exercised.
- `internal/cli` tests were not run.

#### Residual risk

- Age is read with `stat -c %Y`, falling back to `stat -f %m`. A platform where both fail treats the
  record as stale and re-runs; this path is unverified.
- `stop_hook_active` detection is line-based. A payload that splits the key and the value across
  lines would not be detected, and the re-delivery would then happen on that turn.
- State writes swallow failures to keep exit 0 on every path. An unwritable `.moai/state` therefore
  degrades to re-gating every turn, never to a silent pass.
- The payload file carries the full stdout of the failing run. That is the bytes already written to
  the log channel, so no new data leaves the project.

### M3 — document and hook-comment wording (AC-010 / AC-011 / AC-012), recorded 2026-09-11 by manager-develop

#### Pre-flight (before any M3 edit)

| # | Command | Verbatim stdout | Exit |
|---|---|---|---|
| P1 | `git rev-parse --show-toplevel` | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t624` | 0 |
| P2 | `git branch --show-current` | `WT-sync-gate-failstate` | 0 |
| P3 | `git rev-parse --short HEAD` | `edecbff54` | 0 |
| P4 | `git status --short` | *(empty)* | 0 |
| P5 | `git diff --stat fa96fe644 edecbff54 -- <both doc copies>` | *(empty)*: the documents were unchanged between the plan baseline and M2 | 0 |

`manifest audit` sat at lines 3 and 480 of both hook copies. The Phase 9 anchor sat at line 160 of both document copies.

#### What changed

- **Documents** (both copies, identical edits, all above the Phase 9 anchor):
  - The Step 0.55.0 scope sentence now describes the hook-side manifest-change observation. The "only check for that drift" claim is gone.
  - The dependency block is now headed "Dependency manifest-change observation (hook-side, informational)". It says the Stop hook runs `git diff` over the detected language's manifests for the HEAD commit and records `deps_modified=1` when that diff is non-empty. It also says the value appears only in the hook output and in `.moai/logs/sync-quality-gate.log`, never drives the block decision, and is not a vulnerability scan.
  - The every-manifest-at-root list and the scan claim are removed. The `moai-ref-supply-chain` injection is kept, reworded as a separate agent-invoked review.
  - Step 0.55.2 opens with a new paragraph, "Relationship to the sync-auditor Security rule". It names Step 0.5.4's rubric (Critical/High → FAIL) as canonical and Phase 8 as an additional lens whose CRITICAL-only gate never clears an earlier sync-auditor FAIL.
  - Line 71 (HARD THRESHOLD) and lines 136-137 / 150 / 156 (CRITICAL blocks, HIGH warns) are untouched.
- **Hooks** (template first, then `cp` to local): the header `# Purpose:` line and the comment above the manifest step. Comment lines only.
- Full diffs: `.moai/reports/t624/m3-docs-diff.txt` (two hunks per copy) and `.moai/reports/t624/m3-hooks-diff.txt`.

#### Wording commit

- Commit `c0e56ab09`:
  `docs(t624): M3 correct manifest-observation and severity wording in sync quality gate doc and hook comments`.
  `git show --stat` lists exactly the four files, `4 files changed, 22 insertions(+), 20 deletions(-)`.
- Trailers: `git log -1 --format='%(trailers)'` → `Authored-By-Agent: manager-develop`.
- Attribution: every check below ran on the uncommitted tree at HEAD `edecbff54` after all four edits and the hook `cp`. The staged set at commit time was those same four files, so the measured bytes are the committed ones.

#### AC rows

Each row was a separate plain invocation with `/usr/bin/grep` and literal paths. The full transcript is in `.moai/reports/t624/m3-ac-rows.txt`.

| AC | Check | Stdout | Exit | Expected |
|---|---|---|---|---|
| 010 | `cmp` template hook vs local hook | *(empty)* | 0 | exit 0 |
| 010 | anchor `grep -c`, template doc / local doc | `1` / `1` | 0 / 0 | 1 each |
| 010 | `cmp` of the through-anchor extracts (`sed '/<anchor>/q'`) between copies | *(empty)* | 0 | identical |
| 010 | `diff` after-anchor extract (`sed '1,/<anchor>/d'`), fa96fe644 vs current, template | *(empty)* | 0 | unchanged |
| 010 | same, local | *(empty)* | 0 | unchanged |
| 011 | scan claim, template / local | `0` / `0` | 1 / 1 | 0, exit 1 |
| 011 | `deps_modified`, template / local | `1` / `1` | 0 / 0 | ≥1, exit 0 |
| 011 | "only check for that drift", template / local | `0` / `0` | 1 / 1 | 0, exit 1 |
| 011 | "Audit ALL of the following manifest files present at project root", template / local | `0` / `0` | 1 / 1 | 0, exit 1 |
| 011 | `manifest audit`, template hook / local hook | `0` / `0` | 1 / 1 | 0, exit 1 |
| 011 | `head -5` template hook \| `grep -E 'manifest[ -]change'` | `# Purpose: Fast sync-phase quality gate (compile/vet checks + dependency manifest-change observation)` | 0 | match |
| 012a | `sync-auditor FAIL`, template / local | `1` / `1` | 0 / 0 | ≥1 |
| 012b | HARD THRESHOLD `grep -n`, template / local | `71:- Security (25%): … HARD THRESHOLD: any Critical/High finding causes overall FAIL regardless of other scores` (both) | 0 / 0 | present |
| 012b | `grep -nE 'sync-auditor FAIL\|CRITICAL\|HIGH'`, both copies (identical) | lines 136 `Only CRITICAL findings block`, 137 `HIGH findings are reported as warnings`, 148 new paragraph, 150 `If CRITICAL findings exist:`, 156 `If no CRITICAL findings` | 0 / 0 | CRITICAL still the only blocking severity |
| 009 | `grep -c "SPEC-"`, template doc / template hook | `0` / `0` | 1 / 1 | 0 |
| 009 | L-19 card-id/date regex, template doc / template hook | `0` / `0` | 1 / 1 | 0 |
| 009 | SHA-token scan, template doc / template hook | *(empty)* / *(empty)* | 1 / 1 | no hits |
| 009 | L-18 jq invocation regex, template hook | `0` | 1 | 0 |

(`\|` inside the table is table escaping; the executed pattern carries a plain `|`.)

AC-012 reviewer read, as author: the 012a token sits in the sentence "Phase 8 … never clears an earlier sync-auditor FAIL". That sentence also states the rubric is canonical, so the token is not in a contrary sentence. The "Step 0.5.4" target exists: `grep -nE '^#+ .*(Step 0\.5|0\.5\.4|Phase 8)'` shows `60:#### Step 0.5.4: Deep Code Review with Auto-Fix`, and line 71 sits under that heading. This is the author's read, not an independent verdict.

AC-011 reviewer read (mutant M9), as author: the removed claims are replaced by a description of `deps_modified` for manifests changed in the HEAD commit, marked informational. The description was checked against the hook: `DEPS_MODIFIED` is set from `git diff "$DIFF_RANGE" -- $DEPS_MANIFESTS` (hook line 504). It is echoed only in the two stdout `printf` messages and the log line (lines 551-556, 575).

#### Hooks

| Check | Command | Stdout | Exit |
|---|---|---|---|
| Syntax, template | `bash -n internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` | *(empty)* | 0 |
| Syntax, local | `bash -n .claude/hooks/moai/sync-phase-quality-gate.sh` | *(empty)* | 0 |
| Parity | `cmp <template hook> <local hook>` | *(empty)* | 0 |
| Comment-only | `grep -cE '^[-+][^-+#]' m3-hooks-diff.txt` (changed lines not starting with `#`) | `0` | 1 |

The diff against `edecbff54` has two hunks per copy. Hunk 1 changes line 3 (`# Purpose:`). Hunk 2 replaces the three-line manifest comment with a four-line one. That is 5 insertions and 4 deletions per copy.

#### Tests (the hook bytes changed, so they were re-run after all edits and the `cp`)

| Check | Command | Exit | Verbatim tail | Full output |
|---|---|---|---|---|
| M1 selector | `unset MOAI_SYNC_GATE_BLOCKING MOAI_AUTONOMY_TIER && go test ./internal/hook/ -run '^TestSyncGateFailState' -count=1 -v` | 0 | `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/hook	119.508s` | `m3-m1-selector.txt` |
| Six guards | `unset MOAI_SYNC_GATE_BLOCKING MOAI_AUTONOMY_TIER && go test ./internal/hook/ ./internal/template/ -run '^(TestHookWrapperCopiesStayIdentical\|TestAC004_SyncGateAdvisoryAtFullyAutonomous\|TestAC002_NonSyncHeadSkipsVetBuild\|TestHookOfficialCompliance_AC002_SyncGateStopHookSpecificOutput\|TestTemplateNoInternalContentLeak\|TestTemplateNeutralityAudit)$' -count=1 -v` | 0 | `ok  	github.com/modu-ai/moai-adk/internal/hook	10.572s` / `ok  	github.com/modu-ai/moai-adk/internal/template	1.626s` | `m3-guards.txt` |

- **M1 selector swept set** (from `m3-m1-selector.txt`): 13 lines match `^--- PASS: TestSyncGateFailState`, the same 13 top-level tests listed under M2. 0 lines match `^--- (FAIL|SKIP)` and 0 lines contain `no tests to run`.
- **Guards**: each of the six printed a `--- PASS:` line by name. The neutrality positive control `grep -c -- '--- PASS: TestTemplateNeutralityAudit '` → `1`.
- **Package runs**: the two full package runs were not repeated, because no guard failed.

#### Gaps (not observed in M3)

- The AC-012 and AC-011 (M9) reviewer reads above are the author's own, not an independent reviewer's verdict.
- The full `go test ./internal/hook/ ./internal/template/` package runs, lint, Windows vet, and `internal/cli` tests were not run in M3. The change is comment and markdown only, and the coordinator's instruction limited the re-run to the selector and the six guards.
- The document was not rendered. Only raw bytes were checked.
- No live Stop event exercised the hook.
- Mutant probes are M4.

#### Residual risk

- "restricted to the HEAD commit" is accurate for the normal `HEAD~1..HEAD` range. On an initial commit the hook's `DIFF_RANGE` is the empty tree, and the diff then compares the empty tree with the working tree.
- The hook captures `git diff` stderr into the same file (`2>&1`). A `git diff` error would therefore also set `deps_modified=1`. "When that diff is non-empty" omits that edge.
- The after-anchor comparison uses fa96fe644. The pre-flight showed the documents unchanged between fa96fe644 and edecbff54, so the result also holds against edecbff54.

### M4 — mutant probes and run-phase AC matrix, recorded 2026-09-11 by manager-develop

#### Tree check and restore baseline (before the first mutant)

| # | Command | Verbatim stdout | Exit |
|---|---|---|---|
| T1 | `git rev-parse --show-toplevel` | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t624` | 0 |
| T2 | `git branch --show-current` | `WT-sync-gate-failstate` | 0 |
| T3 | `git rev-parse --short HEAD` | `989ef144b` | 0 |
| T4 | `git --no-optional-locks status --short` | *(empty)* | 0 |

| File | sha256 (restore baseline) |
|---|---|
| `.claude/hooks/moai/sync-phase-quality-gate.sh` | `3fdce8e697b7afc91a7005d6c371b91759287d08db3e8fefc35988a138fbe866` |
| `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` | same |
| `.claude/skills/moai/workflows/sync/quality-gates-quality.md` | `72d7888b8a5d7809cb0b619d5350fad2543a958e3bd45e7cb208273d75ecbd09` |
| `internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md` | `c0c3310a3d1a4d4b32fe71d2f937ec8669abf0977797e20f5eaa97d609460df3` |

#### Method

- One mutant at a time. Hook mutants edit the LOCAL hook (the copy every `sgfFixture` drives); M9 edits both document copies (the copies the AC-011 grep rows read).
- Each mutant runs only the tests carrying its named §D.16 rows: `unset MOAI_SYNC_GATE_BLOCKING MOAI_AUTONOMY_TIER && go test ./internal/hook/ -run '<named test>' -count=1 -v`, output to `.moai/reports/t624/m4-mutant-MNN.txt` (M9: the AC-011 grep rows).
- Revert by copying a pristine copy back (no `git stash` / `checkout` / `restore`), then prove the restore: sha256 equals the baseline above and `git --no-optional-locks diff --stat` is empty.
- M1-M27 in the first pass were measured on HEAD `989ef144b`. The M18 rerun was measured on HEAD `21216ae6f` plus the S2/S3 test edit, which is the tree of test commit `7791c0e7d`.

#### Mutant table (§D.16; M22 was probed as two mutants, one per clause)

Restore proof for every row: hook sha256 `3fdce8e6…fbe866` (M9: both document hashes above) and an empty `git diff --stat`.

| Mutant | Mutation (one line) | Named rows | Exit | Red assertion (quoted) | Evidence |
|---|---|---|---|---|---|
| M1 | `if stop_hook_active_set; then` → `if false; then` | AC-004 (a) | 1 | F1-F4 (a) `flagged call carries "decision"` | `m4-mutant-M01.txt` |
| M2 | re-emit `tail -n +2` → no-op | AC-001, AC-004 (b), AC-008 A4/A9 | 1 | AC-001 call 2 0 bytes; AC-004 F1-F4 (b)(c); A4, A9 `call 2 stdout is not byte-identical` | `m4-mutant-M02.txt` |
| M3 | `if [ "$PAYLOAD_VALID" = "1" ]` → `if false` | AC-001 (count) | 1 | `stub count after call 2 = 4, after call 1 = 2; want equal` | `m4-mutant-M03.txt` |
| M4 | `if [ "$MODE" != "blocking" ]` → `if false` | AC-008 A5 | 1 | A5 (also A7, A8) `call 2 stdout = … want empty` | `m4-mutant-M04.txt` |
| M5 | fresh-running condition → `if false` | AC-006 a0, AC-015 N2a/N2b | 1 | a0, a50 `stub invoked 2 time(s)…want 0`; N2a `5 of 9 stdouts carry decision`; N2b `1 of 9 … decision` | `m4-mutant-M05.txt` |
| M6 | retry-marker condition → `if false` | AC-006c | 1 | assertions 3, 4, 5 | `m4-mutant-M06.txt` |
| M7 | flag grep → `grep -q 'stop_hook_active'` | AC-004 (c) | 1 | (c) `call1 237 bytes, now 0 bytes` | `m4-mutant-M07.txt` |
| M8 | `if stop_hook_active_set; then exit 0; fi` after `RERUN_OF_RUNNING=0` | AC-007 R10a/R10b | 1 | R10a, R10b `final invocation is not a block; stdout=""` | `m4-mutant-M08.txt` |
| M9 | deps_modified paragraph deleted in both documents | AC-011 `deps_modified` rows | 1 | `deps_modified` template `0` exit 1, local `0` exit 1 | `m4-mutant-M09.txt` |
| M10 | `"$HEAD_SHA fail")` → `*" fail")`, SHA check dropped | AC-007 R8 | 1 | R8 `stub count did not increase on the final invocation (delta 0)` | `m4-mutant-M10.txt` |
| M11 | exhausted-retry notice → no-op | AC-006c, AC-015 N1 | 1 | AC-006c assertion 5 `stdout=""`; N1 `9 of 9 stdouts lack systemMessage` | `m4-mutant-M11.txt` |
| M12 | exhausted notice printed as a `decision` block | AC-015 N1 | 1 | N1 `9 of 9 stdouts carry "decision"; want exactly 0` | `m4-mutant-M12.txt` |
| M13 | `tail -n +2 "$PAYLOAD_FILE"` before the advisory exit | AC-008 A1-A3 | 1 | A1, A2, A3 (A6) `call 2 stdout = {systemMessage WARNING…}; want empty` | `m4-mutant-M13.txt` |
| M14 | flag grep → exactly one space after the colon | AC-004 F2/F3 | 1 | F2, F3 (F4) (a) `flagged call carries decision` | `m4-mutant-M14.txt` |
| M15 | `unset MOAI_SYNC_GATE_BLOCKING` before re-delivery `resolve_gate_mode` | AC-008 A7 | 1 | A7 `call 2 stdout = {block…}; want empty` | `m4-mutant-M15.txt` |
| M16 | `C1_EXIT="$P_C1"` → `C1_EXIT=0` | AC-008 A8 | 1 | A8 `call 2 stdout = {block…}; want empty` | `m4-mutant-M16.txt` |
| M17 | `&& [ "$C2_EXIT" -eq 0 ]` dropped in the automatic branch | AC-008 A9 | 1 | A9 `call 1 is not a block` / `call 2 stdout is not byte-identical` | `m4-mutant-M17.txt` |
| M18 (first pass) | `rm -f "$PAYLOAD_FILE" 2>/dev/null \|\| true` → no-op | AC-013 S1 (as named at `989ef144b`) | **0** | **survived**: `--- PASS` for D1 and S1 | `m4-mutant-M18.txt` |
| M18 (rerun, after amendment) | same mutation | AC-013 S2, S3 (as named at `21216ae6f`) | 1 | S2 `2 of 2 stub invocation(s) during call 2 saw the payload file present; want every observation absent`; S3 `a payload file exists after call 2; want none` and `call 3 stub count unchanged (delta 0); want increased — the checks must re-run` | `m4-mutant-M18-rerun.txt` |
| M19 | `"$HEAD_SHA "*) exit 0 ;;` before `esac` | AC-005 U1 | 1 | U1 (also U2) `stub invoked 0 time(s)`, `stdout is not a block` | `m4-mutant-M19.txt` |
| M20 | fresh comparison `-le 100` | AC-006 b70 | 1 | b70 (also b61) `stub invoked 0 time(s)…want >= 1` | `m4-mutant-M20.txt` |
| M21 | fresh comparison `-le 40` | AC-006 a50 | 1 | a50 `stub invoked 2 time(s)…want 0` | `m4-mutant-M21.txt` |
| M22a | `if [ "$P_KIND" = "advisory" ]` → `if false` | AC-008 A6 | 1 | A6 `call 2 stdout = {systemMessage WARNING…}; want empty` | `m4-mutant-M22a.txt` |
| M22b | `exit 0` before the "fail record without a usable payload" re-gate | AC-005 U5 | 1 | U5 `stub invoked 0 time(s) on this call; want >= 1`, `stdout is not a block; stdout=""` | `m4-mutant-M22b.txt` |
| M23 | advisory payload for HEAD → `exit 0` before the record `case` | AC-005 TA1-TA5 | 1 | TA2, TA3, TA4, TA5-bare, TA5-pass `stub invoked 0 time(s) on this call; want >= 1` (TA1 also FAIL) | `m4-mutant-M23.txt` |
| M24 | `if [ ! -f "$PAYLOAD_FILE" ]; then exit 0; fi` in the running arm | AC-005 TB2 | 1 | TB2 `stub invoked 0 time(s) on this call; want >= 1`, `stdout is not a block; stdout=""` | `m4-mutant-M24.txt` |
| M25 | flag grep `[ ${shas_tab}]*` → `[ ]*` | AC-004 F4 | 1 | F4 (a) `flagged call carries "decision"` | `m4-mutant-M25.txt` |
| M26 | fresh-notice branch rewrites the record to the bare SHA | AC-015 N2a | 1 | N2a `5 of 9 stdouts carry "decision"; want exactly 0`; `record after run 9 = "<HEAD> fail"; want "<HEAD> running"` | `m4-mutant-M26.txt` |
| M27 | fresh comparison `-le 65` | AC-006 b61 | 1 | b61 `stub invoked 0 time(s) on this call; want >= 1` | `m4-mutant-M27.txt` |

(`\|` inside the table is table escaping.)

#### M18: survival, amendment, rerun

- **First pass (HEAD `989ef144b`).** M18 left AC-013 green (`test_exit=0`, D1 and S1 PASS). S1's call 2 is a failing advisory run, and hook lines 564-566 rewrite the payload before the `fail` record, so the stale call-1 payload was overwritten before call 3 read it. manager-develop stopped before the evidence commit and returned a blocker report.
- **Amendment.** The lead chose option A. manager-spec amended AC-013 in `21216ae6f` (rows S2 and S3; the §D.16 M18 row now names S2 and S3, and S1 is explicitly not an M18 target). SPEC status was unchanged and there was no re-audit (lead-approved under the kickoff debt rule).
- **Test commit `7791c0e7d`** (`internal/hook/sync_gate_failstate_test.go` only, 124 insertions, 0 deletions):
  - `sgfStubSpec.probePayload` (line 69) and the guarded probe line in `setStub` (line 168). The probe line is emitted only when the field is set, so every other row's stub script is byte-identical.
  - `sgfPayloadName` (line 41) and `payloadPath()` (line 217).
  - S2 (line 760): the probe records `present`/`absent` per stub invocation. There is a setup assertion that the payload exists before call 2; the Then requires at least one observation and every observation `absent`.
  - S3 (line 805): an `mv` shim in `f.stubBin` refuses only when its last argument is the payload path, otherwise it execs the system `mv` resolved by `exec.LookPath` before the shim is written. Reachability (refused move logged, call-2 stub delta ≥ 1, record `<HEAD> fail`) is asserted separately and stops the row before the Then when it fails. S3 skips on Windows, following the file's U4 convention.
- **Green on the unmutated tree** (`m4-ac013-green.txt`, exit 0):
  - S2 `observations=[absent absent]`, S3 `refused-moves=1`, call 3 `stub delta=2`.
  - `--- PASS` for D1, S1, S2, S3.
- **Rerun** (`m4-mutant-M18-rerun.txt`, exit 1): S2 and S3 FAIL with the lines quoted in the table, D1 and S1 PASS.
  - Restore: both hooks `3fdce8e697b7afc91a7005d6c371b91759287d08db3e8fefc35988a138fbe866`; `git diff --stat -- .claude internal/template` empty.
- **Other mutants re-checked:** none. The fixture change is opt-in, and no other §D.16 row names AC-013.

#### Part 2 checks (tree of `7791c0e7d`; measured on the same bytes just before that commit, docs and hooks byte-identical to `c0e56ab09` per `git diff --stat c0e56ab09 HEAD -- <4 files>` → empty)

| Check | Command | Exit | Verbatim tail / result | Evidence |
|---|---|---|---|---|
| M1 selector | `unset MOAI_SYNC_GATE_BLOCKING MOAI_AUTONOMY_TIER && go test ./internal/hook/ -run '^TestSyncGateFailState' -count=1 -v` | 0 | `ok  	github.com/modu-ai/moai-adk/internal/hook	98.081s`; 13 lines `^--- PASS: TestSyncGateFailState`, 0 `^--- FAIL`, 0 nested `--- FAIL`, 0 `no tests to run`; AC013 lists D1, S1, S2, S3 | `m4-selector-full.txt` |
| Six guards | `unset MOAI_SYNC_GATE_BLOCKING MOAI_AUTONOMY_TIER && go test ./internal/hook/ ./internal/template/ -run '^(TestHookWrapperCopiesStayIdentical\|TestAC004_SyncGateAdvisoryAtFullyAutonomous\|TestAC002_NonSyncHeadSkipsVetBuild\|TestHookOfficialCompliance_AC002_SyncGateStopHookSpecificOutput\|TestTemplateNoInternalContentLeak\|TestTemplateNeutralityAudit)$' -count=1 -v` | 0 | six `--- PASS:` lines by name; `ok  	…/internal/hook	8.127s` / `ok  	…/internal/template	3.005s` | `m4-guards.txt` |
| Both packages (AC-009 When), first run on `7791c0e7d` — superseded, kept as history | `unset MOAI_SYNC_GATE_BLOCKING MOAI_AUTONOMY_TIER && go test ./internal/hook/ ./internal/template/ -count=1` | 1 | `ok  	…/internal/hook	125.585s`; `--- FAIL: TestManifestHashFormat` (`CATALOG_HASH_UNSTABLE: moai stored hash=1d23838d…, computed hash=773956f6… (source=.claude/skills/moai/ (whole tree))`), `--- FAIL: TestCatalogHashCoversSkillSubfiles` (`CATALOG_HASH_SKINNY … run gen-catalog-hashes.go --all`); `FAIL	…/internal/template	27.860s`. Attribution: M2's run at `edecbff54` passed `internal/template` (`m2-packages-run.txt`); `git diff --stat edecbff54 HEAD -- internal/template/templates/.claude/skills/moai/ internal/template/catalog.yaml` shows only `quality-gates-quality.md` (M3 `c0e56ab09`, 6+/6−), so M3's document edit left the `moai` catalog hash stale. M3 did not run the package tests | `m4-ac009-packages-before-regen.txt` |
| Both packages (AC-009 When), re-measured on HEAD `dc9feafb2` after the catalog regen, run alone | same command | 0 | `ok` for `internal/hook` (159.262s) and `ok` for `internal/template` (51.493s); `pkg_exit=0` (verbatim lines in the file) | `m4-ac009-packages.txt` |
| Catalog tests by name, HEAD `dc9feafb2` | `go test ./internal/template/ -run '^(TestManifestHashFormat\|TestCatalogHashCoversSkillSubfiles)$' -count=1 -v` | 0 | `--- PASS: TestCatalogHashCoversSkillSubfiles`, `--- PASS: TestManifestHashFormat`; `audited 45 catalog entries`, `audited 34 directory entries` | `m4-catalog-tests.txt` |

The catalog regen commit `dc9feafb2` changes only `internal/template/catalog.yaml`. The test file and both hooks are unchanged by it, so every other row in this table stays attributable to tree `7791c0e7d`.
| go vet | `go vet ./internal/hook/...` | 0 | *(empty)* | `m4-vet.txt` |
| golangci-lint | `golangci-lint run ./internal/hook/...` | 0 | `0 issues.` | `m4-lint.txt` |
| Windows vet | `GOOS=windows GOARCH=amd64 go vet ./internal/hook/` | 0 | *(empty)* | `m4-vet-windows.txt` |
| Hook parity | `cmp <local hook> <template hook>` | 0 | *(empty)* | `m4-ac010.txt` |
| AC-010 anchors | anchor `grep -c` both docs; through-anchor `diff`; after-anchor `diff` against `fa96fe644`, both copies | 0 | `1` / `1`; `head_diff_exit=0`; `tmpl_tail_diff_exit=0`; `local_tail_diff_exit=0` | `m4-ac010.txt` |
| AC-009/011/012 document rows | the M3 row commands, re-run on this tree with `/usr/bin/grep` | see file | all rows at their green values (below) | `m4-ac-docrows.txt` |

#### AC-009: first failure, cause, catalog regen (history kept)

- **First failure.** On `7791c0e7d` the AC-009 package run exited 1. `internal/template` failed `TestManifestHashFormat` (`CATALOG_HASH_UNSTABLE`) and `TestCatalogHashCoversSkillSubfiles` (`CATALOG_HASH_SKINNY`) (`m4-ac009-packages-before-regen.txt`).
- **Cause.** M3 `c0e56ab09` edited the template copy of `quality-gates-quality.md`, which changed the `moai` skill's whole-tree hash; `catalog.yaml` was not regenerated, and M3's re-run used named selectors only. manager-develop stopped before this evidence commit with a blocker; the coordinator reproduced it and chose regeneration.
- **Regen.** `go run internal/template/scripts/gen-catalog-hashes.go --all` exited 0. `git diff` changed one line, the `moai` hash `1d23838d…` → `773956f641a95d67bb5d49ffb5f77c4736425861b3ce16bfa5e9badf3daf92b7`, and no other file. It was committed alone as `dc9feafb2` (`m4-catalog-regen.txt`).
- **Single-entry equivalence.** The lead's correction (use `--entry moai`, not `--all`) arrived after `dc9feafb2` landed, so the equivalence was checked on a scratch copy instead of rewriting history:
  - the pre-regen catalog has 0 comment lines;
  - `--entry moai --dry-run` computes `773956f641a95d67bb5d49ffb5f77c4736425861b3ce16bfa5e9badf3daf92b7` (64 characters, equal to the test's required value);
  - `--entry moai` changes exactly one line (`diff -U0`: one `-`, one `+`);
  - `cmp` of that result against the `dc9feafb2` bytes exits 0.

  The script's write path (lines 257-269) re-marshals the whole file for `--entry` and `--all` alike; the committed bytes equal the single-entry result.
- **Contended re-run.** The first re-run on `dc9feafb2`, issued in the same turn as other `go run` / `go test` jobs, printed `ok` for `internal/template`. `internal/hook` failed two timing tests: `TestSessionStart_DeferredScanDoesNotBlockReturn` (`blocked 629.1065ms`) and `TestSessionStart_DeferredScanJoinsWithinBound/slow_scan_drops_advisory` (`blocked 767.00875ms; expected return near the 250ms bound`). Evidence: `m4-ac009-packages-after-regen-contended.txt`.
- **Solo re-run.** The same command re-run alone printed `ok` for both packages (`m4-ac009-packages.txt`). The 1-minute load average was 19.73 at its start. Contention is the likely cause of the two timing failures; it is not proven.

#### AC matrix (15 criteria)

| AC | Status | Command | Actual output | Evidence |
|---|---|---|---|---|
| AC-001 | PASS | M1 selector | `--- PASS: TestSyncGateFailState_AC001_FailureRedeliveredOnSameHead (1.84s)` | `m4-selector-full.txt` |
| AC-002 | PASS | M1 selector | `--- PASS: TestSyncGateFailState_AC002_NewFailingHeadBlocksWithNewResult (2.66s)` | `m4-selector-full.txt` |
| AC-003 | PASS | M1 selector | `--- PASS: TestSyncGateFailState_AC003_PassThenSameHeadStaysSilent (1.78s)` | `m4-selector-full.txt` |
| AC-004 | PASS | M1 selector | `--- PASS: TestSyncGateFailState_AC004_StopHookActiveDefersRedelivery (6.87s)` | `m4-selector-full.txt` |
| AC-005 | PASS | M1 selector | `--- PASS: …AC005_UnknownAndLegacyRecordsRegate (9.18s)`, `--- PASS: …AC005_TornWriteNeverSilentPass (10.27s)` | `m4-selector-full.txt` |
| AC-006 | PASS | M1 selector | `--- PASS: …AC006_RunningRecordStaleWindow (4.77s)`, `--- PASS: …AC006c_RetryBoundAndNotice (1.20s)` | `m4-selector-full.txt` |
| AC-007 | PASS | M1 selector | `--- PASS: …AC007_NoPathLooserThanToday (30.32s)` | `m4-selector-full.txt` |
| AC-008 | PASS | M1 selector | `--- PASS: …AC008_RedeliveryFollowsModeResolution (12.61s)` | `m4-selector-full.txt` |
| AC-009 | PASS (packages re-measured on HEAD `dc9feafb2` after the catalog regen; the first run on `7791c0e7d` failed, see the AC-009 history above) | six guards; both packages; catalog tests; M1 selector; template rows | six guards PASS by name; packages `ok`/`ok` on `dc9feafb2`; catalog tests PASS; `SPEC-` `0`/`0`; L-19 `0`/`0`; SHA scan no hits; L-18 jq `0` | `m4-guards.txt`, `m4-ac009-packages.txt`, `m4-catalog-tests.txt`, `m4-ac-docrows.txt` |
| AC-010 | PASS | `cmp` + anchor checks | `cmp_exit=0`; `1`/`1`; three diffs exit 0 | `m4-ac010.txt` |
| AC-011 | PASS | document/hook `grep -c` rows | scan `0`/`0`; `deps_modified` `1`/`1`; "only check" `0`/`0`; "Audit ALL" `0`/`0`; `manifest audit` `0`/`0` | `m4-ac-docrows.txt`; `m3-ac-rows.txt` (header match, reviewer read) |
| AC-012 | PASS | `grep -c` rows | (a) `sync-auditor FAIL` `1`/`1`; (b) HARD THRESHOLD `1`/`1` | `m4-ac-docrows.txt`; `m3-ac-rows.txt` (CRITICAL-only lines) |
| AC-013 | PASS | M1 selector | `--- PASS: …AC013_RetryByDeletionNoStaleAuxState (6.34s)` with D1, S1, S2, S3 PASS | `m4-selector-full.txt`, `m4-ac013-green.txt` |
| AC-014 | PASS | M1 selector | `--- PASS: …AC014_StaleWindowEqualsRegisteredTimeout (0.00s)` | `m4-selector-full.txt` |
| AC-015 | PASS | M1 selector | `--- PASS: …AC015_NoticesNeverBlockOrConsumeCap (9.73s)` | `m4-selector-full.txt` |

#### §D.0 unverified gaps (not claimed)

These three behaviors stay explicit unverified gaps. Closing the card does not claim them:

- the mtime-unreadable fallback;
- nested `stop_hook_active` key discrimination;
- the exact-60 s boundary point.

#### Gaps (not observed in M4)

- M1-M17 and M19-M22b were measured before the S2/S3 test commit. The later test diff is additive (124 insertions, 0 deletions), and the probe line is guarded. They were not re-run on `7791c0e7d`.
- S3 skips on Windows, so on Windows M18 is guarded by S2 alone. No Windows run was observed (Windows vet only).
- The AC-011 header row and the AC-012 reviewer reads are carried from M3 (`m3-ac-rows.txt`); the bytes are unchanged since `c0e56ab09`.
- No `go test ./...`, `internal/cli` tests, `make build`, or push. No live Stop event exercised the hook.

#### Residual risk

- S3 depends on the hook calling `mv` by bare name inside `write_state_file`. A later change to an absolute-path `mv` makes the S3 setup assertion fail. That is reported as a gap by design, never as a pass.
- S2 relies on the go stub being invoked only after the payload invalidation. A future `go` call placed before hook line 372 would read `present` on a correct hook, a false red, not a false green.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-11
run_status: audit-ready
run_commits:
  - 1a3b12ce5   # plan-audit debt paid (NEW-1), pre-M1
  - f611a7060   # M1 RED tests (test-only)
  - 2232e1a5f   # M1 evidence
  - 6f098a45e   # M2 hook outcome record and auxiliary state
  - edecbff54   # M2 evidence
  - c0e56ab09   # M3 wording (docs + hook comments)
  - 989ef144b   # M3 evidence
  - 21216ae6f   # AC-013 amendment (manager-spec)
  - 7791c0e7d   # AC-013 S2/S3 tests guarding mutant M18
  - dc9feafb2   # moai catalog hash regenerated after the M3 template doc edit
  - pending-backfill   # this M4 evidence commit
run_commit_sha: pending-backfill
ac_pass_count: 15
ac_fail_count: 0
mutants_red: 27/27   # M18 red only after the 21216ae6f amendment
d0_unverified_gaps_not_claimed: [mtime-unreadable fallback, nested stop_hook_active key, exact-60s boundary]
new_warnings_or_lints_introduced: 0   # golangci-lint "0 issues."
cross_platform_build:
  go_vet_darwin: pass
  go_vet_windows_amd64: pass
spec_status_changed: false
m1_to_mN_commit_strategy: test-first M1 commit, then per-milestone implementation and evidence commits
```

Definition of Done (§D.17):

| DoD item | State | Where |
|---|---|---|
| All 15 criteria pass, recorded with command, output, tree SHA | met (AC-009 after the `dc9feafb2` catalog regen) | AC matrix above, `m4-ac009-packages.txt`, `m4-catalog-tests.txt` |
| M1 test-only commit precedes every hook/document change, with the four RED elements | met | `f611a7060`; §E.2 M1 |
| All 27 §D.16 mutants turn their named rows red | met (M18 after the `21216ae6f` amendment) | mutant table above |
| Three §D.0 behaviors remain unverified gaps, not claimed | met | §D.0 list above |
| `cmp` of the hook copies exits 0; AC-010 anchor checks hold | met | `m4-ac010.txt` |
| `go vet` and `golangci-lint` clean on `internal/hook` | met | `m4-vet.txt`, `m4-lint.txt` |
| No `go test ./...`, no `make build`, no push; every commit names `t624` | met | commit subjects |

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
