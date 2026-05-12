# plan.md — SPEC-V3R4-STATUS-LIFECYCLE-001

Implementation plan for the 7-Layer Defense in Depth solution to chronic SPEC status drift.

---

## §0. Motivation (Catalyst: PR #871)

Pull Request #871, opened 2026-05-13, is the latest in a recurring pattern of "chore(spec): status drift sync" retrofit PRs. Five such retrofits have merged or opened in the four days preceding this plan. PR #871's own body explicitly identifies the root cause and names this SPEC (`SPEC-V3R4-STATUS-LIFECYCLE-001`) as the planned systemic remedy. This plan operationalizes that proposal.

The drift is not caused by a single bug. It is caused by **seven independent system failures stacking on top of each other**:

```
L1 Standard       ─┐
L2 Trigger        ─┤
L3 Lint           ─┤    Compound failure surface.
L4 CI Gate        ─┼──> Any one alone would be tolerable.
L5 Format         ─┤    All seven together produce predictable drift.
L6 Visibility     ─┤
L7 Ownership      ─┘
```

Fixing any single layer leaves the others to propagate the same failure. The plan accordingly addresses all seven layers in three waves.

---

## §1. Technical Approach

### 1.1 Defense-in-Depth Principle

Each layer is independently failing. Each layer is independently fixable. The plan assigns each layer to a specific Wave and Task so failures can be diagnosed and reverted in isolation.

### 1.2 Wave Sequencing Rationale

Wave 1 (Policy + Lint) ships first because:
- Zero behavioural changes to existing tooling.
- Foundation for Waves 2 and 3 (transition table, canonical enum).
- Reversible via `git revert` without state surgery.

Wave 2 (Hook + Transitions) ships second because:
- Adds the local-dev trigger using Wave 1's transition table.
- Hook is opt-in by default — does not affect CI.

Wave 3 (Automation + Visibility) ships last because:
- CI auto-commits to main are highest blast radius.
- Drift CLI requires Wave 2 transition table.
- 30-day monitoring window starts on Wave 3 merge.

### 1.3 Hook + CI Convergence

Both the PostToolUse hook (`internal/hook/spec_status.go`) and the GitHub Actions workflow (`.github/workflows/spec-status-auto-sync.yml`) read from the same transition table (`internal/spec/transitions.go`, new file). When both fire on the same PR, the second run is idempotent: `updateStatusInContent` returns the input unchanged when the status field is already at the target value.

This dual-trigger design accepts that:
- Local `gh pr merge` invocations fire the hook (best-effort).
- Web UI merges, merge queue, and Dependabot auto-merge do NOT fire the hook.
- Only the CI workflow is authoritative for all merge paths.

---

## §2. Milestones (Priority-based, no time estimates)

### Wave 1 — Policy + Lint (Priority: High)

**Acceptance gate**: All Wave 1 tasks pass acceptance.md criteria AC-STATUS-LIFECYCLE-001-01.* before Wave 2 begins.

| Task | Description | Files Modified | REQ Coverage |
|------|-------------|----------------|--------------|
| W1-T1 | Write canonical 8-value enum to `.claude/rules/moai/workflow/spec-workflow.md` § Status Lifecycle (new section) | `.claude/rules/moai/workflow/spec-workflow.md` (+) | REQ-1 |
| W1-T2 | Add `StatusValueEnumRule` and `StatusCaseNormalizationRule` to `internal/spec/lint.go`; register both in lint rule slice (line 118) | `internal/spec/lint.go`, `internal/spec/lint_test.go` (+) | REQ-3.1, REQ-3.2 |
| W1-T3 | Backfill 21 legacy SPECs to YAML frontmatter using `moai spec status` (dry-run first; user reviews dry-run output before write) | 21 `spec.md` files in `.moai/specs/SPEC-*/` | REQ-6 |
| W1-T4 | Add `.github/workflows/spec-lint.yml` workflow running `moai spec lint --strict` on PR events; mark as `required` after Wave 1 merge stabilizes | `.github/workflows/spec-lint.yml` (new) | REQ-4 |

**Wave 1 risks**:
- W1-T3 may surface a SPEC whose legacy status is semantically wrong (e.g., listed Planned but implemented). Mitigation: dry-run review before write; user explicitly approves each line.
- W1-T2 may flag pre-existing SPECs not in REQ-6 backfill list. Mitigation: lint runs in non-strict mode initially; CI gate is `--strict` opt-in.

**Wave 1 exit criteria**: `moai spec lint --strict` passes on all 180+ SPECs. CI workflow `spec-lint.yml` is green on a test PR.

---

### Wave 2 — Hook + Transitions (Priority: High)

**Acceptance gate**: All Wave 2 tasks pass acceptance.md criteria AC-STATUS-LIFECYCLE-001-02.* before Wave 3 begins.

| Task | Description | Files Modified | REQ Coverage |
|------|-------------|----------------|--------------|
| W2-T1 | Create `internal/spec/transitions.go` with `PrefixToStatus` map; extend `internal/hook/spec_status.go` to call `isGhPrMergeCommand` + `ClassifyPRTitle` | `internal/spec/transitions.go` (new), `internal/hook/spec_status.go` | REQ-2 |
| W2-T2 | Add unit tests covering 4 prefix patterns (`plan(`, `feat(SPEC-X)`, `fix(SPEC-X)`, `docs(sync)`) + revert prefix + ambiguous case | `internal/spec/transitions_test.go` (new), `internal/hook/spec_status_test.go` | REQ-2 |
| W2-T3 | Update `.claude/skills/moai/workflows/plan.md:454` and `.claude/skills/moai/workflows/sync.md` (5 mixed-enum lines) to reference canonical 8-value enum from W1-T1 | `internal/template/templates/.claude/skills/moai/workflows/plan.md`, `sync.md` | REQ-1 |

**Wave 2 risks**:
- Hook PostToolUse may detect `gh pr merge` but the PR title is not yet on the closed-PR record (race condition with `gh pr view`). Mitigation: hook reads title from the local PR ref before invoking `gh pr merge` — title is stable pre-merge.
- W2-T3 template edit requires `make build` to regenerate embedded files. Mitigation: include `make build` in the W2-T3 task checklist.

**Wave 2 exit criteria**: Hook + Actions both consume `transitions.go`. Unit tests >= 90% coverage for transitions logic. Template embedded files regenerated.

---

### Wave 3 — Automation + Visibility (Priority: Medium)

**Acceptance gate**: All Wave 3 tasks pass acceptance.md criteria AC-STATUS-LIFECYCLE-001-03.* before 30-day monitoring begins.

| Task | Description | Files Modified | REQ Coverage |
|------|-------------|----------------|--------------|
| W3-T1 | Add `.github/workflows/spec-status-auto-sync.yml` workflow on `pull_request.closed` (merged == true); auto-commit follow-up with `chore(spec): auto-sync status for #<PR>` title and loop-prevention check | `.github/workflows/spec-status-auto-sync.yml` (new) | REQ-5 |
| W3-T2 | Add `moai spec drift` subcommand in `internal/cli/spec_status.go`; reuse `getSPECIDsFromGitLog()` for cross-reference; produce tabular report + `--json` + `--exit-code-on-drift` flags | `internal/cli/spec_status.go`, `internal/cli/spec_status_test.go` | REQ-7 |
| W3-T3 | Integrate `moai spec drift --count` with SessionStart hook; emit one-line warning if count >= 5 | `internal/hook/session_start.go` (or equivalent SessionStart handler) | REQ-8 |
| W3-T4 | Add `StatusGitConsistencyRule` to `internal/spec/lint.go`; reuse W3-T2 git scan; warning by default, error under `--strict` | `internal/spec/lint.go`, `internal/spec/lint_test.go` | REQ-3.3 |
| W3-T5 | Update `manager-spec.md`, `manager-docs.md`, `manager-develop.md` agent definitions with explicit Responsibility Matrix entries | `internal/template/templates/.claude/agents/manager-spec.md`, `manager-docs.md`, `manager-develop.md` | REQ-9 |
| W3-T6 | Document 30-day monitoring period (M-1/M-2/M-3 criteria from spec.md §7) in this SPEC's `acceptance.md` Definition of Done | `acceptance.md` (existing — already includes §7 reference) | (Verification mechanism) |

**Wave 3 risks**:
- R-2 (auto-sync follow-up commit collides with concurrent PR): mitigated by rebase-on-conflict retry + issue-creation fallback (REQ-5 spec.md §3).
- R-4 (web-UI merges miss hook): accepted — CI workflow is authoritative; hook is best-effort secondary.
- R-5 (30-day monitoring not enforceable mechanically): accepted — future audit SPEC may automate.

**Wave 3 exit criteria**: Both new workflows are green on a test PR. `moai spec drift` returns aligned/drift counts within 2 seconds on the 180+ SPEC repo. Agent Responsibility Matrix entries are visible in agent body markdown.

---

## §3. File Dependencies

```
spec-workflow.md  ◄────── W1-T1 (canonical enum)
       │
       ▼
internal/spec/lint.go  ◄── W1-T2 (StatusValueEnum + StatusCase) ── W3-T4 (StatusGitConsistency)
       │
       ▼
internal/spec/transitions.go (new)  ◄── W2-T1 ── W2-T2 (tests)
       │
       ├──> internal/hook/spec_status.go  (W2-T1 extension)
       │
       └──> .github/workflows/spec-status-auto-sync.yml  (W3-T1)

internal/cli/spec_status.go  ◄── W3-T2 (moai spec drift)
       │
       └──> internal/hook/session_start.go  (W3-T3 integration)

21 legacy SPECs  ◄── W1-T3 (backfill via moai spec status)

.github/workflows/spec-lint.yml (new)  ◄── W1-T4 (CI gate)

manager-spec.md / manager-docs.md / manager-develop.md  ◄── W3-T5 (Responsibility Matrix)
```

---

## §4. Quality Strategy

### 4.1 Testing

| Wave | Coverage Target | Test Categories |
|------|-----------------|-----------------|
| Wave 1 | 90%+ for new lint rules | Unit (rule.Check), integration (full SPEC dir scan), CLI smoke test |
| Wave 2 | 90%+ for transitions.go | Unit (ClassifyPRTitle), integration (hook handle), regression (existing commit trigger still works) |
| Wave 3 | 85%+ for moai spec drift; smoke test only for CI workflow | Unit (drift classifier), integration (drift output), end-to-end test via workflow dispatch |

### 4.2 TRUST 5 Mapping

- **Tested**: Each Wave includes explicit test tasks (W1-T2, W2-T2, W3-T2 test files).
- **Readable**: Transition table is a single map literal in `transitions.go`; no scattered if-else chains.
- **Unified**: Hook and CI consume the same transition table — single source of truth for prefix → status mapping.
- **Secured**: REQ-5 auto-commit uses `moai-bot[bot]` author and `[skip ci]` marker to prevent recursive workflow invocation.
- **Trackable**: Every auto-sync commit carries the source PR number in its title: `chore(spec): auto-sync status for #<PR>`.

### 4.3 LSP Quality Gates

Wave 1 / Wave 2 / Wave 3 each must pass the project's standard LSP gate (`zero errors, zero type errors, zero lint errors`) before merge. Lint additions in W1-T2 and W3-T4 must not break the existing lint suite — verified by W1-T2 acceptance criterion AC-STATUS-LIFECYCLE-001-01.b (lint --strict passes on all 180+ SPECs after backfill).

---

## §5. Worktree Strategy

Per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline:

- **Step 1a (this plan PR)**: plan-in-main. Branch `plan/SPEC-V3R4-STATUS-LIFECYCLE-001`. No worktree. Squash merge.
- **Step 2 (run)**: Each Wave creates a fresh SPEC worktree from plan-merged main HEAD via `moai worktree new SPEC-V3R4-STATUS-LIFECYCLE-001 --base origin/main`. Branch `feat/SPEC-V3R4-STATUS-LIFECYCLE-001-wave-<N>`. Squash merge per Wave.
- **Step 3 (sync)**: Continues in the same worktree as Step 2's last Wave. Branch `sync/SPEC-V3R4-STATUS-LIFECYCLE-001`. Squash merge.
- **Step 4 (cleanup)**: After BOTH run AND sync PRs are merged, `moai worktree done SPEC-V3R4-STATUS-LIFECYCLE-001` from the host checkout.

---

## §6. Conventional Commits

Each Wave produces one squash-merged PR:

| Wave | PR title prefix | Example |
|------|-----------------|---------|
| Plan (this) | `plan(spec):` | `plan(spec): SPEC-V3R4-STATUS-LIFECYCLE-001 — Status Lifecycle Automation (7-Layer Defense)` |
| Wave 1 | `feat(spec):` | `feat(spec): SPEC-V3R4-STATUS-LIFECYCLE-001 Wave 1 — Policy + Lint` |
| Wave 2 | `feat(spec):` | `feat(spec): SPEC-V3R4-STATUS-LIFECYCLE-001 Wave 2 — Hook + Transitions` |
| Wave 3 | `feat(spec):` | `feat(spec): SPEC-V3R4-STATUS-LIFECYCLE-001 Wave 3 — Automation + Visibility` |
| Sync | `docs(sync):` | `docs(sync): SPEC-V3R4-STATUS-LIFECYCLE-001 status=completed + 30-day monitoring entry` |

Each PR body includes:
- Summary of Wave deliverables (2-3 bullets)
- Acceptance gate matrix (W*-T* → AC-STATUS-LIFECYCLE-001-*)
- Test plan checklist
- Reference to PR #871 catalyst (Wave 1 only)
- `🗿 MoAI <email@mo.ai.kr>` footer

---

## §7. Plan Audit Pre-Check (for plan-auditor)

The plan-auditor (if invoked per harness level) should verify:

- [P-1] spec.md §3 contains 10 REQs (REQ-1 through REQ-10) — verified.
- [P-2] Every REQ uses EARS keyword (SHALL, WHEN, IF) — verified (Ubiquitous and Event-Driven patterns).
- [P-3] spec.md §6 Out of Scope contains at least one exclusion — verified (7 exclusions).
- [P-4] Each Wave has at least one task per REQ in its scope — verified in §2 above.
- [P-5] Acceptance criteria (acceptance.md) are hierarchical per SPC-001 schema and depth >= 2 — verified.
- [P-6] research.md grounds every claim in spec.md with file:line evidence — verified (§7 references in research.md).
- [P-7] No timeline estimates ("days", "weeks") — verified (priority labels only).
- [P-8] Plan-in-main discipline followed (branch `plan/SPEC-V3R4-STATUS-LIFECYCLE-001`, no worktree) — verified.
- [P-9] Frontmatter 9-field schema present (id, version, status, created_at, updated_at, author, priority, labels, issue_number) — verified.
- [P-10] SPEC ID matches `^SPEC-[A-Z][A-Z0-9]+-[A-Z]{2,5}-\d{3}$` — verified (`SPEC-V3R4-STATUS-LIFECYCLE-001`).

If plan-auditor is not in scope for this harness level, this section serves as the plan-author's self-check.

---

## §8. Rollback Plan

Each Wave is independently revertible:

- **Wave 1 revert**: `git revert <Wave-1-squash-commit>`. Restores 6-value enum, removes 2 lint rules, restores 21 legacy frontmatter formats. No state surgery.
- **Wave 2 revert**: `git revert <Wave-2-squash-commit>`. Hook reverts to commit-only trigger. transitions.go removed. CI in Wave 3 must be reverted first if Wave 3 already merged.
- **Wave 3 revert**: `git revert <Wave-3-squash-commit>`. Removes auto-sync workflow, drift CLI, SessionStart integration, StatusGitConsistencyRule, agent updates. Hook in Wave 2 remains functional as best-effort fallback.

No data loss in any revert path. SPEC frontmatter changes during operation are committed to main and survive revert.
