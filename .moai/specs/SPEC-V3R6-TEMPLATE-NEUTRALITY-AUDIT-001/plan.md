---
id: SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001
title: "Template Neutrality Audit — Implementation Plan"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: Author Name
priority: P1
phase: "v3.0.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "template-system, audit, plan, migration, ci-guard"
tier: L
---

# SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001 — Implementation Plan

## §1 Milestones

본 Tier L SPEC은 **6 milestones (M1–M6)**로 분할된다. Milestones는 manager-develop의 standard delegation 단위이며, Section A-E (Tier L MANDATORY)로 위임된다. Wave 분할은 본 SPEC scope에서는 불필요 (138 files이나 작업 자체는 단순 grep+rewrite로 sequential 가능; SSE stall threshold 30+ tasks 기준 미해당).

### M1 — SPEC scope finalize + allow-list draft

**Owner**: orchestrator-direct (또는 필요 시 manager-spec follow-up)
**Activity**:
- `migration-matrix.md` 초안 작성 (`.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/migration-matrix.md`)
- 8 카테고리별 detection regex, action policy, allow-list draft, baseline counts 확정
- C8 false positive exclusion regex 확정 (`GOOS=(linux|windows|darwin|freebsd|openbsd|netbsd)`)
- C2 V3R[0-9] 70 files 중 PRESERVE 후보 식별 (rule SSOT cite, decision record, doctrine citation)
- C3 dates 32 files 중 incident allow-list 식별
- M2~M5 분배 plan 확정

**Deliverables**: migration-matrix.md (~400 lines), updated progress.md (M1 status: complete)

### M2 — Critical: macOS-bias fix (C1, 4 files / 8 lines)

**Owner**: manager-develop cycle_type=ddd Section A-E Tier L MANDATORY
**Activity**:
- `worktree-integration.md` 4 lines: `/Users/user/project/...` → `$HOME/project/...` 또는 `~/project/...` 통일
- `run/context-loading.md` 3 lines: 동일 변환 (line 177은 example warning text → 보존 가능, 185/188 example는 변환)
- `moai-foundation-cc/examples.md` 1 line: `/Users/john/projects/...` → `$HOME/projects/...`
- `moai-workflow-loop/examples.md` 1 line: `/Users/project/...` → `$HOME/project/...`
- 각 변환 시 문맥 narrative (e.g., "Read X" example text) 보존
- AC-TNA-001 verify: `grep -rln '/Users/' internal/template/templates/` = 0

**Deliverables**: 4 file modifications. Tests: AC-TNA-001 / AC-TNA-008 / AC-TNA-013 PASS.

### M3 — V3R[0-9] refs classification + fix (C2, 70 files)

**Owner**: manager-develop cycle_type=ddd Section A-E Tier L MANDATORY
**Activity**:
- M1 migration-matrix.md §C2 allow-list에 따라 PRESERVE / GENERALIZE / REMOVE 적용
- PRESERVE 예: `decisions/lsp-client-choice.md` V3R5 decision record, `agent-common-protocol.md` `CONST-V3R5-NNN` registry IDs, `worktree-state-guard.md` CONST-V3R5-029
- GENERALIZE 예: "V3R6 finding F-009" → "a workflow audit finding"
- REMOVE 예: 단순 dev-history 흔적 ("V3R6 cleanup commit `fc47f31a7`로 해결")
- AC-TNA-002 verify: post-fix count ≤ allow-list count

**Deliverables**: 70 files audit + modifications (allow-list 외 모두 fix). Tests: AC-TNA-002 / AC-TNA-008 / AC-TNA-013 PASS.

### M4 — Dates + memory + CLAUDE.local refs (C3+C4+C5, 51 files)

**Owner**: manager-develop cycle_type=ddd Section A-E Tier L MANDATORY
**Activity**:
- C3 dates 32 files: M1 allow-list 외 dates를 month-year granularity로 generalize
- C4 feedback_/memory refs 9 files: rule SSOT citation 외 generalize ("auto-memory") 또는 remove
- C5 `CLAUDE.local.md` refs 10 files: remove or generic-replace ("machine-specific configuration")
- AC-TNA-003 / AC-TNA-004 / AC-TNA-005 verify

**Deliverables**: 59 file audit + modifications (overlap 가능, unique ~50-55 files). Tests: AC-TNA-003 / AC-TNA-004 / AC-TNA-005 / AC-TNA-008 / AC-TNA-013 PASS.

### M5 — PR + commit hash refs + audit script + CI guard (C6+C7+REQ-009+REQ-010)

**Owner**: manager-develop cycle_type=ddd Section A-E Tier L MANDATORY
**Activity**:
- C6 PR #N refs 3 files: remove or generic-replace ("a prior round of the same rule")
- C7 commit hash refs 2 files: replace with doctrine name / rule citation
- 신규 `internal/template/template_neutrality_audit_test.go` 작성:
  - Function `TestTemplateNeutralityAudit`
  - C1/C5/C6/C7 binary FAIL patterns + C2/C3/C4 WARN patterns + C8 exclusion
  - allow-list driven from migration-matrix.md (Go test reads .md or duplicated Go constant)
  - Cross-platform (darwin + linux + windows) PASS
- 신규 `.github/workflows/template-neutrality-check.yaml`:
  - Trigger: `on.pull_request.paths: [internal/template/templates/**]`
  - Run: `go test ./internal/template/... -run TestTemplateNeutralityAudit`
  - Required status check 등록
- AC-TNA-006 / AC-TNA-007 / AC-TNA-009 / AC-TNA-010 / AC-TNA-011 verify

**Deliverables**: 5+ files (3 PR refs fix + 2 commit hash fix + 1 new Go test + 1 new workflow yaml + migration-matrix.md ref). Tests: AC-TNA-006 / AC-TNA-007 / AC-TNA-009 / AC-TNA-010 / AC-TNA-011 PASS.

### M6 — Migration matrix finalize + Template-First guideline + status implemented v0.2.0

**Owner**: orchestrator-direct chore
**Activity**:
- migration-matrix.md M1 draft → final (M2~M5 실측 결과 반영, post-fix counts 기록)
- `CLAUDE.local.md` §2 Template-First Rule에 "Acceptable Content Range for Templates" subsection 추가 (REQ-TNA-012)
- 향후 contributor checklist 추가 (REQ-TNA-013) — `CLAUDE.local.md` §22 또는 별도 subsection
- spec.md frontmatter `status: draft → implemented`, `version: 0.1.0 → 0.2.0`, `updated: <M6 date>`
- progress.md 최종 작성 (M1~M6 evidence 보존, AC-TNA-001~013 PASS 증거)
- AC-TNA-012 / AC-TNA-013 verify

**Deliverables**: migration-matrix.md final, CLAUDE.local.md guideline update, spec.md status implemented, progress.md complete.

## §2 Risks

### R1 — Dev-history justification ambiguity (HIGH)

**Issue**: 70 V3R refs 중 어느 것이 PRESERVE 정당하고 어느 것이 REMOVE 대상인지 case-by-case 판정 필요. 잘못 PRESERVE하면 dev-incident가 distribute됨.

**Mitigation**:
- M1 migration-matrix.md PRESERVE 기준 명문화: (a) rule SSOT citation, (b) decision record file (`.moai/decisions/*.md`), (c) zone-registry.md CONST-V3R5-NNN entries
- M3 manager-develop에 PRESERVE 기준 + allow-list explicit injection
- AC-TNA-002 WARN-level finding으로 표면화하여 미식별 항목 post-hoc 검출

### R2 — False positive in C8 detection

**Issue**: `GOOS=` regex가 Go env var이 아닌 다른 문맥 (예: 색상 코드, 무관한 hex)에 매칭될 가능성.

**Mitigation**:
- M1 regex narrow: `GOOS=(linux|windows|darwin|freebsd|openbsd|netbsd)` (값까지 명시)
- AC-TNA-011 binary verify: post-fix 후 audit script가 보존 4 hits에 대해 FAIL 발생시키지 않음
- AC-TNA-008은 darwin + linux + windows cross-platform PASS 검증

### R3 — Cross-Wave conflict with Sprint 1 in-progress SPECs (MEDIUM)

**Issue**: Sprint 1 잔여 SPEC (SKILL-GEARS-ALIGN-001 머지 대기, CODE-COMMENTS-EN-001 Wave 4 진행) 작업 중 본 SPEC이 동일 template file 수정하면 conflict.

**Mitigation**:
- M2~M5 진행 전 main 최신 fetch + Sprint 1 잔여 SPEC 머지 상태 확인
- 본 SPEC PRESERVE list에 Sprint 1 in-progress SPEC dirs 명시 (Section §2 Non-Goals 참조)
- manager-develop B9 git remote sync prohibition rule 인식 (mid-task rebase autonomous 금지) — `feedback_w3_metaanalysis_lessons.md` 참조

### R4 — docs-site 4-locale parity impact

**Issue**: 본 SPEC이 `docs-site/content/{en,ko,ja,zh}/**` 영향 시 4-locale parity ratio 변동 가능.

**Mitigation**:
- M1 baseline에서 docs-site/ 파일 수 측정 (현재 audit는 `internal/template/templates/`에 집중, docs-site는 sibling)
- 본 SPEC scope 한정: `internal/template/templates/**`만, `docs-site/**`는 별도 SPEC scope
- docs-site touch 미발생 시 별도 parity verification 불필요 (spec.md §2 Non-Goals에서 docs-site 제외)

### R5 — Audit script performance (LOW)

**Issue**: `internal/template/templates/**` 전체 grep 스캔이 느릴 가능성.

**Mitigation**:
- Go test에서 `filepath.WalkDir` + bufio.Scanner 사용 (regex 컴파일 1회)
- 138 files / ~50K LOC 규모는 1초 이내 예상
- 측정 시 timeout 60s buffer 설정

## §3 Dependencies

### Prerequisites (merged ✅)

- **SPEC-V3R6-RULES-PATH-SCOPE-001** (PR #1047 머지 `7d94cbc54`, 2026-05-23) — 4 rule path-scoped 작업 완료, 본 SPEC이 추가 path-scoped rule 작성 시 동일 패턴 사용
- **SPEC-V3R6-CODE-COMMENTS-EN-001** Wave 3 implemented (`bac893173`, 2026-05-23) — Go code Korean → English translation 진행 중, 본 SPEC scope와 disjoint (본 SPEC은 markdown/template, code-comments는 Go code)
- **fc47f31a7** Critical 4 fix commit — `moai/SKILL.md` 2 lines + `spec-frontmatter-schema.md` 2 lines + sprint-round-naming.md SSOT 신규

### Concurrent (cross-Wave caveat)

- **SPEC-V3R6-SKILL-GEARS-ALIGN-001** (plan complete `81d47a445`, 머지 대기) — GEARS notation alignment, 본 SPEC scope (markdown sanitization)과 disjoint
- **SPEC-V3R6-CODE-COMMENTS-EN-001** Wave 4+ (pkg/ + remaining internal/*) — Go code 한정, 본 SPEC `internal/template/templates/` disjoint

### Out of dependency scope

- Sprint 1 in-progress drafts (AGENT-MODEL-ROUTING / PROMPT-CACHE / HOOK-OBSERVE-OPT-IN / HOOK-ASYNC-EXPAND) — 본 SPEC PRESERVE list에 포함, 직접 수정 금지

### Out of Scope

본 plan-phase가 명시적으로 **수행하지 않는** 작업:

- Sprint 1 in-progress SPEC directories (`.moai/specs/SPEC-V3R6-AGENT-MODEL-ROUTING-001/`, `SPEC-V3R6-HOOK-ASYNC-EXPAND-001/`, `SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001/`, `SPEC-V3R6-PROMPT-CACHE-001/`) 직접 수정
- `docs-site/content/{en,ko,ja,zh}/book/` untracked 디렉토리 정리
- `internal/hook/.moai/` working dir leak 정리
- `.moai/specs/.moai/` symlink/dup 정리
- `settings.local.json` runtime-managed 값
- Go code files (`internal/*.go`, `pkg/*.go`) dev-incident 주석 — `internal/template/templates/`만 scope
- Template Go template variable rendering logic 정정 (`internal/template/render_*.go`)
- User-facing string literals preserved per EXCL-CCE-001 pattern (Korean nodejs install URLs, GitLab descriptions, BODP `(권장)` literal)
- `internal/template/templates/` 외 다른 maintenance work (CI fixes, dependency updates, unrelated rule changes)
