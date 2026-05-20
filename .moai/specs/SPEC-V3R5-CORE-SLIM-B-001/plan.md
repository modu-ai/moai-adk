# SPEC-V3R5-CORE-SLIM-B-001 — Implementation Plan

> **LEAN dogfooding (2nd cycle, post WORKFLOW-LEAN-001 PR #1030)**: This plan is Tier S. Section A-E template is OPTIONAL for run-phase delegations. plan-auditor PASS threshold for this SPEC: 0.75. Artifacts: spec.md + plan.md + progress.md (no acceptance.md / design.md / research.md).

## 1. Strategy Summary

**Tier**: S (LEAN — per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier and SPEC-V3R5-WORKFLOW-LEAN-001 precedent).

**Approach**: Pure delete + textual cross-reference cleanup. No new code, no abstraction changes, no behavioral changes. The 4 skills being retired have **0 workflow invocations and 0 agent frontmatter invocations** measured across the codebase (audit §3.2), so the change is mechanical and TRIVIAL-risk.

**Brownfield strategy**: **(b) Preserve all other skills; extend nothing.** This is pure delete + textual cleanup. No baseline extraction, no replacement skill, no behavioral substitute. The 4 retired skills are dead weight per audit §3.2 — there is nothing to preserve behavior of.

**Single-pass delegation**: Tier S minimal delegation prompt (~500-800 tokens per SPEC-V3R5-WORKFLOW-LEAN-001 Applicability clause). The full Section A-E 5-section template is OPTIONAL for Tier S. Goal / Deliverables / Constraints / Self-verification suffice.

**Estimated wall-time**: ~20-30 min run-phase (LEAN dogfooding 2nd cycle target).

**Estimated commits on main during implementation** (per LATE-BRANCH C-CSB-002):
- 1 plan commit (this commit, SPEC artifacts only)
- 1 M1 commit (4 skill directory deletes, template + local mirror, 8 directory removals)
- 1 M2 commit (4 language rule edits: elixir + csharp, local + template mirror)
- 1 M3 commit (3 `agents-reference.md` line cleanups in local + template mirror + `make build` regeneration of `embedded.go`)
- 1 status commit (`draft → implemented`, version `0.1.0 → 0.2.0`) at sync-phase

PR (Phase C of LATE-BRANCH) created via `git switch -c feat/SPEC-V3R5-CORE-SLIM-B-001` + `git push -u origin` + `gh pr create`. Squash merge target.

---

## 2. Milestones

Three milestones, executed sequentially. Each milestone is verified by its mapped ACs before proceeding.

### M1 — Retire 4 Skill Directories (Template + Local Mirror)

**Scope**: 8 directory deletes (4 skills × 2 paths each).

**Tasks**:

| # | Action | Path | REQ | AC |
|---|--------|------|-----|----|
| 1.1 | `rm -rf` | `internal/template/templates/.claude/skills/moai-framework-electron/` | REQ-CSB-001 | AC-CSB-001 |
| 1.2 | `rm -rf` | `.claude/skills/moai-framework-electron/` | REQ-CSB-001 | AC-CSB-001 |
| 1.3 | `rm -rf` | `internal/template/templates/.claude/skills/moai-platform-auth/` | REQ-CSB-002 | AC-CSB-002 |
| 1.4 | `rm -rf` | `.claude/skills/moai-platform-auth/` | REQ-CSB-002 | AC-CSB-002 |
| 1.5 | `rm -rf` | `internal/template/templates/.claude/skills/moai-platform-chrome-extension/` | REQ-CSB-003 | AC-CSB-003 |
| 1.6 | `rm -rf` | `.claude/skills/moai-platform-chrome-extension/` | REQ-CSB-003 | AC-CSB-003 |
| 1.7 | `rm -rf` | `internal/template/templates/.claude/skills/moai-platform-deployment/` | REQ-CSB-004 | AC-CSB-004 |
| 1.8 | `rm -rf` | `.claude/skills/moai-platform-deployment/` | REQ-CSB-004 | AC-CSB-004 |

**Verification**: AC-CSB-001..004 (4 binary `test ! -e ... && test ! -e ...` commands).

**Commit message**: `feat(SPEC-V3R5-CORE-SLIM-B-001): M1 — retire 4 Category B dead-weight skills (1,432 LOC)`

### M2 — Update 2 Language Rules + Their Template Mirrors (Remove `moai-platform-deployment` References)

**Scope**: 4 file edits (2 local + 2 template mirror). Each file has its `moai-platform-deployment` reference line removed entirely (not renamed) because the underlying skill is being retired in M1.

**Plan-time verification**: The template tree at `internal/template/templates/.claude/rules/moai/languages/` DOES mirror all 16 language rules. Both `elixir.md` (line 96) and `csharp.md` (line 107) have matching reference removals required in BOTH local and template paths. Other language rules (`kotlin.md`, `swift.md`, `flutter.md`, etc.) carry zero matches — they are NOT included in M2 scope. M2 enumerates all 4 file edits as mandatory tasks below.

**Tasks**:

| # | Action | File | REQ | AC |
|---|--------|------|-----|----|
| 2.1 | Edit (remove line 96) | `.claude/rules/moai/languages/elixir.md` | REQ-CSB-005 | AC-CSB-005 |
| 2.2 | Edit (remove line 107) | `.claude/rules/moai/languages/csharp.md` | REQ-CSB-005 | AC-CSB-005 |
| 2.3 | Edit (remove line 96, mirror) | `internal/template/templates/.claude/rules/moai/languages/elixir.md` | REQ-CSB-005 | AC-CSB-005 |
| 2.4 | Edit (remove line 107, mirror) | `internal/template/templates/.claude/rules/moai/languages/csharp.md` | REQ-CSB-005 | AC-CSB-005 |

**Verification**: AC-CSB-005 — combined grep across local + template tree returns 0:

```bash
[ $(grep -rcE "moai-platform-deploy" .claude/rules/moai/languages/ internal/template/templates/.claude/rules/moai/languages/ | awk -F: '{sum+=$2} END {print sum+0}') -eq 0 ] && echo PASS
```

**Commit message**: `chore(SPEC-V3R5-CORE-SLIM-B-001): M2 — remove moai-platform-deployment cross-refs in elixir + csharp (local + template mirror)`

### M3 — Cleanup `agents-reference.md` Dead Refs + Regenerate `embedded.go` + Cross-Platform Verification

**Scope**: 2 file edits (3 line-level changes each = 6 line-level changes total) + `make build` + full test suite + cross-platform build verification.

**Plan-time verification (grep ground truth)**: `agents-reference.md` lines 269, 288, 290 reference retired skills in both local and template mirror copies. Line 269 also references `moai-platform-database` which is OUT of this SPEC's scope and MUST be preserved.

**Tasks**:

| # | Action | Detail | REQ | AC |
|---|--------|--------|-----|----|
| 3.1 | Edit (line 269) | `.claude/skills/moai-foundation-core/modules/agents-reference.md` — remove `moai-platform-auth` and `moai-platform-deploy` from the row; preserve `moai-platform-database` | REQ-CSB-006 | AC-CSB-007 |
| 3.2 | Edit (line 288) | `.claude/skills/moai-foundation-core/modules/agents-reference.md` — delete the entire `moai-platform-auth` row | REQ-CSB-006 | AC-CSB-007 |
| 3.3 | Edit (line 290) | `.claude/skills/moai-foundation-core/modules/agents-reference.md` — delete the entire `moai-platform-deploy` row | REQ-CSB-006 | AC-CSB-007 |
| 3.4 | Edit (mirror, line 269) | `internal/template/templates/.claude/skills/moai-foundation-core/modules/agents-reference.md` — same row edit as 3.1 (Template-First per C-CSB-001) | REQ-CSB-006 | AC-CSB-007 |
| 3.5 | Edit (mirror, line 288) | `internal/template/templates/.claude/skills/moai-foundation-core/modules/agents-reference.md` — delete `moai-platform-auth` row (mirrors 3.2) | REQ-CSB-006 | AC-CSB-007 |
| 3.6 | Edit (mirror, line 290) | `internal/template/templates/.claude/skills/moai-foundation-core/modules/agents-reference.md` — delete `moai-platform-deploy` row (mirrors 3.3) | REQ-CSB-006 | AC-CSB-007 |
| 3.7 | Run | `make build` (regenerates `internal/template/embedded.go` via `go:embed`) | REQ-CSB-007 | AC-CSB-006 |
| 3.8 | Run | `go test ./...` (full suite, verify no regression) | — | AC-CSB-006 |
| 3.9 | Run | `GOOS=windows GOARCH=amd64 go build ./...` (cross-platform, known issue B1) | — | AC-CSB-008 |

**Important: Line-number drift after edits**: Lines 288 and 290 are 1-based positions in the pre-edit file. After deleting line 288 (task 3.2), the original line 290 will shift up by one. Implementation MUST identify the rows by **content match** (e.g., `| moai-platform-auth |`, `| moai-platform-deploy |`) rather than re-applying numeric line offsets. Same caveat applies to the template mirror file in tasks 3.5/3.6.

**Verification**: AC-CSB-005 (M2), AC-CSB-006 (`make build && go test ./...`), AC-CSB-007 (agents-reference dead-ref grep across local + template = 0), AC-CSB-008 (cross-platform build).

**Commit message**: `feat(SPEC-V3R5-CORE-SLIM-B-001): M3 — cleanup agents-reference.md dead-refs (local + template mirror) + regenerate embedded.go`

---

## 3. Verification Plan

All 8 ACs from spec.md §3 are binary PASS/FAIL via single shell commands. The orchestrator runs each verbatim from project root at sync-phase. The implementation agent self-verifies via the same commands before signaling completion.

| AC | Command (paste-ready) | Maps to milestone |
|----|----------------------|-------------------|
| AC-CSB-001 | `test ! -e internal/template/templates/.claude/skills/moai-framework-electron && test ! -e .claude/skills/moai-framework-electron && echo PASS` | M1 |
| AC-CSB-002 | `test ! -e internal/template/templates/.claude/skills/moai-platform-auth && test ! -e .claude/skills/moai-platform-auth && echo PASS` | M1 |
| AC-CSB-003 | `test ! -e internal/template/templates/.claude/skills/moai-platform-chrome-extension && test ! -e .claude/skills/moai-platform-chrome-extension && echo PASS` | M1 |
| AC-CSB-004 | `test ! -e internal/template/templates/.claude/skills/moai-platform-deployment && test ! -e .claude/skills/moai-platform-deployment && echo PASS` | M1 |
| AC-CSB-005 | `[ $(grep -rcE "moai-platform-deploy" .claude/rules/moai/languages/ internal/template/templates/.claude/rules/moai/languages/ \| awk -F: '{sum+=$2} END {print sum+0}') -eq 0 ] && echo PASS` | M2 |
| AC-CSB-006 | `make build && go test ./... && echo PASS` | M3 |
| AC-CSB-007 | `[ $(grep -cE "moai-platform-auth\|moai-framework-electron\|moai-platform-chrome-extension\|moai-platform-deploy" .claude/skills/moai-foundation-core/modules/agents-reference.md internal/template/templates/.claude/skills/moai-foundation-core/modules/agents-reference.md \| awk -F: '{sum+=$2} END {print sum+0}') -eq 0 ] && echo PASS` | M3 |
| AC-CSB-008 | `GOOS=windows GOARCH=amd64 go build ./... && echo PASS` | M3 |

**Verification batch** (per `.claude/rules/moai/workflow/verification-batch-pattern.md`): AC-CSB-001..005 and AC-CSB-007 are read-only filesystem/grep checks and can run in a single parallel multi-Bash turn. AC-CSB-006 (`make build && go test ./...`) and AC-CSB-008 (`GOOS=windows GOARCH=amd64 go build`) both depend on M3 file edits and on `embedded.go` regeneration — they MUST run after M3 file edits and after `make build`, but the two builds can themselves parallelize once `embedded.go` is regenerated.

---

## 4. Brownfield Strategy

**Strategy (b) — Preserve other skills; extend nothing.**

Rationale per audit §3.2: the 4 Category B skills have **0 workflow invocations + 0 agent frontmatter invocations** measured. They are not load-bearing for any current MoAI-ADK functionality. Pure delete + textual cleanup is the minimal change.

**What is preserved** (untouched):
- All other skills under `internal/template/templates/.claude/skills/` (moai-foundation-*, moai-workflow-*, moai-domain-*, moai-meta-harness/, etc.) except the targeted 3-line cleanup in `moai-foundation-core/modules/agents-reference.md` (M3)
- All agents under `internal/template/templates/.claude/agents/` and `.claude/agents/`
- All workflow files under `.claude/skills/moai/workflows/`
- All rules under `.claude/rules/` (except the elixir.md + csharp.md edits in M2)
- `moai-platform-database` references — explicitly preserved in `agents-reference.md` line 269 (out of this SPEC's scope)
- `moai-meta-harness/SKILL.md` — completely untouched in iter 2 revision (the previous draft's expert-mobile cleanup was a phantom target; grep verified 0 matches at iter 2)
- `internal/template/embedded.go` is regenerated, not edited — its content is mechanically derived from the template tree

**What is NOT extended**: No replacement skill is created. No baseline content is migrated. No agent learns a new responsibility. The change shrinks the system; it does not redistribute behavior. Category A and C migrations (which DO involve extension) are explicitly out of scope (§6.1).

---

## 5. Known Issues (from `.claude/rules/moai/development/manager-develop-prompt-template.md` § Section B)

Pre-filtered to relevant categories for this Tier S SPEC:

- **B1 Cross-platform Build Tags**: This SPEC deletes Go-free directories (skill markdown only). No syscall code is touched. `GOOS=windows GOARCH=amd64 go build ./...` validates that `embedded.go` regeneration does not introduce platform-specific compilation errors. Verified by AC-CSB-008.
- **B6 spec-lint h3 Out of Scope pattern**: `## Out of Scope` (h2) alone produces `MissingExclusions` ERROR. spec.md §6.1 (`### 6.1 Out of Scope`) satisfies this.
- **B8 Working Tree Hygiene**:
  - DO NOT modify `.moai/harness/usage-log.jsonl` (runtime-managed by SessionEnd hook).
  - DO NOT modify `internal/template/embedded.go` directly — let `make build` regenerate it from the template tree.
  - DO NOT delete `.moai/research/core-slimming-audit-2026-05-20.md` — it is the predecessor doc and must remain intact for cross-referencing from spec.md §7 References.
  - `git add` SHOULD use specific paths (not `git add -A`) to avoid sweeping the runtime usage-log delta into commits.

Other categories (B2 Cross-SPEC policy conflicts, B3 C-HRA-008 subagent boundary, B4 frontmatter canonical, B5 CI 3-tier, B7 observer.go path) are not relevant to this delete-only SPEC.

---

## 6. Rollback Plan

If AC-CSB-006 or AC-CSB-008 fails (`make build`, `go test ./...`, or cross-platform build breaks):

1. `git status` — confirm uncommitted-vs-committed state of each milestone
2. If M3 (the regeneration step) is the failure point: `git restore .` for any uncommitted files; `git reset --hard HEAD~N` only after confirming N committed milestones to drop
3. Investigate the failure mode (likely `embedded.go` regeneration missed a reference, or a test was implicitly depending on a retired skill path)
4. Re-attempt M3 with the fix; do not skip ACs

If a user project break is detected post-merge (R-CSB-001 materializes):

1. The change is NOT reverted on main — the retire is intentional per audit §3.2.
2. Add a deprecation note to CHANGELOG.md for v3.5.0 release tag (C-CSB-003 deferral).
3. Document the migration path (user invokes domain-equivalent skill or accepts loss of dead functionality).
