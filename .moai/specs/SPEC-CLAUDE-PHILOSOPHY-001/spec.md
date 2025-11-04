---
id: CLAUDE-PHILOSOPHY-001
version: 0.1.0
status: implementation-complete
created: 2025-11-04
updated: 2025-11-04
author: @Alfred
priority: high
category: documentation
labels:
  - claude-md
  - philosophy
  - structure
  - skills
related_specs:
  - DOCS-001
scope:
  packages:
    - CLAUDE.md
    - .claude/skills/
  files:
    - /Users/goos/MoAI/MoAI-ADK/CLAUDE.md
    - /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-session-analytics/
    - /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-config-advanced/
---

# @SPEC:CLAUDE-PHILOSOPHY-001: CLAUDE.md 철학 재정렬 및 Skill 분리

## HISTORY

### v0.1.0 (2025-11-04)
- **INITIAL**: CLAUDE.md 철학 재정렬 명세 작성
- **AUTHOR**: @Alfred
- **SCOPE**:
  - CLAUDE.md 구조 재설계 (Tier 1-4 계층화)
  - 2개 Skill 분리 (session-analytics, config-advanced)
  - 부정적 제약 → 긍정적 가이드라인 변환 (20개 이상)
  - 패키지 템플릿 동기화
- **CONTEXT**: CLAUDE.md가 1,000줄 이상으로 비대화되어 핵심 규칙 가독성 저하. 세션 시작 시 필수 정보에 집중하기 위해 구조 재정렬 필요.

---

## Environment (환경 및 전제조건)

### 실행 환경
- **프로젝트**: MoAI-ADK Phase 6 (CLAUDE.md 철학 재정렬)
- **대상 파일**: `/Users/goos/MoAI/MoAI-ADK/CLAUDE.md` (한국어 로컬 버전)
- **패키지 템플릿**: `src/moai_adk/templates/CLAUDE.md` (영어 글로벌 버전)
- **언어**: 한국어 (로컬), 영어 (패키지)
- **동기화 정책**: 로컬과 패키지는 언어만 다르고 구조는 동일

### 기술 스택
- **언어**: Markdown
- **Skill 도구**: Claude Skills (YAML frontmatter + Markdown)
- **검증 도구**: Grep (TAG 검증), Read (파일 참조)

### 제약사항
- **최소 400줄 유지**: CLAUDE.md는 과도한 단순화 금지
- **기존 참조 링크 유지**: 섹션 제목 변경 시 기존 링크가 깨지지 않도록 제목 유지
- **핵심 제약 유지**: 부정적 표현을 100% 긍정으로 바꾸지 않음 (필수 금지사항은 유지)
- **패키지 템플릿 동기화**: 로컬 변경 → 패키지 동기화 (언어만 영어로)

---

## Assumptions (가정사항)

1. **현재 CLAUDE.md 구조 가정**:
   - 섹션 순서: Alfred 소개 → 언어 경계 → 워크플로우 → Sub-agents → Skills → ...
   - 총 1,000줄 이상 (긴 섹션: 언어 경계, Permissions, 에이전트 설명)
   - Skill 링크는 `Skill("name")` 형식으로 명시적 호출

2. **Tier 1-4 우선순위 가정**:
   - **Tier 1 (핵심 규칙)**: 4단계 워크플로우, 언어 경계, Permissions, TRUST 원칙
   - **Tier 2 (실행 가이드)**: Sub-agents 역할, 커맨드 설명, Git 워크플로우
   - **Tier 3 (고급 기능)**: 적응형 페르소나, 자동 수정 프로토콜, 보고 스타일
   - **Tier 4 (참조)**: 프로젝트 정보, 기술 스택, 설정 가이드

3. **Skill 분리 가정**:
   - **moai-alfred-session-analytics**: 세션 분석, 로깅, 메트릭 수집 (현재 CLAUDE.md에 포함)
   - **moai-alfred-config-advanced**: 고급 설정, 타임아웃, Hook 최적화 (현재 CLAUDE.md에 포함)
   - 분리 후 CLAUDE.md는 Skill("moai-alfred-session-analytics") 링크만 포함

4. **부정적 제약 변환 가정**:
   - 최소 20개 이상의 부정적 표현을 긍정적으로 변환
   - 예시: "❌ DO NOT X" → "✅ INSTEAD: Y"
   - 금지사항은 명확히 유지 (예: git push --force는 deny)

5. **패키지 동기화 가정**:
   - 로컬 CLAUDE.md 변경 → `src/moai_adk/templates/CLAUDE.md` 동기화
   - 언어만 영어로 변환 (구조, 섹션 순서, Skill 링크는 동일)

---

## Requirements (EARS 요구사항)

### Ubiquitous Requirements (기본 기능)

**UR-001**: 시스템은 CLAUDE.md를 Tier 1-4 계층 구조로 재정렬해야 한다
- **Tier 1 (핵심 규칙, 항상 읽힘)**:
  - 4단계 워크플로우 (의도 파악 → 계획 수립 → 작업 실행 → 보고 및 커밋)
  - 언어 경계 규칙 (Layer 1: conversation_language, Layer 2: 영어 인프라)
  - Permissions 우선순위 (deny → ask → allow)
  - TRUST 5 원칙 (Test First, Readable, Unified, Secured, Trackable)
- **Tier 2 (실행 가이드)**:
  - Sub-agents 역할 (spec-builder, code-builder, git-manager 등)
  - 4개 커맨드 설명 (/alfred:0-project, :1-plan, :2-run, :3-sync)
  - Git 워크플로우 (TDD 커밋, 브랜치 전략, PR)
- **Tier 3 (고급 기능)**:
  - 적응형 페르소나 시스템
  - 자동 수정 및 병합 충돌 프로토콜
  - 보고 스타일
- **Tier 4 (참조)**:
  - 프로젝트 정보 (버전, 기술 스택)
  - 설정 가이드 (config.json, settings.json)
  - 설치 및 Quick Start

**UR-002**: 시스템은 2개의 새로운 Skill을 생성해야 한다
- **moai-alfred-session-analytics**:
  - 위치: `.claude/skills/moai-alfred-session-analytics/`
  - 내용: 세션 분석, 로깅, 메트릭 수집 (현재 CLAUDE.md 섹션 분리)
  - YAML frontmatter: name, version, status, description, keywords, allowed-tools
  - reference.md: 세션 메트릭 정의, 로깅 정책
  - examples.md: 세션 분석 예시
- **moai-alfred-config-advanced**:
  - 위치: `.claude/skills/moai-alfred-config-advanced/`
  - 내용: 고급 설정 (Hook 타임아웃, 권한 세분화, 메타데이터 최적화)
  - YAML frontmatter: name, version, status, description, keywords, allowed-tools
  - reference.md: 고급 설정 필드 설명
  - examples.md: 고급 설정 예시

**UR-003**: 시스템은 최소 20개 이상의 부정적 제약을 긍정적 가이드라인으로 변환해야 한다
- **변환 전**: "❌ DO NOT create IMPLEMENTATION_GUIDE.md in project root"
- **변환 후**: "✅ CREATE reports in `.moai/docs/` or `.moai/reports/` instead"
- **대상 섹션**: 보고 스타일, 자동 수정 프로토콜, Permissions

**UR-004**: 시스템은 패키지 템플릿과 로컬을 동기화해야 한다
- **로컬**: `/Users/goos/MoAI/MoAI-ADK/CLAUDE.md` (한국어)
- **패키지**: `src/moai_adk/templates/CLAUDE.md` (영어)
- **동기화 항목**: 구조, 섹션 순서, Skill 링크
- **차이점**: 언어만 다름 (한국어 vs 영어)

---

### Event-driven Requirements

**ED-001**: WHEN 사용자가 CLAUDE.md를 열 때, THEN Tier 1 핵심 규칙이 먼저 보여야 한다
- **첫 화면**: 4단계 워크플로우, 언어 경계, Permissions 우선순위
- **스크롤 없이**: 500줄 이내 (핵심만)

**ED-002**: WHEN Alfred가 세션 분석이 필요할 때, THEN Skill("moai-alfred-session-analytics")를 호출한다
- **트리거**: 사용자가 "세션 분석", "로그 확인", "메트릭" 요청
- **응답**: Skill에서 JIT 로드된 세션 분석 정보 제공

**ED-003**: WHEN Alfred가 고급 설정이 필요할 때, THEN Skill("moai-alfred-config-advanced")를 호출한다
- **트리거**: Hook 타임아웃 조정, 권한 세분화, 메타데이터 최적화 요청
- **응답**: Skill에서 JIT 로드된 고급 설정 정보 제공

**ED-004**: WHEN 개발자가 새로운 규칙을 추가할 때, THEN 1개 섹션에만 수정하면 된다
- **중복 제거**: 같은 내용이 여러 섹션에 반복되지 않음
- **Skill 링크**: 상세 내용은 Skill에 위임

**ED-005**: WHEN 로컬 CLAUDE.md가 변경될 때, THEN 패키지 템플릿을 자동 동기화해야 한다
- **트리거**: Git 커밋 후 Hook 실행
- **동기화**: 구조 복사 + 언어 변환 (한국어 → 영어)
- **검증**: YAML frontmatter, Skill 링크, 섹션 제목 일치

---

### State-driven Requirements

**SD-001**: WHILE 문서 구조 재설계 중일 때, THEN 패키지 템플릿과 로컬 동기화를 유지한다
- **동시 작업**: 로컬 변경 후 즉시 패키지 업데이트
- **검증**: diff 도구로 구조 일치 확인

**SD-002**: WHILE Skill 분리 작업 중일 때, THEN 기존 기능성을 보존한다
- **테스트**: 세션 분석, 고급 설정 기능이 Skill 분리 후에도 동작
- **검증**: Skill("name") 호출 시 정상 로드

**SD-003**: WHILE 긍정적 가이드라인 변환 중일 때, THEN 원래 의도는 유지한다
- **검증**: 부정적 표현을 긍정으로 바꾸되 금지사항은 명확히 유지
- **예시**: "git push --force는 deny" → 유지 (필수 금지사항)

---

### Optional Features

**OF-001**: WHERE 부정적 제약을 100% 긍정적으로 변환하지 못할 수 있다
- **허용**: 일부 금지사항은 부정적 표현 유지 (명확성 우선)
- **예시**: "NEVER run git push --force" (명확한 금지)

**OF-002**: WHERE 패키지 템플릿은 로컬과 완전히 동일하지 않을 수 있다
- **차이점**: 언어만 다름 (구조는 동일)
- **허용**: 번역 과정에서 미세한 표현 차이

---

### Unwanted Behaviors (피해야 할 동작)

**UB-001**: IF 재구조화 후에도 기존 섹션 위치가 크게 변경되면 THEN 기존 참조 링크가 깨진다
- **완화 전략**: 섹션 제목 유지, 앵커 링크 보존
- **검증**: Grep으로 모든 `[링크](#섹션)` 참조 검사

**UB-002**: IF Skill 분리로 인해 CLAUDE.md가 너무 단순화되면 THEN 필수 정보가 부족해진다
- **완화 전략**: Tier 1 핵심 규칙은 CLAUDE.md에 유지 (최소 400줄)
- **검증**: 세션 시작 시 필요한 정보가 모두 Tier 1에 있는지 확인

**UB-003**: IF 부정적 제약을 무분별하게 제거하면 THEN 실제 금지사항이 불명확해진다
- **완화 전략**: 핵심 제약은 부정적 표현 유지 (예: git push --force는 deny)
- **검증**: 20개 이상 변환하되 필수 금지사항은 명확히 유지

**UB-004**: IF 패키지 동기화를 수동으로 하면 THEN 누락 및 불일치 발생
- **완화 전략**: Git Hook으로 자동 동기화 (커밋 후 실행)
- **검증**: CI/CD에서 로컬과 패키지 diff 비교

**UB-005**: IF Skill 호출을 자동 트리거에 의존하면 THEN 의도치 않은 로드 발생
- **완화 전략**: 명시적 Skill("name") 호출만 사용
- **검증**: CLAUDE.md에 자동 트리거 로직 없음

---

## Specifications (상세 명세)

### SPEC-001: Tier 1-4 구조 재설계

**목표**: CLAUDE.md를 계층화하여 핵심 규칙 가독성 향상

**재설계 구조**:

```markdown
# MoAI-ADK

## 🎩 Alfred의 핵심 지침 (Tier 1)
### 4단계 워크플로우
### 언어 경계 규칙
### Permissions 우선순위
### TRUST 5 원칙

## 🛠️ Alfred의 실행 가이드 (Tier 2)
### Sub-agents 역할
### 4개 커맨드 설명
### Git 워크플로우

## 🎭 고급 기능 (Tier 3)
### 적응형 페르소나 시스템 → Skill("moai-alfred-personas")
### 자동 수정 프로토콜 → Skill("moai-alfred-autofixes")
### 보고 스타일 → Skill("moai-alfred-reporting")

## 📚 참조 (Tier 4)
### 프로젝트 정보
### 설정 가이드 → Skill("moai-alfred-config-advanced")
### 세션 분석 → Skill("moai-alfred-session-analytics")
```

**변경 사항**:
- Tier 1을 문서 상단 500줄 이내로 제한
- Tier 3 고급 기능은 Skill 링크로 대체
- Tier 4 참조는 필요시 JIT 로드

---

### SPEC-002: 2개 Skill 분리

**moai-alfred-session-analytics** (세션 분석):

```yaml
# .claude/skills/moai-alfred-session-analytics/SKILL.md
---
name: moai-alfred-session-analytics
version: 1.0.0
created: 2025-11-04
updated: 2025-11-04
status: active
description: Alfred 세션 분석, 로깅, 메트릭 수집 가이드
keywords: ['session', 'analytics', 'logging', 'metrics']
allowed-tools:
  - Read
  - Bash
  - Grep
---

# 세션 분석 Skill

## What It Does
- 세션 시작/종료 시간 추적
- 명령 실행 로그 수집
- 에러율, 성공률 메트릭 계산
- 세션 요약 리포트 생성

## When to Use
- WHEN 사용자가 "세션 분석" 요청
- WHEN 메트릭 확인 필요
- WHEN 로그 리뷰 필요
```

**moai-alfred-config-advanced** (고급 설정):

```yaml
# .claude/skills/moai-alfred-config-advanced/SKILL.md
---
name: moai-alfred-config-advanced
version: 1.0.0
created: 2025-11-04
updated: 2025-11-04
status: active
description: MoAI-ADK 고급 설정 가이드 (Hook 타임아웃, 권한 세분화)
keywords: ['config', 'advanced', 'hooks', 'permissions']
allowed-tools:
  - Read
  - Edit
  - Bash
---

# 고급 설정 Skill

## What It Does
- Hook 실행 타임아웃 조정
- Permissions 세분화 (deny/ask/allow)
- 메타데이터 최적화
- Claude Code 설정 고급 커스터마이징

## When to Use
- WHEN Hook 성능 조정 필요
- WHEN 권한 정책 커스터마이징
- WHEN .moai/config.json 고급 필드 설정
```

---

### SPEC-003: 부정적 제약 → 긍정적 가이드라인 변환 (20개 이상)

**변환 목록** (최소 20개):

1. ❌ "DO NOT create IMPLEMENTATION_GUIDE.md in project root"
   → ✅ "CREATE reports in `.moai/docs/` or `.moai/reports/` instead"

2. ❌ "NEVER use time predictions"
   → ✅ "USE priority-based milestones (High/Medium/Low) instead"

3. ❌ "DO NOT batch completions"
   → ✅ "MARK tasks completed IMMEDIATELY after finishing"

4. ❌ "NEVER skip confirmations for complex tasks"
   → ✅ "CONFIRM high-complexity decisions with AskUserQuestion"

5. ❌ "DO NOT use git:* wildcard"
   → ✅ "SPECIFY exact git commands (git status, git log, git diff)"

6. ❌ "NEVER overwrite existing files without reading first"
   → ✅ "READ files first, then use Edit tool for modifications"

7. ❌ "DO NOT rely on auto-triggering"
   → ✅ "USE explicit Skill('name') invocation"

8. ❌ "NEVER commit without user request"
   → ✅ "CREATE commits only when explicitly requested"

9. ❌ "DO NOT use emojis unless requested"
   → ✅ "KEEP output professional; emojis only on explicit request"

10. ❌ "NEVER amend other developers' commits"
    → ✅ "CHECK authorship before amending; create new commit if not yours"

11. ❌ "DO NOT push --force to main/master"
    → ✅ "WARN user and require explicit confirmation for force push to main"

12. ❌ "NEVER skip hooks (--no-verify)"
    → ✅ "RUN hooks unless user explicitly requests --no-verify"

13. ❌ "DO NOT update git config"
    → ✅ "USE existing git config; avoid modifications"

14. ❌ "NEVER use -i flag (interactive mode)"
    → ✅ "USE non-interactive commands for automation"

15. ❌ "DO NOT create empty commits"
    → ✅ "VERIFY changes exist before creating commit"

16. ❌ "NEVER use sequential Bash calls for independent tasks"
    → ✅ "RUN independent commands in parallel for performance"

17. ❌ "DO NOT use cd excessively"
    → ✅ "PREFER absolute paths; maintain current working directory"

18. ❌ "NEVER use find/grep in Bash"
    → ✅ "USE dedicated Glob/Grep tools for file operations"

19. ❌ "DO NOT use echo for file writing"
    → ✅ "USE Write tool for file creation"

20. ❌ "NEVER use cat/head/tail in Bash"
    → ✅ "USE Read tool for file reading"

21. ❌ "DO NOT create documentation files proactively"
    → ✅ "CREATE docs only when explicitly requested by user"

22. ❌ "NEVER assume values for required parameters"
    → ✅ "ASK user for missing values with AskUserQuestion"

23. ❌ "DO NOT ignore test failures"
    → ✅ "KEEP task as in_progress if tests fail; create blocker task"

24. ❌ "NEVER mark tasks completed with partial implementation"
    → ✅ "MARK completed ONLY when fully accomplished (tests pass, no errors)"

25. ❌ "DO NOT hardcode secrets"
    → ✅ "USE environment variables or .env files for sensitive data"

---

### SPEC-004: 패키지 템플릿 동기화

**동기화 워크플로우**:

```bash
# Phase 1: 로컬 CLAUDE.md 변경
1. Edit /Users/goos/MoAI/MoAI-ADK/CLAUDE.md (한국어)
2. Git commit

# Phase 2: 패키지 동기화 (자동 Hook)
3. Hook detects CLAUDE.md change
4. Copy structure to src/moai_adk/templates/CLAUDE.md
5. Translate Korean → English (구조 유지)
6. Verify: diff로 섹션 수, Skill 링크 일치 확인

# Phase 3: 검증
7. Run: grep 'Skill\("' CLAUDE.md src/moai_adk/templates/CLAUDE.md
8. Verify: 모든 Skill 링크 일치
9. Verify: 섹션 제목 수 일치 (한국어 vs 영어)
```

**동기화 검증 체크리스트**:
- [ ] 섹션 수 일치
- [ ] Skill("name") 링크 수 일치
- [ ] Tier 1-4 구조 동일
- [ ] YAML frontmatter 필드 동일 (영어로 번역)
- [ ] 언어만 한국어 vs 영어

---

## Traceability (추적성)

### TAG 체인

```
@SPEC:CLAUDE-PHILOSOPHY-001
  ↓ (drives requirements)
@TEST:CLAUDE-PHILOSOPHY-001 (tests/test_claude_md_structure.py)
  ↓ (tests implementation)
@CODE:CLAUDE-PHILOSOPHY-001 (CLAUDE.md, .claude/skills/moai-alfred-session-analytics/, moai-alfred-config-advanced/)
  ↓ (documented by)
@DOC:CLAUDE-PHILOSOPHY-001 (plan.md, acceptance.md)
```

### 관련 SPEC
- **SPEC-DOCS-001**: VitePress 문서 구조 참조
- **SPEC-SKILL-FACTORY-001**: Skill 생성 패턴 참조 (예정)

---

## Quality Gates

### 필수 검증
1. **구조 검증**: Tier 1이 500줄 이내
2. **Skill 검증**: 2개 Skill 정상 로드 (Skill("name") 호출 성공)
3. **변환 검증**: 최소 20개 부정적 제약 → 긍정적 가이드라인 변환
4. **동기화 검증**: 로컬과 패키지 구조 일치 (언어만 다름)
5. **링크 검증**: 모든 Skill("name") 링크 유효성 확인

### 성공 기준
- [ ] CLAUDE.md Tier 1이 500줄 이하
- [ ] Tier 3 섹션이 Skill 링크로 대체
- [ ] 2개 Skill 생성 완료
- [ ] 최소 20개 부정적 → 긍정적 변환
- [ ] 패키지 템플릿 동기화 완료

---

## Risk Management

### 주요 위험
1. **위험**: 기존 참조 링크 깨짐
   - **완화**: 섹션 제목 유지, 앵커 링크 보존
   - **검증**: Grep으로 모든 링크 참조 검사

2. **위험**: CLAUDE.md 과도한 단순화
   - **완화**: Tier 1 핵심 규칙 최소 400줄 유지
   - **검증**: 세션 시작 시 필요 정보 체크리스트

3. **위험**: 부정적 제약 제거로 금지사항 불명확
   - **완화**: 핵심 제약 유지 (git push --force는 deny)
   - **검증**: 20개 변환 후 필수 금지사항 명확성 확인

4. **위험**: 패키지 동기화 수동 누락
   - **완화**: Git Hook 자동 동기화
   - **검증**: CI/CD diff 비교

---

## Next Steps

1. **Phase 1 실행**: CLAUDE.md 구조 재설계 (Tier 1-4)
2. **Phase 2 실행**: 2개 Skill 분리 작업
3. **Phase 3 실행**: 부정적 제약 → 긍정적 가이드라인 변환
4. **Phase 4 실행**: 패키지 템플릿 동기화
5. **검증**: 전체 Quality Gates 통과 확인
6. **커밋**: TDD 커밋 (RED → GREEN → REFACTOR)

---

_이 SPEC은 `/alfred:1-plan`에 의해 생성되었으며, `/alfred:2-run`으로 구현됩니다._
