# MoAI-ADK 시스템 개요

> Claude Code 2025 표준 기반 Spec-First TDD 개발 시스템
> 개인/팀 모드 통합 + Git 완전 자동화

## 🎯 목표와 범위

### 목적

- **명세 우선 개발**: 명세가 모든 구현을 주도
- **완전 자동화**: AI 에이전트가 구현 작업 수행
- **품질 내재화**: 단계별 자동 검수와 게이트
- **Living Document**: 코드↔문서 실시간 동기화

### 범위

- **신규 프로젝트**: 그린필드 개발
- **기존 프로젝트**: 브라운필드 개선/리팩토링
- **팀 프로젝트**: 개발 가이드 기반 거버넌스
- **개인 프로젝트**: 간소화 모드 지원

## 🏛️ 핵심 원칙

1. **Spec-First**: 명세 없이 코드 없음
2. **TDD-First**: 테스트 없이 구현 없음
3. **Living Doc**: 문서와 코드는 항상 동기화
4. **Full Traceability**: 모든 요구사항은 추적 가능
5. **YAGNI**: 필요한 것만 구현

## 시스템 특징

### 🤖 개인/팀 모드 통합 + Git 자동화
- **모드 자동 선택**: `moai init --personal|--team`, 필요 시 `/moai:0-project update`로 재조정 (CLI 전환은 보조 용도)
- **체크포인트 시스템(개인)**: 파일 변경 감지 + 5분 주기 Annotated Tag 백업 → 안전한 롤백
- **브랜치 전략(팀)**: `feature/SPEC-XXX-{slug}` + Draft PR → Ready 자동화(옵션)
- **스마트 커밋**: RED→GREEN→REFACTOR 7단계 커밋 패턴, 개발 가이드 기반 메시지
- **Git 전용 명령어 5종**: `/moai:git:checkpoint|rollback|branch|commit|sync`

### 🏗️ 중앙 집중식 아키텍처 (v0.1.13+)
- **패키지 내장 리소스**: importlib.resources 기반 "복사" 방식(심볼릭 링크 미의존)으로 안정성 향상
- **전역 리소스 관리**: ~/.claude/moai/ 통합 구조
- **완전한 메모리 시스템**: Claude/MoAI 메모리 파일, 개발 가이드, ADR 템플릿
- **지능형 Git 시스템**: 자동 설치, 저장소 보존, 운영체제별 최적화
- **자동 업데이트 추적**: `.moai/version.json`으로 템플릿 버전을 관리하고 `moai update --check`에서 즉시 확인

### 🔄 신규 4단계 워크플로우 (/moai:0-project → /moai:3-sync)
1. **/moai:0-project**: 프로젝트 문서(product/structure/tech) 갱신 + CLAUDE 메모리 반영
2. **/moai:1-spec**: 프로젝트 문서 기반 SPEC auto 제안 (개인: 로컬 생성 / 팀: GitHub Issue)
3. **/moai:2-build**: TDD 구현 (개인: 자동 체크포인트 / 팀: 7단계 커밋)
4. **/moai:3-sync**: 문서 동기화 + PR 상태 전환 + TAG 인덱스 갱신

### 🤖 핵심 에이전트 & 브리지
- **project-manager**: `/moai:0-project` 인터뷰, 설정/브레인스토밍 옵션 관리
- **cc-manager**: Claude Code 권한/훅/환경 최적화
- **spec-builder**: SPEC 자동 제안/생성
- **code-builder**: TDD RED→GREEN→REFACTOR 실행
- **doc-syncer**: 문서/PR 동기화 및 TAG 인덱스 관리
- **git-manager**: 체크포인트, 브랜치, 커밋, 동기화 전담
- **codex-bridge** · **gemini-bridge**: 선택적으로 Codex/Gemini CLI를 headless 방식으로 호출해 브레인스토밍/디버깅 아이디어를 수집
- `.moai/config.json.brainstorming` 설정으로 외부 브레인스토밍 사용 여부와 제공자를 관리합니다.

### 🏷️ 16-Core @TAG 추적성 시스템
- **Primary**: @REQ → @DESIGN → @TASK → @TEST
- **Steering**: @VISION → @STRUCT → @TECH → @ADR
- **Quality**: @PERF → @SEC → @DEBT → @TODO

### 🛡️ 품질 보증 시스템
- **개발 가이드 5원칙**: Simplicity, Architecture, Testing, Observability, Versioning
- **Hook 시스템**: PreToolUse, PostToolUse, SessionStart 자동 검증 + 파일감시 체크포인트(개인)
- **CI/CD 통합**: GitHub Actions 자동 파이프라인

## 주요 성과 지표

### 정량적 개선
- **개발자 시간 절약**: 30-60분 → 2-3분 (95% 감소)
- **사용자 편의성**: 복잡한 pip 명령어 → 한 번의 moai update
- **디스크 사용량**: 800KB/프로젝트 → 30KB/프로젝트 (96% 절약)
- **설치 방식**: 플랫폼 무관한 안정적인 파일 복사 방식

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
