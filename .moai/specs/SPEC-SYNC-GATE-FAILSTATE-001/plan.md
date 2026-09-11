# Plan — SPEC-SYNC-GATE-FAILSTATE-001

Card: t624 · Tier M · Branch `WT-sync-gate-failstate` · Worktree `.claude/worktrees/t624` ·
Plan-phase tree `fa96fe644` (target files unchanged through `ad0ad6f9b`, acceptance.md §D.0 L-17)

Sections are ordered by how likely a decision is to change: decisions and the state contract
first, mechanical steps last.

## §A Context

- SPEC artifacts: `.moai/specs/SPEC-SYNC-GATE-FAILSTATE-001/{spec,plan,acceptance,progress}.md`.
- Defect evidence: `.moai/reports/t624/h01-repro-develop.md` plus `h01-develop-{call1,call2,ctrlA,ctrlB1,ctrlB2}.out`
  (tree `d5dc42959`; target files identical at `fa96fe644`, acceptance.md §D.0 L-01).
- Plan-audit verdicts, all three:
  - round 1: `.moai/reports/t624/plan-audit.md` (FAIL 0.60, 11 blocking);
  - round 2: `.moai/reports/t624/plan-audit-r2.md` (FAIL 0.86, 4 blocking);
  - round 3: `.moai/reports/t624/plan-audit-r3.md` (FAIL 0.86, 1 blocking, NEW-1).

  Every finding of all three rounds is dispositioned in the current artifacts. After round 3 the
  operator approved Implementation Kickoff as PASS-with-debt, with NEW-1 paid before M1
  (`progress.md §E.1`, provenance record).
- Files in scope:
  - `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` (source) and
    `.claude/hooks/moai/sync-phase-quality-gate.sh` (byte-identical copy).
  - `internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md`
    (source) and `.claude/skills/moai/workflows/sync/quality-gates-quality.md` — edits only above
    the Phase 9 anchor line (REQ-013).
  - New test file(s) under `internal/hook/` (suggested: `sync_gate_failstate_test.go`, test
    names prefixed `TestSyncGateFailState_` so one `-run` selector covers them).
- Guards that already pin this hook and must stay green:
  `TestHookWrapperCopiesStayIdentical`, `TestAC004_SyncGateAdvisoryAtFullyAutonomous`,
  `TestAC002_NonSyncHeadSkipsVetBuild`, `TestHookOfficialCompliance_AC002_SyncGateStopHookSpecificOutput`,
  `TestTemplateNoInternalContentLeak`, `TestTemplateNeutralityAudit`.

## §B Resolved decisions (lead rulings, 2026-09-10)

Every ruling below moves in the "never looser than today" direction.

- **B1 — after the one allowed stale re-run.** The per-HEAD cap stays at ONE re-run for a
  stale `running` record: re-running on every stale record would bring back the per-turn
  re-run the original sentinel existed to prevent. If that single re-run also does not
  complete, later invocations for that HEAD neither re-run the checks nor stay silent. Each
  one emits a NON-BLOCKING `systemMessage` saying this HEAD's gate run has not completed and
  that deleting the state file forces a new gate run (REQ-009). The notice never carries
  `decision`, so it never counts toward the runtime Stop-hook block cap of 8 (AC-015).
- **B2 — mode changes between turns.** The mode is resolved again at re-delivery. The stored
  block is re-emitted verbatim only when that mode resolves to blocking. An advisory first run
  is never retroactively blocked, even if the environment later resolves to blocking
  (REQ-006; AC-008 rows A5-A8).
- **B3 — advisory warnings.** The advisory warning is emitted once, by the run that executed
  the checks, as today. A re-delivery under an advisory resolution, or of an advisory payload,
  writes nothing to stdout (REQ-006; AC-008 rows A1-A3).
- **B4 — when behavioral RED cells are recorded.** Accepted. Each behavioral criterion becomes
  release-blocking at the M1 test-only commit, when its four elements (command, verbatim
  stdout, exit code, tree SHA) are recorded in `progress.md §E.2`. The h01 reproduction stays
  supporting evidence.
- **B5 — no `make build` (record only).** The dispatch excludes `make build` from run-phase
  verification. `go test ./internal/template/` compiles the embedded tree fresh and the hook
  tests read the scripts from disk, so verification loses nothing. Refreshing the installed
  binary is the integration step's job. Recorded so the Template-First "`make build` after a
  template edit" habit is not applied by reflex and then reported as a gap.
- **B6 — doc line 43.** Out of this SPEC. The claim that the Stop hook reads the shared
  diagnostic snapshot is handled by a separate card; spec.md §5 keeps a one-line note.
- **B7 — reconciling the audit's write-ordering choice with B2/B3 (plan-audit round 1, O7).**
  The audit chose: write the block before the `fail` record, and treat a `fail` record with no
  block as unknown, so the gate re-runs. Taken literally, an advisory first run — which emits
  no block — would re-run the checks on every later turn. That repeats the advisory warning
  (breaks B3) and blocks once the environment turns blocking (breaks B2). The payload file
  therefore stores the output of **every** failing run, block or advisory, and is written before
  the `fail` record. A `fail` record with **no payload file** is unknown and re-gates (REQ-007),
  while a `fail` record with an **advisory payload** stays silent (REQ-006). This is a state
  representation inside D1, not new policy.

## §C State contract (the data model)

**Record** — `.moai/state/sync-quality-gate.last`: one line, `<40-hex sha> <outcome>`, with the
outcome from the closed set {`running`, `pass`, `fail`} (REQ-001). A bare SHA is the legacy
format. Anything else — empty, unreadable, unknown token, extra field — is malformed.

**Payload file** — a separate file under `.moai/state/` (suggested
`.moai/state/sync-quality-gate.payload`), bound to its HEAD. It holds the exact stdout bytes of
the failing run, the payload kind (`block` or `advisory`), and the failed-check composition
(C1 and C2 exit codes) that mode resolution needs at re-delivery. The layout is up to
run-phase; the acceptance tests observe it only through behavior.

**Retry marker** — a separate file under `.moai/state/` (suggested
`.moai/state/sync-quality-gate.retry`) holding the SHA of the HEAD whose one stale re-run has
been used. It is **not** an outcome token: REQ-001's closed set stays closed.

**Write ordering and invalidation (REQ-002).** Every write uses a temporary file inside
`.moai/state/` renamed into place. When an invocation runs the checks, the steps happen in this
order:
1. Remove the payload file.
2. Write the retry marker if this is the stale re-run; otherwise remove it.
3. Write `<sha> running`.
4. Run the checks.
5. On failure, write the payload file, then `<sha> fail`. On success, write `<sha> pass`.

Behavior on invocation, for a sync-phase HEAD with a code delta (the unchanged early exits run
**before** any state is read):

| State for this HEAD | stdin `stop_hook_active` | Action | stdout |
|---|---|---|---|
| no record / record for another SHA / `.moai/state` absent | any | run checks (today's path) | today's output |
| `pass` | any | nothing | empty |
| `fail` + block payload, mode resolves blocking | absent / false | re-emit | stored bytes, byte-identical |
| `fail` + block payload, mode resolves blocking | true (any whitespace form) | nothing; state unchanged | no `decision` |
| `fail` + block payload, mode resolves advisory | any | nothing (B2) | empty |
| `fail` + advisory payload | any | nothing (B2, B3) | empty |
| `fail` + **no** payload file | any | run checks; rewrite record (B7) | check output |
| legacy bare SHA / empty / unreadable / unknown token / extra field | any | run checks; rewrite record | check output |
| `running`, age ≤ window (exactly 60 s is fresh) | any | no checks | `systemMessage` notice only ("previous run has not completed") |
| `running`, age > window (stale), no retry marker for this SHA | any | run checks as the stale re-run (marker written) | check output |
| `running`, age > window (stale), retry marker for this SHA | any | no checks (B1) | `systemMessage` notice only ("this HEAD's gate run has not completed; delete the state file to force a re-gate"); never `decision` |

## §D Constraints (DO NOT VIOLATE)

- **Template-First (process note, not a verifiable requirement):** edit the template hook,
  then copy it byte-for-byte to the local hook; for the document, apply the identical edit
  above the Phase 9 anchor line to the local copy. Do **not** touch anything after the anchor
  line in either document copy.
- All new state lives under `.moai/state/` (the stopchain tests reset with
  `os.RemoveAll(".moai/state")`).
- The **first** line containing `printf` + `decision` + `block` must remain the compliant
  `hookSpecificOutput` printf. Re-emit stored bytes with a line that does not contain the word
  `decision` (for example `printf '%s'` over the stored bytes, or `cat`).
- No `jq` invocation, no `moai` binary call, exit 0 on every path, `set -e` safety on every new
  command.
- No SPEC ID, card id, date, or SHA in template content (hook comments or the document).
- Early exits, detection, check commands, and mode resolution for all 16 programming languages
  stay unchanged; the state logic sits after the early exits.
- Verification: `go test ./internal/hook/ ./internal/template/` (targeted `-run` while
  iterating). No `go test ./...`, no `make build`, no background load. Process-group kills for
  the interrupted-run tests go in `t.Cleanup`, and any sleeping stub bounds its own lifetime.
- Commit by explicit pathspec. Every commit message carries `t624`. No push.

## §E Pre-flight (run-phase, before M1)

- `git rev-parse --show-toplevel` → `.claude/worktrees/t624`; `git branch --show-current` →
  `WT-sync-gate-failstate`.
- `cmp` of the two hook copies → exit 0 (baseline L-13).
- `git diff --stat fa96fe644 -- <hook copies> <doc copies>` → empty (no drift since plan-phase).
- `go test ./internal/hook/ -run 'TestAC004_SyncGateAdvisoryAtFullyAutonomous|TestAC002_NonSyncHeadSkipsVetBuild|TestHookWrapperCopiesStayIdentical' -count=1`
  → `ok` (baseline before any change).

## §F Milestones (priority order)

### M1 — RED tests first (Priority High)

- Add `internal/hook/sync_gate_failstate_test.go`, driving the **local** hook copy in a
  `t.TempDir()` git fixture. Reuse `initGoFixtureRepo`, `mustRunGit`, `requireBash`,
  `requireGit`, and the counting-stub pattern. Add:
  - a stub variant whose exit depends on the `vet` / `build` argument, so C1-only, C2-only,
    and both-failing compositions are reproducible without a real toolchain;
  - a Python fixture helper (`pyproject.toml` plus a `.py` change) with a failing `ruff` stub;
  - a sleeping stub that **writes a marker file before it sleeps**, whose sleep is bounded by
    its own timeout.
- **Interrupted-run harness (AC-006c, AC-015 N1).** For each run:
  1. **Delete the stub's marker file before starting the run**, so a marker left by an earlier
     run can never trigger this run's kill.
  2. Start the hook in its own process group.
  3. Wait for **whichever comes first**: the marker file appears, or the hook process exits on
     its own.
  4. On the marker, record "marker observed for this run" and kill the process group
     (registered in `t.Cleanup`). Then age the record file.

  The harness never waits on a state-file token. On today's hook (which never writes `running`),
  a run that short-circuits simply exits: the test records "marker not observed" for that run
  and moves on, so no step can end in a harness timeout. Assertions are **non-fatal**
  (`t.Errorf`, not `t.Fatalf`), so every failing assertion of a row is observed and can be
  compared with the named RED reason.
- Cover AC-001, AC-002, AC-003, AC-004, AC-005, AC-006, AC-007, AC-008, AC-013, AC-014, and
  AC-015 (acceptance.md).
- Commit **the tests alone** (`test(t624): …`), with no hook or document change. Run
  `go test ./internal/hook/ -run '^TestSyncGateFailState' -count=1 -v`, and record the command,
  verbatim stdout, exit code, and commit SHA in `progress.md §E.2` as the RED cells. Confirm:
  - no `[no tests to run]` appears;
  - each release-blocking test fails with the named RED reason given in acceptance.md;
  - every regression-guard row PASSES on this commit.

  This commit is the ordering witness (verification-claim-integrity §2.3). Also measure the
  ledger gaps named in acceptance.md §D.0.

### M2 — outcome record and auxiliary state in the hook (Priority High)

- Implement §C in the template hook, then copy it to the local hook (`cmp` exit 0).
- Declare the stale window as a named variable holding 60 (with a comment pointing at the
  settings timeout), and use that variable in the age comparison.
- Replace the "Once-per-commit" header block and the sentinel comment with a description of
  the record, payload, retry marker, re-delivery, `stop_hook_active`, and delete-to-retry
  (REQ-011, REQ-012).
- Run the M1 selector until it is green, then run the six existing guards and
  `go test ./internal/hook/ ./internal/template/`.

### M3 — document and hook-comment wording (Priority Medium)

- Template `quality-gates-quality.md`, above the Phase 9 anchor only:
  - Rewrite line 116: the dependency step's scope and the "only check for that drift" claim.
  - Rewrite lines 140-146 to describe manifest-change observation via `deps_modified`
    (informational). Remove the scan claim and the every-manifest-at-root claim. Keep the
    supply-chain skill fallback sentence only if it is reworded as an agent-invoked review, not
    a replacement for a hook scan.
  - Add the SX-R05 reconciliation sentence near Phase 8 Step 0.55.1 / 0.55.2, naming
    `sync-auditor FAIL` literally.
- Hook comments (both copies): the header Purpose line and the comment above the manifest step
  (line 287 on `fa96fe644`) describe manifest-change observation; the phrase `manifest audit`
  disappears.
- Apply the identical edit to the local document; verify AC-010.

### M4 — mutant probes and evidence (Priority Medium)

- Run the mutant probes in acceptance.md §D.16. Record each mutant's observed failure, then
  revert it.
- Fill `progress.md §E.2` / `§E.3` with the AC matrix (command, verbatim output, tree SHA).

## §G Technical approach notes

- **Reading stdin.** The hook does not read stdin today. Read it only when a decision depends on
  it, and never block: skip the read when stdin is a terminal (`[ -t 0 ]`), and treat an empty
  or `/dev/null` stdin as "field absent".
- **`stop_hook_active` without jq.** Match the key only in object-key position, never inside a
  string: require that the `"` opening `"stop_hook_active"` is not preceded by `\`. Then allow
  any run of spaces or tabs (including none), `:`, another such run, and `true`. This match does
  **not** tell a top-level key from a nested one, and REQ-005 does not require it to (an explicit
  gap in acceptance.md §D.0). The runtime's actual payload spacing was not observed (plan-audit
  Gaps), so AC-004 probes the compact, single-space, multi-space, and tab forms rather than
  assuming one.
- **mtime portability.** `stat` flags differ between BSD/macOS (`-f %m`) and GNU (`-c %Y`);
  Windows git-bash ships GNU. Prefer a form that works on both (for example, compare against a
  reference file with `find … -newer`, or try one `stat` form and fall back to the other). On
  failure, lean toward "stale", which runs the checks — never toward silence.
- **Mode resolution at re-delivery** reads the failed-check composition from the payload file;
  it never re-runs the checks to find out.
- **Logging.** A re-delivery MAY append a log line (for example `decision=block-redelivered`).
  Acceptance does not assert it.

## §H Risks

| Risk | Impact | Mitigation |
|---|---|---|
| Stale re-run loops on slow checks | per-turn toolchain re-runs return | REQ-009 one-retry bound; AC-006c |
| Exhausted-retry notice repeats every turn | notice noise until a new commit or state-file deletion (accepted by lead ruling B1) | notice is `systemMessage` only, never `decision`, so it never counts toward the Stop-hook block cap of 8; AC-015 |
| A `stop_hook_active` literal in message text is misread | a stored failure is suppressed forever | key-position match; AC-004c |
| Runtime sends a spacing variant not probed | re-delivery under the flag spends the block cap | whitespace-agnostic match; AC-004 compact/multi-space rows |
| A stale payload survives a later check run | an advisory run retroactively blocked; after a partial payload-write failure, a stale block re-delivered without running the checks | REQ-002 invalidation; AC-013 S2 (payload absent while checks run), S3 (partial-write re-gate); mutant M18. S1 guards the advisory path but cannot observe M18 |
| Payload write fails under blocking mode | silent pass | payload written before `fail` record; a `fail` record without a payload re-gates (B7); AC-005 U5 |
| Crash between the payload write and the `fail` record | a torn state reads as a silent pass | torn states read as a notice or a re-gate, never silent; AC-005 torn-write rows TA1-TA5, TB1-TB2; mutants M23, M24 |
| Re-emit printf becomes the compliance first match | `TestHookOfficialCompliance_AC002` fails or passes vacuously | §D printf rule; AC-009 |
| A stored block re-emitted under an advisory resolution | advisory users suddenly blocked | REQ-006; AC-008 rows A5, A7, A8 |
| The 60-second constant drifts from the settings timeout, or the comparison uses another value | fresh runs misread as stale, or the reverse | AC-014 equality; AC-006 50 s / 61 s / 70 s boundary rows |
| mtime unreadable on a platform | a stale `running` record is never re-gated (notice repeats), or the fresh branch is never taken | fall toward stale (runs checks, never silence); recorded as an explicit unverified Gap in acceptance.md §D.0 — no criterion pins it |
| Two sessions share `.moai/state/` | interleaved records | atomic rename; locking out of scope |
| An unbounded stdin read | hook hangs until the 60-second kill | `[ -t 0 ]` guard; AC-007 R11 `/dev/null` run |

## §I Anti-patterns to avoid

- Fixing H01 by re-running the checks on every same-HEAD call (mutant M3).
- Letting `stop_hook_active` suppress a first-time block (mutant M8 against AC-007 R10).
- Encoding the retry state as an extra outcome token (REQ-001's set is closed).
- Deleting the false sentences without writing the true ones (mutant M9 against AC-011).
- Editing either document copy after the Phase 9 anchor line, or hand-syncing the two
  documents by whole-file copy (they intentionally differ after the anchor).

## §J Cross-references

- `.claude/rules/moai/core/hooks-system.md` § Stop Hook Block Cap and `stop_hook_active`
- `.claude/rules/moai/development/hook-independence.md` §4 (row d: this gate is jq-free)
- `.claude/rules/moai/development/verification-completeness.md` §2, §2.1
- `.claude/rules/moai/core/verification-claim-integrity.md` §2.3 (ordering witness)
- `.claude/agents/moai/sync-auditor.md` § Evaluation Dimensions
