# Wave 2 Execution Strategy (Phase 1 Output)

> Audit trail. manager-strategy output for Wave 2 of SPEC-V3R3-CI-AUTONOMY-001.
> Generated: 2026-05-06. Methodology: TDD (RED-GREEN-REFACTOR). Mode: Standard.
> Base: origin/main 0b028bfaa (Wave 1 merged via PR #785 in commit 79f313c4a).

## Architecture Decisions

### Skill Structure (OQ1 RESOLVED — new skill, not sync.md extension)

New skill `moai-workflow-ci-watch` lives at:

```
internal/template/templates/.claude/skills/moai-workflow-ci-watch/
├── SKILL.md                       # Progressive Disclosure (Quick / Implementation / Advanced)
└── modules/
    ├── ci-watch-protocol.md       # When to invoke, timeouts, abort conditions
    └── trigger-handoff.md         # T3 auto-fix metadata schema
```

Mirrored at user-facing path under `.claude/skills/moai-workflow-ci-watch/` after `make build` (Template-First).

Companion shell layer at repo root:

```
scripts/ci-watch/
├── run.sh                         # gh pr checks polling + classification
├── lib/
│   ├── _common.sh                 # Helpers: log_step, status_dump, posix_now (reuse Wave 1 patterns)
│   ├── classify.sh                # required vs auxiliary discrimination via .github/required-checks.yml
│   └── timeout.sh                 # 30-min wall-clock guard + state file management
└── test/
    └── run_test.sh                # Bash test harness with mock gh fixtures
```

### Why `gh pr checks --watch` (not raw API polling)

- `gh pr checks <PR> --watch --json status,conclusion,name,workflow,detailsUrl` returns structured JSON every iteration
- `--watch` blocks until all checks reach terminal state (completed/failure/cancelled/timed_out) — no manual polling loop required for happy path
- Auth and rate limiting handled by `gh` CLI — no token storage in MoAI scripts
- Output stable since `gh` v2.20+ (Q3 2023); we already require `gh` >= 2.40 for Wave 1 branch protection (`gh api -X PUT`)
- Fallback: when `--watch` unavailable, fall back to `gh pr checks <PR> --json ...` polled every 30s by orchestrator

**Honest scope checkpoint (gh CLI version stability)**: the `--json` field set is stable across gh 2.40-2.65, but `workflow` field added in 2.50. We pin `gh >= 2.50` in `moai doctor` for Wave 2; lower versions fall back to name-based heuristics. Documented as soft-blocker in T2 protocol.

### Polling Schema

Top-level state machine (orchestrator side):

```
[idle] → [arm-watch] → [watching] → ┬─→ [all-required-pass] → emit ready-to-merge → [idle]
                                    ├─→ [required-fail] → capture metadata → handoff to T3 → [idle]
                                    ├─→ [auxiliary-fail-only] → emit advisory → [all-required-pass]
                                    └─→ [timeout-30m] → emit blocker → [idle]
```

Each tick (30s while in [watching]):
1. `gh pr checks <PR> --json status,conclusion,name,workflow,detailsUrl` → parse
2. Filter rows by `name` against `.github/required-checks.yml` `required:` list (Wave 1 SSoT)
3. Compute aggregate state: `required_pending`, `required_failed`, `auxiliary_failed`
4. Emit natural-language summary to orchestrator transcript ("Required: 4/6 pass, 2 in_progress; Advisory: 1 fail (claude-code-review)")
5. Update `.moai/state/ci-watch-active.flag` heartbeat (mtime touch) for crash recovery

### State File Location

`.moai/state/ci-watch-active.flag` (gitignored via existing `.moai/state/` rule):

```yaml
# Single-line YAML for atomic write
pr_number: 785
started_at: "2026-05-06T08:30:00Z"
heartbeat_at: "2026-05-06T08:35:00Z"
required_checks: [Lint, "Test (ubuntu-latest)", "Test (macos-latest)", "Test (windows-latest)", "Build (linux/amd64)", CodeQL]
abort_requested: false
```

**Race condition mitigation (honest scope checkpoint)**: only one watch loop per repo at a time (single PR sync model). On concurrent invocation, second invocation reads file, sees recent heartbeat (< 90s), aborts with "watch already active for PR=N, run `moai pr watch --abort` to release". File deleted on terminal state. No fcntl/flock — relies on heartbeat freshness + atomic rename via `tempfile + mv`. Acceptable because watch is orchestrator-driven, not concurrent multi-process.

### Required vs Auxiliary Discrimination

Single source of truth at `.github/required-checks.yml` (created in Wave 1 W1-T07). Wave 2 reuses without modification:

```yaml
version: 1
branches:
  main:
    contexts: [Lint, Test (ubuntu-latest), Test (macos-latest), Test (windows-latest), Build (linux/amd64), CodeQL]
auxiliary:
  - claude-code-review
  - llm-panel
  - docs-i18n-check
```

Loader: `internal/config/required_checks.go::LoadRequiredChecks()` (Wave 1) — Wave 2 adds `internal/config/required_checks.go::IsRequired(checkName, branchName) bool` helper if not already present.

## Module Topology

```
moai-adk/
├── internal/
│   ├── ciwatch/                        # NEW (Wave 2) — Go-side helpers
│   │   ├── state.go                    # state file read/write (atomic, heartbeat)
│   │   ├── state_test.go
│   │   ├── classifier.go               # IsRequired(check, branch) wrapper around required_checks.go
│   │   ├── classifier_test.go
│   │   ├── handoff.go                  # T3 metadata schema (FailedCheck struct, NewHandoff(...))
│   │   └── handoff_test.go
│   ├── config/
│   │   └── required_checks.go          # EXTEND (Wave 1 file) — add IsRequired helper
│   └── cli/
│       └── pr/
│           ├── watch.go                # NEW — `moai pr watch <PR>` and `moai pr watch --abort`
│           └── watch_test.go
├── scripts/
│   └── ci-watch/                       # NEW
│       ├── run.sh
│       ├── lib/_common.sh
│       ├── lib/classify.sh
│       ├── lib/timeout.sh
│       └── test/run_test.sh
└── internal/template/templates/
    ├── .claude/skills/moai-workflow-ci-watch/
    │   ├── SKILL.md
    │   └── modules/
    │       ├── ci-watch-protocol.md
    │       └── trigger-handoff.md
    └── .claude/rules/moai/workflow/
        └── ci-watch-protocol.md        # Mirror of skill module for global rule load
```

## Why-Skill-Not-Sync-Extension (OQ1 Resolution Rationale)

User selected "신규 skill `moai-workflow-ci-watch` 분리" (plan.md §7 OQ1). Reasoning:

1. **Concern separation**: `/moai sync` is one-shot (push + PR create + return). Watch is a long-running orchestrator-side loop. Bundling into sync.md would overload its Quick Reference and force `/moai sync` to remain "active" for up to 30 minutes — hostile to context budget.
2. **Independent disable/test**: skill can be selectively loaded only when watching is needed. sync.md extension would load watch metadata into every `/moai sync` invocation.
3. **Progressive Disclosure fit**: skill body (~5K tokens) at level 2 only loads when `/moai pr watch` is invoked or T2 trigger keyword detected. sync.md extension would inflate sync's level-2 budget unnecessarily.
4. **Reusability**: future SPECs can compose `moai-workflow-ci-watch` with non-sync workflows (e.g., manual `gh pr create` + watch).

## Per-Task TDD Breakdown (W2-T01 .. W2-T10)

### W2-T01 — Skill SKILL.md skeleton

- **RED**: `internal/template/templates/.claude/skills/moai-workflow-ci-watch/SKILL_test.go` (or `scripts/check-skill-frontmatter.sh` reuse) asserts: file exists, YAML frontmatter has `name`, `description`, `version`; SKILL.md body has Quick Reference / Implementation / Advanced sections; line count <= 500.
- **GREEN**: write SKILL.md with Progressive Disclosure 3-tier structure. Quick Reference = trigger conditions + `gh pr checks --watch` one-liner. Implementation = orchestrator state machine + protocol invocations. Advanced = timeout tuning + abort recovery.
- **REFACTOR**: extract two body sections into `modules/ci-watch-protocol.md` and `modules/trigger-handoff.md` if SKILL.md exceeds 500 lines. Cross-reference via "Detailed Reference: modules/<file>.md".

### W2-T02 — Wave 1 mirror script extraction validation

- **RED**: `scripts/ci-watch/test/run_test.sh::test_mirror_script_intact()` asserts that `scripts/ci-mirror/run.sh` (Wave 1) is callable as-is from Wave 2 watch path (no Wave 2 changes break Wave 1 contract). Smoke run with `MOAI_CI_LIB_DIR=/tmp/empty` should yield silent skip exit 0.
- **GREEN**: no code change — this is a contract-preservation test. Wave 1 already produces `scripts/ci-mirror/run.sh`. Test simply asserts it exists and conforms to dispatch interface.
- **REFACTOR**: skip — Wave 1 module ownership preserved.

### W2-T03 — `scripts/ci-watch/run.sh` polling

- **RED**: `scripts/ci-watch/test/run_test.sh::test_polling_loop()` 4 cases:
  1. mock `gh pr checks` returns all-pass → run.sh exits 0
  2. mock returns required-fail → run.sh emits handoff JSON to stdout, exits 2
  3. mock returns auxiliary-only-fail → run.sh emits "advisory" line, exits 0
  4. mock returns mixed (required pending) → run.sh continues looping, sleeps 30s
- **GREEN**: implement run.sh with `gh pr checks <PR> --watch --json status,conclusion,name` loop. Per-tick parse via `jq`. Mock injection via `MOAI_CIWATCH_GH=/path/to/fake-gh`.
- **REFACTOR**: extract classification call to `lib/classify.sh`; extract timeout management to `lib/timeout.sh`.

### W2-T04 — Required-checks SSoT loading

- **RED**: `internal/ciwatch/classifier_test.go::TestIsRequired_TableDriven` 6 cases (main branch, release/* branch, auxiliary, unknown name, empty SSoT, missing file). `scripts/ci-watch/test/test_classify.sh` parallel test for shell consumer (uses yq or jq path-based extraction).
- **GREEN**: `internal/ciwatch/classifier.go::IsRequired(check, branch string) bool` reads `.github/required-checks.yml` via Wave 1 `LoadRequiredChecks`. Shell-side `lib/classify.sh::is_required <check> <branch>` reads same YAML via `yq` (fallback to grep heuristic if yq missing — soft-required).
- **REFACTOR**: cache YAML load per process (single-read pattern).

### W2-T05 — 30s status report formatter

- **RED**: `internal/ciwatch/handoff_test.go::TestFormatStatusUpdate` produces deterministic natural-language output for 3 scenarios (all pending, partial pass, fail-imminent). Asserts no ANSI escape codes (orchestrator chat output), max 200 chars per line.
- **GREEN**: `FormatStatusUpdate(state CIState) string` returns single-line summary like `[ci-watch] PR #785: required 4/6 pass, 2 pending; advisory 0 fail`. Avoid emojis (CLAUDE.md no-emoji rule).
- **REFACTOR**: ensure formatter is pure-function (no I/O) for testability.

### W2-T06 — SKILL.md Implementation/Advanced sections

- **RED**: skill linter (Wave 1 W1 frontmatter check + structural conformance) validates that Implementation section references the protocol module and Advanced section references abort + timeout protocols.
- **GREEN**: write Implementation (300-400 lines) covering orchestrator-side state machine, AskUserQuestion handoff for ready-to-merge, abort path. Advanced (100-200 lines) covers cross-PR concurrent invocation, gh version compat fallback.
- **REFACTOR**: split into modules if Quick+Impl+Adv > 500 lines.

### W2-T07 — `ci-watch-protocol.md` rule file

- **RED**: `internal/template/templates/.claude/rules/moai/workflow/ci-watch-protocol.md` existence + frontmatter schema check (paths frontmatter for auto-load on `/moai sync` and `/moai pr watch`).
- **GREEN**: write rule file with [HARD]/[WARN] markers. Key rules: when to invoke (after `/moai sync` PR create), 30s polling cadence, 30-min hard timeout, abort flag honoring, T3 handoff format.
- **REFACTOR**: cross-reference Wave 1 W1-T07 SSoT and Wave 3 (T3) trigger metadata.

### W2-T08 — Ready-to-merge AskUserQuestion trigger

- **RED**: `internal/cli/pr/watch_test.go::TestReadyToMergeFlow` asserts that on all-pass, the CLI emits a structured handoff document (markdown) suitable for orchestrator AskUserQuestion presentation. Format: PR title, check summary, recommended action with `(권장)` first option.
- **GREEN**: `internal/cli/pr/watch.go::EmitReadyToMergeReport(state CIState) error` writes report to stdout. Orchestrator (skill body) reads stdout and constructs AskUserQuestion. CLI itself does NOT call AskUserQuestion (HARD: orchestrator-only per `.claude/rules/moai/core/askuser-protocol.md`).
- **REFACTOR**: report writer is pure I/O — extract format helpers to `handoff.go`.

### W2-T09 — Failure metadata capture for T3 handoff

- **RED**: `internal/ciwatch/handoff_test.go::TestNewHandoff` 3 cases:
  1. Single required failure → handoff struct populated with run-id, log URL, check name
  2. Multiple required failures → all captured in slice
  3. Auxiliary fails ignored (not in handoff)
- **GREEN**: `internal/ciwatch/handoff.go` defines `type FailedCheck struct { Name, RunID, LogURL, ConclusionDetail string }` and `NewHandoff(state CIState) Handoff`. Serializable to JSON for piping to T3 (Wave 3) `expert-debug` invocation prompt.
- **REFACTOR**: ensure JSON shape stable for T3 consumer (forward compatibility).

### W2-T10 — Timeout + abort path

- **RED**: `internal/ciwatch/state_test.go::TestStateFile_HeartbeatStale` (heartbeat older than 90s → considered stale → new watch may proceed). `TestStateFile_AbortFlag` (file with `abort_requested: true` → run.sh exits 0 with abort confirmation). `scripts/ci-watch/test/test_timeout.sh::test_30min_hard_stop`.
- **GREEN**: state.go atomic write via tempfile + rename. Abort flag via `moai pr watch --abort` writing `abort_requested: true`. 30-min wall-clock guard in run.sh main loop using `posix_now` deltas.
- **REFACTOR**: factor heartbeat update into `state.go::Touch()` for periodic refresh.

## Risk Mitigations (R-CIAUT-2 + Cross-Cutting)

| Risk | Source | Mitigation in Wave 2 |
|------|--------|---------------------|
| R-CIAUT-2 token consumption from long polling | spec.md §8 | 30s polling interval; 30-min hard timeout; orchestrator state-machine emits status only on state-change ticks (not every 30s); abort path returns control to user |
| gh CLI `--json` field churn | gh release notes | Pin `gh >= 2.50`; fallback heuristics in `lib/classify.sh` when `workflow` field missing |
| Concurrent watch invocations corrupting state file | implementation | Atomic write via tempfile+rename; heartbeat freshness check (< 90s = active); explicit `--abort` semantics |
| User Ctrl+C leaves stale flag | UX | Heartbeat staleness reclaim (next invocation auto-clears > 90s old flag); `moai pr watch --abort` for explicit cleanup |
| Auxiliary check intermittent flakiness blocks ready-to-merge incorrectly | T2 contract | Strict separation: auxiliary failures emit advisory line but DO NOT block ready-to-merge; verified by AC-CIAUT-005 |
| Watch invoked on non-existent PR | edge case | First `gh pr checks` call returns error; run.sh exits 1 with diagnostic, no state file written |
| `gh` not authenticated | Wave 1 R3 echo | Reuse Wave 1 graceful exit pattern; print exact `gh auth login` remediation |

## Implementation Order Rationale

1. **W2-T04** (classifier) first — pure-Go, no deps, validates Wave 1 SSoT loader integration; smallest blast radius
2. **W2-T09** (handoff schema) — defines T3 forward contract early; small struct + tests
3. **W2-T05** (status formatter) — pure function, isolated
4. **W2-T10** (state file + abort) — Go-side state mgmt
5. **W2-T03** (run.sh polling) — composes T04 (classifier) + T10 (state); shell-side
6. **W2-T08** (ready-to-merge report) — depends on T05 (formatter) + T09 (handoff)
7. **W2-T01** (SKILL.md skeleton) — frontmatter + structure
8. **W2-T06** (SKILL.md body) — depends on T01 + concrete behavior from T03/T08
9. **W2-T07** (protocol rule) — depends on T06; canonical rule reference
10. **W2-T02** (Wave 1 mirror contract test) — final smoke test; validates no Wave 1 regression

## Commit Pacing (Phase 3 manager-git target — 8 commits)

- **C1 (W2-T04)**: `feat(ciwatch): add IsRequired classifier with SSoT integration`
- **C2 (W2-T09)**: `feat(ciwatch): define FailedCheck handoff schema for T3 consumer`
- **C3 (W2-T05 + W2-T10)**: `feat(ciwatch): add state file + status formatter primitives`
- **C4 (W2-T03)**: `feat(ci-watch): add run.sh polling loop with classification dispatch`
- **C5 (W2-T08)**: `feat(cli): add 'moai pr watch' command and ready-to-merge report`
- **C6 (W2-T01 + W2-T06)**: `feat(skills): add moai-workflow-ci-watch skill (3-tier disclosure)`
- **C7 (W2-T07)**: `feat(rules): add ci-watch-protocol rule with HARD invocation contract`
- **C8 (W2-T02)**: `test(ci-mirror): add Wave 1 contract preservation smoke test + CHANGELOG`

Each commit: Conventional Commits + 🗿 MoAI co-author trailer + green `make ci-local`.

## Honest Scope Checkpoints

1. **gh CLI `--watch` JSON output stability**: confirmed stable from gh 2.50+. Pin via `moai doctor` dependency check. Fallback path documented (heuristic name-matching when `workflow` field absent). **Status: mitigated, not blocker.**
2. **`.moai/state/ci-watch-active.flag` race conditions**: single-PR-at-a-time model + heartbeat staleness + atomic rename = sufficient for orchestrator-driven serial use. Multi-process concurrent watch is OUT OF SCOPE for Wave 2 (would require fcntl, deferred to follow-up SPEC). **Status: documented limitation, not blocker.**
3. **30-min hard timeout vs slow CI**: most MoAI CI runs complete in 8-15 min. 30 min covers tail end + safety margin. If users hit this regularly, T2 follow-up can extend via config knob. **Status: empirical default, revisit after first month.**
4. **AskUserQuestion-only-by-orchestrator**: Wave 2 CLI (`moai pr watch`) emits structured stdout reports for orchestrator consumption. CLI itself NEVER calls AskUserQuestion. Skill body's protocol is the orchestrator surface. **Status: explicit design constraint, no blocker.**

No hard blockers identified. Proceed to Phase 2 (manager-tdd delegation).
