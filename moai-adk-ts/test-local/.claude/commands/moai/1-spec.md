---
name: moai:1-spec
description: EARS 명세 작성 + 브랜치/PR 생성
argument-hint: "제목1 제목2 ... | SPEC-ID 수정내용"
tools: Read, Write, Edit, MultiEdit, Grep, Glob, TodoWrite, Bash
---

# MoAI-ADK 1단계: EARS 명세 작성 + 브랜치/PR 생성

**SPEC 생성 대상**: $ARGUMENTS

## 🚀 SPEC 작성 및 브랜치 생성

프로젝트 문서를 분석하여 SPEC 후보를 제안하고, 선택된 SPEC을 즉시 생성합니다.

## 핵심 기능

- **스마트 분석**: 프로젝트 문서를 분석하여 SPEC 후보 자동 제안
- **Personal 모드**: `.moai/specs/SPEC-XXX/` 디렉터리 생성 + 로컬 브랜치
- **Team 모드**: GitHub Issue 생성 + 원격 브랜치 + PR 템플릿

## 사용법

```bash
/moai:1-spec                      # 자동 제안 (권장)
/moai:1-spec "JWT 인증 시스템"       # 수동 생성
/moai:1-spec SPEC-001 "보안 강화"   # 기존 SPEC 수정
```

## 워크플로우

### 1. 프로젝트 분석
- product/structure/tech.md 문서 스캔
- 기존 SPEC 목록 및 우선순위 검토
- 핵심 요구사항 추출

### 2. SPEC 후보 제안
- 비즈니스 가치 기반 우선순위 설정
- 기술적 제약사항 반영
- 3-5개 후보 리스트 생성

### 3. SPEC 문서 생성
- **EARS 구조**: Environment, Assumptions, Requirements, Specifications
- **3개 파일**: spec.md, plan.md, acceptance.md
- **16-Core TAG**: @REQ → @DESIGN → @TASK → @TEST 체인

### 4. Git 작업 자동화
- **Personal 모드**: 로컬 브랜치 + 체크포인트
- **Team 모드**: GitHub Issue + 원격 브랜치 + PR

## 실행 순서

### 1. spec-builder 에이전트 실행
```bash
@agent-spec-builder --target="$ARGUMENTS"
```
- 프로젝트 문서 분석
- SPEC 후보 제안 및 사용자 선택
- EARS 구조 SPEC 문서 작성
- MultiEdit으로 3개 파일 동시 생성

### 2. git-manager 에이전트 실행
```bash
@agent-git-manager --spec-created --mode="$MODE"
```
- 브랜치 생성 (Personal/Team 모드별)
- GitHub Issue/PR 생성 (Team 모드)
- 초기 커밋 및 태그

## 품질 기준

- **EARS 구조**: Environment, Assumptions, Requirements, Specifications 필수
- **TAG 체인**: @REQ → @DESIGN → @TASK → @TEST 완전성
- **Acceptance Criteria**: Given-When-Then 시나리오 최소 2개
- **16-Core TAG**: 추적성 확보를 위한 TAG 시스템 적용

## 다음 단계

- `/moai:2-build SPEC-XXX`: TDD 구현 시작
- `/moai:3-sync`: 문서 동기화 및 PR 상태 전환
