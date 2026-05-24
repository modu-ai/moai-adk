---
id: SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001
title: "SPEC artifact ownership realignment across manager-spec / manager-develop / manager-docs — Implementation Plan"
version: "0.1.0"
status: draft
created: 2026-05-24
updated: 2026-05-24
author: GOOS행님
priority: P1
phase: "v3.0.0"
module: ".claude/agents/core"
lifecycle: spec-anchored
tags: "agent-ownership, soc, manager-spec, manager-develop, manager-docs, status-transition, schema, audit-tier-2, anthropic-best-practice"
---

# SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001 — Implementation Plan (Tier S minimal Section A-E)

## §A. Context

### §A.1 SPEC artifacts location

- Project root: `/Users/goos/MoAI/moai-adk-go`
- Branch: `main` (Hybrid Trunk 1-person OSS per CLAUDE.local.md §23.7)
- HEAD SHA at plan-phase start: `860fc119f` (audit Tier 1 chore commit: `chore(governance): Anthropic best practices audit Tier 1 — F2 + F4 (F5 blocked)`)
- SPEC artifacts (4 files, plan-phase):
  - `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/spec.md` (canonical SSOT)
  - `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/plan.md` (this file)
  - `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/acceptance.md` (7 ACs canonical matrix)
  - `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/progress.md` (lifecycle + audit-ready signal)
- plan-auditor verdict: TBD at plan-phase write completion (target: PASS ≥ 0.75 Tier S threshold)

### §A.2 Existing infrastructure (PRESERVE vs EXTEND)

**PRESERVE (unchanged across plan-phase commit) — 7 dirty/untracked entries per L45 discipline (verified `git status --porcelain | wc -l` at plan-phase start)**:

| # | Path | Type | Source |
|---|------|------|--------|
| 1 | `.moai/config/sections/git-convention.yaml` | M | local edit (§22 dev settings) |
| 2 | `.moai/config/sections/language.yaml` | M | §22 dev settings |
| 3 | `.moai/config/sections/quality.yaml` | M | §22 dev settings |
| 4 | `.moai/harness/usage-log.jsonl` | M | runtime-managed |
| 5 | `.moai/harness/observations.yaml` | ?? | runtime-managed |
| 6 | `.moai/research/anthropic-best-practices-2026-05-24.md` | ?? | parallel-session audit artifact (origin of this SPEC) |
| 7 | `.moai/research/v3.0-redesign-2026-05-23.md` | ?? | parallel-session research artifact |
| 8 | `i18n-validator` | ?? | parallel-session artifact |

Total `git status --porcelain | wc -l` at plan-phase start: 7-8 entries (verified 2026-05-24 pre-plan).

**EXTEND (modified in plan-phase) — 4 entries exactly**:

| # | Path | Type | Operation |
|---|------|------|-----------|
| 1 | `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/spec.md` | NEW SPEC artifact | created in plan-phase by manager-spec |
| 2 | `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/plan.md` | NEW SPEC artifact | created in plan-phase (this file) |
| 3 | `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/acceptance.md` | NEW SPEC artifact | created in plan-phase |
| 4 | `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/progress.md` | NEW SPEC artifact | created in plan-phase |

**EXTEND (modified in future run-phase) — 7 entries exactly per spec.md §B.1**:

| # | Path | Type | Operation |
|---|------|------|-----------|
| 1 | `.claude/agents/core/manager-spec.md` | operational source | description: update + new `## SPEC Artifact Ownership` section |
| 2 | `.claude/agents/core/manager-develop.md` | operational source | description: update + new `## SPEC Artifact Ownership` section |
| 3 | `.claude/agents/core/manager-docs.md` | operational source | description: update + new `## SPEC Artifact Ownership` section |
| 4 | `internal/template/templates/.claude/agents/core/manager-spec.md` | template mirror | byte-identical mirror cp of #1 |
| 5 | `internal/template/templates/.claude/agents/core/manager-develop.md` | template mirror | byte-identical mirror cp of #2 |
| 6 | `internal/template/templates/.claude/agents/core/manager-docs.md` | template mirror | byte-identical mirror cp of #3 |
| 7 | `.claude/rules/moai/development/spec-frontmatter-schema.md` | schema doc | new `## Status Transition Ownership Matrix` section appended |

### §A.3 plan-auditor expectation

- Tier S threshold: PASS ≥ 0.75 (per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier)
- Target: iter-1 PASS ≥ 0.85 (margin ≥ 0.10) for skip-eligibility at Phase 0.5 in next /moai run invocation
- MP-1 REQ consistency / MP-2 EARS / MP-3 frontmatter / MP-4 lang neutrality expected to pass

## §B. Known Issues (B1-B12 filtered for relevance — Tier S minimal)

Per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability, Section B is filtered for Tier S minimal — only relevant categories enumerated:

### B-relevant.4 — Frontmatter Canonical Schema (B4)

All 4 plan-phase artifacts use canonical 12 fields per `.claude/rules/moai/development/spec-frontmatter-schema.md` SSOT: `id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags`. `created:` / `updated:` / `tags:` used (NOT `created_at` / `updated_at` / `labels`). `tags:` CSV string (NOT YAML array) per `internal/spec/lint.go` `SPECFrontmatter.Tags string yaml:"tags"` binding. The 7-vs-12 plan-auditor finding pattern from SIV-001 plan-phase iter-1 does NOT apply here — this SPEC uses 12 fields throughout.

### B-relevant.5 — CI 3-tier 인지 (B5)

- spec-lint: spec.md emits `✓ No findings`; plan/acceptance/progress MAY emit `MissingExclusions` ERROR (accepted derived-artifact lint surface per §D.1 in spec.md)
- golangci-lint: 0 new issues introduced (plan-phase is pure .md authoring, no .go changes)
- Test (per OS): N/A in plan-phase (no .go changes; run-phase will engage Go tests for mirror invariant)

### B-relevant.6 — spec-lint Heading 규약 (B6)

spec.md uses `## §B. Scope` heading + `### §B.2 Out of Scope (deferred to follow-up SPECs or explicitly NOT done)` h3 sub-section — `MissingExclusions` requirement satisfied on spec.md.

### B-relevant.8 — Working Tree Hygiene (B8)

PRESERVE list 7-8 entries above §A.2. Path-specific `git add .moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/` discipline only. `--no-verify` absolutely forbidden. Runtime-managed files (`.moai/harness/usage-log.jsonl`, `.moai/harness/observations.yaml`) untouched. Parallel-session audit artifacts (`.moai/research/*`) untouched.

### B-relevant.9 — Git Commit + Push 자체 수행 (B9)

For Hybrid Trunk 1-person OSS per CLAUDE.local.md §23.7, plan-phase commit + push is performed by orchestrator (not manager-spec subagent) since this is plan-phase. Conventional Commits format obligation: `feat(SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001): plan-phase artifacts (Tier S minimal Section A-E, 4 artifacts)`. Korean commit body per `git_commit_messages: ko`. `🗿 MoAI <email@mo.ai.kr>` trailer obligatory.

### B-relevant.10 — Untouched Paths PRESERVE — Scope Discipline (B10)

Per L40 + L45 strict enforcement. PRESERVE list 7-8 entries above §A.2 unchanged across plan-phase commits. ALL other paths in working tree (`.claude/`, `internal/`, `cmd/`, etc.) untouched in plan-phase — only `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/` is staged.

### B-relevant.11 — AskUserQuestion 금지 — Subagent Boundary (B11)

manager-spec is a subagent — AskUserQuestion invocation prohibited per CLAUDE.md §8 + askuser-protocol.md §Orchestrator-Subagent Boundary. Blocker 발견 시 structured blocker report 반환 (orchestrator가 AskUserQuestion 수행 + re-delegate). Blocker report 4-option format with change/impact/risk/ETA per agent-common-protocol.md.

### B-relevant.12 — Sync-phase CHANGELOG emission discipline (B12)

Applies to manager-docs at /moai sync phase (NOT this plan-phase). manager-docs MUST: (a) Read every implementation file referenced in plan.md before drafting; (b) `grep -c '<SPEC-ID>' CHANGELOG.md` pre-emission verify = 0; (c) AC count match `acceptance.md` SSOT (7 ACs). For B12 self-test PASS streak preservation.

### L51 self-check (SPEC ID regex pre-write protocol — manager-spec body section, post-SIV-001 run-phase)

Per `.claude/agents/core/manager-spec.md` Step 5 Verification Checklist (post-SIV-001 run M1 `0e103eacc`):

```
SSOT: internal/spec/lint.go:573 → ^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$
Decomposition: SPEC ✓ | V3R6 ✓ | AGENT ✓ | RESPONSIBILITY ✓ | REALIGN ✓ | 001 ✓ → PASS
```

The plan-phase manager-spec performed this check verbatim before any Write (recorded at top of orchestrator's plan-phase response). All 4 plan-phase artifacts use the validated SPEC ID `SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001`.

## §C. Pre-flight Check List (run-phase M1 시작 전 manager-develop 의무 실행)

```bash
# 1. 현재 branch + baseline 확인
git branch --show-current             # expect: main
git rev-parse HEAD                    # expect: post-plan-phase commit (this SPEC's plan commit SHA + descendants)

# 2. L44 HARD pre-spawn fetch
git fetch origin main 2>&1
git rev-list --count --left-right origin/main...HEAD  # expect: 0 0

# 3. Mirror baseline 확인 — 3 agent pair는 byte-identical 인가?
for agent in manager-spec manager-develop manager-docs; do
  src=".claude/agents/core/${agent}.md"
  mirror="internal/template/templates/.claude/agents/core/${agent}.md"
  echo "=== ${agent} ==="
  diff -q "$src" "$mirror" || echo "DRIFT detected pre-edit (must fix mirror first)"
done

# 4. spec-frontmatter-schema.md 현재 상태 — 새 section 위치 식별
grep -n '^## ' .claude/rules/moai/development/spec-frontmatter-schema.md
# expect: locate `## Status Enum (8 values)` to append AFTER it

# 5. Lint baseline 측정 (NEW vs pre-existing 구분)
golangci-lint run --timeout=2m 2>&1 | tail -5  # capture baseline; expect: 0 issues.
go vet ./... 2>&1 | tail -5                     # expect: exit 0

# 6. PRESERVE list snapshot
git status --porcelain | tee /tmp/preserve-list-pre.txt
# expect: 7-8 entries unchanged from plan-phase

# 7. Test baseline — TestRuleTemplateMirrorDrift 현재 상태
go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift' -v 2>&1 | tail -20
# expect: 3 manager agent subtests (if listed in workflowOptMirroredPaths) PASS pre-edit; verify post-edit PASS
go test ./internal/template/ -run 'TestLateBranchTemplateMirror' -v 2>&1 | tail -20
# expect: existing entries PASS; verify post-edit no regressions
```

## §D. Constraints (DO NOT VIOLATE)

### D.1 PRESERVE 대상 enumeration

7-8 entries per §A.2 PRESERVE list. Path-specific `git add` only. `git status --porcelain` post-plan-commit MUST show:
- Pre-existing 7-8 PRESERVE entries unchanged status (still M or ?? as before)
- 0 staged paths after the plan commit pushes
- No new untracked files introduced
- Only `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/*` staged in plan-phase commit
- In run-phase: only the 7 EXTEND files (§A.2 run-phase EXTEND table) + the appropriate progress.md update staged

### D.2 금지 명령

- `--no-verify` (any git command)
- `--amend` (creates new commit per CLAUDE.md anti-amend policy)
- `git reset --hard` (per CLAUDE.local.md §23.5 — use `--keep` if needed)
- Force-push to main (per CLAUDE.local.md §23.7 Hybrid Trunk discipline)
- `git add -A` or `git add .` (path-specific only)
- AskUserQuestion invocation (manager-spec/manager-develop/manager-docs are subagents)
- Modifying any file outside `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/` in plan-phase
- Direct Run-phase work in plan-phase (this is PLAN ONLY — agents are listed as Run-phase targets)
- Snake_case frontmatter aliases (`created_at`, `updated_at`, `labels`)
- AC sub-ID convention `AC-XXX-NNNa/b` for SPEC IDs (canonical regex `\d{3}$` digit-only anchor)

### D.3 사용 의무 명령

- Conventional Commits format: `feat(SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001): plan-phase artifacts (Tier S minimal Section A-E, 4 artifacts)` for plan-phase; appropriate `feat`/`fix` for run-phase
- Korean commit body per `git_commit_messages: ko`
- `🗿 MoAI <email@mo.ai.kr>` trailer obligatory
- HEREDOC for commit message: `git commit -m "$(cat <<'EOF' ... EOF)"`
- Path-specific staging: `git add .moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/` (plan-phase) or explicit 7-file enumeration (run-phase)

### D.4 Binary constraints (grep / diff 0 매치, run-phase)

- `diff -q .claude/agents/core/manager-spec.md internal/template/templates/.claude/agents/core/manager-spec.md` = empty (exit 0) post-edit
- `diff -q .claude/agents/core/manager-develop.md internal/template/templates/.claude/agents/core/manager-develop.md` = empty (exit 0) post-edit
- `diff -q .claude/agents/core/manager-docs.md internal/template/templates/.claude/agents/core/manager-docs.md` = empty (exit 0) post-edit
- `grep -c '^## SPEC Artifact Ownership$' .claude/agents/core/manager-{spec,develop,docs}.md` = 1 per file (3 total)
- `grep -c '^## Status Transition Ownership Matrix$' .claude/rules/moai/development/spec-frontmatter-schema.md` = 1
- `git status --porcelain` PRESERVE 7-8 entries verbatim status preservation
- `go vet ./...` = exit 0
- `golangci-lint run --timeout=2m` = `0 issues.`

## §E. Self-Verification Deliverables (M1 완료 보고 시 manager-develop 필수)

### E.1 AC Binary PASS/FAIL Matrix

| AC | Status | Verification Command | Expected Output |
|----|--------|---------------------|-----------------|
| AC-ARR-001 | TBD | `grep -c '^## SPEC Artifact Ownership$' .claude/agents/core/manager-spec.md` | `1` |
| AC-ARR-002 | TBD | `grep -c '^## SPEC Artifact Ownership$' .claude/agents/core/manager-develop.md` | `1` |
| AC-ARR-003 | TBD | `grep -c '^## SPEC Artifact Ownership$' .claude/agents/core/manager-docs.md` | `1` |
| AC-ARR-004 | TBD | `grep -c '^## Status Transition Ownership Matrix$' .claude/rules/moai/development/spec-frontmatter-schema.md` | `1` |
| AC-ARR-005 | TBD | `for agent in manager-spec manager-develop manager-docs; do diff -q .claude/agents/core/$agent.md internal/template/templates/.claude/agents/core/$agent.md \|\| exit 1; done; echo OK` | `OK` (3 byte-identical pairs) |
| AC-ARR-006 | TBD | `go vet ./... 2>&1; echo "vet_exit=$?"; golangci-lint run --timeout=2m 2>&1 \| tail -1` | `vet_exit=0` AND `0 issues.` |
| AC-ARR-007 | TBD | `grep -E '(manager-spec\|manager-develop\|manager-docs).*owns' .claude/rules/moai/development/spec-frontmatter-schema.md \| wc -l` | `≥ 6` (matrix lists at least 6 ownership rows: draft / draft→in-progress / in-progress→implemented / implemented→completed / →superseded / →archived / →rejected) |

### E.2 Cross-Platform Build 결과

```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

(Run-phase modifies only `.md` files; no Go code changes. Cross-platform builds are sanity-only.)

### E.3 Coverage 측정 (scope = template package mirror tests)

```
$ go test -cover ./internal/template/...
```

Expected delta: no statement coverage change (this SPEC adds no `.go` lines; existing TestRuleTemplateMirrorDrift / TestLateBranchTemplateMirror coverage of manager-{spec,develop,docs}.md mirrors continues to apply if those paths are in the allowlist).

### E.4 Subagent Boundary Grep (C-HRA-008)

```
$ grep -rn 'AskUserQuestion' .claude/agents/core/manager-{spec,develop,docs}.md | grep -v "^[^:]*:[0-9]*:[ \t]*//"
(no output expected — agent bodies do NOT invoke AskUserQuestion)
```

### E.5 Lint Status — NEW vs baseline 구분

```
$ golangci-lint run --timeout=2m
0 issues.
```

Pre/post delta MUST be 0 issues introduced. spec-lint on `spec.md` MUST emit `✓ No findings`.

### E.6 Branch HEAD + Push 상태

- 새 commits SHA: 1 commit (M1 fix bundled all 7 EXTEND files) OR 2-3 commits (M1.a agent ownership sections, M1.b mirror cp, M1.c schema doc append) — manager-develop discretion within Tier S envelope
- `git push origin main` 결과: post-push HEAD verified

### E.7 Blocker Report (if any)

If blocker found (e.g., source file modified between pre-flight and M1 commit, parallel session race detected via L44 pre-commit fetch returning `N 0`, scope expansion needed beyond declared 7 files), structured 4-option blocker report returned (orchestrator-direct fix-forward avoided). No AskUserQuestion invocation.

## §F. Implementation Milestones (M1-M3 — Tier S 3 milestones)

### M1 — manager-spec body section + frontmatter description update + template mirror cp

**Action sequence**:

1. Pre-flight checks per §C above
2. Add new `## SPEC Artifact Ownership` section to `.claude/agents/core/manager-spec.md` (operational source). Section content per REQ-ARR-001 spec.md §F definition. Estimated insertion point: after Step 6 Expert Consultation in the existing Workflow Steps section, OR as a top-level section after Persistent Agent Memory block.
3. Update `description:` frontmatter field per REQ-ARR-007 (1-2 sentence summary referring to new section).
4. Mirror cp:
   ```bash
   cp .claude/agents/core/manager-spec.md \
      internal/template/templates/.claude/agents/core/manager-spec.md
   ```
5. Post-edit verify:
   ```bash
   diff -q .claude/agents/core/manager-spec.md internal/template/templates/.claude/agents/core/manager-spec.md
   # expect: empty (exit 0)
   grep -c '^## SPEC Artifact Ownership$' .claude/agents/core/manager-spec.md
   # expect: 1
   ```
6. Commit + push (optionally bundled with M2 + M3 in a single commit for Tier S simplicity; manager-develop discretion).

### M2 — manager-develop body section + frontmatter description update + template mirror cp

**Action sequence**:

1. Add new `## SPEC Artifact Ownership` section to `.claude/agents/core/manager-develop.md` per REQ-ARR-002 spec.md §F definition. Insertion point: after the existing Workflow / methodology-routing section (DDD vs TDD branch) so that the ownership section reads in immediate context with the methodology section.
2. Update `description:` frontmatter field per REQ-ARR-007.
3. Mirror cp:
   ```bash
   cp .claude/agents/core/manager-develop.md \
      internal/template/templates/.claude/agents/core/manager-develop.md
   ```
4. Post-edit verify:
   ```bash
   diff -q .claude/agents/core/manager-develop.md internal/template/templates/.claude/agents/core/manager-develop.md
   # expect: empty (exit 0)
   grep -c '^## SPEC Artifact Ownership$' .claude/agents/core/manager-develop.md
   # expect: 1
   ```
5. Bundle commit with M1 + M3 OR separate commit (manager-develop discretion).

### M3 — manager-docs body section + frontmatter description update + template mirror cp + schema doc matrix

**Action sequence**:

1. Add new `## SPEC Artifact Ownership` section to `.claude/agents/core/manager-docs.md` per REQ-ARR-003 spec.md §F definition. Insertion point: after the existing Documentation Sync workflow section.
2. Update `description:` frontmatter field per REQ-ARR-007.
3. Mirror cp:
   ```bash
   cp .claude/agents/core/manager-docs.md \
      internal/template/templates/.claude/agents/core/manager-docs.md
   ```
4. Append new `## Status Transition Ownership Matrix` section to `.claude/rules/moai/development/spec-frontmatter-schema.md` per REQ-ARR-004. Insertion point: after the existing `## Status Enum (8 values)` section, before the `## Optional Fields` section (or after, manager-develop discretion within the schema doc structure).
5. Post-edit verify:
   ```bash
   diff -q .claude/agents/core/manager-docs.md internal/template/templates/.claude/agents/core/manager-docs.md
   # expect: empty (exit 0)
   grep -c '^## SPEC Artifact Ownership$' .claude/agents/core/manager-docs.md
   # expect: 1
   grep -c '^## Status Transition Ownership Matrix$' .claude/rules/moai/development/spec-frontmatter-schema.md
   # expect: 1
   ```
6. Final commit + push per §D.3 with bundled message OR M1/M2/M3 separate commits — manager-develop discretion. Recommended for Tier S: single bundled commit:
   ```bash
   git add .claude/agents/core/manager-spec.md \
           .claude/agents/core/manager-develop.md \
           .claude/agents/core/manager-docs.md \
           internal/template/templates/.claude/agents/core/manager-spec.md \
           internal/template/templates/.claude/agents/core/manager-develop.md \
           internal/template/templates/.claude/agents/core/manager-docs.md \
           .claude/rules/moai/development/spec-frontmatter-schema.md \
           .moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/progress.md
   git status  # verify: ONLY 8 paths staged, 7-8 PRESERVE unchanged
   # Pre-commit L44 HARD fetch
   git fetch origin main && git rev-list --count --left-right origin/main...HEAD  # expect: 0 0
   git commit -m "$(cat <<'EOF'
   feat(SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001): M1-M3 — 3-agent ownership realignment + schema matrix

   Audit Tier 2 (F1 P1 Critical + F12 P3 Improvement auto-resolve).
   3 core manager agents (spec/develop/docs) receive new `## SPEC Artifact Ownership` body sections + frontmatter description: updates.
   spec-frontmatter-schema.md receives new `## Status Transition Ownership Matrix` section.

   Scope (7 files):
   - M1: manager-spec.md ownership section + mirror cp
   - M2: manager-develop.md ownership section + mirror cp
   - M3: manager-docs.md ownership section + mirror cp + schema matrix

   7/7 ACs PASS: AC-ARR-001..004 grep count + AC-ARR-005 mirror parity 3 pairs + AC-ARR-006 vet/lint 0 issues + AC-ARR-007 matrix row count ≥ 6.
   PRESERVE 7-8 files verbatim. TMD-001/SIV-001 ownership policy now binding.

   🗿 MoAI <email@mo.ai.kr>
   EOF
   )"
   git push origin main
   git rev-list --count --left-right origin/main...HEAD  # expect: 0 0 post-push
   ```

### M-final acceptance gate

All 7 ACs PASS per §E.1 matrix. manager-develop self-verifies and reports per §E. If any AC FAIL, return blocker report to orchestrator (no orchestrator-direct fix-forward in run-phase).

## §G. Anti-Patterns to avoid

- Modifying agent files OUTSIDE the 3 core managers (spec/develop/docs). Expert and meta agents are out-of-scope per spec.md §B.2.
- Using `git add .` or `git add -A` (D.2 violation, PRESERVE list contamination risk).
- Adding new top-level frontmatter field (e.g., `owns_artifacts:`) instead of body-section approach (per spec.md §C.5 decision rationale — body-section is the chosen interface).
- Editing one mirror without the other (mirror invariant violation; CLAUDE.local.md §2 [HARD]).
- Adding new `.go` code (scope envelope violation; this SPEC is `.md` only).
- Bundling unrelated cleanup into the same commit (scope discipline violation; commit covers exactly the 7 EXTEND files + progress.md).
- Implementing the OPTIONAL hook (REQ-ARR-009) in this SPEC scope (deferred to follow-up SPEC if desired).
- Modifying CLAUDE.md narrative in run-phase (REQ-ARR-008 is OPTIONAL MAY; if attempted, must be a separate commit and out of this SPEC's declared scope).
- Retroactive correction of TMD-001 sync precedent (spec.md §B.2: TMD-001 remains historical record per Lessons Protocol).

## §H. Cross-References

- spec.md §F — REQ-ARR-001..009 canonical SSOT
- spec.md §B.1 — 7-file run-phase EXTEND scope
- spec.md §B.2 — Out of Scope enumeration
- spec.md §C.1 — SSOT hierarchy + decision rules
- spec.md §C.5 — body-section vs frontmatter-field design choice rationale
- spec.md §D — lint surface canonical (spec.md `✓ No findings`)
- spec.md §E — Sprint context (Tier S minimal cohort 5/5)
- spec.md §G — cross-references
- acceptance.md §D — AC-ARR-001..007 binary PASS/FAIL matrix
- progress.md §E — 4-phase Lifecycle Status + Audit-Ready Signal
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier S definition
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability — Tier S minimal form
- `.claude/rules/moai/workflow/mx-tag-protocol.md` §a — Mx Step C SKIP/EVALUATE rules
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12-field canonical SSOT (will receive new `## Status Transition Ownership Matrix` section)
- `.claude/agents/core/manager-spec.md` — operational SSOT (will receive new `## SPEC Artifact Ownership` section)
- `.claude/agents/core/manager-develop.md` — operational SSOT (will receive new section)
- `.claude/agents/core/manager-docs.md` — operational SSOT (will receive new section)
- TMD-001 precedent `.moai/specs/SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001/spec.md` — sync-phase scope expansion archetype that motivated this SPEC
- SIV-001 precedent `.moai/specs/SPEC-V3R6-SPEC-ID-VALIDATION-001/spec.md` — L51 sister SPEC pattern + D-NEW-1 inline-fix pattern
- Tier S minimal cohort precedent: IVB-001 / SARM-001 / TMC-001 / TMD-001 / SIV-001
