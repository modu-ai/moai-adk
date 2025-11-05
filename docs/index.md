# MoAI-ADK 한국어 문서

MoAI-ADK (Agentic Development Kit) 완전한 한국어 온라인 문서에 오신 것을 환영합니다. 이 문서 시스템은 초보자도 쉽게 MoAI-ADK를 이해하고 시작할 수 있도록 구성되었습니다.

## 📋 문서 목차

### 🚀 시작하기
MoAI-ADK를 처음 사용하는 분들을 위한 기본 안내입니다.

- **[설치 가이드](getting-started/installation.md)** - UV, Python, MCP 설정 등 완전한 설치 과정
- **[기본 개념](getting-started/concepts.md)** - SPEC-First, TDD, @TAG, TRUST 5, Alfred 핵심 개념 이해

### 🎯 사용 가이드
실제 개발 과정을 따라가는 상세한 가이드입니다.

#### Alfred 워크플로우
- **[Alfred 워크플로우 개요](guides/alfred/index.md)** - 4단계 개발 사이클 소개
- **[1단계: 계획 (Plan)](guides/alfred/1-plan.md)** - SPEC 작성과 요구사항 정의
- **[2단계: 실행 (Run)](guides/alfred/2-run.md)** - TDD 개발 사이클 완전 가이드
- **[3단계: 동기화 (Sync)](guides/alfred/3-sync.md)** - 문서 자동화 및 추적성 관리
- **[9단계: 피드백 (Feedback)](guides/alfred/9-feedback.md)** - GitHub Issue 자동 생성

#### SPEC 작성
- **[SPEC 작성 기초](guides/specs/basics.md)** - SPEC 문서 구조와 작성 원칙
- **[EARS 문법 상세](guides/specs/ears.md)** - 5가지 패턴과 실전 예제
- **[TAG 시스템 상세](guides/specs/tags.md)** - @TAG 추적성 시스템 완전 가이드
- **[SPEC 예시 모음](guides/specs/examples.md)** - 실제 프로젝트 SPEC 템플릿 모음

#### TDD 개발
- **[RED 단계 상세](guides/tdd/red.md)** - 실패하는 테스트 작성 전략
- **[GREEN 단계 상세](guides/tdd/green.md)** - 최소 구현과 통화 전략
- **[REFACTOR 단계 상세](guides/tdd/refactor.md)** - 코드 개선과 품질 향상

#### 프로젝트 관리
- **[프로젝트 초기화](guides/project/init.md)** - 새 프로젝트 및 기존 프로젝트 설정
- **[프로젝트 설정](guides/project/config.md)** - .moai/config.json 상세 설정 가이드
- **[배포 가이드](guides/project/deploy.md)** - 프로덕션 배포와 CI/CD 통합

### 📚 기술 참고
상세한 기술 문서와 API 참고 자료입니다.

#### CLI 명령어
- **[CLI 명령어 개요](reference/cli/index.md)** - moai-adk 명령어 구조와 기본 사용법
- **[moai-adk 명령어 상세](reference/cli/moai-adk.md)** - init, doctor, update 등 핵심 명령어
- **[서브커맨드 상세](reference/cli/subcommands.md)** - Alfred 명령어와 옵션

#### 에이전트 시스템
- **[에이전트 시스템 개요](reference/agents/index.md)** - Alfred와 Sub-agent 아키텍처
- **[핵심 에이전트 상세](reference/agents/core.md)** - 10개 핵심 Sub-agent 기능과 역할
- **[전문가 에이전트 상세](reference/agents/experts.md)** - 도메인별 전문가 에이전트 활용

#### 스킬 시스템
- **[스킬 시스템 개요](reference/skills/index.md)** - Claude Skills 아키텍처와 로딩
- **[기반 스킬 상세](reference/skills/foundation.md)** - TRUST, TAG, SPEC 기반 스킬
- **[언어 스킬 상세](reference/skills/languages.md)** - Python, JavaScript, Go 등 언어별 스킬
- **[Alfred 스킬 상세](reference/skills/alfred.md)** - 워크플로우 오케스트레이션 스킬

#### Hook 시스템
- **[Hook 시스템 개요](reference/hooks/index.md)** - Claude Code Hook 아키텍처
- **[세션 Hook 상세](reference/hooks/session.md)** - SessionStart/End Hook 기능
- **[도구 Hook 상세](reference/hooks/tool.md)** - PreToolUse/PostToolUse Hook

#### TAG 시스템
- **[TAG 시스템 개요](reference/tags/index.md)** - @TAG 추적성 시스템 구조
- **[TAG 타입 상세](reference/tags/types.md)** - SPEC, TEST, CODE, DOC TAG 종류
- **[추적성 상세](reference/tags/traceability.md)** - TAG 체인과 무결성 검증

### 🔥 고급 주제
심화된 주제와 전문적인 사용법을 다룹니다.

- **[다국어 지원](advanced/i18n.md)** - 완전한 언어 현지화 시스템
- **[성능 최적화](advanced/performance.md)** - 대규모 프로젝트 성능 튜닝
- **[보안](advanced/security.md)** - OWASP Top 10과 보안 best practices
- **[확장](advanced/extensions.md)** - 커스텀 스킬과 에이전트 개발

### 👥 개발자 가이드
MoAI-ADK 기여와 개발에 참여하는 방법을 안내합니다.

- **[기여하기](contributing/index.md)** - 기여 가이드와 커뮤니티 참여
- **[개발 환경](contributing/development.md)** - 로컬 개발 환경 설정
- **[릴리즈 프로세스](contributing/releases.md)** - 버전 관리와 릴리즈 절차
- **[코드 스타일](contributing/style.md)** - 코딩 표준과 문서 스타일

## 🎯 빠른 시작

처음 사용하시는 분이라면 다음 순서로 문서를 읽어보세요:

1. **[설치 가이드](getting-started/installation.md)** - 5분 안에 개발 환경 구축
2. **[기본 개념](getting-started/concepts.md)** - 핵심 원리 이해
3. **[Alfred 워크플로우 개요](guides/alfred/index.md)** - 개발 흐름 파악
4. **[첫 10분 실습](guides/alfred/1-plan.md#첫-10분-실습-hello-world-api)** - 실제로 만들어보기

## 📖 문서 특징

이 문서 시스템은 다음과 같은 특징을 가지고 있습니다:

- ✅ **초보자 친화적**: 개념부터 실전까지 체계적으로 구성
- ✅ **실무 중심**: 실제 프로젝트에서 바로 사용 가능한 내용
- ✅ **완전한 예제**: 모든 코드 예제는 바로 실행 가능
- ✅ **상세한 다이어그램**: Mermaid를 통한 시각적 이해
- ✅ **추적성 보장**: 관련 섹션간의 링크 제공
- ✅ **문제 해결**: 일반적인 문제와 해결책 제공

## 🤝 지원 및 피드백

문서 개선이나 질문이 있으시면 다음 채널을 이용하세요:

- **GitHub Issues**: [문서 개선 제안](https://github.com/modu-ai/moai-adk/issues)
- **GitHub Discussions**: [질문과 토론](https://github.com/modu-ai/moai-adk/discussions)
- **빠른 피드백**: `/alfred:9-feedback` 명령어로 즉시 이슈 생성

---

**MoAI-ADK**으로 신뢰할 수 있는 AI 협력 개발을 시작해보세요! 🚀