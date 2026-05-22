---
id: SPEC-V3R6-AGENT-FOLDER-SPLIT-001
title: "Acceptance Criteria — Agent Folder Split"
version: "0.1.1"
status: draft
created: 2026-05-22
updated: 2026-05-22
author: manager-spec
priority: Medium
phase: "v3.0.0"
module: ".claude/agents, internal/template/templates/.claude/agents, internal/template/catalog.yaml"
lifecycle: spec-anchored
tags: "agent, folder-restructure, acceptance, wave-2"
tier: M
---

# Acceptance Criteria — SPEC-V3R6-AGENT-FOLDER-SPLIT-001

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-22 | manager-spec | Initial 11 binary ACs covering 10 REQs |
| 0.1.1 | 2026-05-22 | manager-spec | iter 2: AC-AFS-006 hedge word "~25" removed (replaced with precise count "41 in 21 in-scope files"). AC-AFS-007 count corrected (8 → 22 occurrences in 19 files). AC-AFS-009 catalog entry count corrected (18 → 19). **NEW AC-AFS-012**: 9 out-of-scope Go files PRESERVE verification (git diff empty). |

본 문서는 SPEC-V3R6-AGENT-FOLDER-SPLIT-001 run-phase 완료의 binary verification matrix를 정의한다. 모든 AC는 단일 verifiable command + expected output으로 구성된다.

## REQ ↔ AC Traceability Matrix

| REQ | Description | AC IDs |
|---|---|---|
| REQ-AFS-001 | Folder creation + agent placement | AC-AFS-001, AC-AFS-002 |
| REQ-AFS-002 | `git mv` history preservation | AC-AFS-003 |
| REQ-AFS-003 | Frontmatter byte-identical preservation | AC-AFS-005 |
| REQ-AFS-004 | In-scope cross-reference update | AC-AFS-006, AC-AFS-007 |
| REQ-AFS-005 | Template-First Rule (local + template mirror) | AC-AFS-001, AC-AFS-002, AC-AFS-008 |
| REQ-AFS-006 | No backward-compat aliasing | AC-AFS-004 |
| REQ-AFS-007 | catalog.yaml path + hash regeneration | AC-AFS-009 |
| REQ-AFS-008 | `harness/` PRESERVE | AC-AFS-010 |
| REQ-AFS-009 | `name:` field unchanged | AC-AFS-005 |
| REQ-AFS-010 (NEW iter 2) | 9 out-of-scope Go files PRESERVE | AC-AFS-012 |
| REQ-AFS-011 | Dirty working tree preservation | AC-AFS-011 |

Total: 11 REQs ↔ 12 ACs. 100% traceability.

## Acceptance Criteria (Binary PASS/FAIL)

### AC-AFS-001: Final folder structure correct (local)

**Verification command**:
```bash
find .claude/agents -maxdepth 1 -type d | sort
```

**Expected output** (exact, after run-phase):
```
.claude/agents
.claude/agents/core
.claude/agents/expert
.claude/agents/harness
.claude/agents/meta
```

**Binary criterion**: 5 lines exactly, no `.claude/agents/moai` entry. Output equals expected → PASS.

### AC-AFS-002: File count per folder correct (local + template)

**Verification command**:
```bash
echo "=== local ===" \
  && ls .claude/agents/core/*.md | wc -l \
  && ls .claude/agents/expert/*.md | wc -l \
  && ls .claude/agents/meta/*.md | wc -l \
  && ls .claude/agents/harness/*.md | wc -l \
  && echo "=== template ===" \
  && ls internal/template/templates/.claude/agents/core/*.md | wc -l \
  && ls internal/template/templates/.claude/agents/expert/*.md | wc -l \
  && ls internal/template/templates/.claude/agents/meta/*.md | wc -l \
  && ls internal/template/templates/.claude/agents/harness/*.md | wc -l
```

**Expected output** (exact):
```
=== local ===
       8
       6
       5
       4
=== template ===
       8
       6
       5
       4
```

**Binary criterion**: 8 (core) + 6 (expert) + 5 (meta) + 4 (harness) = 23 per side. Total 46. PASS if output equals expected.

### AC-AFS-003: Git history continuity via git mv

**Verification command**:
```bash
git log --follow --format="%h %s" -2 .claude/agents/core/manager-spec.md
```

**Expected output**: At least 2 commits returned, with the older commit referencing the file under its previous path `.claude/agents/moai/manager-spec.md` (visible via `git log --follow`).

**Binary criterion**: `--follow` returns ≥2 commits → PASS. If only 1 commit (file appears as new) → FAIL (indicates `Write + rm` instead of `git mv`).

### AC-AFS-004: No backward-compat aliasing

**Verification command**:
```bash
test ! -d .claude/agents/moai && test ! -d internal/template/templates/.claude/agents/moai && echo "PASS" || echo "FAIL"
```

**Expected output**: `PASS`

**Binary criterion**: 둘 다 부재해야 PASS. 옛 `moai/` 폴더가 local 또는 template 어느 한 쪽이라도 존재하면 FAIL.

### AC-AFS-005: Frontmatter byte-identical preservation

**Verification command** (sample 5 agents from different folders):
```bash
for agent in core/manager-spec expert/expert-backend meta/builder-harness meta/plan-auditor core/manager-develop; do
  echo "=== $agent ==="
  awk '/^---$/{c++; if(c==2) exit} {print}' ".claude/agents/${agent}.md"
done
```

**Expected output**: For each sampled agent, the printed frontmatter MUST contain unchanged `name:` field value matching the file basename (e.g., `name: manager-spec` for `.claude/agents/core/manager-spec.md`). All other fields (`tools:`, `model:`, `hooks:`, `skills:`, `effort:`, `permissionMode:`, `memory:`) MUST be preserved as in pre-rename baseline.

**Binary criterion**: 5/5 sampled agents pass frontmatter inspection (manual verification or diff against baseline snapshot saved to progress.md) → PASS.

### AC-AFS-006: Zero stale `.claude/agents/moai/` path refs in active Go code (in-scope files only)

**Verification command** (in-scope subset only — excludes 9 out-of-scope files per REQ-AFS-010):
```bash
grep -rn '\.claude/agents/moai/' internal/ 2>/dev/null \
  | grep -v -E '(internal/template/templates/|internal/template/catalog\.yaml|internal/harness/frozen_guard\.go|internal/harness/safety/frozen_guard\.go|internal/harness/safety/frozen_guard_test\.go|internal/harness/safety_preservation_test\.go|internal/harness/safety/pipeline_test\.go|internal/harness/meta_invocation_test\.go|internal/hook/pre_tool\.go|internal/template/catalog_tier_audit_test\.go|internal/template/agent_frontmatter_audit_test\.go)' \
  | wc -l
```

**Expected output**: `0`

**Binary criterion**: 0 occurrences in 21 in-scope Go files. Pre-existing baseline had **41 occurrences in 21 in-scope Go files** (precise count, no hedge): `internal/cli/agent_lint.go` (2), `internal/cli/plan_audit_d7_d8_test.go` (1), `internal/cli/update.go` (1), `internal/cli/update_preserve_my_harness_test.go` (1), `internal/cli/update_safety_test.go` (1), `internal/core/project/initializer_test.go` (1), `internal/hook/agent_start_test.go` (2), `internal/hook/subagent_start.go` (1), `internal/manifest/types_test.go` (1), `internal/merge/confirm_coverage_test.go` (1), `internal/research/safety/frozen.go` (1), `internal/research/safety/frozen_test.go` (3), `internal/template/builder_skill_path_test.go` (1), `internal/template/deployer_bench_test.go` (1), `internal/template/deployer_mode_test.go` (8), `internal/template/deployer_test.go` (4), `internal/template/embed.go` (2), `internal/template/manager_develop_present_test.go` (3), `internal/template/rule_template_mirror_test.go` (3), `internal/template/slim_guard.go` (1), `internal/template/slim_guard_test.go` (2). Post-implementation: 0 in-scope occurrences. The 9 out-of-scope files (42 occurrences) remain unchanged per AC-AFS-012.

### AC-AFS-007: Zero stale `.claude/agents/moai/<name>.md` path refs in active rules/skills

**Verification command**:
```bash
grep -rn '\.claude/agents/moai/' \
  .claude/rules/ .claude/skills/ \
  internal/template/templates/.claude/rules/ \
  internal/template/templates/.claude/skills/ \
  CLAUDE.md \
  internal/template/templates/CLAUDE.md \
  2>/dev/null | wc -l
```

**Expected output**: `0`

**Binary criterion**: 0 active rule/skill/CLAUDE.md occurrences. Pre-existing baseline had **22 occurrences in 19 unique files** (12 local + 10 template):
- Local files (10): `.claude/rules/moai/development/agent-authoring.md` (1), `.claude/rules/moai/development/model-policy.md` (2), `.claude/rules/moai/workflow/spec-workflow.md` (2), `.claude/rules/moai/workflow/team-protocol.md` (1), `.claude/skills/moai-harness-learner/SKILL.md` (1), `.claude/skills/moai-meta-harness/SKILL.md` (1), `.claude/skills/moai/workflows/harness.md` (1), `.claude/skills/moai/workflows/plan/spec-assembly.md` (1), `.claude/skills/moai/workflows/project/meta-harness.md` (1), `.claude/skills/moai/workflows/release.md` (1, local-only per CLAUDE.local.md §21 — no template mirror)
- Template mirrors (9 files, 10 occurrences): mirrors of all 10 above except `workflows/release.md` (dev-only, excluded from template per §21)

### AC-AFS-008: Template-First Rule — local ↔ template mirror identical

**Verification command**:
```bash
diff -rq .claude/agents/core/ internal/template/templates/.claude/agents/core/ && \
diff -rq .claude/agents/expert/ internal/template/templates/.claude/agents/expert/ && \
diff -rq .claude/agents/meta/ internal/template/templates/.claude/agents/meta/ && \
diff -rq .claude/agents/harness/ internal/template/templates/.claude/agents/harness/ && \
echo "ALL_PASS"
```

**Expected output**: `ALL_PASS` (no diff output before final echo).

**Binary criterion**: 4 diff commands all exit 0 with no output → PASS. Any difference → FAIL (Template-First Rule violation).

> **Note**: pre-existing baseline drift (plan-auditor.md / hooks-system.md 등 10 files per `project_v3r6_template_mirror_drift_audit_2026_05_22`) is out of this SPEC's scope. AC-AFS-008 verifies that **agents/** specifically has zero drift after run-phase.

### AC-AFS-009: catalog.yaml path entries updated + hash regenerated

**Verification command**:
```bash
grep -c 'templates/\.claude/agents/moai/' internal/template/catalog.yaml && \
grep -c 'templates/\.claude/agents/\(core\|expert\|meta\)/' internal/template/catalog.yaml
```

**Expected output**:
```
0
19
```

**Binary criterion**: 0 `agents/moai/` path refs + 19 new `agents/{core,expert,meta}/` path refs in catalog.yaml (iter 1 acceptance had "18" — correction: builder-harness.md at line 341 nested under harness category brings total to 19). Hash regeneration verified separately via `TestManifestHashFormat` (CATALOG-SSOT-001 mechanism — `make build` runs `gen-catalog-hashes.go --all`).

### AC-AFS-010: `.claude/agents/harness/` PRESERVE (no change)

**Verification command**:
```bash
ls .claude/agents/harness/*.md | xargs -I{} basename {} .md | sort
```

**Expected output**:
```
cli-template-specialist
hook-ci-specialist
quality-specialist
workflow-specialist
```

**Binary criterion**: 4 files, exact names match (HARNESS-RENAME-001 baseline). PASS if output equals expected.

### AC-AFS-011: Dirty working tree preservation

**Verification command** (run before commit):
```bash
git status --short | grep -E '^( M|MM|\?\?)' | grep -E '(usage-log\.jsonl|docs-site/hugo\.toml|docs-site/layouts|SPEC-V3R5-GIT-STRATEGY-SCHEMA-001|SPEC-V3R5-INIT-WIZARD-EXPANSION-001|SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001|SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001|research/config-audit|research/lsp-yaml-v2-audit|docs-site/data|docs-site/scripts/gen_menu\.py|internal/hook/\.moai)' | wc -l
```

**Expected output**: `≥10` (matches pre-run-phase baseline count of PRESERVE files).

**Binary criterion**: PRESERVE list files still in modified/untracked state (unchanged). PASS if count matches pre-run baseline.

### AC-AFS-012 (NEW iter 2): 9 out-of-scope Go files byte-identical PRESERVE

**Verification command** (run after run-phase commits, before final report):
```bash
git diff origin/main -- \
  internal/harness/frozen_guard.go \
  internal/harness/safety/frozen_guard.go \
  internal/harness/safety/frozen_guard_test.go \
  internal/harness/safety_preservation_test.go \
  internal/harness/safety/pipeline_test.go \
  internal/harness/meta_invocation_test.go \
  internal/hook/pre_tool.go \
  internal/template/catalog_tier_audit_test.go \
  internal/template/agent_frontmatter_audit_test.go \
  | wc -l
```

**Expected output**: `0` (empty diff)

**Binary criterion**: 9 out-of-scope files byte-identical with origin/main → PASS. Any line of diff output → FAIL (REQ-AFS-010 PRESERVE violation, follow-up SPEC `FROZEN-PREFIX-REALIGN-001` scope leaked into this SPEC).

**Additional grep verification**:
```bash
# Verify out-of-scope files still contain '.claude/agents/moai/' literal (preservation evidence)
out_of_scope_count=$(grep -c '\.claude/agents/moai/' \
  internal/harness/frozen_guard.go \
  internal/harness/safety/frozen_guard.go \
  internal/harness/safety/frozen_guard_test.go \
  internal/harness/safety_preservation_test.go \
  internal/harness/safety/pipeline_test.go \
  internal/harness/meta_invocation_test.go \
  internal/hook/pre_tool.go \
  internal/template/catalog_tier_audit_test.go \
  internal/template/agent_frontmatter_audit_test.go 2>/dev/null | awk -F: '{sum += $2} END {print sum}')
echo "Out-of-scope preserved occurrences: $out_of_scope_count (expected: 42)"
```

**Expected**: `Out-of-scope preserved occurrences: 42`

## Edge Cases

### EC-1: One of 19 agents has a corrupted frontmatter pre-existing

If pre-flight discovers any agent with invalid YAML frontmatter (parse error), run-phase MUST NOT mask it. Report as blocker. Resolution: separate SPEC scope.

### EC-2: `git mv` fails on case-insensitive filesystem (macOS default)

If `.claude/agents/moai/foo.md` and `.claude/agents/Moai/foo.md` collide (hypothetical), `git mv` fails. Pre-flight `git config --get core.ignorecase` check expected `true` on macOS — no collision risk since target dirs (`core/`, `expert/`, `meta/`) have entirely distinct names.

### EC-3: Template mirror has drift before this SPEC

Pre-existing baseline (`project_v3r6_template_mirror_drift_audit_2026_05_22`) reports 10 drifted files (plan-auditor.md, hooks-system.md, etc.) outside `agents/`. Run-phase MUST NOT inadvertently fix these unrelated drifts. AC-AFS-008 scope is **agents/** only.

### EC-4: Active SPEC frontmatter `module:` field references `.claude/agents/moai/...`

Multiple active SPECs (SPEC-V3R5-LINT-CLEAN-001, SPEC-AGENT-002, SPEC-V3R4-HARNESS-001, etc.) have `module:` containing `.claude/agents/moai/` string. Per spec.md §6.2.1 Out of Scope, these are historical artifacts — NOT updated. AC-AFS-006/007 scope explicitly excludes `.moai/specs/` directory.

### EC-5: New SPEC PR added between plan-phase and run-phase

If another contributor adds a new agent file under `.claude/agents/moai/` between plan and run, pre-flight pre-run-phase grep MUST detect it and report as blocker. Run-phase rebases against latest main before starting `git mv` operations.

### EC-6 (NEW iter 2): Accidental modification of out-of-scope Go file

If run-phase accidentally edits any of the 9 out-of-scope files (e.g., manager-develop uses too-broad sed pattern), AC-AFS-012 will FAIL with non-zero diff. Resolution:
1. `git diff origin/main -- <file>` to identify the accidental change
2. `git restore --source=origin/main -- <file>` to revert
3. Re-run AC-AFS-012 to confirm 0 diff
4. Document the near-miss in progress.md for future prevention

## Quality Gate Criteria

### TRUST 5 Validation

- **Tested**: 12 binary ACs cover all 11 REQs. Verification commands are reproducible.
- **Readable**: Folder names (`core/`, `expert/`, `meta/`, `harness/`) self-documenting. Agent classification rationale documented in §3. Out-of-scope rationale in §6.2.2.
- **Unified**: Template-First Rule strictly enforced (AC-AFS-008 diff check).
- **Secured**: No backward-compat aliasing (AC-AFS-004) prevents reference-leak vulnerabilities. 9 out-of-scope Go files (frozen_guard family) PRESERVE prevents accidental FROZEN policy weakening (AC-AFS-012).
- **Trackable**: All REQs traced to ACs (matrix above). 38 `git mv` rename operations preserved in git history (AC-AFS-003). Follow-up SPEC `FROZEN-PREFIX-REALIGN-001` (가칭) explicit reference in §6.2.2 + 8.3.

### CI Gate Criteria

- **spec-lint**: 0 NEW regression. Pre-existing baseline 22 NEW preserved.
- **golangci-lint**: 0 NEW regression. Pre-existing baseline preserved.
- **Test (per OS)**: Cross-platform PASS (`go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0).
  - Pre-existing baseline 3 FAIL preserved (`TestRuleTemplateMirrorDrift`, `TestLateBranchTemplateMirror`, `TestSkillsContainPlanAuditGateMarkers`).
  - **NEW out-of-scope deferred FAIL** (acceptable, follow-up SPEC scope): `TestAllAgentsInCatalog` (WalkDir `.claude/agents/moai` returns empty post-move), `TestAgentFrontmatterAudit` (3 WalkDir + per-agent fs.Stat checks fail similarly). 이 두 FAIL은 progress.md에 "out-of-scope deferred residual" 카테고리로 명시.

### Definition of Done

- [ ] AC-AFS-001 ~ AC-AFS-012: All 12 PASS
- [ ] spec.md status: draft → implemented, version 0.1.1 → 0.2.0
- [ ] progress.md with run-phase completion summary + baseline 3 FAIL + out-of-scope deferred FAIL 분류 명시
- [ ] Cross-platform build PASS evidence in progress.md
- [ ] Commit message follows Conventional Commits + `🗿 MoAI <email@mo.ai.kr>` trailer
- [ ] PRESERVE list files unchanged (AC-AFS-011)
- [ ] 9 out-of-scope Go files byte-identical (AC-AFS-012)
- [ ] Follow-up SPEC `SPEC-V3R6-FROZEN-PREFIX-REALIGN-001` (가칭) plan-phase 후보 명시 (progress.md)

## Given-When-Then Scenarios

### Scenario 1: Successful folder split (scope-reduced)

**Given**:
- `.claude/agents/moai/` 19 files exist
- `.claude/agents/harness/` 4 files exist
- 9 out-of-scope Go files contain 42 `.claude/agents/moai/` occurrences (FROZEN guard + walker tests + SentinelHarnessFrozenAgent)
- 21 in-scope Go files contain 41 `.claude/agents/moai/` occurrences
- 19 active rule/skill files contain 22 `.claude/agents/moai/` occurrences

**When**: `/moai run SPEC-V3R6-AGENT-FOLDER-SPLIT-001` executed

**Then**:
- `.claude/agents/{core,expert,meta}/` 3 folders created with 8/6/5 files respectively
- `.claude/agents/harness/` 4 files unchanged
- `.claude/agents/moai/` completely removed
- `internal/template/templates/.claude/agents/` mirrors local structure exactly
- `internal/template/catalog.yaml` 19 path entries updated to new prefixes (iter 1 "18"은 오기, 실제 19개)
- 21 in-scope Go files: 0 `.claude/agents/moai/` occurrences (AC-AFS-006)
- 19 active rule/skill files: 0 `.claude/agents/moai/` occurrences (AC-AFS-007)
- **9 out-of-scope Go files: 42 `.claude/agents/moai/` occurrences UNCHANGED** (AC-AFS-012)
- Git history shows `git mv` rename for all 38 file movements
- All 12 ACs PASS

### Scenario 2: Run-phase pre-flight detects new file in moai/

**Given**: Between plan-phase and run-phase, another contributor merged PR adding `.claude/agents/moai/new-agent.md`
**When**: Run-phase pre-flight `ls .claude/agents/moai/*.md | wc -l` returns 20 (not 19)
**Then**: manager-develop returns blocker report identifying the 20th file. Re-delegation required after user decides classification (core/expert/meta) or whether to defer integration.

### Scenario 3: catalog.yaml hash regeneration

**Given**: 19 catalog.yaml entries have new paths but old hashes
**When**: `make build` executed
**Then**: `gen-catalog-hashes.go --all` runs (CATALOG-SSOT-001 self-healing manifest gate). All 19 hash values updated to reflect new file contents (file contents unchanged, but path-based hash input changes). `TestManifestHashFormat` PASS.

### Scenario 4 (NEW iter 2): Accidental out-of-scope file edit detection

**Given**: manager-develop applies a broad sed pattern that inadvertently rewrites `.claude/agents/moai/` → `.claude/agents/meta/` in `internal/hook/pre_tool.go:720` (SentinelHarnessFrozenAgent table)
**When**: AC-AFS-012 verification executes `git diff origin/main -- internal/hook/pre_tool.go`
**Then**: diff output non-empty (e.g., 2 lines for the literal change). AC-AFS-012 FAILS. Manager reverts via `git restore --source=origin/main -- internal/hook/pre_tool.go`. Re-verify: AC-AFS-012 PASS. Document near-miss in progress.md for future grep filter discipline.

### Scenario 5 (NEW iter 2): Follow-up SPEC scope handoff

**Given**: This SPEC implemented (status: completed). `TestAllAgentsInCatalog` + `TestAgentFrontmatterAudit` in baseline residual FAIL state. 9 out-of-scope Go files preserved.
**When**: User initiates `/moai plan SPEC-V3R6-FROZEN-PREFIX-REALIGN-001` (가칭)
**Then**: Follow-up SPEC plan-phase analyzes:
- `pre_tool.go:720` SentinelHarnessFrozenAgent table — should `.claude/agents/moai/` → `.claude/agents/core|expert|meta/` OR keep as semantic legacy (multi-prefix support)?
- `catalog_tier_audit_test.go` + `agent_frontmatter_audit_test.go` walker logic — update WalkDir paths to walk 3 new folders OR consolidate via fs.Glob pattern?
- `frozen_guard_test.go` expected-array — update to new prefixes OR multi-prefix support?

Follow-up SPEC defines exact scope, REQs, ACs, and migrates 9 out-of-scope files. Total ~17 LOC change estimated.
