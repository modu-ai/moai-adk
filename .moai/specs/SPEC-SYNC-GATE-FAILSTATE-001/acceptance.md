# Acceptance — SPEC-SYNC-GATE-FAILSTATE-001

> Harness: **standard**. Document-level pin: every observation below that carries no pin of
> its own was measured on tree `fa96fe644fcff8a15ac336833a4e816fc0a46fe3`.
> Verification scope: `go test ./internal/hook/ ./internal/template/`. No local full suite, no
> `make build`.

## Classification legend

- **release-blocking** — the criterion has an adopted RED-now cell (four elements, per
  `verification-completeness.md` §2.1) and a green path. Behavioral criteria become
  release-blocking at the M1 test-only commit, when their four elements (command, verbatim
  stdout, exit code, tree SHA) are recorded in `progress.md §E.2` (lead ruling B4, plan.md §B).
  The h01 reproduction stays supporting evidence. Until that record exists, a behavioral
  criterion is not release-blocking and cannot be recorded as a pass.
- **regression-guard** — the behavior already holds on `fa96fe644`, so no honest RED-now cell
  exists. The criterion guards against the fix breaking the behavior. Its baseline is observed
  as a PASS on the M1 commit, before the fix.

## §D AC Matrix

| AC | REQ | Class | RED / baseline source | Green path |
|---|---|---|---|---|
| AC-001 | 001·002·003 | release-blocking | h01 call2 0 bytes (supporting); M1 RED cell | M2 |
| AC-002 | 004 | regression-guard | h01 control A blocks | M2 keeps it |
| AC-003 | 004 | regression-guard | h01 controls B1/B2 silent | M2 keeps it |
| AC-004 | 005 | release-blocking (a, b); regression-guard (c) | M1 RED cell | M2 |
| AC-005 | 007 | release-blocking | M1 RED cell | M2 |
| AC-006 | 008·009 | release-blocking (a, c); regression-guard (b) | M1 RED cell | M2 |
| AC-007 | 010 | regression-guard | M1 baseline PASS (today blocks) | M2 keeps it |
| AC-008 | 006 | release-blocking (A4); regression-guard (A1-A3, A5, A6) | M1 RED cell / baseline | M2 |
| AC-009 | 012 | regression-guard | existing guards green on `fa96fe644` | M2, M3 keep them |
| AC-010 | 013 | regression-guard | L-13, L-15 | M2, M3 keep them |
| AC-011 | 014 | release-blocking | L-02..L-05, L-08..L-12 | M3 |
| AC-012 | 015 | release-blocking (a); regression-guard (b) | L-06, L-07, L-16 | M3 |
| AC-013 | 011 | regression-guard | deleting the sentinel re-runs today | M2 keeps it |
| AC-014 | 008 | release-blocking | M1 RED cell (no named window exists) | M2 |
| AC-015 | 008·009 | release-blocking | M1 RED cell (exhausted state is silent today) | M2 |

15 criteria (Tier M ceiling 16).

## §D.0 Evidence ledger (plan-phase observations)

Each command ran as a single invocation. Its exit code was captured by an `echo "exit=$?"`
appended in the same call; that echo is not part of the cited command or its stdout. Tree:
`fa96fe644fcff8a15ac336833a4e816fc0a46fe3` unless stated.

| Id | Command | Verbatim stdout | Exit |
|---|---|---|---|
| L-01 | `git diff --stat d5dc42959 fa96fe644 -- .claude/hooks/moai/sync-phase-quality-gate.sh internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md .claude/skills/moai/workflows/sync/quality-gates-quality.md internal/hook/ internal/template/hook_official_compliance_test.go` | *(empty)* | 0 |
| L-02 | `grep -c "dependency vulnerability scan runs automatically via the Stop hook" internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md` | `1` | 0 |
| L-03 | `grep -c "dependency vulnerability scan runs automatically via the Stop hook" .claude/skills/moai/workflows/sync/quality-gates-quality.md` | `1` | 0 |
| L-04 | `grep -c "deps_modified" internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md` | `0` | 1 |
| L-05 | `grep -c "deps_modified" .claude/skills/moai/workflows/sync/quality-gates-quality.md` | `0` | 1 |
| L-06 | `grep -c "sync-auditor FAIL" internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md` | `0` | 1 |
| L-07 | `grep -c "sync-auditor FAIL" .claude/skills/moai/workflows/sync/quality-gates-quality.md` | `0` | 1 |
| L-08 | `grep -c "Skipping the whole of Phase 8 would remove the only check for that drift" internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md` | `1` | 0 |
| L-09 | `grep -c "Skipping the whole of Phase 8 would remove the only check for that drift" .claude/skills/moai/workflows/sync/quality-gates-quality.md` | `1` | 0 |
| L-10 | `grep -c "Audit ALL of the following manifest files present at project root" internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md` | `1` | 0 |
| L-11 | `head -5 internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` | see block below | 0 |
| L-12 | `grep -c "manifest audit" internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` | `2` | 0 |
| L-13 | `cmp .claude/hooks/moai/sync-phase-quality-gate.sh internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` | *(empty)* | 0 |
| L-14 | `grep -c "SPEC-" internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` | `0` | 1 |

L-11 verbatim stdout:

```
#!/bin/bash
# Hook: sync-phase-quality-gate
# Purpose: Fast sync-phase quality gate (compile/vet + dependency manifest audit)
# Trigger: Stop event when the current session's HEAD is a sync-phase commit
#
```

Non-single-invocation baselines (regression-guard support only, not RED cells):

- **L-15** — a process-substitution `diff` of `sed -n 1,161p` over the two document copies
  printed nothing and reached the `&& echo DOC_1_161_IDENTICAL` branch. Both files measured
  (`wc -l`): local 326 lines, template 322 lines.
- **L-16** — a line-filtered `grep -n` sweep of the template document (lines 1-161) showed
  line 71 carrying `HARD THRESHOLD: any Critical/High finding causes overall FAIL regardless of
  other scores`, and lines 136-137 and 156 carrying the CRITICAL-only / HIGH-warning wording.

**Gaps in this ledger:** L-10's phrase was not measured on the local document copy. L-15
implies the local copy carries it too (lines 1-161 identical), but that is an inference; M1
measures it directly.

## §D.1 AC-001 — a failure on a HEAD is re-delivered on the next call (release-blocking)

- **Given** a fixture repository with `go.mod` and a code file, HEAD subject
  `docs(x): sync-phase …`, a stub `go` on `PATH` that exits 1 for `vet`, and no `.moai/state`.
- **When** the local hook runs twice on the same HEAD with stdin `{}` and
  `MOAI_SYNC_GATE_BLOCKING` / `MOAI_AUTONOMY_TIER` unset.
- **Then** call 1's stdout contains `"hookSpecificOutput"` and `"decision":"block"`; call 2's
  stdout is **byte-identical** to call 1's; the stub `go` invocation count after call 2 equals
  the count after call 1 (no re-run); both calls exit 0.
- **RED reason:** on `fa96fe644`, call 2 writes 0 bytes, because the SHA-only sentinel written
  before the checks short-circuits (h01 call2: `h01-develop-call2.out`, 0 bytes; call1: 237
  bytes). Not red for any fixture reason — control A blocks on a new HEAD in the same setup.
- **Green path:** M2's outcome record and re-delivery. Passing output: call 2 prints the same
  237-class block JSON.

## §D.2 AC-002 — a failure followed by a new failing HEAD blocks with the new result (regression-guard)

- **Given** the AC-001 fixture after call 1 (failure recorded), then a new commit (new HEAD,
  still sync-phase, still failing).
- **When** the hook runs once.
- **Then** stdout contains a block, the stub `go` count increased (checks ran for the new
  HEAD), and the record names the new SHA.
- **Baseline:** h01 control A (`h01-develop-ctrlA.out`, 237-byte block) — already true today.

## §D.3 AC-003 — a pass followed by a same-HEAD call stays silent and does not re-run (regression-guard)

- **Given** a fixture whose stub `go` exits 0 for every subcommand.
- **When** the hook runs twice on the same HEAD.
- **Then** both stdouts are empty, and the stub `go` count after call 2 equals the count after
  call 1.
- **Baseline:** h01 controls B1/B2 (0 bytes each; B2 wrote no log line).
- **Mutant guard:** a fix that re-runs checks on every call keeps stdout empty but fails the
  count assertion.

## §D.4 AC-004 — `stop_hook_active` defers re-delivery without losing it

- **(a, release-blocking) Given** the AC-001 state after call 1. **When** the hook runs with
  stdin `{"stop_hook_active": true}`. **Then** stdout contains no `"decision"`, and the stub
  `go` count is unchanged.
- **(b, release-blocking) When** the hook then runs again with stdin `{}`. **Then** stdout is
  byte-identical to call 1's block. *RED reason:* on `fa96fe644` this third call writes 0
  bytes, since nothing is ever re-delivered; that also makes (a) vacuously green there, which
  is why (a) is only adopted together with (b).
- **(c, regression-guard) Given** the AC-001 state after call 1. **When** stdin is
  `{"stop_hook_active": false, "last_assistant_message": "note: \"stop_hook_active\": true"}`
  (the key false; the literal appears escaped inside a string). **Then** stdout is
  byte-identical to call 1's block — the escaped literal did not suppress re-delivery.
- **Green path:** M2's jq-free key detection (plan.md §G).

## §D.5 AC-005 — a legacy SHA-only record triggers one more run (release-blocking)

- **Given** a failing fixture and `.moai/state/sync-quality-gate.last` containing only the
  current HEAD's 40-hex SHA plus a newline.
- **When** the hook runs once.
- **Then** the stub `go` count is at least 1, stdout contains the block, and the record now
  carries an outcome token (`<sha> fail`).
- **RED reason:** on `fa96fe644` a SHA-only record equal to HEAD exits 0 before any check — the
  same short-circuit as h01 call2 — so the count is 0 and stdout is empty.
- **Green path:** M2 (REQ-007).

## §D.6 AC-006 — `running` records

- **(a, release-blocking) Given** a failing fixture and a record `<HEAD> running` whose mtime
  is now. **When** the hook runs. **Then** the stub `go` count is 0; stdout contains
  `"systemMessage"`; stdout contains no `"decision"`. *RED reason:* on `fa96fe644` the record
  content differs from the bare SHA, so the checks run (count > 0) and a block is emitted.
- **(b, regression-guard) Given** the same record with mtime 120 s in the past. **When** the
  hook runs. **Then** the count is ≥ 1 and stdout contains a block. *Baseline:* true on
  `fa96fe644`, since the record differs from the bare SHA.
- **(c, release-blocking — the retry bound and its notice) Given** a stub `go` that sleeps
  (bounded by its own timeout) and a hook process killed as a process group once the record
  reads `running`, aged to 120 s; then a second run killed the same way and aged again (the
  one allowed re-run is now spent). **When** the hook runs a third time with a stub `go` that
  exits 1. **Then** the stub count for the third run is 0; stdout contains `"systemMessage"`
  whose text says this HEAD's gate run has not completed and names deleting the state file as
  the way to force a re-gate; stdout contains no `"decision"`. *RED reason:* on `fa96fe644`
  the first interrupted run leaves a SHA record, so the third run is silent: count 0 and no
  decision already hold, but stdout is empty, so the notice assertion fails. The count and
  no-decision parts are the retry bound, which must stay green. Unix-only; skipped on Windows.
  Kills are registered in `t.Cleanup`.
- **Green path:** M2 (REQ-008, REQ-009).

## §D.7 AC-007 — no path is looser than today (regression-guard matrix)

Each row runs once on a fresh fixture. The row passes when stdout contains
`"hookSpecificOutput"` and `"decision":"block"` and the exit code is 0. **All rows must pass on
the M1 commit (today's hook) and after M2.**

| Row | Environment | Stub `go` fails | State before the call | stdin |
|---|---|---|---|---|
| R1 | both env vars unset | `vet` only | no `.moai/state` | `{}` |
| R2 | both unset | `build` only | none | `{}` |
| R3 | `MOAI_SYNC_GATE_BLOCKING=1` | `vet` | none | `{}` |
| R4 | `MOAI_AUTONOMY_TIER=semi-auto` | `vet` | none | `{}` |
| R5 | `MOAI_AUTONOMY_TIER=automatic` | `build` | none | `{}` |
| R6 | `MOAI_AUTONOMY_TIER=bogus` | `vet` | none | `{}` |
| R7 | both unset | `vet` | record for a **different** SHA with outcome `pass` (or bare SHA) | `{}` |
| R8 | both unset | `vet` | record + persisted block for a **different** SHA | `{}` — stdout must name the current run, not replay the old bytes |
| R9 | both unset | `vet` | `.moai` exists, `.moai/state` absent | `{}` |
| R10 | both unset | `vet` | none | `{"stop_hook_active": true}` |
| R11 | both unset | `vet` | none | `/dev/null` |
| R12 | both unset | `vet` | none, **initial commit** (no `HEAD~1`) | `{}` |

- **Failure meaning:** any row failing after M2 means the change loosened the gate (D2
  violation).
- **R8 detail:** a fix that replays any persisted block regardless of SHA would pass stdout
  shape but fail the "current run" check. The test asserts that the stub `go` count increased.

## §D.8 AC-008 — re-delivery follows mode resolution

Each row calls the hook twice on the same failing HEAD.

| Row | Call 1 env | Call 2 env | Stub fails | Call 2 expectation | Class |
|---|---|---|---|---|---|
| A1 | tier `fully-autonomous` | same | `build` | call 1 carries no `"decision"`; call 2 stdout empty; no re-run | regression-guard |
| A2 | tier `automatic` | same | `vet` only | call 1 carries no `"decision"`; call 2 stdout empty; no re-run | regression-guard |
| A3 | `MOAI_SYNC_GATE_BLOCKING=0` | same | `vet` | call 1 carries no `"decision"`; call 2 stdout empty; no re-run | regression-guard |
| A4 | tier `automatic` | same | `build` | call 2 stdout byte-identical to call 1's block | release-blocking (RED: call 2 is 0 bytes today) |
| A5 | both unset | tier `fully-autonomous` | `vet` | call 2 stdout empty; no re-run | regression-guard |
| A6 | tier `fully-autonomous` | both unset | `vet` | call 2 stdout empty; no re-run (an advisory first run is never retroactively blocked) | regression-guard |

- **Green path:** M2 (REQ-006; lead rulings B2 and B3). A5 is the direct probe of "re-emit only
  while blocking". A1-A3 pin the once-only advisory warning. A6 pins "never retroactively
  blocked". Baseline for every regression-guard row: on `fa96fe644`, call 2 writes 0 bytes.

## §D.9 AC-009 — existing guards and invariants stay green (regression-guard)

- **When** `go test ./internal/hook/ ./internal/template/ -count=1` runs after M2 and after M3.
- **Then** it exits 0, and the verbose `-run` selection reports these by name:
  `TestHookWrapperCopiesStayIdentical`, `TestAC004_SyncGateAdvisoryAtFullyAutonomous`,
  `TestAC002_NonSyncHeadSkipsVetBuild`,
  `TestHookOfficialCompliance_AC002_SyncGateStopHookSpecificOutput`, and every
  `TestSyncGateFailState_*` test. A selector that matches zero tests is not a pass.
- **And** `grep -c "jq" <template hook>` prints `0`, and `grep -c "SPEC-" <template hook>` prints
  `0` (L-14 baseline).

## §D.10 AC-010 — copy parity (regression-guard)

- **Then** `cmp` of the two hook copies exits 0 (L-13 baseline).
- **And** lines 1-161 of the two document copies are identical (L-15 baseline).
- **And** lines 162 onward of each document copy equal the same range at `fa96fe644`: the
  output of `tail -n +162` on the working file matches `git show fa96fe644:<path>` piped
  through `tail -n +162`, for both paths.

## §D.11 AC-011 — H03 wording corrected in both copies (release-blocking)

| Check | RED on `fa96fe644` | Green after M3 |
|---|---|---|
| scan claim, template doc | L-02: `1`, exit 0 | `0`, exit 1 |
| scan claim, local doc | L-03: `1`, exit 0 | `0`, exit 1 |
| `deps_modified`, template doc | L-04: `0`, exit 1 | ≥ `1`, exit 0 |
| `deps_modified`, local doc | L-05: `0`, exit 1 | ≥ `1`, exit 0 |
| "only check for that drift", template | L-08: `1`, exit 0 | `0`, exit 1 |
| "only check for that drift", local | L-09: `1`, exit 0 | `0`, exit 1 |
| "Audit ALL … present at project root", template | L-10: `1`, exit 0 | `0`, exit 1 |
| same phrase, local | measured in M1 (ledger gap) | `0`, exit 1 |
| hook header (`head -5`, template) | L-11: contains `manifest audit` | contains no `manifest audit`; matches `manifest[ -]change` |

- **RED reason:** the false claims are present, and the name of the value the hook actually
  records is absent.
- **Reviewer read (mutant guard M9):** lines 116 and 140-146 describe the hook as recording
  `deps_modified` for manifests changed in the HEAD commit, informational only. Deleting the
  claims without adding that description fails this check.

## §D.12 AC-012 — SX-R05 reconciliation in both copies

- **(a, release-blocking)** `grep -c "sync-auditor FAIL"` prints ≥ `1` for both document copies
  (RED: L-06 / L-07 print `0`, exit 1). Reviewer read: the sentence carrying it states that the
  sync-auditor rubric (Critical/High → FAIL) is canonical and that Phase 8's CRITICAL-only gate
  never clears a sync-auditor FAIL. The token inside a contrary sentence fails.
- **(b, regression-guard)** line 71's `HARD THRESHOLD: any Critical/High finding causes overall
  FAIL regardless of other scores` is still present in both copies (L-16 baseline), and the Phase
  8 blocking rule still names CRITICAL as the only blocking severity (no behavior change).

## §D.13 AC-013 — deleting the state file forces a retry (regression-guard)

- **Given** the AC-001 state after call 1. **When** `.moai/state/sync-quality-gate.last` is
  removed and the hook runs. **Then** the stub `go` count increased and stdout contains a block.
- **And** the hook's leading comment block names `.moai/state/sync-quality-gate.last` together
  with deletion as the way to force a re-run (reviewer read).
- **Baseline:** on `fa96fe644`, removing the sentinel re-runs the checks, which is why the
  stopchain tests reset this way.

## §D.14 AC-014 — the stale window equals the registered timeout (release-blocking)

- **Given** the template hook and `internal/template/templates/.claude/settings.json.tmpl`.
- **When** a Go test reads the hook's named stale-window variable (a single assignment of an
  integer literal) and the `timeout` of the settings entry whose args name
  `sync-phase-quality-gate.sh`.
- **Then** both values exist and are equal (60 today).
- **RED reason:** no stale-window variable exists on `fa96fe644`, so the test fails at M1 for
  lack of the variable, not for any settings reason (the settings entry reads `"timeout": 60`).

## §D.15 AC-015 — gate notices never block and never consume the Stop-hook block cap (release-blocking)

- **Given** two fixtures, both on a failing sync-phase HEAD with a counting stub `go`:
  - **(N1)** the exhausted-retry state of AC-006c (the one allowed stale re-run is spent);
  - **(N2)** a record `<HEAD> running` whose mtime is refreshed to now before every call (the
    fresh-running notice of REQ-008).
- **When** the hook runs **9 consecutive times** on each fixture (one more than the runtime
  Stop-hook block cap of 8), alternating stdin between `{}` and `{"stop_hook_active": true}`,
  with `MOAI_SYNC_GATE_BLOCKING` and `MOAI_AUTONOMY_TIER` unset (the blocking default, where a
  block would be emitted if one were possible).
- **Then**, for each fixture:
  - all 9 runs exit 0;
  - all 9 stdouts contain `"systemMessage"`;
  - the number of stdouts containing `"decision"` is exactly `0`, so no invocation can count
    toward the block cap;
  - the stub `go` count after run 9 equals the count before run 1 (the notices never re-run
    the checks).
  - For N1, every notice also says that this HEAD's gate run has not completed and names
    deleting the state file as the way to force a re-gate.
- **Swept-set check:** the test reports 18 invocations (9 × 2). A loop that executes fewer
  than 9 per fixture is a partial sweep, not a pass.
- **RED reason:** on `fa96fe644`, N1 is silent (the SHA record short-circuits), and N2 runs the
  checks and emits a block (the record content differs from the bare SHA). N1 fails the
  `"systemMessage"` assertion; N2 fails both the `"decision"`-count and stub-count assertions.
  Neither fails for a fixture reason.
- **Green path:** M2 (REQ-008, REQ-009; lead ruling B1).

## §D.16 Mutant probes (run in M4; each must turn at least one named criterion red)

| Mutant | Change | Must fail |
|---|---|---|
| M1 | Re-emit the stored block and ignore `stop_hook_active` | AC-004a |
| M2 | Never re-emit (today's behavior) | AC-001, AC-004b, AC-008 A4 |
| M3 | Re-run the checks on a same-HEAD `fail` instead of re-emitting | AC-001 (count), AC-003 unaffected |
| M4 | Re-emit the stored block whatever the resolved mode | AC-008 A5 |
| M5 | Treat a fresh `running` record as stale | AC-006a, AC-015 N2 |
| M6 | Re-run on every stale `running` record (no retry bound) | AC-006c |
| M7 | Detect `stop_hook_active` with a plain substring match | AC-004c |
| M8 | Let `stop_hook_active` suppress every block | AC-007 R10 |
| M9 | Delete the false document sentences and add nothing | AC-011 (`deps_modified` rows) |
| M10 | Replay any persisted block regardless of SHA | AC-007 R8 |
| M11 | After the retry is spent, stay silent (the previous default) | AC-006c, AC-015 N1 |
| M12 | Emit the exhausted-retry notice together with `"decision":"block"` | AC-015 N1 (`decision` count) |
| M13 | Repeat the advisory warning on re-delivery | AC-008 A1-A3 (call 2 stdout empty) |

A mutant that turns nothing red means the criterion it targets is too shallow; tighten the
criterion before closing. Each mutant is reverted after its measurement.

## §D.17 Definition of Done

- All 15 criteria pass. Each is recorded in `progress.md §E.2` with the command, its verbatim
  output (or the persisted file holding it), and the tree SHA measured.
- M1's test-only commit precedes every hook or document change, and it records the four
  elements (command, verbatim stdout, exit code, tree SHA) that make the behavioral criteria
  release-blocking (lead ruling B4).
- All thirteen mutants of §D.16 are observed turning their named criteria red.
- `cmp` of the hook copies exits 0; document lines 1-161 are identical, and lines 162 onward are
  unchanged against `fa96fe644`.
- `go vet ./internal/hook/...` and `golangci-lint run ./internal/hook/...` are clean on the
  touched package.
- No `go test ./...`, no `make build`, no push. Every commit names `t624`.
