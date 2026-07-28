# acceptance.md — SPEC-WORKTREE-ENTRY-STRATEGY-001

> Acceptance criteria (Given-When-Then) + quality-gate criteria + Definition
> of Done. Every AC is observable (test output, file existence, grep evidence,
> or command exit code). Every REQ in spec.md §C has ≥1 matching AC here.

## §A.2 AC format — Given-When-Then convention note

The 16 ACs below use **Given-When-Then (GWT)** format, which is the de facto
project standard for acceptance criteria. The sibling SPEC
`SPEC-WORKTREE-BRANCH-GUARD-001/acceptance.md` (status: completed, PR #1192
merged `e89d01461`) uses the identical GWT form: each AC names a `**Given**`
precondition, a `**When` trigger, and a `**Then**` observable outcome,
followed by a `**Verification command**` (run-phase) whose verbatim output
decides PASS/FAIL. The GEARS/EARS notation is reserved for the *requirements*
in `spec.md §C` (REQ-WES-XXX), not for acceptance criteria. This convention
note records the repo-local policy so iter-2 audit can ratify GWT as the
accepted AC form rather than flagging it as an MP-2 divergence.

## §A. AC Matrix (coverage table)

| REQ | AC(s) | Severity | Verification kind |
|-----|-------|----------|-------------------|
| REQ-WES-001 (EnterWorktree-first) | AC-WES-001a, AC-WES-001b | MUST | Grep + doc inspection |
| REQ-WES-002 (no bare-cd Block 0) | AC-WES-002 | MUST | Grep |
| REQ-WES-003 (L1 vs L2 distinction) | AC-WES-003 | MUST | Grep + doc inspection |
| REQ-WES-004 (web auto-toggles default OFF) | AC-WES-004a, AC-WES-004b, AC-WES-004c | MUST | Go test (defaults assertion) |
| REQ-WES-005 (parallel-session auto-isolation) | AC-WES-005a, AC-WES-005b | MUST | Doc + procedure trace |
| REQ-WES-006 (`moai cc -w` role preserved) | AC-WES-006 | MUST | Grep + behavior preserved |
| REQ-WES-007 (init/update no auto-entry) | AC-WES-007 | MUST | CLI help text / README |
| REQ-WES-008 (doc surfaces reflect policy) | AC-WES-008 | MUST | Multi-file grep |
| REQ-WES-009 (sprawl mitigation) | AC-WES-009 | SHOULD | Doc + baseline cited |
| REQ-WES-010 (launcher `-w` L2 absolute-path extension) | AC-WES-010a, AC-WES-010b, AC-WES-010c | MUST | Go test (launcher resolver) |

## §B. Acceptance Criteria (Given-When-Then)

### AC-WES-001a — `EnterWorktree` documented as canonical current-session entry

**Given** the doc surface at
`.claude/rules/moai/workflow/worktree-integration.md` § `EnterWorktree` /
`ExitWorktree` Tools has been updated per M2;

**When** a reader runs
`grep -n -A2 "EnterWorktree" .claude/rules/moai/workflow/worktree-integration.md`;

**Then** the output contains the phrase "canonical mechanism for entering
an existing worktree in the current session" (or equivalent semantic
phrasing) within 5 lines of an `EnterWorktree` mention.

**Verification command**:
```bash
grep -n -B1 -A5 "EnterWorktree" .claude/rules/moai/workflow/worktree-integration.md | \
  grep -c "current session"
# expected: ≥ 1
```

### AC-WES-001b — bare `cd` deprecated in orchestrator-emitted guidance

**Given** the doc surfaces listed in spec.md §B.6 have been updated per M2-M5;

**When** a reader runs
`grep -nE '^[^#].*\\bcd\\b.*<worktree' .claude/rules/moai/workflow/session-handoff.md
.claude/rules/moai/workflow/session-handoff-examples.md
.claude/rules/moai/workflow/worktree-integration.md`;

**Then** no match appears outside H4/H5 code-block contexts explicitly
documenting the legacy form (deprecated examples must be wrapped in a
`> DEPRECATED` callout or moved to a "Legacy form" subsection).

**Verification command**:
```bash
grep -nE 'cd .*\\.claude/worktrees|cd .*\\.moai/worktrees' \
  .claude/rules/moai/workflow/session-handoff.md \
  .claude/rules/moai/workflow/session-handoff-examples.md \
  .claude/rules/moai/workflow/worktree-integration.md \
  | grep -v "DEPRECATED" | grep -v "^.*#"
# expected: 0 matches outside deprecated callouts
```

### AC-WES-002 — paste-ready Block 0 uses canonical forms

**Given** the Worktree-Anchored Resume Pattern (session-handoff.md §
Worktree-Anchored Resume Pattern + session-handoff-examples.md §
Worktree-Anchored) has been updated per M3;

**When** a reader inspects Block 0 of the paste-ready resume example;

**Then** Block 0 contains EITHER:
- (a) `moai cc -w <worktree-name>` (the new-session launcher form), OR
- (b) `EnterWorktree(<path>)` (the current-session re-entry form);

AND does NOT contain a bare `cd <path> && <launcher>` instruction as the
canonical form (the `cd` form may appear only in an explicitly-deprecated
"Legacy shell form" subsection).

**Verification command**:
```bash
# Canonical forms appear
grep -c "moai cc -w\|EnterWorktree" .claude/rules/moai/workflow/session-handoff-examples.md
# expected: ≥ 2 (both forms documented)

# Bare-cd not canonical outside deprecated callouts
grep -nE '^[^#].*\\bcd .*\\.moai/worktrees' .claude/rules/moai/workflow/session-handoff-examples.md | grep -v DEPRECATED
# expected: 0 matches
```

### AC-WES-003 — `Agent(isolation: "worktree")` L1 vs L2 distinction documented

**Given** the worktree-integration.md § `isolation: worktree` in Agent
Frontmatter (lines 82-99 originally) has been updated per M2;

**When** a reader runs
`grep -n "L1 ephemeral" .claude/rules/moai/workflow/worktree-integration.md`;

**Then** the output contains ≥1 match that explicitly distinguishes L1
ephemeral (created by `Agent(isolation: "worktree")`) from L2 persistent
(created by `moai worktree new`), with an explicit statement that
`Agent(isolation: "worktree")` is NOT a re-entry mechanism for existing L2
worktrees.

**Verification command**:
```bash
grep -n "L1 ephemeral" .claude/rules/moai/workflow/worktree-integration.md | wc -l
# expected: ≥ 2 (the distinction is named in both the glossary and the expanded EnterWorktree section)

grep -n "NOT a re-entry mechanism\|NOT.*re-enter.*L2" .claude/rules/moai/workflow/worktree-integration.md | wc -l
# expected: ≥ 1
```

### AC-WES-004a — `AutoCleanup` default `false`

**Given** `internal/config/defaults.go` has been mutated per M1;

**When** the test suite runs
`go test ./internal/config/... -run TestNewDefaultWorkflowConfig`;

**Then** the assertion `AutoCleanup == false` passes.

**Verification command**:
```bash
go test ./internal/config/... -run TestNewDefaultWorkflowConfig -v 2>&1 | grep -c "AutoCleanup.*false"
# expected: ≥ 1 (the assertion is present and green)
```

### AC-WES-004b — `AutoCreate` default `false` (unchanged)

**Given** `internal/config/defaults.go` `AutoCreate` was already `false`
and remains `false` per M1;

**When** the test suite runs the defaults assertion;

**Then** `AutoCreate == false` passes.

**Verification command**:
```bash
go test ./internal/config/... -run TestNewDefaultWorkflowConfig -v 2>&1 | grep -c "AutoCreate.*false"
# expected: ≥ 1
```

### AC-WES-004c — `AutoMerge` default `false`

**Given** `internal/config/defaults.go` has been mutated per M1 (was `true`);

**When** the test suite runs the defaults assertion;

**Then** `AutoMerge == false` passes.

**Verification command**:
```bash
go test ./internal/config/... -run TestNewDefaultWorkflowConfig -v 2>&1 | grep -c "AutoMerge.*false"
# expected: ≥ 1
```

### AC-WES-005a — parallel-session auto-isolation procedure documented

**Given** worktree-integration.md has been updated per M2 to include the
auto-isolation procedure under Worktree Selection Rules (or a new sub-
section);

**When** a reader runs
`grep -n "auto-.*-.*<spec-id>\|auto-isolation\|parallel-session.*branch.*conflict" .claude/rules/moai/workflow/worktree-integration.md`;

**Then** the output contains ≥1 match documenting the procedure: when
worktree entry is chosen AND another active session is on the same branch,
the orchestrator auto-creates an `auto-<session-short>-<spec-id>` worktree.

**Verification command**:
```bash
grep -nE "auto-<session-short>-<spec-id>|parallel-session branch conflict|auto-isolation" \
  .claude/rules/moai/workflow/worktree-integration.md | wc -l
# expected: ≥ 2 (procedure named + naming scheme cited)
```

### AC-WES-005b — auto-isolation respects branch-guard exemption

**Given** the auto-isolation procedure (REQ-WES-005 / M6) creates a NEW
worktree under `.claude/worktrees/` or `~/.moai/worktrees/`;

**When** the procedure executes in the presence of the sibling
SPEC-WORKTREE-BRANCH-GUARD-001 PreToolUse guard;

**Then** the guard's discriminant (`git rev-parse --git-dir` ≠
`git rev-parse --git-common-dir`) classifies the operation as a worktree
context (NOT primary checkout), and the deny is suppressed — the auto-
isolation proceeds without tripping the branch-guard.

**Verification** (procedure trace + grep evidence):
```bash
# The procedure doc explicitly names worktree paths (NOT primary checkout)
grep -nE "\\.claude/worktrees/|~/\\.moai/worktrees/" .claude/rules/moai/workflow/worktree-integration.md | \
  grep -E "auto-<session|auto-isolation" | wc -l
# expected: ≥ 1
```

### AC-WES-006 — `moai cc -w` role preserved (new-session launcher)

**Given** the doc surfaces have been updated per M2-M3 AND the launcher
`-w` L2 path-resolution extension (REQ-WES-010 / M3a) has landed;

**When** a reader runs
`grep -n "moai cc -w" .claude/rules/moai/workflow/worktree-integration.md
.claude/rules/moai/workflow/session-handoff-examples.md`;

**Then** the output contains ≥2 matches documenting `moai cc -w` as the
canonical launcher for starting a NEW Claude Code session inside a worktree
(the Block 0 new-terminal / post-`/clear` launcher role).

**AND** the legacy `.claude/worktrees/<name>` resolution path in
`internal/cli/launcher.go` `normalizeWorktreeFlag` remains behaviorally
intact (verified via the existing launcher test suite green — no
regression to L1 / short-name resolution). The REQ-WES-010 L2 absolute-
path extension is ADDITIVE; it does NOT rewrite the legacy resolution
path.

**Verification command**:
```bash
grep -n "moai cc -w" .claude/rules/moai/workflow/worktree-integration.md .claude/rules/moai/workflow/session-handoff-examples.md | wc -l
# expected: ≥ 2

# Legacy L1/short-name resolution not regressed
go test ./internal/cli/... -run 'TestNormalizeWorktreeFlag' -v 2>&1 | tail -20
# expected: existing tests PASS (no behavioral regression to .claude/worktrees/<name>)
```

### AC-WES-007 — `moai init` / `moai update` no auto-entry

**Given** the existing `moai init` / `moai update` commands do NOT enter a
worktree on the user's behalf (verified per research.md §G);

**When** a reader runs `moai init --help` and `moai update --help`;

**Then** the help text either (a) carries no worktree-auto-entry claim, OR
(b) explicitly states worktree entry is opt-in via `--worktree` flag or
subsequent `moai worktree new <SPEC-ID>` invocation.

**Verification command**:
```bash
moai init --help 2>&1 | grep -iE "worktree" | grep -iE "auto|enter|switch" | wc -l
# expected: 0 (no auto-entry claim)
```

### AC-WES-008 — doc surfaces reflect EnterWorktree-first policy

**Given** all doc surfaces listed in spec.md §B.6 have been updated;

**When** a reader runs a multi-file grep for the policy keyword
"EnterWorktree-first" (or equivalent semantic phrasing);

**Then** the output contains ≥1 match in EACH of:
- `.claude/rules/moai/workflow/session-handoff.md`
- `.claude/rules/moai/workflow/worktree-integration.md`
- `.claude/rules/moai/workflow/session-handoff-examples.md`
- `CLAUDE.local.md` §22.8 (the new addendum per M4)

**Verification command**:
```bash
for f in .claude/rules/moai/workflow/session-handoff.md \
         .claude/rules/moai/workflow/worktree-integration.md \
         .claude/rules/moai/workflow/session-handoff-examples.md \
         CLAUDE.local.md; do
  echo "$f: $(grep -cE "EnterWorktree|enter.*existing.*worktree" "$f")"
done
# expected: each file ≥ 1
```

### AC-WES-009 — worktree sprawl baseline cited + mitigation documented

**Given** spec.md §A (Problem Statement) cites the 2026-07-28 baseline (58
worktrees / 31 `agent-*` uncleaned);

**When** a reader runs
`grep -n "58 worktrees\|31.*agent-\|sprawl baseline" .moai/specs/SPEC-WORKTREE-ENTRY-STRATEGY-001/spec.md`;

**Then** the output contains ≥1 match citing the baseline as motivation for
REQ-WES-009 (sprawl mitigation via re-entry preference + web defaults OFF).

**Verification command**:
```bash
grep -nE "58 worktrees|31 .agent-|sprawl baseline" .moai/specs/SPEC-WORKTREE-ENTRY-STRATEGY-001/spec.md | wc -l
# expected: ≥ 1
```

### AC-WES-010a — launcher `-w` accepts `~/.moai/worktrees/` absolute paths (L2 re-entry)

**Given** the `internal/cli/launcher.go` `-w` / `--worktree` flag
resolution has been extended per M3a (Decision Point 1 Resolution OQ-1 →
REQ-WES-010);

**When** a user runs `moai cc -w <abs-path-under-~/.moai/worktrees/>` (or
the `glm` / `cg` variants) with an absolute path under
`~/.moai/worktrees/<project>/...`;

**Then** the launcher resolves the path as an L2 worktree re-entry target
(NOT creating a new `.claude/worktrees/` worktree) and starts a new Claude
Code session inside the L2 worktree.

**Verification command**:
```bash
go test ./internal/cli/... -run 'TestNormalizeWorktreeFlag.*L2\|TestLauncherWorktreeL2Abs' -v 2>&1 | tail -30
# expected: NEW test case asserting an L2 absolute path under
#           ~/.moai/worktrees/ resolves to the L2 worktree (NOT a new
#           .claude/worktrees/ dir) — PASS
```

### AC-WES-010b — launcher `-w` legacy `.claude/worktrees/` token-normalization preserved

**Given** the legacy token-normalization behavior of
`normalizeWorktreeFlag` (`internal/cli/launcher.go:665-702` — rewriting
`-w` / `--worktree[=name]` forms into the canonical two-token
`--worktree [name]` form for claude pass-through, which then resolves the
value against `.claude/worktrees/<name>` for short-name inputs);

**When** the existing launcher test suite runs against the post-M3a
launcher;

**Then** all pre-existing launcher tests pass (no behavioral regression
to the short-name token-normalization + claude pass-through path — the
REQ-WES-010 extension is ADDITIVE: a new MoAI-side pre-resolution step
handles `~/.moai/worktrees/` absolute paths before the legacy
token-normalization runs; the legacy token-normalization path itself is
NOT rewritten). If the pre-resolution step is implemented as an inlined
early-return branch within `normalizeWorktreeFlag`, the function body
MAY gain lines; what MUST be preserved behaviorally is the
short-name-input → `--worktree [name]` two-token pass-through output for
non-L2-absolute-path inputs.

**Verification command**:
```bash
go test ./internal/cli/... -run 'TestNormalizeWorktreeFlag' -v 2>&1 | tail -30
# expected: pre-existing tests PASS (no regression for short-name inputs)
```

### AC-WES-010c — launcher `-w` rejects non-worktree absolute paths

**Given** the launcher `-w` extension accepts both `.claude/worktrees/`
and `~/.moai/worktrees/` paths;

**When** a user runs `moai cc -w <abs-path-NOT-under-either-prefix>`;

**Then** the launcher rejects the path with a clear error message (no
silent fall-through to creating a new worktree under either prefix).

**Verification command**:
```bash
go test ./internal/cli/... -run 'TestNormalizeWorktreeFlag.*Reject\|TestLauncherWorktreeReject' -v 2>&1 | tail -20
# expected: NEW test case asserting an out-of-prefix absolute path is
#           rejected with a non-nil error — PASS
```

## §C. Edge Cases

- **Edge-1**: An L2 worktree is created at a non-standard path (not under
  `~/.moai/worktrees/`). The auto-isolation procedure MUST still detect the
  branch conflict via the active-sessions registry, not via path-prefix
  matching.
- **Edge-2**: The active-sessions registry is stale (recorded PID has exited).
  Per the Decision Point 1 Resolution OQ-2 conservative predicate (FIRM),
  the procedure fires anyway (false-positive-tolerant). The info log notes
  the registry entry's age. A 30-day false-positive audit (§F Forward-
  Looking Checks) MAY motivate a follow-up SPEC tightening the predicate.
- **Edge-3**: Multiple foreign sessions are on the same branch simultaneously
  (≥2). The procedure auto-creates N worktrees, one per session.
- **Edge-4**: The branch-guard exemption (`MOAI_BRANCH_GUARD_EXEMPT=1`) is
  set. The procedure does NOT need the exemption (it operates in worktree
  paths); the exemption is reserved for `manager-git` Late-Branch closure.

## §D. Quality Gate Criteria (Definition of Done)

- All MUST ACs PASS (AC-WES-001 through AC-WES-008, plus the new
  AC-WES-010a/b/c added per Decision Point 1 Resolution OQ-1); AC-WES-009
  SHOULD PASS.
- `go build ./...` exit 0 on linux/darwin; `GOOS=windows GOARCH=amd64 go
  build ./...` exit 0.
- `go test ./internal/config/... -run TestNewDefaultWorkflowConfig` green.
- `golangci-lint run --timeout=2m` — no NEW findings vs baseline.
- Subagent-boundary grep: 0 matches for `AskUserQuestion` in `internal/cli/`
  non-test code (C-HRA-008 preserved).
- Sanitized-pair mirror parity: template-neutrality CI green after `make
  build`.
- SPEC frontmatter `status: completed` carried by the single sync commit
  per the 3-phase close contract.

## §E. Indirect Verification (cross-cutting)

- **Branch-guard sibling compatibility**: confirmed via research.md §E.3
  (worktree paths are exempt from the deny); no runtime test needed because
  the discriminant is unchanged.
- **`normalizeWorktreeFlag` token-normalization behavior preserved**: the
  short-name token-normalization behavior (rewriting `-w` / `--worktree[=name]`
  into the canonical `--worktree [name]` form for claude pass-through) is
  preserved for non-L2-absolute-path inputs. Per REQ-WES-010 / AC-WES-010b,
  the L2 absolute-path extension MAY be implemented as an inlined
  early-return branch within `normalizeWorktreeFlag` (in which case the
  function body gains lines) OR as a sibling pre-resolution function (in
  which case `normalizeWorktreeFlag` itself is byte-identical). The
  preserved-behavior assertion is verified by pre-existing launcher tests
  remaining green — NOT by `git diff` byte-identity of the function body.
- **i18n key completeness**: confirmed via research.md §B.3 (the 3 toggle
  key pairs exist in all 4 locales; no new keys added).

## §F. Forward-Looking Checks (post-merge audit)

- After 30 days, re-run `git worktree list | wc -l` and compare to the
  2026-07-28 baseline (58). The expectation is a decrease or stabilization,
  NOT continued growth, attributable to the EnterWorktree-first policy +
  web defaults OFF.
- After 30 days, audit the active-sessions registry false-positive rate
  (how many auto-isolation events fired against genuinely-active foreign
  sessions vs stale entries). If false-positive rate > 50%, a follow-up
  SPEC should tighten the predicate (OQ-2 strict mode).

## §G. Closure Gates

- Implementation Kickoff Approval (plan→run HUMAN GATE) — MANDATORY per
  CLAUDE.local.md §19.1 + spec-workflow.md; NOT bypassed by plan-auditor
  PASS or skip-eligibility.
- Decision Point 1 Resolutions (formerly the 4 open-question items in
  plan.md §I) are FIRM as of 2026-07-28 — recorded in spec.md §B.2 / §C
  REQ-WES-005 / REQ-WES-010 / §F Out of Scope / HISTORY 0.2.0, plan.md
  §E M1 / M3 / M3a / M6 / §I, and this acceptance.md §A matrix. No
  further AskUserQuestion round is required for Decision Point 1 before
  M1 commit.
- plan-auditor verdict ≥ 0.85 (Tier L threshold) before proceeding to run.
