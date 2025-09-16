# MoAI-ADK 시스템 개요

> Claude Code 2025 최신 표준 기반 완전 자동화 Spec-First TDD 개발 시스템
> Python 패키지 기반 - 완전 자동화 버전 관리 시스템 및 통합 구조 관리 - v0.1.16 최신 버전

## 🎯 목표와 범위

### 목적

- **명세 우선 개발**: 명세가 모든 구현을 주도
- **완전 자동화**: AI 에이전트가 구현 작업 수행
- **품질 내재화**: 단계별 자동 검수와 게이트
- **Living Document**: 코드↔문서 실시간 동기화

### 범위

- **신규 프로젝트**: 그린필드 개발
- **기존 프로젝트**: 브라운필드 개선/리팩토링
- **팀 프로젝트**: Constitution 기반 거버넌스
- **개인 프로젝트**: 간소화 모드 지원

## 🏛️ 핵심 원칙

1. **Spec-First**: 명세 없이 코드 없음
2. **TDD-First**: 테스트 없이 구현 없음
3. **Living Doc**: 문서와 코드는 항상 동기화
4. **Full Traceability**: 모든 요구사항은 추적 가능
5. **YAGNI**: 필요한 것만 구현

## 시스템 특징

### 🤖 완전 자동화 (v0.1.16)
- **버전 관리**: 24개 파일 하드코딩 버전 일괄 동기화
- **사용자 업데이트**: `moai update` 한 번으로 패키지 + 리소스 자동 업그레이드
- **개발자 생산성**: `moai update-version` 명령어로 95% 시간 절약
- **안전장치**: 드라이 런, 백업, 검증 시스템으로 무사고 업데이트

### 🏗️ 중앙 집중식 아키텍처 (v0.1.13)
- **전역 리소스 관리**: ~/.claude/moai/ 통합 구조로 96% 공간 절약
- **심볼릭 링크 시스템**: 10배 빠른 설치, N-프로젝트 효율성
- **완전한 메모리 시스템**: Claude/MoAI 메모리 파일, Constitution, ADR 템플릿
- **지능형 Git 시스템**: 자동 설치, 저장소 보존, 운영체제별 최적화

### 🔄 4단계 파이프라인 워크플로우
1. **SPECIFY**: EARS 형식 명세 작성
2. **PLAN**: Constitution Check 및 계획 수립
3. **TASKS**: TDD 태스크 분해
4. **IMPLEMENT**: Red-Green-Refactor 구현

### 🤖 11개 전문 에이전트 시스템
- **steering-architect**: Steering 문서 생성
- **spec-manager**: SPEC 문서 관리
- **plan-architect**: 계획 수립 및 ADR 관리
- **task-decomposer**: 작업 분해
- **code-generator**: TDD 기반 코드 생성
- **test-automator**: TDD 자동화
- **doc-syncer**: Living Document 동기화
- **tag-indexer**: 16-Core @TAG 시스템 관리
- **integration-manager**: 외부 서비스 연동
- **deployment-specialist**: 배포 전략 및 자동화
- **claude-code-manager**: MoAI-Claude 통합 전문가

### 🏷️ 16-Core @TAG 추적성 시스템
- **Primary**: @REQ → @DESIGN → @TASK → @TEST
- **Steering**: @VISION → @STRUCT → @TECH → @ADR
- **Quality**: @PERF → @SEC → @DEBT → @TODO

### 🛡️ 품질 보증 시스템
- **Constitution 5원칙**: Simplicity, Architecture, Testing, Observability, Versioning
- **Hook 시스템**: PreToolUse, PostToolUse, SessionStart 자동 검증
- **CI/CD 통합**: GitHub Actions 자동 파이프라인

## 주요 성과 지표

### 정량적 개선
- **개발자 시간 절약**: 30-60분 → 2-3분 (95% 감소)
- **사용자 편의성**: 복잡한 pip 명령어 → 한 번의 moai update
- **디스크 사용량**: 800KB/프로젝트 → 30KB/프로젝트 (96% 절약)
- **설치 속도**: 파일 복사 → 심볼릭 링크 (10배 개선)

### 정성적 개선
- **개발자 경험**: 반복 작업 제거로 핵심 개발에 집중
- **사용자 경험**: 간단한 명령어로 항상 최신 상태 유지
- **품질 보장**: 자동 검증으로 일관성 보장
- **확장성**: 새로운 파일 패턴 쉽게 추가

## 핵심 가치

MoAI-ADK는 단순히 개발 도구가 아닌, **AI 시대의 새로운 개발 패러다임**을 제시합니다:

- **명세 중심**: 모든 개발이 명세에서 시작
- **자동화 우선**: 반복 작업은 AI가 처리
- **품질 내재**: 품질이 개발 과정에 자연스럽게 통합
- **완전 추적**: 모든 요구사항과 구현의 연결고리 보장

이를 통해 개발자는 창의적이고 핵심적인 작업에 집중할 수 있으며, 팀은 일관된 품질의 소프트웨어를 빠르게 개발할 수 있습니다.

---

*Based on Claude Code Official Documentation v2025, with enterprise-grade development environment and intelligent automation*
