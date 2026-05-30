---
id: SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001
title: "Template Neutrality Audit — Implementation Plan"
version: "0.1.2"
status: in-progress
created: 2026-05-23
updated: 2026-05-30
author: Author Name
priority: P1
phase: "v3.0.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "template-system, audit, plan, migration, ci-guard"
tier: L
related_specs: [SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001]
---

# SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001 — Implementation Plan

## §1 Milestones

본 Tier L SPEC은 **6 milestones (M1–M6)**로 분할된다. Milestones는 manager-develop의 standard delegation 단위이며, Section A-E (Tier L MANDATORY)로 위임된다. Wave 분할은 본 SPEC scope에서는 불필요 (작업 자체는 단순 grep+rewrite로 sequential 가능; SSE stall threshold 30+ tasks 기준 미해당).

**Rescope (v0.1.1, plan-audit iter-1 remediation)**: kept classes = **C1/C2/C4/C5/C6/C8** (NEUTRALITY-unique). C3 (dates) / C7 (commit hash)는 shipped sibling SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001의 `internal_content_leak_test.go` strict-tier classes로 **deferred** (§3 Dependencies 참조). 따라서 milestone 분배가 조정됐다: M3은 C2 유지; M4는 (C3+C4+C5)→**C4+C5만**; M5는 (C6+C7+test)→**C6+audit-script+CI-guard만** (C7 deferred).

### M1 — SPEC scope finalize + allow-list draft — **[COMPLETE, commit `367a84715`]**

**Owner**: orchestrator-direct (또는 필요 시 manager-spec follow-up)
**Status**: **COMPLETE**. `migration-matrix.md` (232 lines, 8/8 sections)가 commit `367a84715` (`chore(SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001): M1 migration-matrix.md draft`)으로 작성·머지되었다. AC-TNA-010 (matrix 8-section header count) PASS.
**Activity (완료됨)**:
- `migration-matrix.md` 초안 작성 (`.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/migration-matrix.md`) ✅
- 8 카테고리별 detection regex, action policy, allow-list draft, baseline counts 확정 ✅
- C8 false positive exclusion regex 확정 (`GOOS=(linux|windows|darwin|freebsd|openbsd|netbsd)`) ✅
- C2 V3R[0-9] PRESERVE 후보 식별 (rule SSOT cite, decision record, doctrine citation) ✅
- v0.1.1 rescope: C3/C7 matrix sections를 DEFERRED action policy로 전환 (M6 chore 또는 run-phase 진입 전 반영)
- M2~M5 분배 plan 확정 (rescope 반영) ✅

**Deliverables**: migration-matrix.md (232 lines, committed `367a84715`). Run-phase는 M2부터 진입.

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

### M3 — V3R[0-9] **bare-narrative** sigils classification + fix (C2, 7 files) — **[NARROWED v0.1.2, blocker resolution]**

**Owner**: manager-develop cycle_type=ddd Section A-E Tier L MANDATORY

**Blocker context (v0.1.2)**: M3 진입 시 broad `V3R[0-9]` (341 hits)의 ≤18 / 73-file target이 구조적으로 달성 불가능함이 발견됐다 (299/341 = 88%가 `SPEC-V3R…`/`CONST-V3R…`/`REQ-V3R…` ID-embedded, ISOLATION-001 소유). Option A (user-approved): C2를 **bare-narrative sigil (7 files)** 로 narrow. 상세는 spec.md REQ-TNA-002 § C2 detection scope + migration-matrix.md §C2.

**Activity**:
- migration-matrix.md §C2 allow-list (6 bare-narrative PRESERVE entries)에 따라 PRESERVE / GENERALIZE 적용
- PRESERVE 예 (6 files): `zone-registry.md` V3R2/V3R5 namespace + CONST section headers, `manager-spec.md` SPEC-ID decomposition self-check 예시 `V3R6`, harness `V3R4 Self-Evolving` authoritative-SPEC citations (`moai-harness-learner/SKILL.md`, `moai/SKILL.md`, `harness.md`, `moai-meta-harness/SKILL.md` — 4개가 harness PRESERVE group)
- GENERALIZE 예 (1 file): `manager-develop-prompt-template.md` 2 hits — `V3R4 HARNESS retirement` → "a prior harness retirement", `다른 V3R6 SPEC` → "다른 SPEC". **F3 mirror caveat**: 이 파일은 `rule_template_mirror_test.go` byte-parity allow-list 대상 — template + `.claude/` 양쪽 동시 수정 (또는 `.claude/` 측이 이미 generic form인지 확인 후 template만)
- REMOVE: 없음 (bare-narrative set의 7 files 모두 PRESERVE 6 + GENERALIZE 1)
- Pre-fix re-measure (run-phase): `grep -rlP '(?<![A-Za-z0-9-])V3R[0-9]' internal/template/templates/ | wc -l` (baseline 7 at 2026-05-30; M3 re-measures at run-phase HEAD)
- **DO NOT** touch the 299 ID-embedded `SPEC-V3R…`/`CONST-V3R…`/`REQ-V3R…` matches (ISOLATION-001 domain — cross-SPEC scope bleed 금지)
- AC-TNA-002 verify: post-fix `actual` ≤ allow-list count (6, computable via corrected awk; bare-narrative grep)

**Deliverables**: 7 bare-narrative files audit (6 PRESERVE + 1 GENERALIZE). Tests: AC-TNA-002 / AC-TNA-008 (isolated `-run`) / AC-TNA-013 PASS.

> **Run-phase note (v0.1.2)**: `internal/template` package는 이미 13 pre-existing 실패 test로 RED (본 SPEC scope 외 — spec.md §3.4). AC-TNA-008 검증은 package-wide green이 아니라 `go test ./internal/template/... -run TestTemplateNeutralityAudit` isolated form으로만 수행. package-wide RED은 본 SPEC closure를 block하지 않는다.

### M4 — memory + CLAUDE.local refs (C4+C5, ~12 files) — **[C3 DEFERRED, dropped from M4]**

**Owner**: manager-develop cycle_type=ddd Section A-E Tier L MANDATORY
**Activity**:
- ~~C3 dates~~ — **DEFERRED to ISOLATION-001** (leak-test strict-tier `S1-internal-date`); NOT in M4 scope (avoids dual-allow-list drift)
- C4 feedback_/memory refs 9 files: M1 allow-list 외 generalize ("auto-memory") 또는 remove. **NEUTRALITY-unique** — leak test does NOT cover the `feedback_`/`memory.md` substring class (only memory *path* class, which is disjoint)
- C5 `CLAUDE.local.md` refs 3 files (re-measured 2026-05-30, was 10): remove or generic-replace ("machine-specific configuration")
- Pre-fix re-measure (run-phase): C4 `grep -rln 'feedback_\|memory\.md'` + C5 `grep -rln 'CLAUDE\.local\.md'`
- AC-TNA-004 / AC-TNA-005 verify (AC-TNA-003 is deferred — no verification command)

**Deliverables**: C4 (9 files) + C5 (3 files) audit + modifications (overlap 가능, unique ~12 files). Tests: AC-TNA-004 / AC-TNA-005 / AC-TNA-008 / AC-TNA-013 PASS.

### M5 — PR refs + audit script + CI guard (C6+REQ-009+REQ-010) — **[C7 DEFERRED, dropped from M5]**

**Owner**: manager-develop cycle_type=ddd Section A-E Tier L MANDATORY
**Activity**:
- C6 PR #N refs 3 files: remove or generic-replace ("a prior round of the same rule")
- ~~C7 commit hash refs~~ — **DEFERRED to ISOLATION-001** (leak-test strict-tier `S2-short-sha-sentence-final`); NOT in M5 scope. This also resolves the original D3 problem (FP-saturated broad regex + dependency on a not-yet-existing test).
- 신규 `internal/template/template_neutrality_audit_test.go` 작성 (**NEW disjoint test file, NOT an extension of `internal_content_leak_test.go`** — different severity semantics + ownership lifecycle, per REQ-TNA-009 DECISION note):
  - Function `TestTemplateNeutralityAudit`
  - C1/C5/C6 binary FAIL patterns + C2/C4 WARN patterns + C8 exclusion
  - **C3/C7 NOT scanned** (owned by `internal_content_leak_test.go` strict-tier — pattern set MUST be disjoint from the leak test; no class enforced by both files)
  - allow-list driven from migration-matrix.md (Go test reads .md or duplicated Go constant)
  - Cross-platform (darwin + linux + windows) PASS
- 신규 `.github/workflows/template-neutrality-check.yaml`:
  - Trigger: `on.pull_request.paths: [internal/template/templates/**]`
  - Run: `go test ./internal/template/... -run TestTemplateNeutralityAudit`
  - Required status check 등록
- AC-TNA-006 / AC-TNA-008 / AC-TNA-009 / AC-TNA-011 verify (AC-TNA-007 is deferred — no verification command)

**Deliverables**: 5 files (3 PR refs fix + 1 new Go test + 1 new workflow yaml + migration-matrix.md ref). Tests: AC-TNA-006 / AC-TNA-008 / AC-TNA-009 / AC-TNA-011 PASS.

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

### R1 — Dev-history justification ambiguity (MEDIUM, narrowed v0.1.2)

**Issue**: bare-narrative V3R sigils (7 files) 중 어느 것이 PRESERVE 정당하고 어느 것이 GENERALIZE 대상인지 case-by-case 판정 필요. (v0.1.2 narrow로 ambiguity 크게 축소 — 299 ID-embedded matches는 ISOLATION-001 소유로 본 SPEC scope 밖, 판정 대상에서 제외.)

**Mitigation**:
- migration-matrix.md §C2 PRESERVE 기준 명문화: (a) rule SSOT / zone-registry namespace decision record, (b) named-doctrine citation (`V3R4 Self-Evolving Harness` authoritative-SPEC), (c) SPEC-ID decomposition self-check 예시
- M3 manager-develop에 6-entry allow-list + 1 GENERALIZE target explicit injection (bare-narrative set 7 files 전수 enumerate)
- bare-narrative two-pass detection (RE2 lookbehind 부재 대응)으로 ID-embedded false-positive 원천 차단
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

### Overlapping / sibling SPEC — SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001 (status `completed`)

[HARD] **SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001** (status `completed`, 2026-05-25) ships `internal/template/internal_content_leak_test.go` (16810B) in the SAME Go package (`internal/template/`) that this SPEC's audit script (REQ-TNA-009) targets. It enforces moai-adk dev-internal-content classes (SPEC IDs, REQ/AC tokens, audit citations, dates, memory paths, commit SHAs) over `internal/template/templates/**`. **This SPEC was rescoped (v0.1.1) to deconflict with it.**

**C3 / C4 / C7 ↔ ISOLATION partition** (leak-test pattern catalog verified 2026-05-30):

| NEUTRALITY class | leak-test enforcement | partition |
|---|---|---|
| C3 generic ISO date `2026-0[5-9]` | strict-tier `S1-internal-date` `\b202[6-9]-[0-1][0-9]-[0-3][0-9]\b` (opt-in `MOAI_TEMPLATE_LEAK_STRICT=1`, §25.1 evolution tier) | **DEFER to ISOLATION** |
| C4 `feedback_` / `memory.md` substring | **NOT enforced** (default OR strict). leak-test C5 enforces only memory *paths* (`~/.claude/projects/-Users-` / `.moai/backups/agent-archive-`) — disjoint pattern | **KEEP in NEUTRALITY** |
| C7 commit hash | strict-tier `S2-short-sha-sentence-final` `\b[0-9a-f]{7,8}([\s\.,;:!?]\|$)` (opt-in strict) | **DEFER to ISOLATION** |

**REQ-TNA-009 scope (NEW disjoint test, NOT an extension)**: the new `template_neutrality_audit_test.go` scans ONLY the kept classes {C1 `/Users/`, C2 `V3R[0-9]`, C4 `feedback_`/`memory.md`, C5 `CLAUDE.local.md`, C6 `PR #N`, C8 `GOOS=` exclude}. It does NOT re-scan C3/C7 (those belong to `internal_content_leak_test.go`). The two test files' pattern sets are DISJOINT — no class is enforced by both. A NEW test file (not an extension of the leak test) is chosen because the two tests have different severity semantics (NEUTRALITY C2/C4 are WARN-level advisory vs the leak test's all-FAIL) and different ownership lifecycles (NEUTRALITY is the active SPEC; ISOLATION is `completed`/frozen). This eliminates the dual-allow-list drift risk that plan-audit iter-1 D1 identified (two test files independently scanning the same tree for overlapping date/SHA classes with divergent allow-lists → flaky CI on the next template date edit).

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
