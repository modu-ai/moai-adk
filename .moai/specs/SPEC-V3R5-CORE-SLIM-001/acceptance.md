# Acceptance — SPEC-V3R5-CORE-SLIM-001

Test scenarios, verification matrix, and quality gate criteria for the V3R5 W2 LR-08 rule refinement + expert agent foundation skill symmetry.

## Verification Matrix

One row per AC leaf with concrete shell-verifiable probe and binary PASS/FAIL outcome.

| AC ID | Given (precondition) | When (probe command) | Then (expected outcome) | Pass/Fail decision |
|-------|---------------------|---------------------|------------------------|----|
| AC-CSLM-001.a | Track A rule fix applied | `./bin/moai agent lint --strict 2>&1 \| grep "LR-08" \| grep "moai-domain-backend" \| wc -l` | Output `0` | exact `0` → PASS |
| AC-CSLM-001.b | (same) | `./bin/moai agent lint --strict 2>&1 \| grep "LR-08" \| grep "moai-domain-database" \| wc -l` | Output `0` | exact `0` → PASS |
| AC-CSLM-001.c | (same) | `./bin/moai agent lint --strict 2>&1 \| grep "LR-08" \| grep "moai-domain-frontend" \| wc -l` | Output `0` | exact `0` → PASS |
| AC-CSLM-001.d | (same) | `./bin/moai agent lint --strict 2>&1 \| grep "LR-08" \| grep "moai-design-system" \| wc -l` | Output `0` | exact `0` → PASS |
| AC-CSLM-002 | Track B preload addition applied | `./bin/moai agent lint --strict 2>&1 \| grep "LR-08" \| grep "moai-foundation-quality" \| wc -l` | Output `0` | exact `0` → PASS |
| AC-CSLM-003.a | Track B source + mirror edits committed | `diff .claude/agents/moai/expert-backend.md internal/template/templates/.claude/agents/moai/expert-backend.md` | Empty output, exit 0 | empty + exit 0 → PASS |
| AC-CSLM-003.b | (same) | `diff .claude/agents/moai/expert-frontend.md internal/template/templates/.claude/agents/moai/expert-frontend.md` | Empty output, exit 0 | empty + exit 0 → PASS |
| AC-CSLM-003.c | (same) | `diff .claude/agents/moai/expert-refactoring.md internal/template/templates/.claude/agents/moai/expert-refactoring.md` | Empty output, exit 0 | empty + exit 0 → PASS |
| AC-CSLM-003.d | (same) | `diff .claude/agents/moai/expert-devops.md internal/template/templates/.claude/agents/moai/expert-devops.md` | Empty output, exit 0 | empty + exit 0 → PASS |
| AC-CSLM-004 | Mirror parity established | `go run internal/template/scripts/gen-catalog-hashes.go --all && go test ./internal/template/...` | Tool succeeds, tests PASS | exit 0 → PASS |
| AC-CSLM-005.a | Track A code + tests applied | `go test ./internal/cli/... -run "TestSkillPreloadDriftExemption_DomainSkills" -v` | Output contains `--- PASS` | PASS line present → PASS |
| AC-CSLM-005.b | (same) | `go test ./internal/cli/... -run "TestSkillPreloadDriftExemption_FoundationSkills" -v` | Output contains `--- PASS` | PASS line present → PASS |
| AC-CSLM-005.c | (same) | `go test ./internal/cli/... -run "TestSkillPreloadDriftExemption_WorkflowSkills" -v` | Output contains `--- PASS` | PASS line present → PASS |
| AC-CSLM-005.d | (same) | `go test ./internal/cli/... -run "TestSkillPreloadDriftExemption_EdgeCases" -v` | Output contains `--- PASS` | PASS line present → PASS |
| AC-CSLM-006 | Both Track A + Track B merged on `main` | `./bin/moai agent lint --strict 2>&1 \| grep "LR-08" \| wc -l` | Output `0` | exact `0` → PASS (primary metric) |
| AC-CSLM-007 | (same) | `./bin/moai agent lint --strict 2>&1 \| grep -E "^! \[LR-" \| grep -v "^! \[LR-08\]" \| wc -l` | Output **equals** pre-merge complement baseline = **0** (auto-tracks ALL LR rules except LR-08 without manual enumeration; baseline measured 2026-05-20 on `plan/SPEC-V3R5-CORE-SLIM-001`) | numeric == 0 → PASS |
| AC-CSLM-008 | spec.md §5 EC-6 documents the sentinel | `grep -c "MIRROR_MISSING_BLOCKER" .moai/specs/SPEC-V3R5-CORE-SLIM-001/spec.md` | Output ≥ 1 | numeric ≥ 1 → PASS |

## Manual Test Scenarios

### Scenario 1 — Rule Fix Unit Tests (Track A, happy path)

**Given** the run-phase implementer has applied Track A changes to `internal/cli/agent_lint.go` (constant + helper + skip clause) and added 4 new test cases to `internal/cli/agent_lint_test.go`.

**When** the orchestrator runs:

```bash
go test ./internal/cli/... -run "TestSkillPreloadDrift" -v
```

**Then**:
- All 4 new test cases PASS:
  - `TestSkillPreloadDriftExemption_DomainSkills` PASS (exemption verified)
  - `TestSkillPreloadDriftExemption_FoundationSkills` PASS (enforcement verified)
  - `TestSkillPreloadDriftExemption_WorkflowSkills` PASS (enforcement verified)
  - `TestSkillPreloadDriftExemption_EdgeCases` PASS (4 sub-cases)
- All pre-existing tests in `agent_lint_test.go` continue to PASS (no regression).
- Final test summary line includes `PASS` and `ok` markers.

### Scenario 2 — Agent Edit + Lint (Track B, happy path)

**Given** Track B edits have been committed: `moai-foundation-quality` added to 4 source agents + 4 template mirrors + catalog refreshed.

**When** the orchestrator runs:

```bash
./bin/moai agent lint --strict 2>&1 | grep "LR-08" | grep "moai-foundation-quality" | wc -l
```

**Then**: Output is `0`. AC-CSLM-002 PASS.

### Scenario 3 — Combined Full Sweep (Tracks A + B, primary metric)

**Given** the user invokes `/moai run SPEC-V3R5-CORE-SLIM-001` on `feat/SPEC-V3R5-CORE-SLIM-001` branched from `plan/SPEC-V3R5-CORE-SLIM-001` after plan PR merges.

**When** the orchestrator delegates to `manager-develop` (`cycle_type` per `quality.yaml` `development_mode` setting), which:
1. Executes Track A (Go rule fix + tests)
2. Executes Track B (agent metadata edits + mirror sync + catalog refresh)
3. Runs `make build`
4. Verifies AC-CSLM-001 (all 4 leaves) PASS
5. Verifies AC-CSLM-002 PASS
6. Verifies AC-CSLM-003 (all 4 leaves) PASS
7. Verifies AC-CSLM-004 PASS
8. Verifies AC-CSLM-005 (all 4 leaves) PASS
9. Verifies AC-CSLM-006 PASS (primary)
10. Verifies AC-CSLM-007 PASS (orthogonal-gates check)

**Then**:
- 11 modified files staged: 1 Go src + 1 Go test + 4 source agents + 4 mirrors + 1 catalog
- `internal/template/embedded.go` regenerated (auto-generated, do not edit)
- Commit message: `feat(SPEC-V3R5-CORE-SLIM-001): W2 — LR-08 rule refinement + foundation-quality symmetry (LR-08: 12 → 0)`
- PR opened against `main`
- CI Tier 1 required checks pass: Lint, Test (ubuntu-latest), Build (linux/amd64), CodeQL
- PR squash-merged to `main`
- Post-merge lint verification (`./bin/moai agent lint --strict | grep LR-08 | wc -l`) returns `0`

### Scenario 4 — Non-Regression Check (Track A logic refactor)

**Given** Track A is applied but Track B is NOT yet applied (Order 1 mid-state).

**When** the orchestrator runs:

```bash
./bin/moai agent lint --strict 2>&1 | grep "LR-08" | wc -l
```

**Then**:
- Output is `2` (the 2 remaining `moai-foundation-quality` warnings — one for source path on expert-performance/security path-set, one for mirror path).

  > Note: The actual residual count depends on `moai-foundation-quality` only being preloaded by 2 of 6 expert agents (performance + security). LR-08 fires for both agents per the current rule logic — one warning per (agent file, missing skill) pair, separately for source vs mirror. The exact residual count after Track A alone may be 2 or 4 lines depending on how the rule attributes the drift (per research.md §3.1, the warning fires on the file that HAS the skill against the missing peers). Final dissolution requires Track B.

- This intermediate state confirms Track A correctly removes 10 of 12 LR-08 lines (the domain-prefix-mentioning ones).

### Scenario 5 — Idempotent Replay

**Given** the SPEC is merged and a second `/moai run SPEC-V3R5-CORE-SLIM-001` invocation is initiated (hypothetical replay).

**When** the orchestrator re-executes Track A + Track B steps.

**Then**:
- Track A: `isDomainExemptSkill` helper already exists; modifying `agent_lint.go` is a no-op or rejected (file unchanged).
- Track B: `grep -c "moai-foundation-quality" <agent-file>` returns ≥ 1 for all 4 target files; edit is no-op.
- Catalog refresh produces no diff in `internal/template/catalog.yaml` (idempotent).
- AC-CSLM-006 already at 0, no work needed.
- No PR created (no changes to commit).
- Orchestrator reports "SPEC already implemented, no action needed."

### Scenario 6 — Mirror Missing (EC-6 path)

**Given** between plan-phase and run-phase, an external process deletes one of the template mirrors (e.g., `internal/template/templates/.claude/agents/moai/expert-refactoring.md`).

**When** the run-phase Step B.3 attempts to mirror the source edit.

**Then**:
- The implementer (manager-develop subagent) detects the missing target file.
- Halts with sentinel `MIRROR_MISSING_BLOCKER`.
- Returns a structured blocker report to the orchestrator per `agent-common-protocol.md` §Blocker Report Format.
- Orchestrator presents the blocker to the user via AskUserQuestion: (a) re-run plan-phase to capture new baseline, (b) restore mirror from git, (c) abort SPEC.
- No partial edit is committed.

## CI / Lint Gates

The following gates MUST be green on the run-phase PR for the SPEC to be merge-eligible:

| Gate | Command | Required outcome |
|------|---------|------------------|
| `moai spec lint --strict` | `./bin/moai spec lint --strict` | ✓ exit 0, no SPEC frontmatter regression |
| LR-08 dissolution (primary) | `./bin/moai agent lint --strict 2>&1 \| grep "LR-08" \| wc -l` | `0` |
| Track A unit tests | `go test ./internal/cli/... -run "TestSkillPreloadDriftExemption"` | PASS for all 4 new test cases |
| Track A non-regression on existing tests | `go test ./internal/cli/...` | All pre-existing tests still PASS |
| `moai agent lint --strict` (complement-style — all LR rules except LR-08) | `./bin/moai agent lint --strict 2>&1 \| grep -E "^! \[LR-" \| grep -v "^! \[LR-08\]" \| wc -l` | == pre-merge baseline (= 0; auto-tracks ALL non-LR-08 LR rules without enumeration) |
| AC-CSLM-008 (EC-6 sentinel documented) | `grep -c "MIRROR_MISSING_BLOCKER" .moai/specs/SPEC-V3R5-CORE-SLIM-001/spec.md` | ≥ 1 |
| `golangci-lint` Tier 1 | `golangci-lint run ./...` | 0 new issues |
| `make build` | `make build` | success |
| `go test ./internal/template/...` (catalog hash audit) | `go test ./internal/template/...` | PASS |
| `go test ./...` (full suite) | `go test ./...` | PASS across all packages |
| Branch protection required checks | (CI workflow) | Lint, Test (ubuntu-latest), Build (linux/amd64), CodeQL all green per `.github/required-checks.yml` |

## Definition of Done

The SPEC is COMPLETE when all of the following are TRUE simultaneously on `main` HEAD:

- [ ] All 4 leaves of AC-CSLM-001 (001.a through 001.d) PASS — no LR-08 warning mentions any of the 4 domain-prefix skills (verified by Track A rule fix)
- [ ] AC-CSLM-002 PASS — no LR-08 warning mentions `moai-foundation-quality` (verified by Track B preload addition)
- [ ] All 4 leaves of AC-CSLM-003 (003.a through 003.d) PASS — source/mirror byte-identical diffs for all 4 modified agents
- [ ] AC-CSLM-004 PASS — catalog refreshed, template tests green
- [ ] All 4 leaves of AC-CSLM-005 (005.a through 005.d) PASS — new unit tests bidirectionally verify exemption + enforcement
- [ ] AC-CSLM-006 PASS (primary) — total LR-08 count across all categories = 0
- [ ] AC-CSLM-007 PASS — complement-style regex output equals pre-merge baseline (= 0; auto-tracks ALL LR rules except LR-08 without manual enumeration)
- [ ] AC-CSLM-008 PASS — `MIRROR_MISSING_BLOCKER` sentinel contract durably documented in spec.md §5 EC-6 (verified by `grep -c "MIRROR_MISSING_BLOCKER" spec.md` ≥ 1)
- [ ] Definition of Done from `.claude/rules/moai/workflow/spec-workflow.md` § Phase Transitions met (run PR merged into main + tests passing)
- [ ] Sync PR merged: HISTORY updated, version 0.3.0 → 0.4.0, status `implemented → completed`, project memory entry added (`project_v3r5_w2_core_slim_001_complete.md`)
- [ ] Memory `project_v3r5_w1_constitution_complete` entry receives `[SUPERSEDED by project_v3r5_w2_core_slim_001_complete]` prefix or is replaced per Lessons Protocol
- [ ] Sibling SPEC-V3R5-LINT-CLEAN-001 W2-deferred manifest at `.moai/state/lint-w2-deferred.json` is **deleted** (`rm .moai/state/lint-w2-deferred.json`) after AC-CSLM-006 PASS verified. The empirical lint output is the authoritative source of truth; the manifest is stale-by-construction per research.md §5.3 (composition mismatch between manifest and empirical baseline) and chicken-and-egg dissolution is complete only when the file is removed wholesale, not partially updated.

## Out of Acceptance Scope

The following are explicitly NOT verified by this acceptance.md (deferred or out of scope):

- Behavioral verification that `moai-foundation-quality` preload produces different agent runtime behavior on `expert-{backend,frontend,refactoring,devops}` (this is a metadata-only change for Track B; no behavior change is intended or measured).
- Performance benchmarks for skill preload time (negligible; sub-millisecond per skill).
- Cross-agent context-window-impact analysis (1 added skill per agent, no realistic overflow risk).
- LR-08 dissolution on non-expert categories beyond the natural exemption-list extension behavior covered by AC-CSLM-007.
- Future addition of new domain-skill prefixes beyond the canonical 6 listed in `domainExemptPrefixes`.
- Reverting Track A to restore strict uniform symmetry (alternative path NOT recommended; rejected at plan-phase).

## Cross-References

- spec.md §3 (EARS Requirements) — REQ-CSLM-001 through 005
- spec.md §4 (Acceptance Criteria EARS Hierarchical) — AC-CSLM-001 through 007
- plan.md §Track A + §Track B — implementation steps mapped to ACs
- research.md §3 — LR-08 implementation analysis (verbatim source)
- research.md §4 — EC-4 Discovery Log (v0.1.0 → v0.2.0 scope pivot rationale)
- `internal/cli/agent_lint.go:892-977` — LR-08 rule authoritative implementation (Track A target)
- `internal/cli/agent_lint_test.go` — existing test scaffolding (Track A append target)
- `.claude/rules/moai/workflow/spec-workflow.md` § Completion Markers — `<moai>DONE</moai>` and `<moai>COMPLETE</moai>` conventions
- `.github/PULL_REQUEST_TEMPLATE.md` — Merge Strategy checkbox (`--squash` for feat/* per CLAUDE.local.md §18.3)
