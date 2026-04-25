---
id: SPEC-DESIGN-CONST-AMEND-001
version: 0.3.0
status: draft
created_at: 2026-04-20
updated_at: 2026-04-21
author: moai-adk-go
priority: High
labels: [design, constitution, frozen, amendment, brand-context, design-brief]
issue_number: null
---

# SPEC-DESIGN-CONST-AMEND-001: Design Constitution Section 3 개정

## HISTORY

- 2026-04-21 v0.3.0: plan-auditor iteration 2 FAIL 후속 수정. N1 BLOCKING (AC-5 vs Section 3.2 중복) 해결: Section 3.2 기계 산출물 불릿에서 parenthetical 파일명 목록 제거하여 canonical list 단일화, REQ-012 신설로 Section 3.2 내 각 예약 파일명 정확히 1회 출현 강제, AC-5 기대값을 exact 1로 유지. N2 MEDIUM (AC-9 single-line regex 과제약) 해결: `### 3.1/3.2/3.3` heading 아래 FROZEN 진술을 per-subsection 개별 검증(bullet 스타일 허용)으로 완화. N3 LOW (Expected 주석) 해결: AC-3, AC-5, AC-6, AC-7, AC-8, AC-9, AC-10, AC-11의 모든 `# Expected: >= N` / `# Expected: N` 주석을 `test "$(...)" -ge N` 또는 `[ "$(...)" -eq N ]` 결정적 shell 어서션으로 변환.
- 2026-04-21 v0.2.0: plan-auditor iteration 1 FAIL 후속 수정. Frontmatter MoAI 표준 정렬, REQ-AC traceability 완성, rationale 보강, 감사 리포트 모든 결함 해결. F1 (예약 파일명 통일: `tokens.json`, `components.json`, `assets/`, `import-warnings.json`, `brief/BRIEF-*.md`), F2 (pencil-plan.md 정식 명명), F3 (frontmatter), F4 (육안 검토 → 결정적 검증), F5 (Section 3.3도 FROZEN), F6 (절대경로 제거), F7 (REQ-011 제목 변경 신설), F8 (토큰 예산 기본값 inline fallback), F9 (AC-5 정확한 매치 수), F10/F11 (REQ 트리거/HISTORY 시각 해상도) 반영. Target file/section/classification은 frontmatter body 하단의 Metadata block으로 이동.
- 2026-04-20 v0.1.0: SPEC 최초 작성. 대상: `.claude/rules/moai/design/constitution.md` Section 3 v3.2.0 → v3.3.0 개정. FROZEN 영역 정식 수정 절차에 따라 인간 승인 필수 (Section 2 요구사항).

---

## Metadata Block

- Target File: `.claude/rules/moai/design/constitution.md`
- Target Section: Section 3 (plus Section 2 FROZEN list, HISTORY, Version footer)
- Classification: FROZEN_AMENDMENT
- Related SPECs: SPEC-DESIGN-DOCS-001, SPEC-DESIGN-ATTACH-001, SPEC-DB-CMD-001

---

## Background (배경)

### moai-studio 사례 연구

사용자는 현재 `~/moai/moai-studio/.moai/design/`에서 4개의 인간 저작 파일을 활용하여 반복적 UI/UX 리디자인 작업을 수행 중이다:

| 파일명 (moai-studio 원본) | 라인 수 | 용도 | 본 SPEC의 정식 명명 |
|---------------------------|---------|------|---------------------|
| research.md | 214L | 사용자 리서치 및 경쟁사 분석 | research.md (동일) |
| system.md | 341L | 디자인 시스템 정의 | system.md (동일) |
| spec.md | 528L | 화면별 상세 스펙 | spec.md (동일) |
| pencil-redesign-plan.md | 315L | Pencil MCP 기반 `.pen` 파일 배치 작업 계획 | **pencil-plan.md** (rename) |

**정식 명명에 관한 주석 (F2 대응)**: 본 SPEC이 프로젝트 전역 표준으로 승격되면서 moai-studio의 `pencil-redesign-plan.md`는 `pencil-plan.md`로 단축 개명된다. 헌법 Section 3.2 예약 파일명 목록 및 우선순위 문자열에서 사용되는 **정식 명명은 오직 `pencil-plan.md`뿐**이다. 기존 moai-studio 프로젝트는 첫 `/moai design` 실행 시 마이그레이션 훅이 파일명을 재작성한다 (구현은 SPEC-DESIGN-DOCS-001 범위).

이 패턴을 MoAI-ADK-Go의 프로젝트 전역 표준으로 승격하기 위해 다음 4가지를 수행해야 한다:

1. `.moai/design/` 폴더 구조를 템플릿에 정의
2. `.moai/design/*.md` 파일들을 `/moai design` 워크플로우 컨텍스트에 자동 로드
3. `.pen` 파일 배치 작업을 위한 Pencil MCP 통합
4. 기존 `moai-workflow-design-import` 산출물 (`tokens.json`, `components.json`, `assets/`, `import-warnings.json`)과의 공존

### FROZEN 개정 요구 근거

`.claude/rules/moai/design/constitution.md` Section 2는 해당 파일 전체를 **FROZEN Zone**으로 명시한다. 따라서 Section 3을 수정하려면 다음 절차가 필수이다:

- 정식 SPEC 작성 (본 문서)
- 인간 개발자 승인 (FROZEN 규정에 따라)
- 변경 범위 최소화 (Scope Discipline)

본 SPEC은 위 절차를 충족시키기 위한 High 우선순위의 공식 개정 제안서이다.

### 현재 Section 3 원문 (verbatim)

```markdown
## 3. Brand Context as Constitutional Principle

Brand context is not optional decoration. It is a constitutional constraint that flows through every phase:

- [HARD] manager-spec MUST load brand context before generating BRIEF documents
- [HARD] moai-domain-copywriting MUST adhere to brand voice, tone, and terminology from brand-voice.md
- [HARD] moai-domain-brand-design MUST use brand color palette, typography, and visual language from visual-identity.md
- [HARD] expert-frontend MUST implement design tokens derived from brand context
- [HARD] evaluator-active MUST score brand consistency as a must-pass criterion

Brand context is stored in `.moai/project/brand/` and initialized through the brand interview process on first run. Context updates require explicit user approval.
```

### 제안된 개정 Section 3 (v3.3.0)

```markdown
## 3. Brand Context and Design Brief as Constitutional Principles

### 3.1 Brand Context (constitutional parent)

Brand context is not optional decoration. It is a constitutional constraint that flows through every phase:

- [HARD] manager-spec MUST load brand context before generating BRIEF documents
- [HARD] moai-domain-copywriting MUST adhere to brand voice, tone, and terminology from brand-voice.md
- [HARD] moai-domain-brand-design MUST use brand color palette, typography, and visual language from visual-identity.md
- [HARD] expert-frontend MUST implement design tokens derived from brand context
- [HARD] evaluator-active MUST score brand consistency as a must-pass criterion

Brand context is stored in `.moai/project/brand/` and initialized through the brand interview process on first run. Context updates require explicit user approval.

### 3.2 Design Brief (execution scope)

Iteration-specific design briefs are stored in `.moai/design/`:

- [HARD] `/moai design` MUST auto-load human-authored design documents (research.md, system.md, spec.md, pencil-plan.md) when present and not _TBD_
- [HARD] Design briefs MUST NOT override brand context — brand remains the constitutional parent
- [HARD] `moai-workflow-design-import` continues to write machine-generated artifacts to `.moai/design/`; the exact set of reserved file paths is enumerated below — human-authored files must not collide with them
- [HARD] Reserved file paths (canonical list): `tokens.json`, `components.json`, `assets/`, `import-warnings.json`, `brief/BRIEF-*.md`
- [HARD] Token budget for auto-loading is bounded by `.moai/config/sections/design.yaml` `design_docs.token_budget`; when the key is absent, the system MUST default to 20000
- [HARD] Priority order when truncation is needed: spec.md > system.md > research.md > pencil-plan.md

### 3.3 Relationship

- Brand (`.moai/project/brand/`) = WHO the brand is (long-lived, rarely changes)
- Design (`.moai/design/`) = WHAT each iteration produces (per-project, evolves with redesign cycles)
- When both are present, brand constraints win on conflict.
```

---

## Requirements (EARS)

본 SPEC은 다음 12개의 EARS 형식 요구사항을 정의한다. 모든 요구사항은 `.claude/rules/moai/design/constitution.md` 파일의 Section 3 개정 작업에 적용된다.

### REQ-CONST-AMEND-001 (Event-Driven)

**WHEN** the amended `constitution.md` is committed, Section 3.1 **SHALL** contain all five [HARD] rules from the v3.2.0 Section 3 verbatim (manager-spec brand context load, moai-domain-copywriting adherence, moai-domain-brand-design visual language, expert-frontend design tokens, evaluator-active must-pass brand scoring).

**Rationale**: 기존 5개 [HARD] 규칙은 헌법의 핵심 제약이므로 단 한 글자도 변경되어서는 안 된다. 개정은 "추가"이지 "교체"가 아니다. Post-commit 관측 가능 서술어로 표현하여 mechanical 검증 가능하게 한다 (F10 대응).

### REQ-CONST-AMEND-002 (Event-Driven)

**WHEN** the amended `constitution.md` is committed, Section 3.2 **SHALL** exist as a new subsection while Sections 1, 2 (except the FROZEN list entry required by REQ-009), 4-14 remain byte-identical to their v3.2.0 content.

**Rationale**: Scope Discipline 원칙. 개정 범위는 Section 3과 그에 종속된 메타데이터(Section 2 FROZEN 목록, HISTORY, Version footer)로 엄격히 제한한다.

### REQ-CONST-AMEND-003 (Event-Driven)

**WHEN** the amended `constitution.md` is committed, Section 3.2 **SHALL** list the canonical reserved file paths verbatim: `tokens.json`, `components.json`, `assets/`, `import-warnings.json`, `brief/BRIEF-*.md`.

**Rationale**: `.moai/design/` 폴더에 인간 저작 `.md` 파일과 기계 생성 `.json`/`assets/` 산출물이 공존하므로, 예약 경로를 명시하여 충돌을 사전 차단한다. F1 해결: REQ-003, Section 3.2 본문, AC-5 모두 동일 canonical 목록을 사용한다.

### REQ-CONST-AMEND-004 (Event-Driven)

**WHEN** auto-loading design briefs, the system **SHALL** enforce the token budget declared in `.moai/config/sections/design.yaml` key `design_docs.token_budget`. **IF** the key is absent or unparseable, **THEN** the system **SHALL** default to 20000.

**Rationale**: 디자인 브리프 4개 파일 합산 1398 라인은 토큰 예산 초과 위험이 있다. 설정 기반 예산 강제는 컨텍스트 폭주를 방지한다. Inline fallback으로 SPEC-DESIGN-DOCS-001 머지 선후 무관하게 본 SPEC이 stand-alone으로 enforceable하도록 보완한다 (F8 대응).

### REQ-CONST-AMEND-005 (Event-Driven)

**WHEN** token budget is exceeded, the system **SHALL** truncate in priority order verbatim: `spec.md > system.md > research.md > pencil-plan.md`.

**Rationale**: 구현 주도 문서인 spec.md와 system.md가 가장 높은 우선순위를 갖는다. research.md는 배경 정보, pencil-plan.md는 도구 특화 계획이므로 뒷순위로 배치한다.

### REQ-CONST-AMEND-006 (State-Driven)

**WHILE** brand context is being loaded, the system **SHALL NOT** permit design briefs to override brand constraints on conflict.

**Rationale**: 브랜드는 헌법적 부모(constitutional parent), 디자인 브리프는 반복 단위 실행 범위(execution scope)이다. 충돌 시 브랜드가 승리한다는 계층 구조는 Section 3.3에 명시된다.

### REQ-CONST-AMEND-007 (Event-Driven)

**WHEN** the amended `constitution.md` is committed, the file's HISTORY block **SHALL** contain at least two entries: the existing `2026-04-20` relocation entry verbatim, and a new entry distinguished either by ISO-8601 time suffix (e.g., `2026-04-20T12:00Z`) or by explicit SPEC prefix (e.g., `2026-04-20 (SPEC-DESIGN-CONST-AMEND-001)`) noting Section 3 → v3.3.0.

**Rationale**: 헌법 파일 자체의 HISTORY 블록은 추적성(Trackable) 원칙의 핵심이다. 개정 내역이 HISTORY에 기록되지 않으면 미래의 감사 추적이 불가능하다. 동일 날짜 중복 시 시각 해상도 또는 SPEC 식별자로 구분하여 auditor-friendly하게 한다 (F11 대응).

### REQ-CONST-AMEND-008 (Event-Driven)

**WHEN** the amended `constitution.md` is committed, the file's Version footer **SHALL** contain the literal string `Version: 3.3.0` and **SHALL NOT** contain `Version: 3.2.0`.

**Rationale**: 시맨틱 버저닝. 새 기능(Section 3.2 추가)은 minor bump에 해당한다. 3.2.0 → 3.3.0.

### REQ-CONST-AMEND-009 (Event-Driven)

**WHEN** the amended `constitution.md` is committed, the FROZEN zone list in Section 2 **SHALL** explicitly cover Section 3.1, Section 3.2, AND Section 3.3 (all three subsections FROZEN).

**Rationale**: Section 3이 세 개의 하위 섹션으로 분리되면, 기존 "Section 3 전체가 FROZEN" 진술은 모호해질 수 있다. 세 하위 섹션이 모두 FROZEN임을 명시적으로 선언한다. Section 3.3은 brand-wins-on-conflict 규칙(실질적 헌법 주장)을 인코딩하므로 FROZEN 바깥에 두면 개정 절차 없이 수정될 위험이 있다 (F5 대응).

### REQ-CONST-AMEND-010 (Unwanted Behavior)

**IF** any file outside `.claude/rules/moai/design/constitution.md` is modified during this amendment, **THEN** the change **SHALL** be rejected at review time.

**Rationale**: 본 SPEC의 유일한 대상 파일은 `constitution.md`이다. drive-by refactor나 연쇄 수정은 FROZEN 개정의 범위 원칙을 위반하므로 리뷰 단계에서 즉시 거부되어야 한다.

### REQ-CONST-AMEND-011 (Event-Driven)

**WHEN** the amended `constitution.md` is committed, the Section 3 heading **SHALL** be renamed from `## 3. Brand Context as Constitutional Principle` to `## 3. Brand Context and Design Brief as Constitutional Principles` (plural, adding "and Design Brief").

**Rationale**: 제목 변경은 Section 3이 Brand Context 단독이 아닌 Brand + Design Brief 복합 규범을 다룸을 신호한다. 이전 SPEC v0.1.0에서는 IN SCOPE에만 명시되고 REQ로 뒷받침되지 않아 traceability 공백이 있었다. REQ로 승격하여 AC-8이 결정적으로 검증한다 (F7 대응).

### REQ-CONST-AMEND-012 (Event-Driven)

**WHEN** the amended `constitution.md` is committed, each canonical reserved file path (`tokens.json`, `components.json`, `assets/`, `import-warnings.json`, `brief/BRIEF-*.md`) **SHALL** appear exactly once within the Section 3.2 text range (from the `### 3.2` heading up to but not including the `### 3.3` heading) and **SHALL NOT** appear anywhere else in the file.

**Rationale**: N1 BLOCKING 해결 (iteration 2 audit). 이전 Section 3.2 본문은 "machine-generated artifacts (tokens.json, components.json, assets/, import-warnings.json)" parenthetical 열거와 바로 다음 줄의 canonical list 두 곳에서 4개 파일명을 중복 출현시켜 AC-5의 exact-1 기대와 충돌했다. REQ-012는 canonical list를 단일 source of truth로 고정하고, 구현 단계에서 본문의 parenthetical 열거가 재도입되지 않도록 mechanical 어서션 근거를 제공한다. AC-5의 `grep -F -c` exact-1 기대와 `awk` Section-3.2-scoped 어서션이 REQ-012의 직접 검증 수단이다.

---

## Acceptance Criteria

본 SPEC이 `/moai run`을 통해 정확히 구현되었는지 검증하는 12개의 결정적 기준을 정의한다. 각 AC는 최소 1개 이상의 REQ와 대응되며, 모두 결정적 shell 명령 또는 linter 기반으로 검증 가능하다 (F4 대응). 모든 수치 어서션은 `test "$(...)" -ge N` 또는 `[ "$(...)" -eq N ]` 형식의 exit-code 기반 비교로 인코딩되어 CI가 mechanical pass/fail 게이트로 활용할 수 있다 (N3 대응).

REQ ↔ AC 추적 매트릭스:

| REQ | 대응 AC |
|-----|---------|
| REQ-001 (5개 [HARD] 규칙 verbatim) | AC-9 |
| REQ-002 (기타 Section 불변) | AC-1, AC-4 |
| REQ-003 (예약 파일명 verbatim) | AC-5 |
| REQ-004 (토큰 예산 참조 + 기본값 20000) | AC-10 |
| REQ-005 (우선순위 순서) | AC-6 |
| REQ-006 (brand wins on conflict) | AC-11 |
| REQ-007 (HISTORY 2개 이상 entry) | AC-3 |
| REQ-008 (Version 3.3.0) | AC-8 |
| REQ-009 (FROZEN 3.1, 3.2, 3.3) | AC-9 |
| REQ-010 (파일 단일성) | AC-4 |
| REQ-011 (제목 변경) | AC-7 |
| REQ-012 (Section 3.2 내 예약 파일명 exactly-once) | AC-5, AC-12 |

### AC-1: Diff 범위 제한

**기준**: constitution.md v3.2.0과 v3.3.0 간의 diff는 오직 다음 4개 영역만 포함해야 한다:

1. Section 3 (확장)
2. Section 2 FROZEN 목록 (업데이트)
3. HISTORY 블록 (신규 entry 추가)
4. Version footer (3.2.0 → 3.3.0)

**검증 방법** (결정적, 프로젝트 루트에서 실행, N3 exit-code 기반):
```bash
[ "$(git diff --name-only HEAD~1 HEAD -- .claude/rules/moai/design/)" = ".claude/rules/moai/design/constitution.md" ] && echo PASS || echo FAIL
# Expected: PASS (exit 0 — diff contains exactly constitution.md)
```

### AC-2: Section 3 하위 구조 검증

**기준**: 개정 후 `grep -c "^### 3\." constitution.md` 명령은 **3**을 반환해야 한다 (Section 3.1, 3.2, 3.3 각 1회).

**검증 방법** (결정적, F6 절대경로 제거 + N3 exit-code 기반):
```bash
[ "$(grep -c '^### 3\.' .claude/rules/moai/design/constitution.md)" -eq 3 ] && echo PASS || echo FAIL
# Expected: PASS (exit 0 — exactly 3 subsection headings: 3.1, 3.2, 3.3)
```

### AC-3: HISTORY 블록 보존 및 확장

**기준**: `.claude/rules/moai/design/constitution.md` HISTORY 블록은 2026-04-20 날짜의 entry를 **2개 이상** 포함하고, 신규 entry에 `v3.3.0` 마커가 verbatim으로 포함되어야 한다.

**검증 방법** (결정적, N3 대응 — 모든 수치 어서션이 exit-code 기반):
```bash
# 날짜 entry 2개 이상 존재
test "$(grep -c '^- 2026-04-20' .claude/rules/moai/design/constitution.md)" -ge 2 && echo PASS || echo FAIL
# Expected: PASS (exit 0)

# v3.3.0 마커 존재
test "$(grep -F -c 'v3.3.0' .claude/rules/moai/design/constitution.md)" -ge 1 && echo PASS || echo FAIL
# Expected: PASS (exit 0)

# 기존 relocation entry 보존 (regression check, 정확히 1회)
[ "$(grep -F -c 'Relocated from' .claude/rules/moai/design/constitution.md)" -eq 1 ] && echo PASS || echo FAIL
# Expected: PASS (exit 0)
```

### AC-4: 파일 단일성 보장

**기준**: `.claude/rules/moai/` 하위의 다른 어떤 `.md` 파일도 수정되지 않아야 한다.

**검증 방법** (결정적, N3 exit-code 기반 — 다른 파일이 없으면 pipeline은 빈 출력):
```bash
offenders=$(git diff --name-only HEAD~1 HEAD .claude/rules/moai/ | grep -v "^\.claude/rules/moai/design/constitution\.md$" | wc -l | tr -d ' ')
[ "$offenders" -eq 0 ] && echo PASS || echo FAIL
# Expected: PASS (exit 0 — zero offending files outside constitution.md)
```

### AC-5: 예약 파일명 verbatim 등재 (정확한 매치 수)

**기준**: 다음 canonical 예약 파일명이 Section 3.2 범위 내에서 **정확히 각 1회씩** 나타나야 한다:

- `tokens.json`
- `components.json`
- `assets/`
- `import-warnings.json`
- `brief/BRIEF-*.md`

**검증 방법** (결정적, F9 대응 + N1 해결 + N3 exit-code 기반):
```bash
# 전체 파일 기준 각 토큰 출현 횟수 = exactly 1 (REQ-012 단일 출처 보장)
[ "$(grep -F -c 'tokens.json' .claude/rules/moai/design/constitution.md)" -eq 1 ] && echo PASS || echo FAIL
[ "$(grep -F -c 'components.json' .claude/rules/moai/design/constitution.md)" -eq 1 ] && echo PASS || echo FAIL
[ "$(grep -F -c 'assets/' .claude/rules/moai/design/constitution.md)" -eq 1 ] && echo PASS || echo FAIL
[ "$(grep -F -c 'import-warnings.json' .claude/rules/moai/design/constitution.md)" -eq 1 ] && echo PASS || echo FAIL
[ "$(grep -F -c 'brief/BRIEF-*.md' .claude/rules/moai/design/constitution.md)" -eq 1 ] && echo PASS || echo FAIL
# Expected: PASS x5 (exit 0 each)

# 네거티브 주장: 각 파일명이 Section 3.2 내부(### 3.2 heading ~ 다음 ### 3.3 heading 사이)에 정확히 1회
for f in tokens.json components.json assets/ import-warnings.json 'brief/BRIEF-*.md'; do
  count=$(awk '/^### 3\.2/,/^### 3\.3/' .claude/rules/moai/design/constitution.md | grep -F -c "$f")
  [ "$count" -eq 1 ] && echo "PASS $f" || echo "FAIL $f (count=$count)"
done
# Expected: PASS x5 lines
```

### AC-6: 우선순위 순서 verbatim 등재

**기준**: 다음 문자열이 Section 3.2에 **verbatim**으로 나타나야 한다:

`spec.md > system.md > research.md > pencil-plan.md`

**검증 방법** (결정적, N3 exit-code 기반):
```bash
[ "$(grep -F -c 'spec.md > system.md > research.md > pencil-plan.md' .claude/rules/moai/design/constitution.md)" -eq 1 ] && echo PASS || echo FAIL
# Expected: PASS (exit 0)
```

### AC-7: Markdown 구조 및 제목 변경 검증

**기준**: 개정 후 constitution 파일은 결정적 검증에서 다음을 모두 통과해야 한다:
1. heading 계층 `## 3.` → `### 3.1` → `### 3.2` → `### 3.3` 순서가 파일 내에서 정확히 1회씩 존재
2. Section 3 제목이 `## 3. Brand Context and Design Brief as Constitutional Principles`로 변경
3. 구 제목 `## 3. Brand Context as Constitutional Principle`은 존재하지 않음

**검증 방법** (결정적, F4 육안 검토 제거 + N3 exit-code 기반):
```bash
# 제목 변경 (REQ-011)
[ "$(grep -F -c '## 3. Brand Context and Design Brief as Constitutional Principles' .claude/rules/moai/design/constitution.md)" -eq 1 ] && echo PASS || echo FAIL
[ "$(grep -F -c '## 3. Brand Context as Constitutional Principle' .claude/rules/moai/design/constitution.md)" -eq 0 ] && echo PASS || echo FAIL
# Expected: PASS x2 (구 제목 제거; "Principles"와의 부분 일치 방지를 위해 -F 사용)

# heading 순서 검증: 정확한 정규 시퀀스가 파일에 등장
# 기대 순서: ## 3. ..., ### 3.1 ..., ### 3.2 ..., ### 3.3 ...
order=$(grep -nE "^(## 3\. |### 3\.[123] )" .claude/rules/moai/design/constitution.md | awk -F':' '{print $1}')
prev=0
ok=1
for n in $order; do
  if [ "$n" -le "$prev" ]; then ok=0; fi
  prev=$n
done
[ "$ok" -eq 1 ] && echo PASS || echo FAIL
# Expected: PASS (line numbers strictly ascending)

# Markdown linter (pinned config)
markdownlint .claude/rules/moai/design/constitution.md && echo PASS || echo FAIL
# Expected: PASS (exit code 0 — zero heading errors)
```

### AC-8: Version Footer 업데이트

**기준**: Version footer가 정확히 `Version: 3.3.0`으로 변경되고, `Version: 3.2.0`은 잔존하지 않아야 한다.

**검증 방법** (결정적, N3 exit-code 기반):
```bash
[ "$(grep -F -c 'Version: 3.3.0' .claude/rules/moai/design/constitution.md)" -eq 1 ] && echo PASS || echo FAIL
[ "$(grep -F -c 'Version: 3.2.0' .claude/rules/moai/design/constitution.md)" -eq 0 ] && echo PASS || echo FAIL
# Expected: PASS x2 (exit 0 each)
```

### AC-9: FROZEN 목록 확장 및 HARD 규칙 보존

**기준**: Section 2 FROZEN 목록이 3.1, 3.2, 3.3 세 하위 섹션을 모두 명시하고, Section 3.1 본문이 v3.2.0의 5개 [HARD] 규칙을 verbatim 포함해야 한다.

**검증 방법** (결정적, N2 loosening + N3 exit-code 기반):
```bash
# FROZEN 목록에 3.1, 3.2, 3.3 모두 명시 — bullet-per-subsection 또는 single-line 양쪽 허용
# N2 해결: per-subsection 개별 검증으로 bullet 스타일(`- [FROZEN] Section 3.1 ...`) 포용.
# Section 2 FROZEN 범위를 `## 2.`부터 `## 3.` 직전까지 awk로 추출한 후, 각 subsection 레퍼런스가 최소 1회 등장하는지 확인
frozen_zone=$(awk '/^## 2\./,/^## 3\./' .claude/rules/moai/design/constitution.md)
for sub in "Section 3.1" "Section 3.2" "Section 3.3"; do
  count=$(printf '%s\n' "$frozen_zone" | grep -F -c "$sub")
  [ "$count" -ge 1 ] && echo "PASS $sub" || echo "FAIL $sub (count=$count)"
done
# Expected: PASS x3 lines

# 5개 HARD 규칙 verbatim 존재 — 정확히 1회씩
[ "$(grep -F -c 'manager-spec MUST load brand context before generating BRIEF documents' .claude/rules/moai/design/constitution.md)" -eq 1 ] && echo PASS || echo FAIL
[ "$(grep -F -c 'moai-domain-copywriting MUST adhere to brand voice, tone, and terminology from brand-voice.md' .claude/rules/moai/design/constitution.md)" -eq 1 ] && echo PASS || echo FAIL
[ "$(grep -F -c 'moai-domain-brand-design MUST use brand color palette, typography, and visual language from visual-identity.md' .claude/rules/moai/design/constitution.md)" -eq 1 ] && echo PASS || echo FAIL
[ "$(grep -F -c 'expert-frontend MUST implement design tokens derived from brand context' .claude/rules/moai/design/constitution.md)" -eq 1 ] && echo PASS || echo FAIL
[ "$(grep -F -c 'evaluator-active MUST score brand consistency as a must-pass criterion' .claude/rules/moai/design/constitution.md)" -eq 1 ] && echo PASS || echo FAIL
# Expected: PASS x5 (exit 0 each)
```

### AC-10: Token Budget 참조 및 Fallback

**기준**: Section 3.2 본문이 `design.yaml` 및 `design_docs.token_budget` 키를 verbatim 참조하고, 기본값 `20000`이 문자열로 등장해야 한다.

**검증 방법** (결정적, N3 exit-code 기반):
```bash
test "$(grep -F -c 'design_docs.token_budget' .claude/rules/moai/design/constitution.md)" -ge 1 && echo PASS || echo FAIL
test "$(grep -F -c '.moai/config/sections/design.yaml' .claude/rules/moai/design/constitution.md)" -ge 1 && echo PASS || echo FAIL
test "$(grep -E -c 'default[s]?\s+to\s+20000|default\s+20000' .claude/rules/moai/design/constitution.md)" -ge 1 && echo PASS || echo FAIL
# Expected: PASS x3 (exit 0 each)
```

### AC-11: Brand Wins on Conflict 명시

**기준**: Section 3.3 본문에 브랜드가 충돌 시 승리한다는 규칙이 verbatim으로 명시되어야 한다.

**검증 방법** (결정적, N3 exit-code 기반):
```bash
[ "$(grep -F -c 'brand constraints win on conflict' .claude/rules/moai/design/constitution.md)" -eq 1 ] && echo PASS || echo FAIL
# Expected: PASS (exit 0)
```

### AC-12: Section 3.2 canonical list 단일성 (N1 해결)

**기준** (REQ-012): canonical 예약 파일명 5개(`tokens.json`, `components.json`, `assets/`, `import-warnings.json`, `brief/BRIEF-*.md`)가 Section 3.2 텍스트 범위 내에 **정확히 1회씩**만 등장하며, Section 3.2 바깥 범위에는 등장하지 않아야 한다. 이는 v0.2.0에서 Section 3.2 본문의 parenthetical 열거와 canonical list가 같은 파일명을 두 번 싣던 N1 BLOCKING 결함을 근본 해결하는 어서션이다.

**검증 방법** (결정적, N3 exit-code 기반):
```bash
# Section 3.2 내부 (### 3.2 heading 이후 ### 3.3 heading 직전까지)
inside=$(awk '/^### 3\.2/,/^### 3\.3/' .claude/rules/moai/design/constitution.md)

# Section 3.2 외부 (나머지 영역)
# Full file에서 inside 라인 수를 차감하여 외부 존재 여부를 파일 전역 count - inside count로 계산
for f in tokens.json components.json assets/ import-warnings.json 'brief/BRIEF-*.md'; do
  total=$(grep -F -c "$f" .claude/rules/moai/design/constitution.md)
  inside_count=$(printf '%s\n' "$inside" | grep -F -c "$f")
  outside=$((total - inside_count))
  [ "$inside_count" -eq 1 ] && [ "$outside" -eq 0 ] && echo "PASS $f" || echo "FAIL $f (inside=$inside_count outside=$outside)"
done
# Expected: PASS x5 lines (각 파일명이 Section 3.2 내부 정확히 1회, 외부 0회)
```

---

## Scope

### IN SCOPE

본 SPEC의 구현(`/moai run SPEC-DESIGN-CONST-AMEND-001`) 단계에서 허용되는 변경 범위:

1. **Section 3 제목 변경 및 확장**: `## 3. Brand Context as Constitutional Principle` → `## 3. Brand Context and Design Brief as Constitutional Principles`로 제목 변경 (REQ-011에 의해 정식 규정됨) 후 3.1, 3.2, 3.3 하위 섹션 추가
2. **Section 2 FROZEN 목록 업데이트**: Section 3.1, 3.2, 3.3 세 하위 섹션이 모두 FROZEN임을 명시적으로 선언 (REQ-009)
3. **HISTORY 블록 신규 entry 추가**: 2026-04-20 v3.3.0 개정 내역 기록, 기존 entry와 시각 해상도 또는 SPEC prefix로 구분 (REQ-007)
4. **Version footer bump**: `Version: 3.2.0` → `Version: 3.3.0` (REQ-008)

### OUT OF SCOPE

다음 작업은 본 SPEC에서 명시적으로 제외된다. 별도 SPEC으로 다룰 것:

1. **템플릿 파일 생성** → `SPEC-DESIGN-DOCS-001`이 담당
   - `.moai/design/` 폴더의 research.md, system.md, spec.md, pencil-plan.md 템플릿 파일 생성
   - moai-studio `pencil-redesign-plan.md` → `pencil-plan.md` 마이그레이션 훅 구현
   - `design.yaml`에 `design_docs.token_budget` 키 추가 (본 SPEC은 fallback 20000으로 stand-alone 동작 보장, REQ-004)
2. **Skill 구현** → `SPEC-DESIGN-ATTACH-001`이 담당
   - `/moai design` 워크플로우의 자동 로드 로직 구현
3. **`/moai db` 커맨드** → `SPEC-DB-CMD-001`이 담당
   - 데이터베이스 관련 별도 CLI 커맨드
4. **`db` 폴더** → `SPEC-DB-TEMPLATES-001`이 담당
   - 데이터베이스 템플릿 파일 구조

---

## Risks and Mitigations

### R-1: FROZEN 규정에 따른 리뷰어 거부

**Risk**: constitution.md는 Section 2에서 FROZEN으로 명시되어 있어, 일반적 개정 요청은 자동 거부되어야 한다. 리뷰어가 "FROZEN 파일은 수정 불가"라는 이유로 본 SPEC 기반의 변경을 거부할 수 있다.

**Mitigation**:
- 본 SPEC은 Section 2의 FROZEN 규정이 요구하는 "정식 SPEC + 인간 승인" 절차를 명시적으로 따르고 있음을 문서 상단 Metadata block 및 HISTORY에 명기
- Priority를 High로 설정하여 중요성을 표시
- Metadata block에 `Classification: FROZEN_AMENDMENT`를 명시하여 리뷰어가 예외적 개정 절차임을 즉시 인지할 수 있도록 함

### R-2: 예약 파일명 충돌

**Risk**: 기존 `.moai/design/` 폴더에 `moai-workflow-design-import`가 생성하는 `tokens.json`, `components.json`, `assets/`, `import-warnings.json` 등의 기계 산출물과 새로 도입하는 인간 저작 파일(research.md, system.md, spec.md, pencil-plan.md)이 같은 폴더에 공존한다. 향후 확장 시 파일명 충돌 가능성이 존재한다.

**Mitigation**:
- Section 3.2에 canonical 예약 파일 목록을 **verbatim**으로 명시 (REQ-CONST-AMEND-003): `tokens.json`, `components.json`, `assets/`, `import-warnings.json`, `brief/BRIEF-*.md`
- `brief/BRIEF-*.md` 패턴도 예약으로 등재하여 향후 BRIEF 기반 파일명도 보호
- 인간 저작 파일은 확장자 `.md`로 통일, 기계 산출물은 `.json` 및 서브 폴더(`assets/`)로 분리하여 자연스러운 충돌 방지
- AC-5의 negative assertion (`awk`로 Section 3.2 범위 안에서 정확히 1회)으로 drive-by 중복 유입 차단

### R-3: 토큰 예산 초과

**Risk**: moai-studio 사례에서 4개 파일 합산 1398 라인은 LLM 컨텍스트 윈도우에 부담이 될 수 있다. 무제한 자동 로드 시 `/moai design` 워크플로우가 컨텍스트 폭주로 실패할 수 있다.

**Mitigation**:
- `design.yaml`의 `design_docs.token_budget` 설정으로 예산 상한 강제 (REQ-004). 본 SPEC 구현 시점에 해당 키가 아직 병합되지 않았을 수 있으므로 inline fallback 20000을 REQ-004 및 Section 3.2에 명시하여 stand-alone 동작 보장
- 우선순위 기반 truncation 규칙을 Section 3.2에 **verbatim** 코드화: `spec.md > system.md > research.md > pencil-plan.md` (REQ-CONST-AMEND-005)
- 구현 주도 문서(spec.md, system.md)는 우선 로드, 배경 및 도구 특화 문서는 후순위로 배치

### R-4: 동일 날짜 HISTORY entry 혼동 (F11 대응 - LOW, 수용 처리)

**Risk**: 기존 relocation entry(`2026-04-20`)와 신규 amendment entry가 같은 날짜를 공유하여 독자가 중복으로 오인할 수 있다.

**Mitigation**:
- REQ-007이 신규 entry에 시각 접미사(`2026-04-20T12:00Z`) 또는 SPEC 식별자 prefix(`2026-04-20 (SPEC-DESIGN-CONST-AMEND-001)`) 중 하나를 강제
- AC-3가 `v3.3.0` 마커 존재 + 기존 "Relocated from" entry 보존을 독립적으로 검증하여 구 entry 변조/삭제를 방지

### R-5: Cross-SPEC 의존성 (F8 관련 - MEDIUM, deferral justification)

**Risk**: REQ-004가 참조하는 `design.yaml` `design_docs.token_budget` 키의 정식 정의는 SPEC-DESIGN-DOCS-001 소관이다. SPEC-DESIGN-DOCS-001이 본 SPEC보다 나중에 머지되거나 해당 키를 빼먹으면, REQ-004는 runtime 시점에 unsatisfiable할 위험이 있다.

**Mitigation (인라인 해결 선택)**:
- REQ-004에 inline fallback 조항(`IF the key is absent or unparseable, THEN the system SHALL default to 20000`) 추가로 본 SPEC이 cross-SPEC precondition 없이 stand-alone enforceable
- SPEC-DESIGN-DOCS-001이 `design_docs.token_budget` 키를 명시적으로 도입하면 fallback 경로는 자연스럽게 죽은 코드가 되지만, stand-alone 보장은 유지
- 본 SPEC에서 별도 precondition 섹션을 도입하지 않은 이유: 헌법 개정(본 SPEC)과 skill 구현(SPEC-DESIGN-DOCS-001) 사이의 순서 제약은 scope creep이며, 시스템 fallback이 더 강건한 해결책이기 때문

---

## Implementation Note

**본 SPEC은 대상 개정 내용을 정의할 뿐이다**. 실제 `.claude/rules/moai/design/constitution.md` 파일의 편집은 다음 단계에서 수행된다:

```
/moai run SPEC-DESIGN-CONST-AMEND-001
```

본 SPEC 작성 단계에서는 **절대로** constitution.md를 수정하지 않는다. 이는 다음 원칙을 준수하기 위함이다:

- **Scope Discipline**: SPEC 작성과 구현은 분리되어야 한다
- **FROZEN 규정**: 구현 단계에서 별도 인간 승인(AskUserQuestion)이 수반되어야 한다
- **Plan-Run-Sync 분리**: Plan 단계(현재)는 "무엇을" 정의하고, Run 단계는 "어떻게" 실행한다

구현 단계에서 `manager-ddd` 또는 `manager-tdd`가 본 SPEC을 로드하여 위 11개 REQ와 11개 AC를 모두 충족하는 방식으로 constitution.md를 편집할 것이다.

---

Version: 0.3.0
Classification: FROZEN_AMENDMENT
SPEC Status: draft (awaiting human approval for /moai run)
Target File: .claude/rules/moai/design/constitution.md (v3.2.0 → v3.3.0)
