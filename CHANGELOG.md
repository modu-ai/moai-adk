# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [v0.3.5] - 2025-10-10

### 🏗️ Architecture

- **Hooks 소스 관리 독립화**
  - `moai-adk-ts` 패키지 제거 준비를 위한 hooks 소스 구조 재설계
  - 새로운 디렉토리 구조: `src/hooks/` → `.claude-plugin/hooks/scripts/`
  - TypeScript 소스를 MoAI-ADK 루트에서 직접 관리
  - `tsup` 기반 빌드 시스템으로 CommonJS(.cjs) 파일 자동 생성

### ✨ Added

- **빌드 시스템**:
  - `tsconfig.hooks.json` - Hooks 전용 TypeScript 설정
  - `tsup.hooks.config.ts` - tsup 빌드 설정
  - `package.json` - 빌드 의존성 (tsup, typescript, @types/node)
  - `build-hooks.sh` - 빌드 자동화 스크립트

- **Hooks 소스 디렉토리** (`src/hooks/`):
  - `session-notice/` - 세션 시작 알림 (버전 표시 개선 포함)
  - `tag-enforcer/` - TAG 규칙 검증
  - `pre-write-guard/` - 파일 쓰기 보안 검증
  - `policy-block/` - 위험 명령어 차단
  - `types.ts` - 공통 타입 정의

### 🔧 Changed

- **개발 워크플로우**:
  - Hooks 수정: `src/hooks/` 편집 → `./build-hooks.sh` 실행
  - 빌드 결과: `.claude-plugin/hooks/scripts/*.cjs` 자동 생성
  - 플러그인 배포: `.claude-plugin/` 폴더만 복사하면 완료

### 📝 Technical Details

- **독립적인 빌드 시스템**: MoAI-ADK 패키지에서 hooks 빌드 완전 분리
- **자동화된 빌드**: `npm run build:hooks` 또는 `./build-hooks.sh`
- **향후 계획**: `moai-adk-ts` 패키지 제거 예정

---

## [v0.3.4] - 2025-10-10

### 🐛 Fixed

- **Session Notice Version Display**
  - `getMoAIVersion()` 함수가 이제 실제 설치된 MoAI-ADK 플러그인 버전을 올바르게 표시합니다
  - 이전에는 `.moai/config.json`의 프로젝트 버전(0.0.3)을 잘못 표시했음
  - 이제 `~/.claude/plugins/marketplaces/moai-adk/.claude-plugin/plugin.json`에서 플러그인 버전(0.3.4)을 정확히 읽어옴
  - GitHub 최신 버전과 비교하여 업데이트 가능 여부를 올바르게 판단

### 📝 Technical Details

- **우선순위 체계**:
  1. `~/.claude/plugins/marketplaces/moai-adk/.claude-plugin/plugin.json` (설치된 플러그인)
  2. `~/.claude/plugins/cache/moai-adk/.claude-plugin/plugin.json` (캐시된 플러그인)
  3. `node_modules/moai-adk/package.json` (npm 패키지, 개발용)
  4. `'unknown'` (fallback)

---

## [v0.3.0] - 2025-10-10

### 🎉 Major Features

- **Claude Code Plugin System** (SPEC-PLUGIN-001)
  - `.claude-plugin/` 폴더 구조 추가
  - `plugin.json` - 플러그인 매니페스트
  - `marketplace.json` - 마켓플레이스 정의 (GitHub 저장소 형식)
  - Alfred SuperAgent 및 9개 전문 에이전트
  - 4개 Alfred 커맨드 (`/alfred:1-spec`, `/alfred:2-build`, `/alfred:3-sync`, `/alfred:8-project`)
  - 4개 보안/품질 후크 (tag-enforcer, pre-write-guard, policy-block, session-notice)

### ✨ Added

- **완전한 프로젝트 템플릿 제공**
  - `templates/.claude/` - Claude Code 설정 템플릿
    - agents/alfred/ - 9개 Alfred 에이전트
    - commands/alfred/ - 4개 Alfred 커맨드
    - hooks/alfred/ - 4개 보안/품질 후크
    - output-styles/alfred/ - 4개 출력 스타일 (alfred-pro, beginner-learning, pair-collab, study-deep)
    - settings.json - Claude Code 기본 설정
  - `templates/.moai/` - MoAI-ADK 프로젝트 템플릿
    - config.json - 프로젝트 설정
    - memory/ - development-guide.md, spec-metadata.md
    - project/ - product.md, structure.md, tech.md
    - reports/, specs/ - 리포트/SPEC 저장 폴더
  - `templates/.github/` - GitHub 템플릿
    - workflows/moai-gitflow.yml - GitFlow 자동화
    - PULL_REQUEST_TEMPLATE.md
  - `templates/CLAUDE.md` - 프로젝트 지침 템플릿
  - `templates/.gitignore` - Git 무시 목록 템플릿

### 🔧 Changed

- **`.gitignore` 최적화**
  - 루트 폴더만 무시하도록 패턴 수정 (/ 접두사 추가)
  - `/.moai`, `/.claude/`, `/AGENTS.md`, `/CLAUDE.md` - 루트만 무시
  - `templates/.moai`, `templates/.claude` - GitHub에 포함
  - 로컬 전용 파일/폴더 명확히 분리

- **플러그인 표준 준수**
  - `marketplace.json` - GitHub 저장소 형식으로 변경
  - `plugin.json`, `marketplace.json` - author 필드 객체화
  - Claude Code 공식 문서 기준 적용

### 📦 Installation

플러그인 마켓플레이스 등록 완료:

```bash
# 마켓플레이스 추가
/plugin marketplace add modu-ai/moai-adk

# 플러그인 설치
/plugin install moai-adk@moai-adk

# 설치 확인
/plugin list
```

### 🎯 Breaking Changes

없음 (v0.2.6 → v0.3.0 호환성 유지)

---

## [v0.2.6] - 2025-10-06

### Added (SPEC-INSTALL-001)

- **Install Prompts Redesign - 개발자 경험 개선**
  - 개발자 이름 프롬프트 추가 (Git `user.name` 기본값 제안)
  - Git 필수 검증 (OS별 설치 안내 메시지)
  - SPEC Workflow 프롬프트 (Personal 모드 전용)
  - Auto PR/Draft PR 프롬프트 (Team 모드 전용)
  - Alfred 환영 메시지 (페르소나 일관성)
  - Progressive Disclosure 흐름 (인지 부담 최소화)

### Implementation Details

- `@CODE:INSTALL-001:DEVELOPER-INFO` - 개발자 정보 수집 (`src/cli/prompts/developer-info.ts`)
- `@CODE:INSTALL-001:GIT-VALIDATION` - Git 검증 로직 (`src/utils/git-validator.ts`)
- `@CODE:INSTALL-001:SPEC-WORKFLOW` - SPEC 워크플로우 프롬프트 (`src/cli/prompts/spec-workflow.ts`)
- `@CODE:INSTALL-001:PR-CONFIG` - PR 설정 프롬프트 (`src/cli/prompts/pr-config.ts`)
- `@CODE:INSTALL-001:WELCOME-MESSAGE` - Alfred 환영 메시지 (`src/cli/prompts/welcome-message.ts`)
- `@CODE:INSTALL-001:INSTALL-FLOW` - 설치 흐름 오케스트레이션 (`src/cli/commands/install-flow.ts`)

### Tests

- `@TEST:INSTALL-001` - 6개 테스트 파일 (100% 커버리지)
  - 개발자 정보 수집 테스트
  - Git 검증 테스트
  - SPEC Workflow 프롬프트 테스트
  - PR 설정 테스트
  - 환영 메시지 테스트
  - 통합 테스트 (E2E)

### Fixed

- **테스트 안정화** (8개 테스트 수정)
  - Vitest 모킹 호이스팅 이슈 해결 (`init-noninteractive.test.ts`)
  - 환경 변수 격리 패턴 구현 (`path-validator.test.ts`)
  - 인터페이스 필드 일치성 수정 (`optional-deps.test.ts`)
  - fs 모듈 완전 모킹 (`session-notice.test.ts`)
  - 테스트 통과율: 91.9% → 100% (753/753 tests) ✅

- **VERSION 파일 일치성 유지**
  - VERSION 파일과 package.json 버전 동기화
  - 버전 추적성 100% 확보

### Changed

- **문서 동기화 및 품질 검증**
  - SPEC-INSTALL-001 상태 업데이트 (draft → completed, v0.1.0 → v0.2.0)
  - 동기화 보고서 생성 (`.moai/reports/sync-report-INSTALL-001.md`)
  - TAG 체인 무결성 검증 (32개 TAG, 14개 파일, 100% 추적성)
  - TRUST 5원칙 준수율: 72% → 92% ✅

- **패키지 배포 전략 문서화**
  - AI Agent 시간 기반 타임라인 추가 (Phase 1-3, 3.5-7시간)
  - v0.2.x 버전 정책 명시 (v1.0.0 사용자 승인 필수)
  - 언어별 배포 명령어 가이드 (NPM, PyPI, Maven, Go)
  - 품질 게이트 검증 기준 정의

### Documentation

- SPEC-INSTALL-001 완료 보고서 (`.moai/specs/SPEC-INSTALL-001/spec.md`)
- 동기화 보고서 생성 (`.moai/reports/sync-report-INSTALL-001.md`)
- 배포 전략 가이드 추가 (`CLAUDE.md`, `moai-adk-ts/templates/CLAUDE.md`)
- HISTORY 섹션 업데이트 (v0.2.0 구현 완료 기록)

### Impact

- ✅ 설치 경험 대폭 개선 (Progressive Disclosure)
- ✅ Git 필수화로 버전 관리 보장
- ✅ SPEC Workflow Personal 모드 선택 가능
- ✅ Team 모드 PR 자동화 옵션 제공
- ✅ Alfred 페르소나 일관성 유지
- ✅ 테스트 100% 통과 (프로덕션 배포 준비 완료)
- ✅ TAG 체인 무결성 100% (고아 TAG 없음)

---

## [v0.0.3] - 2025-10-06

### Changed (CONFIG-SCHEMA-001)

- **config.json 스키마 통합 및 표준화**
  - TypeScript 인터페이스와 템플릿 JSON 구조 통합
  - MoAI-ADK 철학 반영: `constitution`, `git_strategy`, `tags`, `pipeline`
  - `locale` 필드 추가 (CLI 다국어 지원)
  - CODE-FIRST 원칙 명시적 보존 (`tags.code_scan_policy.philosophy`)

### Implementation Details

- `@CODE:CONFIG-STRUCTURE-001` - 템플릿 구조 정의 (`templates/.moai/config.json`)
- `src/core/config/types.ts` - MoAIConfig 인터페이스 전면 재정의
- `src/core/config/builders/moai-config-builder.ts` - 빌더 로직 통합
- `src/core/project/template-processor.ts` - 프로세서 인터페이스 통합

### Impact

- ✅ 템플릿 ↔ TypeScript 인터페이스 100% 일치
- ✅ 자기 문서화 config (철학/원칙 명시)
- ✅ 타입 안전성 확보 (컴파일 에러 0개)
- ✅ 하위 호환성 유지 (기존 config 마이그레이션 불필요)

### Documentation

- 스키마 분석 보고서 생성 (`.moai/reports/config-template-analysis.md`)
- 6개 파일 수정 (+273 -51 LOC)

---

## [v0.0.2] - 2025-10-06

### Added (SPEC-INIT-001)

- **TTY 자동 감지 및 비대화형 모드 지원**
  - CI/CD, Docker, Claude Code 등 비대화형 환경 자동 감지
  - `process.stdin.isTTY` 검증을 통한 환경 인식
  
- **`moai init --yes` 플래그 추가**
  - 프롬프트 없이 기본값으로 즉시 초기화
  - 대화형 환경에서도 자동화 가능
  
- **의존성 자동 설치 기능**
  - Git, Node.js 등 필수 의존성 플랫폼별 자동 설치
  - macOS: Homebrew 기반
  - Linux: apt 기반
  - Windows: winget 기반 (또는 수동 설치 가이드)
  - nvm 우선 사용 (sudo 회피)
  
- **선택적 의존성 분리**
  - Git LFS, Docker는 선택적 의존성으로 분류
  - 누락 시 경고만 표시하고 초기화 계속 진행

### Implementation Details

- `@CODE:INIT-001:TTY` - TTY 감지 로직 (`src/utils/tty-detector.ts`)
- `@CODE:INIT-001:INSTALLER` - 의존성 자동 설치 (`src/core/installer/dependency-installer.ts`)
- `@CODE:INIT-001:HANDLER` - 대화형/비대화형 핸들러 (`src/cli/commands/init/*.ts`)
- `@CODE:INIT-001:ORCHESTRATOR` - 전체 오케스트레이션 (`src/cli/commands/init/index.ts`)
- `@CODE:INIT-001:DOCTOR` - 선택적 의존성 분리

### Tests

- `@TEST:INIT-001` - 전체 테스트 커버리지 85%+
- 비대화형 환경 시나리오 테스트 완료
- TTY 감지 로직 단위 테스트
- 의존성 설치 통합 테스트

### Changed (SPEC-BRAND-001)

- **CLAUDE.md 브랜딩 통일**
  - "Claude Code 워크플로우" → "MoAI-ADK 워크플로우"
  - "Claude Code 설정" → "MoAI-ADK 설정"
  - 프로젝트 정체성 강화

### Fixed (SPEC-REFACTOR-001)

- **Git Manager TAG 체인 수정 및 통일**
  - `@CODE:REFACTOR-001:BRANCH` - Git branch operations
  - `@CODE:REFACTOR-001:COMMIT` - Git commit operations
  - `@CODE:REFACTOR-001:PR` - Pull Request operations
  - TAG 추적성 매트릭스 완성

### Documentation

- TAG 추적성 매트릭스 업데이트 (`.moai/reports/tag-traceability-INIT-001.md`)
- 동기화 보고서 생성 (`.moai/reports/sync-report-INIT-001.md`)
- CHANGELOG.md 신규 생성

---

## [v0.0.1] - 2025-09-15

### Added

- **초기 MoAI-ADK 프로젝트 설정**
  - Alfred SuperAgent 및 9개 전문 에이전트 생태계 구축
  - SPEC-First TDD 워크플로우 구현
  - @TAG 시스템 기반 추적성 보장
  - TRUST 5원칙 자동 검증
  - 다중 언어 지원 (TypeScript, Python, Java, Go, Rust, Dart, Swift, Kotlin)
  - Personal/Team 모드 지원
  - Claude Code 통합

### CLI Commands

- `/alfred:1-spec` - EARS 명세 작성
- `/alfred:2-build` - TDD 구현
- `/alfred:3-sync` - 문서 동기화
- `/alfred:8-project` - 프로젝트 초기화

### Foundation

- Development Guide (`development-guide.md`) 작성
- TRUST 5원칙 (Test First, Readable, Unified, Secured, Trackable) 정의
- CODE-FIRST @TAG 시스템 구현
- GitFlow 통합 전략 수립

---

## Upgrade Guide

### v0.0.1 → v0.0.2

**Breaking Changes**: 없음

**New Features**:
- `moai init` 명령어가 이제 비대화형 환경을 자동으로 감지합니다
- `--yes` 플래그를 사용하여 자동화된 초기화가 가능합니다

**Migration Steps**:
1. Claude Code 플러그인 업데이트: `/plugin` 메뉴에서 업데이트 또는 재설치
2. (선택적) CI/CD 스크립트에서 `moai init --yes` 사용 (레거시)

---

## Roadmap

### Future

- **프로젝트 초기화 개선**: Windows 환경 지원 강화
  - WSL 지원 전략
  - Windows 멀티 플랫폼 테스트

- Living Document 자동 생성 강화
- TAG 검색 및 네비게이션 도구
- 웹 UI 대시보드
- VS Code Extension

---

**참고 자료**:
- [GitHub Repository](https://github.com/modu-ai/moai-adk)
- [Documentation](https://docs.moai-adk.dev)
- [SPEC 디렉토리](.moai/specs/)
- [Development Guide](.moai/memory/development-guide.md)

**기여하기**:
- [Issues](https://github.com/modu-ai/moai-adk/issues)
- [Discussions](https://github.com/modu-ai/moai-adk/discussions)
- [Contributing Guide](CONTRIBUTING.md)
