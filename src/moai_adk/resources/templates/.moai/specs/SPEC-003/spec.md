# SPEC-003: cc-manager 중심 Claude Code 최적화

## @REQ:CC-OPTIMIZATION-003 프로젝트 컨텍스트

### 배경

MoAI-ADK는 Claude Code 환경에서 완전한 Agentic Development Kit를 제공합니다. 현재 Claude Code 공식 문서 표준과 MoAI-ADK 구조 간의 불일치로 인해 에이전트/커맨드 파일의 표준화가 필요한 상황입니다.

### 문제 정의

- **현재 상태**: Claude Code 공식 구조와 불일치하는 YAML frontmatter
- **핵심 문제**: cc-manager가 단순 설정 관리자 역할에 국한됨
- **비즈니스 영향**: 표준 미준수로 인한 확장성과 유지보수성 저하

### 목표

1. **cc-manager를 Claude Code 표준화의 중앙 관제탑으로 강화**
2. **커맨드/에이전트 템플릿을 cc-manager 내부 지침으로 통합**
3. **Claude Code 공식 문서 100% 준수하는 표준 구조 적용**

## @DESIGN:CC-MANAGER-ARCH-003 아키텍처 설계

### 새로운 cc-manager 역할

```
cc-manager (중앙 관제탑)
├── 📐 내부 템플릿 지침
│   ├── 커맨드 표준 구조
│   └── 에이전트 표준 구조
├── 🔍 표준화 검증 로직
├── ⚙️ 설정 파일 최적화
├── 📝 파일 생성/수정 관리
└── 🎯 다른 에이전트 지도/감독
```

### Claude Code 공식 표준 준수

**커맨드 표준 구조**:

```yaml
---
name: command-name
description: Clear one-line description
argument-hint: [param1] [param2]
allowed-tools: Tool1, Tool2, Bash(cmd:*)
model: sonnet
---
```

**에이전트 표준 구조**:

```yaml
---
name: agent-name
description: Use PROACTIVELY for [specific task]
tools: Read, Write, Edit, MultiEdit, Bash, Glob, Grep
model: sonnet
---
```

## @TASK:IMPLEMENT-003 요구사항 명세

### R1. cc-manager 강화

**WHEN** Claude Code 파일 생성/수정이 필요할 때,
**THE SYSTEM SHALL** cc-manager가 표준 템플릿을 적용하여 일관된 구조를 제공해야 함

**상세 요구사항:**

- cc-manager.md에 커맨드/에이전트 템플릿 지침 내장
- **Claude Code 공식 문서 핵심 내용 완전 통합** (외부 참조 불필요)
- 표준 YAML frontmatter 구조 강제
- 자동 검증 로직 포함
- 파일 생성/수정 시 표준 준수 확인
- **중구난방 지침으로 인한 오류 방지를 위한 완전한 가이드 제공**

### R2. 기존 파일 표준화

**WHEN** 기존 .claude/commands/moai/_.md와 .claude/agents/moai/_.md 파일을 업데이트할 때,
**THE SYSTEM SHALL** Claude Code 공식 구조에 맞게 변경해야 함

**상세 요구사항:**

- 5개 커맨드 파일 YAML frontmatter 표준화
- 7개 에이전트 파일 YAML frontmatter 표준화
- 프로액티브 트리거 조건 명확화
- 도구 권한 최소화 원칙 적용

### R3. 핵심 파일 최적화

**WHEN** CLAUDE.md, development-guide.md, settings.json 파일을 개선할 때,
**THE SYSTEM SHALL** cc-manager 중심 워크플로우를 반영해야 함

**상세 요구사항:**

- CLAUDE.md에서 cc-manager 역할 강조
- development-guide.md에 Claude Code 표준 추가
- settings.json 권한 최적화 (WebSearch, BashOutput 등 추가)
- 불필요한 중복 제거

### R4. 검증 도구 개발

**WHEN** 표준 준수를 확인해야 할 때,
**THE SYSTEM SHALL** 자동화된 검증 도구를 제공해야 함

**상세 요구사항:**

- validate_claude_standards.py 스크립트 개발
- YAML frontmatter 유효성 검사
- 필수 필드 존재 확인
- 통합 테스트 프레임워크

## @TEST:ACCEPTANCE-003 Acceptance Criteria

### AC1. cc-manager 템플릿 지침 내장

**Given** cc-manager.md 파일이 존재할 때
**When** 파일을 열어 내용을 확인하면
**Then** 커맨드와 에이전트 표준 템플릿 지침이 포함되어 있어야 하고, 파일 생성/수정 가이드라인이 명시되어야 함

### AC2. Claude Code 표준 준수

**Given** .claude/commands/moai/\*.md 파일들이 존재할 때
**When** YAML frontmatter를 검사하면
**Then** name, description, argument-hint, allowed-tools, model 필드가 모두 존재해야 함

**Given** .claude/agents/moai/\*.md 파일들이 존재할 때
**When** YAML frontmatter를 검사하면
**Then** name, description, tools, model 필드가 모두 존재하고, "Use PROACTIVELY" 패턴이 description에 포함되어야 함

### AC3. 검증 도구 동작

**Given** validate_claude_standards.py 스크립트가 존재할 때
**When** 스크립트를 실행하면
**Then** 모든 Claude Code 파일의 표준 준수 여부를 확인하고 문제점을 리포트해야 함

### AC4. 통합 워크플로우

**Given** cc-manager 에이전트를 호출할 때
**When** 새로운 커맨드나 에이전트 파일 생성을 요청하면
**Then** 표준 템플릿에 따라 올바른 구조로 파일이 생성되어야 함

## 범위 및 모듈

### In Scope

- cc-manager.md 강화 (템플릿 지침 내장)
- 5개 커맨드 파일 표준화 (.claude/commands/moai/)
- 7개 에이전트 파일 표준화 (.claude/agents/moai/)
- 핵심 문서 최적화 (CLAUDE.md, development-guide.md, settings.json)
- 검증 도구 개발 (.moai/scripts/validate_claude_standards.py)

### Out of Scope

- Claude Code 내부 API 수정 (외부 제어 불가)
- 새로운 에이전트 개발 (기존 7개만 최적화)
- Python 모듈 구조 변경 (Claude Code 파일만 대상)
- 다른 AI 플랫폼 호환성 (Claude Code 전용)

## 기술 노트

### 구현 기술

- **파일 형식**: Markdown + YAML frontmatter
- **검증 도구**: Python 스크립트 (YAML 파싱)
- **테스팅**: pytest 기반 통합 테스트
- **문서화**: Claude Code 공식 문서 준수

### 의존성

- 기존 .claude/ 디렉토리 구조
- Claude Code 에이전트 실행 환경
- YAML 처리 라이브러리 (PyYAML)
- 기존 MoAI-ADK 워크플로우

### 아키텍처 원칙

- **단일 진실의 원천**: cc-manager가 모든 표준의 중심
- **점진적 마이그레이션**: 기존 파일을 단계적으로 최적화
- **실용적 접근**: 과도한 추상화 없이 실제 사용 패턴 기반
- **표준 강제**: cc-manager가 자동으로 표준 준수 검증

### 보안 고려사항

- 최소 권한 원칙: 에이전트별 필요 도구만 허용
- 입력 검증: 사용자 제공 템플릿 파라미터 검증
- 파일 접근 제한: 민감한 파일(.env, secrets) 접근 차단
- 권한 설정 최적화: settings.json 보안 강화

## 추적성

### 연결된 요구사항

- @TASK:AGENT-GUIDE-001: 에이전트별 입력/출력 스키마 정의 → cc-manager 템플릿으로 해결
- @REQ:USER-001: 3차 사용자(생태계 확장자) 지원 → 표준 템플릿 제공
- @STRUCT:ARCHITECTURE-001: Claude Extensions 계층 표준화 → Claude Code 공식 구조 적용

### 구현 우선순위

1. **cc-manager 강화** (High) - 중앙 관제탑 역할 확립
2. **기존 파일 표준화** (High) - Claude Code 공식 구조 적용
3. **검증 도구 개발** (Medium) - 자동화된 품질 관리
4. **핵심 문서 최적화** (Medium) - 워크플로우 통합

### 테스트 전략

- **단위 테스트**: 각 표준화 규칙별 검증 테스트
- **통합 테스트**: cc-manager를 통한 파일 생성/수정 테스트
- **E2E 테스트**: 전체 워크플로우 실행 테스트

## 성공 지표

### 정량적 지표

- Claude Code 표준 준수율: 100% (모든 커맨드/에이전트 파일)
- 자동 검증 성공률: ≥95% (validate_claude_standards.py)
- 파일 생성 시 표준 적용률: 100% (cc-manager 통해 생성 시)

### 정성적 지표

- cc-manager 중심 워크플로우 확립
- 표준 템플릿 기반 일관된 파일 구조
- Claude Code 공식 문서와의 완전한 호환성
- 유지보수성 및 확장성 향상

## 다음 단계

1. **TDD 구현**: Red-Green-Refactor 사이클로 구현
2. **점진적 마이그레이션**: cc-manager 강화 → 파일 표준화 → 검증 도구
3. **품질 검증**: 모든 변경사항에 대한 테스트 커버리지 확보
4. **문서 동기화**: /moai:3-sync로 Living Document 업데이트
