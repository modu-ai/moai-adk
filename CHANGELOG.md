# v1.1.0 - Comprehensive Enhancement: Backup System, TAG System v2.0, Performance & Quality (2026-01-13)

## Summary

Major feature release introducing comprehensive backup system improvements with metadata tracking and automatic cleanup, TAG System v2.0 for flexible validation, 85%+ test coverage achievement, and critical performance optimizations including parallel execution as default. This release also includes important bug fixes for Windows emoji encoding and statusline UTF-8 support.

## Added

### Backup System Enhancement
- **feat(backup)**: Improve backup system with metadata and auto-cleanup
  - Add `backup_metadata.json` to track backup contents and excluded items
  - Unify backup exclusion paths across all modules (specs, reports, project, config/sections)
  - Add `list_backups()` and `cleanup_old_backups()` methods to `TemplateBackup`
  - Integrate automatic backup cleanup in `moai update` command (keep last 5 backups)
  - Add comprehensive tests for new backup functionality (43 tests)

### TAG System v2.0
- **feat(tag-system)**: Implement TAG System v2.0 with flexible validation
  - Flexible validation engine supporting multi-language tags
  - Pre-commit tag validation hook for quality assurance
  - Multi-language tag support (KO, JA, ZH, EN)
  - Improved tag parsing and linkage validation
  - Integration with `moai init` and `moai update` workflows

### Performance & Workflow
- **feat(workflow)**: Parallel execution is now the default mode (#255)
  - All independent tasks execute in parallel by default
  - Add `--sequential` option to opt-out and run tasks sequentially when needed
  - Significantly improves workflow performance for multi-step operations
  - Better utilization of system resources
  - Related PR: #255

- **feat(worktree)**: Add `moai-wt done` command for streamlined workflow completion
  - One-command completion: checkout main → merge branch → remove worktree
  - Optional `--push` flag to push merged changes to remote
  - Automatic feature branch cleanup after merge
  - Simplifies Phase 3 (Merge and Cleanup) workflow

### Test Coverage
- **test(coverage)**: Achieve 85%+ coverage across core modules
  - Add comprehensive TDD tests for `init.py` (88.12% coverage)
  - Add integration tests for core commands
  - Add edge case tests for statusline and config modules
  - Add model allocator tests with comprehensive coverage

### Statusline & CLI
- **feat(statusline)**: Add comprehensive statusline enhancements
  - UTF-8 encoding support for international characters
  - Enable memory and directory display
  - Enhanced output style detection
  - Improved multilingual support

- **feat(cli)**: Add multilingual prompt translations
  - KO/JA/ZH prompt translations for init workflow
  - Improved localization support

## Fixed

### Hooks & Performance
- **fix(hooks)**: Prevent session_start hook from hanging on slow git operations (#254)
  - Fixed blocking issue where slow git commands caused startup delays
  - Improved timeout handling for git operations
  - Enhanced reliability of session initialization
  - Related PR: #254

- **fix(hooks)**: Run session_end hook in background to prevent exit delays
  - Session exit now completes instantly without waiting for cleanup
  - Background processing for auto-cleanup and rank submission
  - Eliminates ~3 second delay when closing Claude Code sessions

### CLI & Encoding
- **fix(cli)**: Fix Windows emoji encoding error (#256)
  - Resolve emoji display issues on Windows platforms
  - Ensure consistent emoji rendering across platforms

- **fix(statusline)**: Add UTF-8 encoding support
  - Fix encoding issues in statusline display
  - Ensure proper character encoding for multilingual content

- **fix(commands)**: Align tool permissions with CLAUDE.md Command Types policy
  - Ensured all commands follow documented permission policies
  - Improved security and consistency across command execution
  - Better alignment with Type A, Type B, and Type C command classifications

### Type Safety
- **fix(types)**: Resolve mypy type checking errors
  - Fix type annotations across core modules
  - Improve type safety and IDE support

### Git Worktree
- **fix(gitignore)**: Add llm-configs to tracked directories
  - Fix `.gitignore` configuration to properly track `llm-configs/` directory in git worktrees
  - Resolve issue where `moai glm` command failed in worktree environments

## Changed

### Configuration
- **chore(config)**: Sync template and local configurations
  - Synchronize `.moai/config/` with latest templates
  - Update `system.yaml`, `quality.yaml`, `language.yaml` configurations
  - Add new configuration options for TAG system and coverage targets

### Documentation
- **docs(tag)**: Add TAG system activation step to installation wizard
  - Document TAG system setup process
  - Add TAG system usage examples

- **docs(project)**: Sync project documentation with TAG System v2.0
  - Update `product.md`, `structure.md`, `tech.md`
  - Document new TAG system features

- **docs(readme)**: Add Step 2 session sync to MoAI Rank guide (EN/JA/ZH)
  - Document session synchronization workflow

- **docs(release)**: Improve CHANGELOG generation guide
  - Update CHANGELOG generation process

## Installation & Update

```bash
# Install
uv tool install moai-adk
pip install moai-adk==1.1.0

# Update existing installation
uv tool update moai-adk
pip install --upgrade moai-adk

# Verify version
moai --version
```

## Migration Guide

No breaking changes. Existing workflows will automatically benefit from:
- Automatic backup cleanup (keeps last 5 backups)
- Enhanced backup metadata for better tracking
- Parallel execution as default (use `--sequential` to opt-out)
- TAG System v2.0 validation (opt-in via configuration)

## Quality

- All 43 new backup tests passing (100% pass rate)
- 85%+ test coverage achieved for core modules
- Comprehensive integration test suite added
- Type safety verified through mypy

## Documentation

- Updated CHANGELOG generation guide
- Added TAG system documentation
- Updated multilingual README files (KO, JA, ZH)
- Added TESTING_GUIDE.md for contributors

---

# v1.1.0 - 포괄적 개선: 백업 시스템, TAG 시스템 v2.0, 성능 및 품질 (2026-01-13)

## 요약

백업 시스템의 메타데이터 추적 및 자동 정리 기능이 포함된 포괄적인 개선과 TAG 시스템 v2.0의 유연한 검증 기능을 도입한 주요 기능 릴리스입니다. 또한 85%+ 테스트 커버리지 목표를 달성했으며, 병렬 실행을 기본값으로 하는 중요한 성능 최적화가 포함되어 있습니다. Windows 이모지 인코딩 및 statusline UTF-8 지원에 대한 중요한 버그 수정도 포함되어 있습니다.

## 추가됨

### 백업 시스템 개선
- **feat(backup)**: 백업 시스템 개선 (메타데이터 및 자동 정리)
  - 백업 내용 및 제외 항목 추적을 위한 `backup_metadata.json` 추가
  - 모든 모듈에서 백업 제외 경로 통일 (specs, reports, project, config/sections)
  - `TemplateBackup`에 `list_backups()` 및 `cleanup_old_backups()` 메서드 추가
  - `moai update` 명령어에 자동 백업 정리 통합 (최근 5개 백업 유지)
  - 새로운 백업 기능에 대한 포괄적인 테스트 추가 (43개 테스트)

### TAG 시스템 v2.0
- **feat(tag-system)**: TAG 시스템 v2.0 구현 (유연한 검증)
  - 다국어 태그 지원 유연한 검증 엔진
  - 품질 보증을 위한 pre-commit 태그 검증 훅
  - 다국어 태그 지원 (KO, JA, ZH, EN)
  - 향상된 태그 파싱 및 연결 검증
  - `moai init` 및 `moai update` 워크플로우와의 통합

### 성능 및 워크플로우
- **feat(workflow)**: 병렬 실행이 이제 기본 모드입니다 (#255)
  - 모든 독립적인 작업이 기본적으로 병렬로 실행됩니다
  - 필요할 때 순차 실행을 위한 `--sequential` 옵션 추가
  - 다단계 작업의 워크플로우 성능 대폭 향상
  - 시스템 리소스 활용 개선
  - 관련 PR: #255

- **feat(worktree)**: 워크플로우 완료를 위한 `moai-wt done` 명령어 추가
  - 한 번의 명령으로 완료: checkout main → 브랜치 병합 → worktree 제거
  - 병합된 변경사항을 원격에 푸시하는 `--push` 옵션
  - 병합 후 자동 feature 브랜치 정리
  - Phase 3 (병합 및 정리) 워크플로우 간소화

### 테스트 커버리지
- **test(coverage)**: 핵심 모듈에서 85%+ 커버리지 달성
  - `init.py`를 위한 포괄적인 TDD 테스트 (88.12% 커버리지)
  - 핵심 명령어에 대한 통합 테스트 추가
  - statusline 및 config 모듈에 대한 엣지 케이스 테스트
  - 포괄적인 커버리지를 갖는 모델 할당자 테스트

### Statusline 및 CLI
- **feat(statusline)**: 포괄적인 statusline 개선
  - 국제 문자를 위한 UTF-8 인코딩 지원
  - 메모리 및 디렉토리 표시 활성화
  - 향상된 출력 스타일 감지
  - 개선된 다국어 지원

- **feat(cli)**: 다국어 프롬프트 번역 추가
  - init 워크플로우를 위한 KO/JA/ZH 프롬프트 번역
  - 개선된 현지화 지원

## 수정됨

### 훅 및 성능
- **fix(hooks)**: 느린 git 작업으로 인한 session_start hook hang 방지 (#254)
  - 느린 git 명령어가 시작 지연을 유발하는 블로킹 문제 수정
  - git 작업에 대한 타임아웃 처리 개선
  - 세션 초기화의 안정성 향상
  - 관련 PR: #254

- **fix(hooks)**: 종료 지연 방지를 위해 session_end hook을 백그라운드에서 실행
  - 세션 종료가 정리 작업을 기다리지 않고 즉시 완료됨
  - auto-cleanup 및 rank 제출의 백그라운드 처리
  - Claude Code 세션 종료 시 ~3초 지연 제거

### CLI 및 인코딩
- **fix(cli)**: Windows 이모지 인코딩 오류 수정 (#256)
  - Windows 플랫폼에서의 이모지 표시 문제 해결
  - 모든 플랫폼에서 일관된 이모지 렌더링 보장

- **fix(statusline)**: UTF-8 인코딩 지원 추가
  - statusline 표시의 인코딩 문제 수정
  - 다국어 콘텐츠를 위한 적절한 문자 인코딩 보장

- **fix(commands)**: CLAUDE.md Command Types 정책에 맞춰 도구 권한 정렬
  - 모든 명령어가 문서화된 권한 정책을 따르도록 보장
  - 명령어 실행 전반의 보안 및 일관성 개선
  - Type A, Type B, Type C 명령어 분류와의 더 나은 정렬

### 타입 안전성
- **fix(types)**: mypy 타입 검사 오류 해결
  - 핵심 모듈의 타입 어노테이션 수정
  - 타입 안전성 및 IDE 지원 개선

### Git Worktree
- **fix(gitignore)**: llm-configs를 추적 디렉토리에 추가
  - git worktree에서 `llm-configs/` 디렉토리를 올바르게 추적하도록 `.gitignore` 구성 수정
  - worktree 환경에서 `moai glm` 명령어 실패 문제 해결

## 변경됨

### 구성
- **chore(config)**: 템플릿 및 로컬 구성 동기화
  - 최신 템플릿과 `.moai/config/` 동기화
  - `system.yaml`, `quality.yaml`, `language.yaml` 구성 업데이트
  - TAG 시스템 및 커버리지 목표를 위한 새로운 구성 옵션 추가

### 문서
- **docs(tag)**: 설치 마법사에 TAG 시스템 활성화 단계 추가
  - TAG 시스템 설정 프로세스 문서화
  - TAG 시스템 사용 예제 추가

- **docs(project)**: TAG 시스템 v2.0으로 프로젝트 문서 동기화
  - `product.md`, `structure.md`, `tech.md` 업데이트
  - 새로운 TAG 시스템 기능 문서화

- **docs(readme)**: MoAI Rank 가이드에 Step 2 세션 동기화 추가 (EN/JA/ZH)
  - 세션 동기화 워크플로우 문서화

- **docs(release)**: CHANGELOG 생성 가이드 개선
  - CHANGELOG 생성 프로세스 업데이트

## 설치 및 업데이트

```bash
# 설치
uv tool install moai-adk
pip install moai-adk==1.1.0

# 기존 설치 업데이트
uv tool update moai-adk
pip install --upgrade moai-adk

# 버전 확인
moai --version
```

## 마이그레이션 가이드

Breaking change 없음. 기존 워크플로우는 다음 기능의 혜택을 자동으로 받습니다:
- 자동 백업 정리 (최근 5개 백업 유지)
- 향상된 추적을 위한 백업 메타데이터
- 병렬 실행 기본값 (순차 실행 필요 시 `--sequential` 사용)
- TAG 시스템 v2.0 검증 (구성을 통해 opt-in)

## 품질

- 43개의 새로운 백업 테스트 모두 통과 (100% 통과율)
- 핵심 모듈에서 85%+ 테스트 커버리지 달성
- 포괄적인 통합 테스트 스위트 추가
- mypy를 통한 타입 안전성 검증

## 문서

- CHANGELOG 생성 가이드 업데이트
- TAG 시스템 문서 추가
- 다국어 README 파일 업데이트 (KO, JA, ZH)
- 기여자를 위한 TESTING_GUIDE.md 추가

---

# v1.0.0 - Production Ready Release (2026-01-12)

## Summary

Initial production-ready release of MoAI-ADK, featuring SPEC-First TDD workflow with Alfred SuperAgent, unified moai-core-* skills, and comprehensive project management capabilities.

## Added

- ** Alfred SuperAgent **: Strategic orchestration engine for automated SPEC-Plan-Run-Sync workflow
- ** SPEC-First TDD **: Complete Test-Driven Development methodology with EARS format requirements
- ** moai-core-* Skills **: Unified skill system for domain-specific expertise
- ** Project Management **: Full project lifecycle management from init to documentation
- ** Multilingual Support **: KO/EN/JA/ZH language support
- ** CI/CD Integration **: GitHub Actions workflows for automated testing and deployment

## Installation

```bash
pip install moai-adk==1.0.0
uv tool install moai-adk
```

---

# v1.0.0 - 프로덕션 준비 릴리스 (2026-01-12)

## 요약

MoAI-ADK의 초기 프로덕션 준비 릴리스입니다. Alfred SuperAgent를 통한 자동화된 SPEC-Plan-Run-Sync 워크플로우, 통합 moai-core-* 스킬, 포괄적인 프로젝트 관리 기능을 특징으로 합니다.

## 추가됨

- ** Alfred SuperAgent **: 자동화된 SPEC-Plan-Run-Sync 워크플로우를 위한 전략적 오케스트레이션 엔진
- ** SPEC-First TDD **: EARS 형식 요구사항을 포함한 완전한 테스트 주도 개발 방법론
- ** moai-core-* 스킬 **: 도메인별 전문 지식을 위한 통합 스킬 시스템
- ** 프로젝트 관리 **: init부터 문서화까지 전체 프로젝트 라이프사이클 관리
- ** 다국어 지원 **: KO/EN/JA/ZH 언어 지원
- ** CI/CD 통합 **: 자동화된 테스트 및 배포를 위한 GitHub Actions 워크플로우

## 설치

```bash
pip install moai-adk==1.0.0
uv tool install moai-adk
```

---
