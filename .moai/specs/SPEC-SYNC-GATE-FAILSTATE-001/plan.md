# Plan — SPEC-SYNC-GATE-FAILSTATE-001

Card: t624 · Tier M · Branch `WT-sync-gate-failstate` · Worktree `.claude/worktrees/t624` ·
Plan-phase tree `fa96fe644`

Sections are ordered by how likely a decision is to change: open decisions and the state
contract first, mechanical steps last.

## §A Context

- SPEC artifacts: `.moai/specs/SPEC-SYNC-GATE-FAILSTATE-001/{spec,plan,acceptance,progress}.md`.
- Defect evidence: `.moai/reports/t624/h01-repro-develop.md` plus `h01-develop-{call1,call2,ctrlA,ctrlB1,ctrlB2}.out`
  (tree `d5dc42959`; target files identical at `fa96fe644`, acceptance.md §D.0 L-01).
- Files in scope:
  - `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` (source) and
    `.claude/hooks/moai/sync-phase-quality-gate.sh` (byte-identical copy).
  - `internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md`
    (source) and `.claude/skills/moai/workflows/sync/quality-gates-quality.md` (lines 1-161 only).
  - New test file(s) under `internal/hook/` (suggested: `sync_gate_failstate_test.go`, test
    names prefixed `TestSyncGateFailState_` so one `-run` selector covers them).
- Guards that already pin this hook and must stay green:
  `TestHookWrapperCopiesStayIdentical`, `TestAC004_SyncGateAdvisoryAtFullyAutonomous`,
  `TestAC002_NonSyncHeadSkipsVetBuild`, `TestHookOfficialCompliance_AC002_SyncGateStopHookSpecificOutput`.

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
  (REQ-006; AC-008 rows A5, A6).
- **B3 — advisory warnings.** The advisory warning is emitted once, by the run that executed
  the checks, as today. A re-delivery under an advisory mode writes nothing to stdout
  (REQ-006; AC-008 rows A1-A3).
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

## §C State contract (the data model — decided by D1/D2)

**Record:** `.moai/state/sync-quality-gate.last` holds one line, `<40-hex sha> <outcome>`,
with outcome in {`running`, `pass`, `fail`}. A line with only a SHA is the legacy format.

**Persisted block:** kept under `.moai/state/`, bound to its HEAD (for example
`.moai/state/sync-quality-gate.block`, whose first line is the SHA and whose remaining bytes
are the stdout that was emitted). The exact shape is up to run-phase; the acceptance tests do
not read it directly.

**Stale-window marker for B1:** run-phase picks a representation (for example a
`running-retry` outcome token) for "the one allowed stale re-run has been used". Tests observe
it only through behavior (AC-006c).

Behavior on invocation, for a sync-phase HEAD with a code delta:

| Record for this HEAD | stdin `stop_hook_active` | Action | stdout |
|---|---|---|---|
| none / other SHA / file absent | any | run checks (today's path) | today's output |
| `pass` | any | nothing | empty |
| `fail` + persisted block, mode blocking | absent / false | re-emit | persisted bytes, byte-identical |
| `fail` + persisted block, mode blocking | true | nothing; record unchanged | no `decision` |
| `fail` + persisted block, mode advisory | any | nothing (B2, B3) | empty |
| `fail`, no persisted block | any | nothing (B2, B3) | empty |
| legacy SHA-only | any | run checks once; rewrite record | check output |
| `running`, fresh (≤ window) | any | no checks | `systemMessage` notice only ("previous run has not completed") |
| `running`, stale, retry unused | any | run checks; mark retry used | check output |
| `running`, stale, retry used | any | no checks (B1) | `systemMessage` notice only ("this HEAD's gate run has not completed; delete the state file to force a re-gate"); never `decision` |

The non-sync-HEAD, no-language-marker, and zero-code-delta early exits run **before** any
state is read, exactly as today.

## §D Constraints (DO NOT VIOLATE)

- **Template-First:** edit the template copy, then copy it byte-for-byte to the local hook;
  for the document, apply the identical lines-1-161 edit to the local copy. Do **not** touch
  lines 162 onward of either document copy.
- All new state lives under `.moai/state/` (the stopchain tests reset with `os.RemoveAll(".moai/state")`).
- The **first** line containing `printf` + `decision` + `block` must remain the compliant
  `hookSpecificOutput` printf. Re-emit the persisted bytes with a line that does not contain
  the word `decision` (for example `printf '%s' "$STORED"` or `cat`).
- No `jq`, no `moai` binary call, exit 0 on every path, `set -e` safety on every new command.
- No SPEC ID, card id, date, or SHA in template content (hook comments or the document).
- Detection and check commands for all 16 programming languages stay unchanged.
- Verification: `go test ./internal/hook/ ./internal/template/` (targeted `-run` while
  iterating). No `go test ./...`, no `make build`, no background load. Process-group kills for
  the interrupted-run test go in `t.Cleanup`, and any sleeping stub bounds its own lifetime.
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
  `t.TempDir()` git fixture (reuse `initGoFixtureRepo`, `mustRunGit`, `requireBash`,
  `requireGit`, and the counting-stub pattern; add a stub variant whose exit depends on the
  `vet` / `build` argument, so C1-only and C2-only failures are reproducible without a real
  toolchain).
- Cover AC-001, AC-004, AC-005, AC-006, AC-007, AC-008, AC-013, AC-014, and AC-015
  (acceptance.md).
- Commit **the tests alone** (`test(t624): …`), with no hook or document change. Run
  `go test ./internal/hook/ -run '^TestSyncGateFailState' -count=1 -v`, and record the command,
  verbatim stdout, exit code, and commit SHA in `progress.md §E.2` as the RED cells. Confirm
  that no `[no tests to run]` appears and that each release-blocking test fails for the stated
  reason; regression-guard rows are expected to PASS on this commit. This commit is the ordering
  witness (verification-claim-integrity §2.3).

### M2 — outcome record in the hook (Priority High)

- Implement §C in the template hook, then copy it to the local hook (`cmp` exit 0).
- Replace the "Once-per-commit" header block and the sentinel comment with a description of the
  outcome record, re-delivery, `stop_hook_active`, the stale window (a named variable holding
  60 with a comment pointing at the settings timeout), and delete-to-retry (REQ-011, REQ-012).
- Run the M1 selector until it is green, then run the four existing guards and
  `go test ./internal/hook/ ./internal/template/`.

### M3 — document and header wording (Priority Medium)

- Template `quality-gates-quality.md`: rewrite line 116 (the dependency step's scope and the
  "only check for that drift" claim) and lines 140-146 (manifest-change observation via
  `deps_modified`, informational; no scan claim; no every-manifest-at-root claim; keep the
  supply-chain skill fallback sentence only if it is reworded as an agent-invoked review, not
  a replacement for a hook scan). Add the SX-R05 reconciliation sentence near Phase 8 Step
  0.55.1 / 0.55.2, naming `sync-auditor FAIL` literally.
- Hook header Purpose line (both copies): describe compile/vet plus manifest-change
  observation.
- Apply the identical lines-1-161 edit to the local document; verify lines 162 onward are
  unchanged against `fa96fe644`.

### M4 — mutant probes and evidence (Priority Medium)

- Run the mutant probes in acceptance.md §D.16. Record each mutant's observed failure, then
  revert it.
- Fill `progress.md §E.2` / `§E.3` with the AC matrix (command, verbatim output, tree SHA).

## §G Technical approach notes

- **Reading stdin.** The hook does not read stdin today. Read it only when a decision depends
  on it, and never block: skip the read when stdin is a terminal (`[ -t 0 ]`), and treat an
  empty or `/dev/null` stdin as "field absent".
- **`stop_hook_active` without jq.** Match the key only where it is an object key, never a
  quote escaped inside a string. For example, require that the `"` opening
  `"stop_hook_active"` is not preceded by `\`, then allow optional whitespace, `:`, optional
  whitespace, and `true`. AC-004c probes the escaped-literal case.
- **mtime portability.** `stat` flags differ between BSD/macOS (`-f %m`) and GNU (`-c %Y`);
  Windows git-bash ships GNU. Prefer a form that works on both (for example, compare against a
  reference file with `find … -newer`, or try one `stat` form and fall back to the other), and
  on failure lean toward "stale", which runs the checks — never toward silence.
- **Atomic writes.** Write the record and the persisted block to a temporary file inside
  `.moai/state/`, then `mv` it into place, so a concurrent reader never sees a half-written
  line.
- **Mode resolution at re-delivery** needs the failed-check composition (C1 versus C2) for the
  `automatic` tier. Persist the composition with the block, or derive it from the stored
  record; do not re-run the checks to find out.
- **Logging.** A re-delivery MAY append a log line (for example `decision=block-redelivered`).
  Acceptance does not assert it.

## §H Risks

| Risk | Impact | Mitigation |
|---|---|---|
| Stale re-run loops on slow checks | per-turn toolchain re-runs return | REQ-009 one-retry bound; AC-006c |
| Exhausted-retry notice repeats every turn | notice noise until a new commit or state-file deletion (accepted by lead ruling B1) | notice is `systemMessage` only, never `decision`, so it never counts toward the Stop-hook block cap of 8; AC-015 |
| A `stop_hook_active` literal in message text is misread | a stored failure is suppressed forever | key-position match; AC-004c |
| Re-emit printf becomes the compliance first match | `TestHookOfficialCompliance_AC002` fails or passes vacuously | §D printf rule; AC-009 |
| A stored block re-emitted under an advisory tier | advisory users suddenly blocked | REQ-006; AC-008 flip rows |
| The 60-second constant drifts from the settings timeout | fresh runs misread as stale, or the reverse | AC-014 equality test |
| mtime unreadable on a platform | the fresh-running branch never taken | fall toward stale (runs checks, never silence) |
| Two sessions share `.moai/state/` | interleaved records | atomic `mv`; locking out of scope |
| An unbounded stdin read | hook hangs until the 60-second kill | `[ -t 0 ]` guard; AC-009 `/dev/null` run |

## §I Anti-patterns to avoid

- Fixing H01 by re-running the checks on every same-HEAD call. The mutant probe M3 exists to
  catch exactly this.
- Letting `stop_hook_active` suppress a first-time block (mutant M8 against AC-007 row R10).
- Deleting the false sentences without writing the true ones (mutant M9 against AC-011).
- Editing the local document past line 161, or hand-syncing the two documents by whole-file
  copy (the copies intentionally differ from line 162 on).

## §J Cross-references

- `.claude/rules/moai/core/hooks-system.md` § Stop Hook Block Cap and `stop_hook_active`
- `.claude/rules/moai/development/hook-independence.md` §4 (row d: this gate is jq-free)
- `.claude/rules/moai/development/verification-completeness.md` §2, §2.1
- `.claude/rules/moai/core/verification-claim-integrity.md` §2.3 (ordering witness)
- `.claude/agents/moai/sync-auditor.md` § Evaluation Dimensions
