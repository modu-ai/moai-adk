---
id: SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001
title: "Template Neutrality Audit — Migration Matrix (M1 Draft)"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: Author Name
priority: P1
phase: "v3.0.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "template-system, audit, migration-matrix, allow-list, m1-draft"
tier: L
---

# SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001 — Migration Matrix

> **M1 Draft (2026-05-23)**. M2–M5에서 manager-develop가 본 matrix를 참조하여 PRESERVE / GENERALIZE / REMOVE 정책을 적용한다. M6 chore에서 실측 결과를 반영하여 final로 승격한다.

## §1 Overview

본 matrix는 `internal/template/templates/` 하위 138개 file에 잔존하는 dev-incident / OS-local 흔적을 8 카테고리(C1–C8)로 분류하고 각 카테고리별 정책을 정의한다. 각 섹션은 다음 5개 필드를 포함한다:

1. **Detection regex** — audit script (`internal/template/template_neutrality_audit_test.go`)가 사용할 패턴
2. **Action policy** — PRESERVE (allow-list 유지) / GENERALIZE (패턴 표현으로 변경) / REMOVE (단순 제거) / PRESERVE-ALL (false positive)
3. **Baseline (2026-05-23)** — 본 plan-phase 측정 시점의 hit 수
4. **Post-fix expected** — M2–M5 작업 완료 후 예상되는 hit 수
5. **Allow-list** — `^- ` bullet 형식의 file path enumeration (또는 "Empty (no exceptions)")

[ZONE:Evolvable] [HARD] `^- ` bullet은 **오직 allow-list 파일 경로 한정**. 기타 필드는 bold prose (`**Field**:`) 형식을 사용한다. AC-TNA-002~005가 `awk '/^### CN /,/^### C[0-9]+/' | grep -c '^- '`로 allow-list count를 산출하므로 bullet 오용은 false count를 유발한다.

### 1.1 Severity Class

| Class | Categories | Audit script behavior |
|-------|------------|------------------------|
| **Binary FAIL** | C1, C5, C6, C7 | 1+ hit (allow-list 외) → `t.Errorf` → CI red |
| **Advisory WARN** | C2, C3, C4 | hit > allow-list → `t.Logf` WARN, CI green with annotations |
| **False positive PRESERVE** | C8 | 항상 PRESERVE, audit script `continue` |

### 1.2 PRESERVE 기준 (3 categories)

본 matrix에서 PRESERVE 정당화 기준은 다음 셋 중 하나에 해당해야 한다:

1. **Rule SSOT citation** — `.claude/rules/**/*.md` rule file이 자체 doctrine / canonical identifier 인용
2. **Decision record** — `.moai/decisions/*.md` 또는 zone-registry CONST-V3R5-NNN entry
3. **Doctrine name reference** — 명명된 incident / pattern (예: "2026-04-25 stream-stall incident")의 doctrine name 인용

위 셋 중 어느 것에도 해당하지 않는 V3R / date / memory ref는 GENERALIZE 또는 REMOVE 대상이다.

---

## §2 Category Matrix

### C1 macOS-bias Absolute Path Placeholder

**Detection regex**: `/Users/`

**Action policy**: REMOVE. `/Users/`, `/home/` 등 OS-specific 경로 placeholder를 `$HOME/...` 또는 `~/...` 통일 형식으로 변환한다. 변환 시 원문 예제 narrative (e.g., "Read X" example context)는 보존하고, path 접두사만 교체한다.

**Severity**: Binary FAIL (4 files / 8 lines 즉시 fix 의무, M2)

**Baseline (2026-05-23)**: 4 files / 8 lines

**Post-fix expected**: 0 files (allow-list empty per Binary FAIL semantics)

**Allow-list**: Empty (no exceptions).

**M2 fix targets (informational; not an allow-list, line-level enumeration in `plan.md` §M2)**: `worktree-integration.md` (lines 260, 261, 270, 273), `run/context-loading.md` (lines 177, 185, 188), `moai-foundation-cc/references/examples.md` (line 646), `moai-workflow-loop/references/examples.md` (line 396).

### C2 V3R[0-9] Dev Version Refs

**Detection regex**: `V3R[0-9]`

**Action policy**: case-by-case classify per PRESERVE 기준 §1.2.
- PRESERVE: rule SSOT citation, zone-registry CONST-V3R5-NNN, decision record
- GENERALIZE: incident-specific finding ID 인용을 패턴 표현으로 변환 (예: "V3R5 W3 meta-analysis" → "a prior workflow audit")
- REMOVE: 단순 dev-history 흔적 (예: "V3R6 cleanup commit으로 해결됨")

**Severity**: Advisory WARN (audit script logs > allow-list, CI passes)

**Baseline (2026-05-23)**: 70 files

**Post-fix expected**: ≤ 25 files (allow-list size, M3 실측 후 확정)

**Allow-list (PRESERVE — initial draft, M3 refinement)**:

- `internal/template/templates/.claude/rules/moai/core/zone-registry.md` — CONST-V3R2-NNN + CONST-V3R5-NNN registry IDs (127 hits in this file, all SSOT)
- `internal/template/templates/.moai/decisions/lsp-client-choice.md` — V3R5 LSP client choice decision record
- `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` — cites CONST-V3R5-001..039 boundary rules
- `internal/template/templates/.claude/rules/moai/workflow/worktree-state-guard.md` — CONST-V3R5-029 escalation path
- `internal/template/templates/.claude/rules/moai/workflow/ci-watch-protocol.md` — CONST-V3R5-014..021 protocol clauses
- `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` — CONST-V3R5-022..026 threshold rules
- `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md` — CONST-V3R5-039 worktree-anchored resume
- `internal/template/templates/.claude/rules/moai/development/branch-origin-protocol.md` — CONST-V3R5-030..036 BODP HARD rules
- `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` — CONST-V3R5-027..028 step ordering rules
- `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` — CONST-V3R5-037 tools field format
- `internal/template/templates/.claude/rules/moai/development/skill-authoring.md` — CONST-V3R5-038 allowed-tools format
- `internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md` — Section A-E origin (W3 HARNESS-AUTONOMY-001 meta-analysis)
- `internal/template/templates/.claude/rules/moai/workflow/verification-batch-pattern.md` — SPEC-V3R5-WORKFLOW-OPT-001 Layer D origin clause
- `internal/template/templates/.claude/rules/moai/workflow/agent-teams-pattern.md` — V3R5 derivation context
- `internal/template/templates/.claude/rules/moai/workflow/ci-autofix-protocol.md` — CONST-V3R5-004..013 auto-fix loop clauses

**GENERALIZE candidates (M3 manager-develop)**: ~30 files referencing V3R5/V3R6 round/sprint identifiers in narrative context (workflow/skill descriptions, harness references, project documentation). Replace specific version sigils with pattern names ("a prior workflow optimization SPEC", "the harness rename SPEC").

**REMOVE candidates (M3 manager-develop)**: ~25 files containing single-mention dev-history traces with no doctrinal value.

### C3 ISO Date Refs (2026-0[5-9]-XX)

**Detection regex**: `2026-0[5-9]`

**Action policy**: case-by-case classify per PRESERVE 기준 §1.2.
- PRESERVE: canonical incident dates documented in active doctrine (2026-04-25 stream-stall, 2026-05-09 model-specific threshold revision, 2026-05-17 worktree opt-in policy, 2026-05-20 W3 meta-analysis)
- GENERALIZE: incident-specific dates → month-year granularity ("2026 mid-May") or doctrine name reference
- REMOVE: SPEC frontmatter dates outside allow-list (these are template-managed, not incident references)

**Severity**: Advisory WARN

**Baseline (2026-05-23)**: 32 files

**Post-fix expected**: ≤ 22 files (allow-list size, M4 실측 후 확정)

**Allow-list (PRESERVE — canonical incident date citations)**:

- `internal/template/templates/CLAUDE.md` — 2026-05-17 worktree opt-in policy reference in §14
- `internal/template/templates/.claude/rules/moai/design/constitution.md` — design constitution dates (TBD M4 line-level audit)
- `internal/template/templates/.claude/rules/moai/core/zone-registry.md` — 2026-05-09 model-specific threshold revision context for CONST-V3R5-022..026
- `internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md` — 2026-05-20 W3 HARNESS-AUTONOMY-001 meta-analysis date (Section A-E origin)
- `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` — 2026-05-09 threshold revision (model-specific)
- `internal/template/templates/.claude/rules/moai/workflow/agent-teams-pattern.md` — 2026-05-17 user policy citation
- `internal/template/templates/.claude/rules/moai/workflow/verification-batch-pattern.md` — 2026-05-20 W3 meta-analysis origin
- `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md` — 2026-04-25 stream-stall incident + 2026-05-09 threshold revision
- `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` — 2026-05-17 worktree opt-in policy
- `internal/template/templates/.claude/rules/moai/workflow/worktree-state-guard.md` — 2026-05-17 policy context for Wave 5 primitive
- `internal/template/templates/.claude/skills/moai/team/run.md` — 2026-05-17 worktree opt-in for team mode

**GENERALIZE candidates (M4 manager-develop)**: ~11 files (sync/run workflow templates, harness skills) referencing dates in narrative context without canonical incident value.

### C4 feedback_/memory.md Refs

**Detection regex**: `feedback_\|memory\.md`

**Action policy**: case-by-case classify per PRESERVE 기준 §1.2.
- PRESERVE: rule SSOT citation of canonical memory file (e.g., `feedback_w3_metaanalysis_lessons.md`, `feedback_large_spec_wave_split.md`, `feedback_worktree_autonomous.md`) when cited in active doctrine
- GENERALIZE: `feedback_*.md` reference → "auto-memory" generic phrasing
- REMOVE: incident-specific memory ref with no doctrinal anchor

**Severity**: Advisory WARN

**Baseline (2026-05-23)**: 9 files

**Post-fix expected**: ≤ 7 files (allow-list size, M4 실측 후 확정)

**Allow-list (PRESERVE — canonical memory citations in rule doctrine)**:

- `internal/template/templates/CLAUDE.md` — §16 Context Search Protocol references auto-memory pattern
- `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` — `feedback_large_spec_wave_split.md` doctrine reference (2026-04-25 incident)
- `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md` — `feedback_large_spec_wave_split.md` cross-reference for wave-split rationale
- `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` — `feedback_worktree_autonomous` memory citation (2026-05-17 policy)
- `internal/template/templates/.claude/rules/moai/workflow/worktree-state-guard.md` — `feedback_worktree_autonomous` memory citation
- `internal/template/templates/.claude/rules/moai/workflow/verification-batch-pattern.md` — `feedback_w3_metaanalysis_lessons.md` origin reference
- `internal/template/templates/.claude/skills/moai/team/run.md` — `feedback_worktree_autonomous` memory citation

**GENERALIZE / REMOVE candidates (M4 manager-develop)**: 2 files outside the allow-list (`run/mode-orchestration.md`, `workflows/sync/quality-gates-context.md`, `workflows/sync/delivery.md`, `workflows/fix.md`, `workflows/brain.md`, `workflows/project/doc-generation.md`, `workflows/harness.md` 등에서 일부 — case-by-case M4 audit).

### C5 CLAUDE.local.md Refs

**Detection regex**: `CLAUDE\.local\.md`

**Action policy**: REMOVE. `CLAUDE.local.md`는 maintainer-only local file이므로 template tree에서 참조 금지. 주변 narrative가 local-only configuration을 abstract하게 논의하면 "local override file" 또는 "machine-specific configuration"으로 generic-replace.

**Severity**: Binary FAIL (1+ hit → audit FAIL, CI red)

**Baseline (2026-05-23)**: 10 files

**Post-fix expected**: 0 files

**Allow-list**: Empty (no exceptions). 10 files 모두 M4에서 remove or generic-replace 대상이다.

**Fix targets (M4 manager-develop)**: `lsp.yaml.tmpl`, `output-styles/moai/moai.md`, `agent-authoring.md`, `skill-authoring.md`, `branch-origin-protocol.md`, `moai-memory.md`, `workflows/loop.md`, `workflows/project/doc-generation.md`, `moai-meta-harness/SKILL.md`, `moai-workflow-loop/SKILL.md`, `moai-workflow-loop/references/reference.md`, `moai-workflow-loop/references/examples.md` — line-level audit는 M4에서 수행.

### C6 PR #N Refs

**Detection regex**: `PR #[0-9]+`

**Action policy**: REMOVE. 특정 PR 번호는 dev-history specific이며 template 배포 부적합. doctrine precedent 인용이 필요하면 generic phrase ("a prior round of the same rule", "the originating PR") 사용.

**Severity**: Binary FAIL

**Baseline (2026-05-23)**: 3 files

**Post-fix expected**: 0 files

**Allow-list**: Empty (no exceptions). 3 files 모두 M5에서 remove or generic-replace 대상이다.

**Fix targets (M5 manager-develop)**: `CLAUDE.md`, `quality.yaml.tmpl`, `moai-workflow-ci-loop/SKILL.md` — line-level audit는 M5에서 수행.

### C7 Commit Hash Refs

**Detection regex**: `\b[a-f0-9]{7,40}\b` (audit script filters via allow-list to exclude color codes, SHA test fixtures, non-commit hex)

**Action policy**: REMOVE. 특정 commit hash는 dev-history specific. doctrinal decision 인용이 필요하면 doctrine name / rule citation으로 대체.

**Severity**: Binary FAIL (audit script handles allow-list logic; manual grep approximation insufficient — see AC-TNA-007 for go test verification)

**Baseline (2026-05-23)**: ~2 files (post-deduplication of false-positive hex matches)

**Post-fix expected**: 0 files (typical) or allow-list size

**Allow-list**: Empty (no exceptions).

**M5 candidate files (informational; line-level audit confirms commit-hash vs hex-example classification)**: `moai/workflows/project/mode-detection.md`, `moai-workflow-testing/references/pr-review-multi-agent.md`. 본 2 files은 fix target이며, allow-list이 아니다. M5 audit script가 hex match를 allow-list (color codes / SHA test fixtures)와 commit hash로 구분한다.

### C8 GOOS= Go Env Var (False Positive Preservation)

**Detection regex (strict)**: `GOOS=(linux|windows|darwin|freebsd|openbsd|netbsd)` (extended regex; BRE escape: `GOOS=\(linux\|...\)`)

**Action policy**: PRESERVE-ALL. `GOOS=` Go cross-compile environment variable은 maintainer-personal name "GOOS Kim"과 무관한 legitimate Go build context. audit script (`internal/template/template_neutrality_audit_test.go`)는 본 패턴을 explicit `continue` 처리하여 violation report에서 제외한다.

**Severity**: False positive PRESERVE (audit script `continue`, no violation emitted)

**Baseline (2026-05-23)**: 3 files / 4 hits

**Post-fix expected**: 3 files / 4 hits (unchanged — PRESERVE-ALL)

**Allow-list (False-positive preservation entries)**:

- `internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md` — `GOOS=windows GOARCH=amd64 go build ./...` cross-platform build verification in Section B1
- `internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` — Go cross-compile example in sync delivery workflow
- `internal/template/templates/scripts/ci-mirror/cross-compile.sh` — CI cross-compile shell script (legitimate Go env var usage)

**Audit script behavior**: `TestTemplateNeutralityAudit/C8_false_positive` subtest verifies these 3 files preserve their `GOOS=` substring AND the audit emits NO violation against them (REQ-TNA-008, AC-TNA-011).
