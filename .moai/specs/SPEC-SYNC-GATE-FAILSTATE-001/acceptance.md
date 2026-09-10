# Acceptance — SPEC-SYNC-GATE-FAILSTATE-001

> Harness: **standard**. Document-level pin: every observation below that carries no pin of
> its own was measured on tree `fa96fe644fcff8a15ac336833a4e816fc0a46fe3`. Ledger rows L-15 to
> L-19 carry their own pin, `ad0ad6f9b`, where the target files are unchanged from `fa96fe644`
> (L-17).
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
- A criterion may hold rows of both classes; each row states its class. A release-blocking row
  names its **RED reason** — the defect signal M1 must observe — and that reason is never a
  harness wait or timeout.

## §D AC Matrix

| AC | REQ | Class | RED / baseline source | Green path |
|---|---|---|---|---|
| AC-001 | 001·002·003 | release-blocking | h01 call2 0 bytes (supporting); M1 RED cell | M2 |
| AC-002 | 004 | regression-guard | h01 control A blocks | M2 keeps it |
| AC-003 | 004 | regression-guard | h01 controls B1/B2 silent | M2 keeps it |
| AC-004 | 005 | release-blocking (a, b, c — all stdin forms) | M1 RED cell | M2 |
| AC-005 | 001·002·007 | release-blocking (L1, TA1, TA4, TB1; record-format sub-assertion of U-rows); regression-guard (U1-U5 run-and-block, TA2, TA3, TA5, TB2) | M1 RED cell / baseline | M2 |
| AC-006 | 008·009 | release-blocking (a0, a50, c); regression-guard (b70, b120) | M1 RED cell / baseline | M2 |
| AC-007 | 010 | regression-guard | M1 baseline PASS (today blocks) | M2 keeps it |
| AC-008 | 006 | release-blocking (A4, A9); regression-guard (A1-A3, A5-A8) | M1 RED cell / baseline | M2 |
| AC-009 | 012 | regression-guard | existing guards green; L-14, L-18, L-19 | M2, M3 keep them |
| AC-010 | 013 | regression-guard | L-13, L-15, L-16, L-17 | M2, M3 keep them |
| AC-011 | 014 | release-blocking | L-02..L-05, L-08..L-12 | M3 |
| AC-012 | 015 | release-blocking (a); regression-guard (b) | L-06, L-07, L-20 | M3 |
| AC-013 | 002·011 | regression-guard | deleting the sentinel re-runs today | M2 keeps it |
| AC-014 | 008 | release-blocking | M1 RED cell (no named window exists) | M2 |
| AC-015 | 008·009 | release-blocking | M1 RED cell (named reasons, §D.15) | M2 |

15 criteria (Tier M ceiling 16).

## §D.0 Evidence ledger (plan-phase observations)

Each command ran as a single invocation. Where an exit code is shown, it was captured either by
an `echo "exit=$?"` appended in the same call (not part of the cited command or stdout) or by
the tool reporting no error (exit 0).

| Id | Tree | Command | Verbatim stdout | Exit |
|---|---|---|---|---|
| L-01 | pin | `git diff --stat d5dc42959 fa96fe644 -- .claude/hooks/moai/sync-phase-quality-gate.sh internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md .claude/skills/moai/workflows/sync/quality-gates-quality.md internal/hook/ internal/template/hook_official_compliance_test.go` | *(empty)* | 0 |
| L-02 | pin | `grep -c "dependency vulnerability scan runs automatically via the Stop hook" internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md` | `1` | 0 |
| L-03 | pin | `grep -c "dependency vulnerability scan runs automatically via the Stop hook" .claude/skills/moai/workflows/sync/quality-gates-quality.md` | `1` | 0 |
| L-04 | pin | `grep -c "deps_modified" internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md` | `0` | 1 |
| L-05 | pin | `grep -c "deps_modified" .claude/skills/moai/workflows/sync/quality-gates-quality.md` | `0` | 1 |
| L-06 | pin | `grep -c "sync-auditor FAIL" internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md` | `0` | 1 |
| L-07 | pin | `grep -c "sync-auditor FAIL" .claude/skills/moai/workflows/sync/quality-gates-quality.md` | `0` | 1 |
| L-08 | pin | `grep -c "Skipping the whole of Phase 8 would remove the only check for that drift" internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md` | `1` | 0 |
| L-09 | pin | `grep -c "Skipping the whole of Phase 8 would remove the only check for that drift" .claude/skills/moai/workflows/sync/quality-gates-quality.md` | `1` | 0 |
| L-10 | pin | `grep -c "Audit ALL of the following manifest files present at project root" internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md` | `1` | 0 |
| L-11 | pin | `head -5 internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` | see block below | 0 |
| L-12 | pin | `grep -c "manifest audit" internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` | `2` | 0 |
| L-13 | pin | `cmp .claude/hooks/moai/sync-phase-quality-gate.sh internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` | *(empty)* | 0 |
| L-14 | pin | `grep -c "SPEC-" internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` | `0` | 1 |
| L-15 | `ad0ad6f9b` | `/usr/bin/grep -n "^Purpose: Ensure code has appropriate @MX annotations for AI agent context" internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md` (and the same on the local copy) | `160:Purpose: Ensure code has appropriate @MX annotations for AI agent context. Supports all 16 MoAI-ADK languages.` (identical for both copies) | 0 |
| L-17 | `ad0ad6f9b` | `git diff --stat fa96fe644 -- internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md .claude/skills/moai/workflows/sync/quality-gates-quality.md` | *(empty)* | 0 |
| L-18 | `ad0ad6f9b` | `/usr/bin/grep -cE '(^\|[;&\|(]\|\$\()[[:space:]]*jq([[:space:]]\|$)' internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` | `0` | 1 |
| L-19 | `ad0ad6f9b` | `/usr/bin/grep -cE '(^\|[^A-Za-z0-9_])t[0-9]{2,4}([^A-Za-z0-9_]\|$)\|20[0-9]{2}-[0-9]{2}-[0-9]{2}' internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md` | `…/sync-phase-quality-gate.sh:0` and `…/quality-gates-quality.md:0` | 1 |

(In the L-18 and L-19 cells, `\|` renders a literal `|` inside the table; the executed regex
contains a plain `|`.)

L-11 verbatim stdout:

```
#!/bin/bash
# Hook: sync-phase-quality-gate
# Purpose: Fast sync-phase quality gate (compile/vet + dependency manifest audit)
# Trigger: Stop event when the current session's HEAD is a sync-phase commit
#
```

Non-single-invocation baselines (regression-guard support only, not RED cells):

- **L-16** (`ad0ad6f9b`) — a process-substitution `diff` of each document copy through the
  anchor line (`sed '/^Purpose: Ensure code has appropriate @MX annotations for AI agent
  context/q'`) printed nothing, exit 0.
- **L-18 control** — the L-18 regex fed three lines (`x=$(echo "$in" | jq -r .a)`, a comment
  line naming jq, and `jq . file`) printed `2`. It matches the two invocations and skips the
  comment.
- **L-20** — a line-filtered `grep -n` sweep of the template document showed line 71 carrying
  `HARD THRESHOLD: any Critical/High finding causes overall FAIL regardless of other scores`,
  and lines 136-137 and 156 carrying the CRITICAL-only / HIGH-warning wording.

**Ledger gaps, all measured in M1:**
- L-10's phrase, `grep -c "manifest audit"`, `grep -c "SPEC-"`, and the L-19 pattern were not
  measured on the local copies. L-13 and L-16 imply equal values, but that is an inference.
- Which tier of `TestTemplateNoInternalContentLeak` covers these two template paths for dates
  and SHAs was not verified (its SHA class is package-restricted).

**Unverified behavior — explicit gap, no criterion.** The fallback "record-file mtime cannot be
read → treat the `running` record as stale" (plan.md §G) is **not verified by any criterion in
this document**.
- **Why no row:** the age-reading mechanism (a `stat` form, or a `find -newer` reference-file
  comparison) is a run-phase choice. A test cannot make mtime unreadable without knowing which
  tool to break, so any row written now would rest on an implementation guess.
- **What this means:** a green AC-006 says nothing about this fallback. A mutant that falls
  toward "fresh" (notice, no checks) whenever mtime is unreadable passes every row here.
- **Residual risk:** on a platform where mtime is unreadable, a genuinely stale `running` record
  could keep emitting the fresh-running notice instead of re-gating. That is still non-blocking
  and never a silent pass, but it is never re-gated.

## §D.1 AC-001 — a failure on a HEAD is re-delivered on the next call (release-blocking)

- **Given** a fixture repository with `go.mod` and a code file, HEAD subject
  `docs(x): sync-phase …`, a stub `go` on `PATH` that exits 1 for `vet`, and no `.moai/state`.
- **When** the local hook runs twice on the same HEAD with stdin `{}` and
  `MOAI_SYNC_GATE_BLOCKING` / `MOAI_AUTONOMY_TIER` unset.
- **Then** call 1's stdout contains `"hookSpecificOutput"` and `"decision":"block"`; call 2's
  stdout is **byte-identical** to call 1's; the stub `go` invocation count after call 2 equals
  the count after call 1 (no re-run); both calls exit 0.
- **RED reason:** call 2 writes 0 bytes — the stored failure is not re-delivered, because the
  bare-SHA sentinel written before the checks short-circuits (h01 call2: 0 bytes; call1: 237
  bytes).
- **Green path:** M2's outcome record and re-delivery.

## §D.2 AC-002 — a failure followed by a new failing HEAD blocks with the new result (regression-guard)

- **Given** the AC-001 fixture after call 1 (failure recorded), then a new commit (new HEAD,
  still sync-phase, still failing).
- **When** the hook runs once.
- **Then** stdout contains a block, the stub `go` count increased (checks ran for the new
  HEAD), and the record names the new SHA.
- **Baseline:** h01 control A (`h01-develop-ctrlA.out`, 237-byte block). Written in M1.

## §D.3 AC-003 — a pass followed by a same-HEAD call stays silent and does not re-run (regression-guard)

- **Given** a fixture whose stub `go` exits 0 for every subcommand.
- **When** the hook runs twice on the same HEAD.
- **Then** both stdouts are empty, and the stub `go` count after call 2 equals the count after
  call 1.
- **Baseline:** h01 controls B1/B2 (0 bytes each; B2 wrote no log line). Written in M1.

## §D.4 AC-004 — `stop_hook_active` defers re-delivery without losing it (release-blocking, a/b/c together)

Run the (a)→(b) pair once for each stdin form of the flag, each on a fresh fixture:

| Form | stdin for step (a) |
|---|---|
| F1 single space | `{"stop_hook_active": true}` |
| F2 compact | `{"stop_hook_active":true}` |
| F3 multi-space | `{"stop_hook_active"   :   true}` |

- **(a) Given** the AC-001 state after call 1. **When** the hook runs with the form's stdin.
  **Then** stdout contains no `"decision"`, and the stub `go` count is unchanged.
- **(b) When** the hook then runs again with stdin `{}`. **Then** stdout is byte-identical to
  call 1's block.
- **(c) Given** the AC-001 state after call 1. **When** stdin is
  `{"stop_hook_active": false, "last_assistant_message": "note: \"stop_hook_active\": true"}`
  (the key false; the literal appears escaped inside a string). **Then** stdout is
  byte-identical to call 1's block.
- **RED reason (b and c):** the stored failure is not re-delivered on a call without the flag
  — today every same-HEAD call after call 1 writes 0 bytes. On `fa96fe644`, (a) passes
  vacuously because nothing is ever re-delivered, which is why (a), (b), and (c) are adopted
  together (B4).
- **Green path:** M2's whitespace-agnostic, key-position detection (plan.md §G).

## §D.5 AC-005 — unknown and legacy records trigger one more run

Each row starts from a failing fixture (stub `go` exits 1 for `vet`, modes unset), with the
record file prepared as shown, and runs the hook once.

| Row | Record prepared as | Class |
|---|---|---|
| L1 | the current HEAD's 40-hex SHA plus a newline (legacy) | release-blocking |
| U1 | `<HEAD> bogus` (unknown token) | regression-guard (run/block) |
| U2 | `<HEAD> fail extra` (extra field) | regression-guard (run/block) |
| U3 | empty file | regression-guard (run/block) |
| U4 | `<HEAD> fail`, file mode 0200 (writable, unreadable); skipped as root and on Windows | regression-guard (run/block) |
| U5 | `<HEAD> fail`, no payload file in `.moai/state/` | regression-guard (run/block) |

- **Then**, for every row: the stub `go` count is ≥ 1; stdout contains a block; the record now
  reads `<HEAD> fail` (a closed-set token, REQ-001).
- **RED reason:**
  - **L1:** the checks never run and stdout is empty — a bare SHA equal to HEAD short-circuits
    (the same defect as h01 call2).
  - **U1-U5:** the run-and-block part already holds today, because the content differs from
    the bare SHA. Their record-format sub-assertion is red today for the REQ-001 reason (today
    rewrites a bare SHA), and it flips in M2.
- **Green path:** M2 (REQ-007; REQ-002 write ordering for U5).

### AC-005 torn-write rows — a crash between the two writes never reads as a silent pass

Write ordering (REQ-002: payload first, then the `fail` record) is the safety mechanism, so each
torn state it can leave must read as a re-gate or as a non-blocking notice — **never as a silent
pass**. The payload file's layout is left to run-phase, so every row builds its torn state
behaviorally: the gate writes a genuine payload, and the test changes only the record or removes
only the auxiliary files.

**Torn state (a) — payload present, record not yet `fail`.**
- **Setup:** run the hook once with tier `fully-autonomous` and `go vet` failing, so a genuine
  **advisory** payload is stored. Then replace only `.moai/state/sync-quality-gate.last` as shown,
  leaving the payload file in place.
- **Final call:** modes unset, `go vet` failing, stdin `{}`.

| Row | `.last` after the replacement | Expected final call | Class |
|---|---|---|---|
| TA1 | `<HEAD> running`, age 0 s (crash after the payload write) | stub count 0; stdout contains `"systemMessage"`, no `"decision"` (REQ-008 notice) | release-blocking |
| TA2 | `<HEAD> running`, age 120 s | stub count ≥ 1; stdout contains a block (REQ-009 re-gate) | regression-guard |
| TA3 | file removed (record absent) | stub count ≥ 1; stdout contains a block (REQ-004) | regression-guard |
| TA4 | the current HEAD's bare SHA (legacy content) | stub count ≥ 1; stdout contains a block (REQ-007) | release-blocking |
| TA5 | a bare SHA or `<sha> pass` naming an earlier commit (previous content) | stub count ≥ 1; stdout contains a block (REQ-004) | regression-guard |

**Torn state (b) — neither payload nor `fail` record written.**
- **Setup:** run the hook once with modes unset and `go vet` failing. Then remove every file under
  `.moai/state/` except `sync-quality-gate.last` (payload and retry marker gone), and set `.last`
  as shown.
- **Final call:** modes unset, `go vet` failing, stdin `{}`.

| Row | `.last` | Expected final call | Class |
|---|---|---|---|
| TB1 | `<HEAD> running`, age 0 s (crash while the checks ran) | stub count 0; stdout contains `"systemMessage"`, no `"decision"` (REQ-008 notice) | release-blocking |
| TB2 | `<HEAD> running`, age 120 s | stub count ≥ 1; stdout contains a block (REQ-009 re-gate) | regression-guard |
| TB3 | `<HEAD> fail` (the payload removed before the record flipped) | covered by U5: re-gate and block (REQ-007) | — |
| TB4 | `.moai/state` removed entirely | covered by AC-007 R9 and AC-013 D1: run and block | — |

- **RED reason (TA1, TB1):** the fresh-running notice is absent. Today's hook reads
  `<HEAD> running` as a foreign record, runs the checks, and blocks — the same signal as AC-006
  a0.
- **RED reason (TA4):** the checks never run and stdout is empty. The bare SHA equal to HEAD
  short-circuits, the same defect as L1.
- **Baseline (TA2, TA3, TA5, TB2):** true today, because the record is absent or differs from the
  bare HEAD SHA, so the checks run and block. None of these rows may become a silent pass after
  M2.
- **Mutants that would turn a torn state into a silent pass:**
  - **M23** — decide from the payload file alone: an advisory payload present for HEAD reads as a
    completed advisory `fail`, whatever the record says. It turns TA1-TA5 silent.
  - **M24** — read a stale `running` record with no payload file as a completed pass. It turns
    TB2 silent.

## §D.6 AC-006 — `running` records and the stale window

Each row starts from a failing fixture with the record `<HEAD> running`, its file mtime set to
the age shown, modes unset, and stdin `{}`.

| Row | Record age | Then | Class |
|---|---|---|---|
| a0 | 0 s | stub count 0; stdout contains `"systemMessage"`, no `"decision"` | release-blocking |
| a50 | 50 s | stub count 0; stdout contains `"systemMessage"`, no `"decision"` | release-blocking |
| b70 | 70 s | stub count ≥ 1; stdout contains a block | regression-guard |
| b120 | 120 s | stub count ≥ 1; stdout contains a block | regression-guard |

- **RED reason (a0, a50):** the fresh-running notice is absent. Today's hook treats
  `<HEAD> running` as a foreign record, runs the checks (count > 0), and blocks.
- **Baseline (b70, b120):** true today for the same reason, since the record differs from the
  bare SHA.
- **Boundary purpose:** a50 and b70 pin the 60 s window by behavior. A window of 40 s fails a50;
  a window of 100 s fails b70.

### AC-006c — the retry bound and its notice (release-blocking)

- **Given** a sync-phase Go fixture; a stub `go` that writes a marker file, then sleeps
  (bounded by its own timeout); and the interrupted-run harness of plan.md §F M1, which kills
  the hook's process group when the marker appears or proceeds when the hook exits on its own,
  whichever comes first.
  - **Run 1:** interrupted by that harness. The record file is then aged to 120 s.
  - **Run 2:** interrupted the same way. The record file is then aged to 120 s again.
- **When** run 3 executes with the stub switched to exit 1 immediately.
- **Then:**
  1. run 1 observed the marker (setup sanity);
  2. the stub count for run 3 is 0;
  3. run 3's stdout contains no `"decision"`;
  4. run 3's stdout contains `"systemMessage"`, whose text says this HEAD's gate run has not
     completed and names deleting the state file as the way to force a re-gate.
- **Named RED reason:** *the non-blocking exhausted-retry notice is absent* — run 3's stdout is
  empty, so assertion 4 fails.
  - On `fa96fe644`, run 1 writes the bare SHA before invoking the stub, so the marker still
    appears and the kill happens.
  - Runs 2 and 3 find that SHA, short-circuit, and exit 0 without invoking the stub. The
    harness proceeds on process exit, so no wait or timeout occurs.
  - Assertions 1-3 already pass there. They are the re-run cap, which must stay green.
- Unix-only; skipped on Windows. Kills are registered in `t.Cleanup`.
- **Green path:** M2 (REQ-008, REQ-009).

## §D.7 AC-007 — no path is looser than today (regression-guard, REQ-010)

Each row's final invocation passes when stdout contains `"hookSpecificOutput"` and
`"decision":"block"`, the exit code is 0, and — for history rows — the stub count increased on
that invocation. **All rows must pass on the M1 commit (today's hook) and after M2.** Rows are
one representative per equivalence class of REQ-010's argument.

| Row | Class covered | Environment | Failing check | Starting state / history | stdin |
|---|---|---|---|---|---|
| R1 | composition: C1 only | both unset | `go vet` | no `.moai/state` | `{}` |
| R2 | composition: C2 only | both unset | `go build` | none | `{}` |
| R3 | blocking opt-out: legacy `1` | `MOAI_SYNC_GATE_BLOCKING=1` | `go vet` | none | `{}` |
| R4 | tier: semi-auto | `MOAI_AUTONOMY_TIER=semi-auto` | `go vet` | none | `{}` |
| R5 | tier: automatic, C2 | `MOAI_AUTONOMY_TIER=automatic` | `go build` | none | `{}` |
| R6 | tier: unrecognized | `MOAI_AUTONOMY_TIER=bogus` | `go vet` | none | `{}` |
| R7 | history: another HEAD passed | both unset | `go vet` (on HEAD 2) | gate run on an earlier sync commit with passing stubs, then a new failing sync commit | `{}` |
| R8 | history: another HEAD failed | both unset | `go vet` (both HEADs) | gate run on an earlier failing sync commit (blocked), then a new failing sync commit; stdout must be the new run's output (stub count increased), not a replay | `{}` |
| R9 | state dir absent | both unset | `go vet` | `.moai` exists, `.moai/state` absent | `{}` |
| R10a | stdin: flag set, spaced | both unset | `go vet` | none | `{"stop_hook_active": true}` |
| R10b | stdin: flag set, compact | both unset | `go vet` | none | `{"stop_hook_active":true}` |
| R11 | stdin: empty | both unset | `go vet` | none | `/dev/null` |
| R12 | initial commit | both unset | `go vet` | none, no `HEAD~1` | `{}` |
| R13 | language: non-Go C1-only branch | both unset | `ruff` (Python fixture: `pyproject.toml` + `.py` change) | none | `{}` |
| R14 | subject: `chore: sync` pattern | both unset | `go vet` | none; HEAD subject `chore: sync docs` | `{}` |
| R15 | tier: automatic, C1 and C2 | `MOAI_AUTONOMY_TIER=automatic` | `go vet` and `go build` | none | `{}` |
| R16 | blocking opt-out: empty | `MOAI_SYNC_GATE_BLOCKING=` (set, empty) | `go vet` | none | `{}` |
| R17 | history: legacy start | both unset | `go vet` | a legacy bare-SHA record for an earlier commit, then a new failing sync commit | `{}` |

- **Failure meaning:** any row failing after M2 means the change loosened the gate (D2
  violation).
- **Not rows here, by design:** test-written same-HEAD records such as `<HEAD> pass` or a fresh
  `<HEAD> running`. Today they block only because the content differs from the bare SHA, which
  no gate-written history produces; REQ-010 scopes them out, and AC-003, AC-005, and AC-006
  define their behavior.

## §D.8 AC-008 — re-delivery follows mode resolution

Each row calls the hook twice on the same failing HEAD.

| Row | Call 1 env | Call 2 env | Stub fails | Expectation | Class | Mutant it closes |
|---|---|---|---|---|---|---|
| A1 | tier `fully-autonomous` | same | `build` | call 1 carries no `"decision"`; call 2 stdout empty; no re-run | regression-guard | M13 |
| A2 | tier `automatic` | same | `vet` only | call 1 carries no `"decision"`; call 2 stdout empty; no re-run | regression-guard | M13 |
| A3 | `MOAI_SYNC_GATE_BLOCKING=0` | same | `vet` | call 1 carries no `"decision"`; call 2 stdout empty; no re-run | regression-guard | M13 |
| A4 | tier `automatic` | same | `build` | call 2 stdout byte-identical to call 1's block | release-blocking | M2 |
| A5 | both unset | tier `fully-autonomous` | `vet` | call 2 stdout empty; no re-run | regression-guard | M4 |
| A6 | tier `fully-autonomous` | both unset | `vet` | call 2 stdout empty; no re-run (advisory first run never retroactively blocked) | regression-guard | M22 |
| A7 | both unset | `MOAI_SYNC_GATE_BLOCKING=0` | `vet` | call 2 stdout empty; no re-run | regression-guard | M15 |
| A8 | both unset | tier `automatic` | `vet` only | call 2 stdout empty; no re-run | regression-guard | M16 |
| A9 | tier `automatic` | same | `vet` and `build` | call 2 stdout byte-identical to call 1's block | release-blocking | M17 |

- **RED reason (A4, A9):** call 2 writes 0 bytes — the stored block is not re-delivered under a
  tier that resolves to blocking.
- **Baseline for the regression-guard rows:** on `fa96fe644`, call 2 writes 0 bytes.
- **Green path:** M2 (REQ-006; lead rulings B2 and B3).

## §D.9 AC-009 — existing guards and invariants stay green (regression-guard)

- **When** `go test ./internal/hook/ ./internal/template/ -count=1` runs after M2 and after M3.
- **Then** it exits 0, and a verbose `-run` selection reports each of these by name:
  - `TestHookWrapperCopiesStayIdentical`
  - `TestAC004_SyncGateAdvisoryAtFullyAutonomous`
  - `TestAC002_NonSyncHeadSkipsVetBuild`
  - `TestHookOfficialCompliance_AC002_SyncGateStopHookSpecificOutput`
  - `TestTemplateNoInternalContentLeak`
  - `TestTemplateNeutralityAudit`
  - every `TestSyncGateFailState_*` test

  A selector that matches zero tests is not a pass.
- **And (no jq invocation):** the L-18 invocation-pattern regex prints `0` for the template
  hook. A comment naming the tool does not count, and the L-18 control shows the regex does match
  real invocations.
- **And (template neutrality):** for both the template hook and the template document:
  - `grep -c "SPEC-"` prints `0`;
  - the L-19 card-id and date regex prints `0` for each file.

  Commit-SHA-shaped tokens are delegated to `TestTemplateNoInternalContentLeak`; which tier of
  that guard covers these paths is a ledger gap measured in M1.

## §D.10 AC-010 — copy parity (regression-guard)

- **Then** `cmp` of the two hook copies exits 0 (L-13 baseline).
- **And** `grep -c "^Purpose: Ensure code has appropriate @MX annotations for AI agent context"`
  prints `1` for each document copy (L-15 baseline).
- **And** each document copy printed through the anchor line
  (`sed '/^Purpose: Ensure code has appropriate @MX annotations for AI agent context/q'`) is
  identical between the two copies (L-16 baseline).
- **And**, for each copy, the text after the anchor line
  (`sed '1,/^Purpose: Ensure code has appropriate @MX annotations for AI agent context/d'`)
  equals the same range of that copy at `fa96fe644`. Obtain that version with
  `git show fa96fe644:<path>` into a scratch file first, then `sed` and `diff` as separate
  invocations.
- **Why a content anchor:** the D6 reconciliation sentence adds lines above the anchor, which
  shifts every line number after it. A line-number check would fail on a correct edit.

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
| `manifest audit`, template hook (header and manifest-step comment) | L-12: `2`, exit 0 | `0`, exit 1 |
| `manifest audit`, local hook | measured in M1 (ledger gap) | `0`, exit 1 |
| hook header (`head -5`, template) | L-11: contains `manifest audit` | matches `manifest[ -]change` |

- **RED reason:** the false claims are present, and the name of the value the hook actually
  records is absent.
- **Reviewer read (mutant M9):** lines 116 and 140-146 describe the hook as recording
  `deps_modified` for manifests changed in the HEAD commit, informational only. Deleting the
  claims without adding that description fails this check.

## §D.12 AC-012 — SX-R05 reconciliation in both copies

- **(a, release-blocking)** `grep -c "sync-auditor FAIL"` prints ≥ `1` for both document copies
  (RED: L-06 / L-07 print `0`, exit 1). Reviewer read: the sentence carrying it states that the
  sync-auditor rubric (Critical/High → FAIL) is canonical and that Phase 8's CRITICAL-only gate
  never clears a sync-auditor FAIL. The token inside a contrary sentence fails.
- **(b, regression-guard)** line 71's `HARD THRESHOLD: any Critical/High finding causes overall
  FAIL regardless of other scores` is still present in both copies (L-20 baseline), and the
  Phase 8 blocking rule still names CRITICAL as the only blocking severity (no behavior change).

## §D.13 AC-013 — retry by deletion, and no stale auxiliary state (regression-guard)

| Row | Sequence | Then |
|---|---|---|
| D1 | AC-001 state after call 1 → remove `.moai/state/sync-quality-gate.last` → run with modes unset | stub count increased; stdout contains a block |
| S1 | call 1 with modes unset, `go vet` failing (blocks; block payload stored) → remove `.moai/state/sync-quality-gate.last` → call 2 with tier `fully-autonomous` (advisory run) → call 3 with modes unset | call 2 stub count increased and stdout carries no `"decision"`; **call 3 stdout is empty and its stub count is unchanged** — the call-1 block is not re-delivered |

- **And** the hook's leading comment block names `.moai/state/sync-quality-gate.last` together
  with deletion as the way to force a re-run (reviewer read).
- **Baseline:** on `fa96fe644`, D1 re-runs (which is why the stopchain tests reset this way);
  S1 call 3 is silent.
- **S1 failure meaning:** a stale payload survived a check-running invocation and blocked an
  advisory run retroactively (REQ-002 violated; mutant M18).

## §D.14 AC-014 — the stale window equals the registered timeout (release-blocking)

- **Given** the template hook and `internal/template/templates/.claude/settings.json.tmpl`.
- **When** a Go test reads the hook's named stale-window variable (a single assignment of an
  integer literal) and the `timeout` of the settings entry whose args name
  `sync-phase-quality-gate.sh`.
- **Then** both values exist and are equal (60 today).
- **RED reason:** the named stale-window variable does not exist on `fa96fe644`. The settings
  entry reads `"timeout": 60`.
- **Scope note:** this criterion pins the declaration only. That the comparison actually uses
  it is pinned by AC-006 rows a50 and b70.

## §D.15 AC-015 — gate notices never block and never consume the Stop-hook block cap (release-blocking)

- **Given** two fixtures, both on a failing sync-phase HEAD with a counting stub `go`:
  - **(N1)** the exhausted-retry state of AC-006c after run 2 (built with the same
    marker-or-exit harness);
  - **(N2)** a record `<HEAD> running` whose mtime is refreshed to now before every call (the
    fresh-running notice of REQ-008).
- **When** the hook runs **9 consecutive times** on each fixture (one more than the runtime
  Stop-hook block cap of 8), alternating stdin between `{}` and `{"stop_hook_active":true}`,
  with `MOAI_SYNC_GATE_BLOCKING` and `MOAI_AUTONOMY_TIER` unset (the blocking default).
- **Then**, for each fixture:
  - all 9 runs exit 0;
  - all 9 stdouts contain `"systemMessage"`;
  - the number of stdouts containing `"decision"` is exactly `0`;
  - the stub `go` count after run 9 equals the count before run 1.
  - For N1, every notice also says this HEAD's gate run has not completed and names deleting
    the state file as the way to force a re-gate.
- **Swept-set check:** the test reports 18 invocations (9 × 2). Fewer is a partial sweep, not
  a pass.
- **Named RED reason, N1:** *the repeated non-blocking exhausted-retry notice is absent* — all
  9 stdouts are empty, so the `"systemMessage"` assertion fails. Today's hook short-circuits on
  the bare-SHA record left by run 1 and exits 0 without invoking the stub; the harness proceeds
  on process exit, so no wait or timeout occurs. The `decision`-count and stub-count assertions
  already pass there.
- **Named RED reason, N2:** *the fresh-running notice is absent* — today's hook treats each
  refreshed `<HEAD> running` record as a foreign record, runs the checks, and blocks. The
  `decision` count is 9 and the stub count grows.
- **Green path:** M2 (REQ-008, REQ-009; lead ruling B1).

## §D.16 Mutant probes (run in M4; each must turn at least one named row red)

| Mutant | Change | Must fail |
|---|---|---|
| M1 | Re-emit the stored block and ignore `stop_hook_active` | AC-004 (a), every form |
| M2 | Never re-emit (today's behavior) | AC-001, AC-004 (b), AC-008 A4 and A9 |
| M3 | Re-run the checks on a same-HEAD `fail` instead of re-emitting | AC-001 (count) |
| M4 | Re-emit the stored block whatever the resolved tier | AC-008 A5 |
| M5 | Treat a fresh `running` record as stale | AC-006 a0, AC-015 N2 |
| M6 | Re-run on every stale `running` record (no retry bound) | AC-006c |
| M7 | Detect `stop_hook_active` with a plain substring match | AC-004 (c) |
| M8 | Let `stop_hook_active` suppress every block | AC-007 R10a, R10b |
| M9 | Delete the false document sentences and add nothing | AC-011 (`deps_modified` rows) |
| M10 | Replay any stored block regardless of SHA | AC-007 R8 |
| M11 | After the retry is spent, stay silent | AC-006c, AC-015 N1 |
| M12 | Emit the exhausted-retry notice together with `"decision":"block"` | AC-015 N1 (`decision` count) |
| M13 | Repeat the advisory warning on re-delivery | AC-008 A1-A3 (call 2 stdout empty) |
| M14 | Detect `stop_hook_active` only with exactly one space after the colon | AC-004 F2 and F3 |
| M15 | Ignore `MOAI_SYNC_GATE_BLOCKING` when resolving the mode at re-delivery | AC-008 A7 |
| M16 | Treat tier `automatic` as blocking at re-delivery regardless of the failed-check composition | AC-008 A8 |
| M17 | Treat tier `automatic` as advisory whenever `vet` failed, ignoring the `build` failure | AC-008 A9 |
| M18 | Leave the payload file in place when an invocation runs the checks | AC-013 S1 |
| M19 | Treat `<HEAD> <any token>` other than `fail` as a silent pass | AC-005 U1 |
| M20 | Declare the stale window as 60 but compare against 100 | AC-006 b70 |
| M21 | Declare the stale window as 60 but compare against 40 | AC-006 a50 |
| M22 | Re-deliver an advisory payload as a block when the mode now resolves blocking, or treat a `fail` record without a payload as a silent pass | AC-008 A6, AC-005 U5 |
| M23 | Decide from the payload file alone: an advisory payload present for HEAD reads as a completed advisory `fail` and passes silently, whatever the record says (torn state a) | AC-005 TA1-TA5 |
| M24 | Read a stale `running` record with no payload file as a completed pass (torn state b) | AC-005 TB2 |

A mutant that turns nothing red means the criterion it targets is too shallow; tighten the
criterion before closing. Each mutant is reverted after its measurement.

## §D.17 Definition of Done

- All 15 criteria pass, every row included. Each is recorded in `progress.md §E.2` with the
  command, its verbatim output (or the persisted file holding it), and the tree SHA measured.
- M1's test-only commit precedes every hook or document change, and it records the four
  elements (command, verbatim stdout, exit code, tree SHA) that make the behavioral criteria
  release-blocking (lead ruling B4). The observed RED reasons match the named reasons above,
  and every regression-guard row passes on that commit.
- All twenty-four mutants of §D.16 are observed turning their named rows red.
- The mtime-unreadable fallback remains an explicit unverified gap (§D.0); closing the card
  does not claim it.
- `cmp` of the hook copies exits 0; AC-010's anchor checks hold.
- `go vet ./internal/hook/...` and `golangci-lint run ./internal/hook/...` are clean on the
  touched package.
- No `go test ./...`, no `make build`, no push. Every commit names `t624`.
