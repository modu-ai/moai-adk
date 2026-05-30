---
id: SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001
title: "Template Neutrality Audit — Migration Matrix (M1 Draft)"
version: "0.1.2"
status: in-progress
created: 2026-05-23
updated: 2026-05-30
author: Author Name
priority: P1
phase: "v3.0.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "template-system, audit, migration-matrix, allow-list, m1-draft"
tier: L
related_specs: [SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001]
---

# SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001 — Migration Matrix

> **M1 Draft (2026-05-23), rescoped v0.1.1 (2026-05-30)**. M2–M5에서 manager-develop가 본 matrix를 참조하여 PRESERVE / GENERALIZE / REMOVE 정책을 적용한다. M6 chore에서 실측 결과를 반영하여 final로 승격한다.
>
> **Rescope (v0.1.1)**: kept classes = C1/C2/C4/C5/C6/C8 (NEUTRALITY-unique). **C3 (dates) / C7 (commit hash)** 는 shipped sibling SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001 (`internal/template/internal_content_leak_test.go` strict-tier `S1-internal-date` / `S2-short-sha-sentence-final`)로 **DEFERRED**. 본 matrix의 C3/C7 sections는 DEFERRED action policy로 표기되며 audit script (REQ-TNA-009)가 scan하지 않는다. Baselines re-measured 2026-05-30.
>
> **C2 narrow (v0.1.2, M3 blocker resolution)**: C2 detection을 broad `V3R[0-9]` (341 hits, 88% ID-embedded)에서 **bare-narrative `V3R[0-9]` (ID-prefix token에 직접 선행되지 않는 sigil, 7 files)** 으로 narrow. 299 ID-embedded matches (`SPEC-V3R…`/`CONST-V3R…`/`REQ-V3R…`)는 ISOLATION-001의 `C1-spec-id` leak-test class 소유 — C2는 침범하지 않는다 (cross-SPEC disjointness). 상세는 §C2 + spec.md REQ-TNA-002 § C2 detection scope. **Pre-existing package RED 인지**: run-phase baseline에서 `internal/template` package는 이미 13개 pre-existing 실패 test로 RED 상태 (본 SPEC scope 외 — spec.md §3.4). audit script는 isolation (`-run TestTemplateNeutralityAudit`)으로 검증.

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
| **Binary FAIL** | C1, C5, C6 | 1+ hit (allow-list 외) → `t.Errorf` → CI red |
| **Advisory WARN** | C2 (bare-narrative only), C4 | hit > allow-list → `t.Logf` WARN, CI green with annotations |
| **False positive PRESERVE** | C8 | 항상 PRESERVE, audit script `continue` |
| **DEFERRED to ISOLATION** | C3, C7 | NOT scanned by `template_neutrality_audit_test.go`; owned by `internal_content_leak_test.go` strict-tier (`S1-internal-date` / `S2-short-sha-sentence-final`) |

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

### C2 V3R[0-9] Bare-Narrative Dev Version Sigils — **[NARROWED v0.1.2: bare-narrative only]**

**Detection regex (bare-narrative)**: `(?<![A-Za-z0-9-])V3R[0-9]` (perl PCRE negative-lookbehind). Go audit script (RE2, no lookbehind) implements the equivalent **two-pass exclusion** — see spec.md REQ-TNA-002 § "Two-pass detection approach": Pass 1 `\bV3R[0-9]`; Pass 2 drop `SPEC-V3R`/`CONST-V3R`/`REQ-V3R` ID-embedded + preceding-rune `[A-Za-z0-9-]`.

**Why narrowed (v0.1.2 M3 blocker resolution)**: the original broad `V3R[0-9]` regex matched **341 occurrences**, of which **299 (88%) are ID-embedded substrings** inside identifiers — `SPEC-V3R[0-9]`=165, `CONST-V3R[0-9]`=130, `REQ-V3R[0-9]`=4 (orchestrator-measured 2026-05-30 HEAD `1162b0de8`). Those 299 belong to the SPEC-ID / CONST-registry-ID / REQ-ID domain **owned by sibling SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001's `C1-spec-id` leak-test class**. Targeting them would force this SPEC to sanitize SPEC-IDs (ISOLATION's job — forbidden cross-SPEC scope bleed), making the broad ≤18 / 73-file target structurally unachievable (the M3 blocker root cause). C2 therefore owns ONLY the **bare-narrative sigil** (the version mention NOT part of a larger identifier) — keeping C2 disjoint from ISOLATION, the same discipline already applied to C3/C7.

**Action policy**: case-by-case classify per PRESERVE 기준 §1.2. (NOTE: action-policy items use `  · ` prefix, NOT `- `, so they are excluded from the AC-TNA-002 allow-list bullet count per §1.2 [HARD] hygiene rule — only file-path `- ` bullets are counted.)
  · PRESERVE: rule SSOT citation, zone-registry namespace/section-header decision record, named-doctrine citation (e.g. `V3R4 Self-Evolving Harness` authoritative-SPEC reference), SPEC-ID decomposition self-check example
  · GENERALIZE: incident-specific version sigil → pattern phrasing (예: "a prior harness retirement" 표현으로 일반화)
  · REMOVE: 단순 dev-history 흔적

**Severity**: Advisory WARN (audit script logs > allow-list, CI passes)

**Baseline (orchestrator-measured 2026-05-30 at HEAD `1162b0de8`, bare-narrative grep `grep -rlP '(?<![A-Za-z0-9-])V3R[0-9]'`)**: **7 bare-narrative files** (replaces the stale 70/73 broad counts). Broad `V3R[0-9]`=341 hits; 299 ID-embedded (ISOLATION-owned, NOT a C2 target). Point-in-time; run-phase M3 re-measures the bare-narrative set before fixing.

**Post-fix expected**: ≤ 6 files (allow-list size below; M3 실측 후 확정). After `manager-develop-prompt-template.md` is GENERALIZEd, the 6 bare-narrative PRESERVE files remain.

**Allow-list (PRESERVE — bare-narrative sigils with doctrinal anchor; 6 file-path entries, computable via corrected awk `awk '/^### C2 /{f=1;next} /^### C[0-9] /{f=0} f' | grep -c '^- '` = 6)**:

- `internal/template/templates/.claude/rules/moai/core/zone-registry.md` — `V3R2`/`V3R5` namespace policy + `V3R5-NNN..NNN` CONST-registry section headers (decision-record / SSOT; bare-narrative namespace prose)
- `internal/template/templates/.claude/agents/moai/manager-spec.md` — `V3R6` segment inside the SPEC-ID Pre-Write decomposition self-check **example** (`decomposition: SPEC ✓ | V3R6 ✓ | … → PASS`); doctrinal example, PRESERVE
- `internal/template/templates/.claude/skills/moai-harness-learner/SKILL.md` — `V3R4`/`V3R3` Harness Learning Subsystem decision-record citation (authoritative harness SPEC reference, named-doctrine)
- `internal/template/templates/.claude/skills/moai/SKILL.md` — `V3R4 Self-Evolving Harness` lifecycle authoritative-SPEC citation (named-doctrine, decision record)
- `internal/template/templates/.claude/skills/moai/workflows/harness.md` — `V3R4 Self-Evolving Harness` lifecycle workflow decision-record citations (authoritative harness SPEC, named-doctrine; multiple bare-narrative mentions all doctrinal)
- `internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md` — `V3R4` meta-harness contract decision-record citation (named-doctrine)

**GENERALIZE candidate (M3 manager-develop)**: `internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md` — 2 bare-narrative hits: `W3 ↔ V3R4 HARNESS retirement` example (generalize to "a prior harness retirement") + `다른 V3R6 SPEC plan-phase artifacts` (generalize to "다른 SPEC plan-phase artifacts"). **F3 mirror caveat**: this file is on the `rule_template_mirror_test.go` byte-parity allow-list — edit both the template and the `.claude/` mirror (or verify the `.claude/` side already matches the intended generic form).

**REMOVE candidates (M3 manager-develop)**: none in the bare-narrative set — all 7 bare-narrative files are either PRESERVE (6) or GENERALIZE (1).

**ID-embedded 299 matches (NOT a C2 target — deferred to ISOLATION-001)**: `SPEC-V3R[0-9]`=165 + `CONST-V3R[0-9]`=130 + `REQ-V3R[0-9]`=4. SPEC-ID / REQ-ID sanitization is owned by ISOLATION-001's `C1-spec-id` leak-test class; CONST-registry IDs are zone-registry SSOT (legitimately preserved as the canonical registry). This SPEC's audit script does NOT count these.

### C3 ISO Date Refs (2026-0[5-9]-XX) — **[DEFERRED to SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001]**

**Detection regex**: `2026-0[5-9]` (informational only — NOT scanned by this SPEC's audit script)

**Action policy**: **DEFERRED**. The generic ISO-date class is owned by the shipped sibling SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001, whose `internal/template/internal_content_leak_test.go` strict-tier `S1-internal-date` class (`\b202[6-9]-[0-1][0-9]-[0-3][0-9]\b`, opt-in `MOAI_TEMPLATE_LEAK_STRICT=1`, tracked under its §25.1 evolution policy) enforces it. This SPEC's `template_neutrality_audit_test.go` does NOT scan C3 — re-scanning would create a second, divergent date allow-list in the same `internal/template/` Go package (dual-allow-list drift).

**Severity**: DEFERRED (owned by leak-test strict tier; NOT a NEUTRALITY binary/advisory class)

**Baseline (re-measured 2026-05-30 at HEAD `ecda4ef04`)**: 39 files (was 32 at 2026-05-23; +7 drift) — informational only.

**Post-fix expected**: n/a (enforcement owned by ISOLATION strict tier).

**Allow-list**: Not applicable — deferred. No NEUTRALITY-owned allow-list bullets (to avoid awk mis-count for adjacent kept categories). Date-class allow-listing, if needed, lives in the leak-test `skipPaths` / strict-tier evolution policy of ISOLATION-001.

### C4 feedback_/memory.md Refs — **[KEPT — NEUTRALITY-unique]**

**Detection regex**: `feedback_\|memory\.md`

**Why kept (not deferred)**: `internal/template/internal_content_leak_test.go` does NOT enforce the `feedback_` / `memory.md` *substring reference* class (default OR strict). Its C5 class enforces only memory *paths* (`~/.claude/projects/-Users-` / `.moai/backups/agent-archive-`), a disjoint pattern. Deferring C4 would silently drop enforcement, so C4 remains NEUTRALITY-owned (verified 2026-05-30).

**Action policy**: case-by-case classify per PRESERVE 기준 §1.2. (NOTE: action-policy items use `  · ` prefix, NOT `- `, so they are excluded from the AC-TNA-004 allow-list bullet count per §1.2 [HARD] hygiene rule — only file-path `- ` bullets are counted.)
  · PRESERVE: rule SSOT citation of canonical memory file (e.g., `feedback_w3_metaanalysis_lessons.md`, `feedback_large_spec_wave_split.md`, `feedback_worktree_autonomous.md`) when cited in active doctrine
  · GENERALIZE: `feedback_*.md` reference → "auto-memory" generic phrasing
  · REMOVE: incident-specific memory ref with no doctrinal anchor

**Severity**: Advisory WARN

**Baseline (re-measured 2026-05-30 at HEAD `ecda4ef04`)**: 9 files (unchanged from 2026-05-23)

**Post-fix expected**: ≤ 7 files (allow-list size, M4 실측 후 확정)

**Allow-list (PRESERVE — canonical memory citations in rule doctrine; 7 file-path entries)**:

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

**Baseline (re-measured 2026-05-30 at HEAD `ecda4ef04`)**: 3 files (was 10 at 2026-05-23; −7 partial prior cleanup)

**Post-fix expected**: 0 files

**Allow-list**: Empty (no exceptions). 3 files 모두 M4에서 remove or generic-replace 대상이다.

**Fix targets (M4 manager-develop)**: `lsp.yaml.tmpl`, `output-styles/moai/moai.md`, `agent-authoring.md`, `skill-authoring.md`, `branch-origin-protocol.md`, `moai-memory.md`, `workflows/loop.md`, `workflows/project/doc-generation.md`, `moai-meta-harness/SKILL.md`, `moai-workflow-loop/SKILL.md`, `moai-workflow-loop/references/reference.md`, `moai-workflow-loop/references/examples.md` — line-level audit는 M4에서 수행.

### C6 PR #N Refs

**Detection regex**: `PR #[0-9]+`

**Action policy**: REMOVE. 특정 PR 번호는 dev-history specific이며 template 배포 부적합. doctrine precedent 인용이 필요하면 generic phrase ("a prior round of the same rule", "the originating PR") 사용.

**Severity**: Binary FAIL

**Baseline (2026-05-23)**: 3 files

**Post-fix expected**: 0 files

**Allow-list**: Empty (no exceptions). 3 files 모두 M5에서 remove or generic-replace 대상이다.

**Fix targets (M5 manager-develop)**: `CLAUDE.md`, `quality.yaml.tmpl`, `moai-workflow-ci-loop/SKILL.md` — line-level audit는 M5에서 수행.

### C7 Commit Hash Refs — **[DEFERRED to SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001]**

**Detection regex**: informational only — NOT scanned by this SPEC's audit script. The original broad `\b[a-f0-9]{7,40}\b` proposal was FP-saturated (45 raw hits including color codes / SHA test fixtures) with no written discrimination rule (plan-audit iter-1 D3).

**Action policy**: **DEFERRED**. The commit-hash class is owned by the shipped sibling SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001, whose `internal/template/internal_content_leak_test.go` strict-tier `S2-short-sha-sentence-final` class (`\b[0-9a-f]{7,8}([\s\.,;:!?]|$)`, opt-in `MOAI_TEMPLATE_LEAK_STRICT=1`) enforces it with a deliberately conservative, false-positive-aware detector. This SPEC's `template_neutrality_audit_test.go` does NOT scan C7 — re-scanning would create a second, divergent SHA allow-list in the same `internal/template/` Go package (dual-allow-list drift).

**Severity**: DEFERRED (owned by leak-test strict tier; NOT a NEUTRALITY binary class)

**Baseline (re-measured 2026-05-30 at HEAD `ecda4ef04`)**: ~2 files — informational only.

**Post-fix expected**: n/a (enforcement owned by ISOLATION strict tier).

**Allow-list**: Not applicable — deferred. No NEUTRALITY-owned allow-list bullets. SHA-class discrimination (commit hash vs color code vs SHA test fixture) is handled by the leak-test strict tier of ISOLATION-001, which resolves the original D3 FP-saturation problem.

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
