---
id: SKILL-REFACTOR-001
version: 0.1.0
status: completed
created: 2025-10-19
updated: 2025-10-21
author: @Goos
priority: high
category: refactor
labels:
  - skills
  - standardization
  - yaml
  - code-quality
---

# @SPEC:SKILL-REFACTOR-001: Claude Code Skills 표준화

## HISTORY

### v0.0.1 (2025-10-19)
- **INITIAL**: Claude Code Skills 표준화 SPEC 최초 작성
- **AUTHOR**: @Goos
- **SCOPE**: 50개 스킬 파일명, YAML 필드, allowed-tools 표준화
- **CONTEXT**: Anthropic 공식 표준 준수 및 코드 품질 향상
- **TASKS**:
  1. skill.md → SKILL.md 파일명 변경 (50개)
  2. 중복 CC 템플릿 삭제 (5개)
  3. YAML 필드 정리 (version, author, license, tags, model 제거)
  4. allowed-tools 필드 추가 (25개)

### v0.1.0 (2025-10-21)
- **COMPLETED**: 전체 구현 및 검증 완료
- **AUTHOR**: @Claude (구현)
- **CHANGES**:
  1. ✅ skill.md → SKILL.md 파일명 변경 완료 (55개 스킬)
  2. ✅ 중복 CC 템플릿 5개 삭제 완료
  3. ✅ YAML 필드 정리 완료 (version, author, license, tags, model 0개)
  4. ✅ allowed-tools 필드 추가 완료 (55개 스킬 + 55개 템플릿)
  5. ✅ 종합 검증 완료 (AC-001 ~ AC-004 모두 통과)
- **DELIVERABLES**:
  - .claude/skills/: 55개 SKILL.md (모두 allowed-tools 포함)
  - src/moai_adk/templates/.claude/skills/: 55개 SKILL.md (동기화 완료)
  - scripts/sync_allowed_tools.py: 템플릿 동기화 자동화 스크립트
- **VALIDATION**:
  - skill.md 파일: 0개 (모두 표준화)
  - SKILL.md 파일: 55개 (.claude) + 55개 (templates)
  - allowed-tools 필드: 55개 (.claude) + 55개 (templates)
  - YAML 불필요 필드: 0개 (version, model, author, license, tags)

---

## Environment (환경)

**프로젝트 구조**:
- Skills 위치: `.claude/skills/*/`
- 템플릿 위치: `src/moai_adk/templates/.claude/skills/*/`
- 전체 스킬 수: 51개

**현재 상태**:
- SKILL.md (표준): 1개 (moai-claude-code)
- skill.md (비표준): 50개
- 중복 CC 템플릿: 5개
- 불필요한 YAML 필드: 174개 (50개 스킬 × 평균 3.5개)
- allowed-tools 누락: 25개

**도구**:
- Python 3.x (YAML 파싱)
- Bash (파일명 변경, 디렉토리 삭제)
- ruamel.yaml (YAML 필드 수정)

---

## Assumptions (가정)

1. **파일명 변경**:
   - Git이 mv 명령으로 파일명 변경을 자동 추적
   - 기존 skill.md 파일 내용은 변경 없음

2. **YAML 필드 정리**:
   - Skills는 name, description, allowed-tools만 필요
   - model 필드는 Agent 전용 (Skills에 불필요)

3. **allowed-tools 추가**:
   - 스킬 유형별 도구 권한 일관성 유지
   - Alfred 에이전트: 복잡한 작업 (Read, Write, Edit, Bash, TodoWrite)
   - Lang/Domain 스킬: 참조 및 실행 (Read, Bash)

4. **중복 템플릿**:
   - moai-claude-code가 5개 CC 템플릿을 대체
   - 삭제 후 복구 불필요

---

## Requirements (요구사항)

### Ubiquitous Requirements (기본 요구사항)

- 시스템은 모든 Skills가 SKILL.md (대문자) 파일명을 사용해야 한다
- 시스템은 Anthropic 공식 표준에 맞는 YAML 필드만 포함해야 한다
- 시스템은 중복된 템플릿을 제거해야 한다
- 시스템은 모든 Skills에 명시적 도구 권한(allowed-tools)을 부여해야 한다

### Event-driven Requirements (이벤트 기반)

- WHEN skill.md (소문자) 파일이 발견되면, 시스템은 SKILL.md (대문자)로 파일명을 변경해야 한다
- WHEN 중복 CC 템플릿 디렉토리가 발견되면, 시스템은 해당 디렉토리를 삭제해야 한다
- WHEN 비표준 YAML 필드(version, author, license, tags, model)가 발견되면, 시스템은 해당 필드를 제거해야 한다
- WHEN allowed-tools 필드가 누락된 스킬이 발견되면, 시스템은 스킬 유형에 맞는 도구를 추가해야 한다

### State-driven Requirements (상태 기반)

- WHILE 파일명 변경 작업을 수행할 때, 시스템은 Git이 변경 사항을 추적하도록 해야 한다
- WHILE YAML 필드를 수정할 때, 시스템은 파일 구조를 손상시키지 않아야 한다
- WHILE 템플릿 디렉토리를 동기화할 때, 시스템은 `.claude/skills/`와 `src/moai_adk/templates/.claude/skills/`를 모두 처리해야 한다

### Constraints (제약사항)

- IF Skills 파일이라면, model 필드를 포함하지 않아야 한다 (Agent 전용 필드)
- IF YAML frontmatter를 수정한다면, name과 description 필드는 반드시 유지해야 한다
- IF 파일명을 변경한다면, 파일 내용은 변경하지 않아야 한다
- 모든 작업은 `.claude/skills/`와 `src/moai_adk/templates/.claude/skills/` 양쪽에 동일하게 적용되어야 한다

---

## Traceability (@TAG)

- **SPEC**: @SPEC:SKILL-REFACTOR-001
- **TEST**: tests/integration/test_skills_structure.py
- **CODE**: scripts/standardize_skills.py
- **DOC**: docs/skills/standardization-guide.md

---

## Implementation Plan (구현 계획 참조)

상세 구현 계획은 `plan.md` 참조

---

## Acceptance Criteria (인수 기준 참조)

상세 인수 기준은 `acceptance.md` 참조

---

**최종 업데이트**: 2025-10-19
**작성자**: @Goos
