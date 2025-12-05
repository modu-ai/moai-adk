# v0.32.12.1 - Test Coverage Release CI/CD Fix (2025-12-05)

## Summary

Patch release to fix CI/CD deployment issue for v0.32.12.

### Fixes

- **fix**: Remove numpy dependency from test files
  - Fixed import error in test_comprehensive_monitoring_system_coverage.py
  - Replaced numpy arrays with Python lists
  - Ensures all tests run in CI environment

## Previous Improvements (from v0.32.12)

The v0.32.12 release achieved the 95% test coverage target through comprehensive test additions across critical modules, significantly improving code quality and reliability.

## Changes

### Quality Improvements

- **feat**: Achieve 95% test coverage across the codebase
  - Added comprehensive test suites for low-coverage modules
  - Increased from ~90% to 95% overall coverage
  - Total of 1,100+ additional test cases added

### Coverage Improvements

- **comprehensive_monitoring_system.py**: 84.34% → 88.06% (+3.72%)
  - Added 69 test cases covering monitoring, metrics, and alerts
  - Full coverage of data classes and core functionality

- **enterprise_features.py**: 80.13% → 87.37% (+7.24%)
  - Added 125 test cases for enterprise features
  - Comprehensive testing of multi-tenant, deployment, and audit features

- **ears_template_engine.py**: 67.76% → 99.07% (+31.31%)
  - Added 101 test cases covering template generation
  - Near-complete coverage of SPEC generation logic

### Previous Improvements (from v0.32.11)

- confidence_scoring.py: 11.03% → 99.63% (+88.60%)
- worktree/registry.py: 48.70% → 100% (+51.30%)
- language_validator.py: 55.02% → 100% (+44.98%)
- template_variable_synchronizer.py: 64.56% → 98.10% (+33.54%)
- selective_restorer.py: 59.43% → 96.23% (+36.80%)
- error_recovery_system.py: 59.32% → 82.15% (+22.83%)
- jit_enhanced_hook_manager.py: 60.64% → 80.89% (+20.25%)
- realtime_monitoring_dashboard.py: 57.33% → 80.89% (+23.56%)
- event_driven_hook_system.py: 47.06% → 82.05% (+34.99%)

### Configuration

- **config**: Set coverage gate to 95% in pyproject.toml
  - Enforces high code quality standards
  - All new code must maintain 95%+ coverage

## Quality Metrics

- Total test files: 14 dedicated coverage test files
- Total test cases added: 1,100+
- Lines of test code: 16,000+
- Coverage improvement: 14+ percentage points
- Quality gate: 95% (achieved)

## Breaking Changes

None

## Migration Guide

No migration required. This is a quality improvement release.

---

# v0.32.12.1 - 테스트 커버리지 릴리즈 CI/CD 수정 (2025-12-05)

## 요약

v0.32.12의 CI/CD 배포 문제를 수정하는 패치 릴리즈입니다.

### 수정 사항

- **fix**: 테스트 파일에서 numpy 의존성 제거
  - test_comprehensive_monitoring_system_coverage.py import 오류 수정
  - numpy 배열을 Python 리스트로 대체
  - CI 환경에서 모든 테스트 실행 보장

## v0.32.12 개선사항

v0.32.12은 95% 테스트 커버리지 목표를 달성했습니다.

## 변경 사항

### 품질 개선

- **feat**: 코드베이스 전체에 95% 테스트 커버리지 달성
  - 낮은 커버리지 모듈에 대한 포괄적인 테스트 스위트 추가
  - 전체 커버리지를 ~90%에서 95%로 향상
  - 총 1,100개 이상의 추가 테스트 케이스 추가

### 커버리지 개선

- **comprehensive_monitoring_system.py**: 84.34% → 88.06% (+3.72%)
  - 69개 테스트 케이스 추가 (모니터링, 메트릭, 알림)
  - 데이터 클래스와 핵심 기능의 전체 커버리지

- **enterprise_features.py**: 80.13% → 87.37% (+7.24%)
  - 125개 테스트 케이스 추가 (엔터프라이즈 기능)
  - 멀티테넌트, 배포, 감사 기능의 포괄적인 테스트

- **ears_template_engine.py**: 67.76% → 99.07% (+31.31%)
  - 101개 테스트 케이스 추가 (템플릿 생성)
  - SPEC 생성 로직의 거의 완벽한 커버리지

### v0.32.11의 개선사항

- confidence_scoring.py: 11.03% → 99.63% (+88.60%)
- worktree/registry.py: 48.70% → 100% (+51.30%)
- language_validator.py: 55.02% → 100% (+44.98%)
- template_variable_synchronizer.py: 64.56% → 98.10% (+33.54%)
- selective_restorer.py: 59.43% → 96.23% (+36.80%)
- error_recovery_system.py: 59.32% → 82.15% (+22.83%)
- jit_enhanced_hook_manager.py: 60.64% → 80.89% (+20.25%)
- realtime_monitoring_dashboard.py: 57.33% → 80.89% (+23.56%)
- event_driven_hook_system.py: 47.06% → 82.05% (+34.99%)

### 설정

- **config**: pyproject.toml에서 커버리지 게이트를 95%로 설정
  - 높은 코드 품질 표준 시행
  - 모든 새 코드는 95%+ 커버리지 유지 필요

## 품질 메트릭

- 총 테스트 파일: 14개 전용 커버리지 테스트 파일
- 총 추가 테스트 케이스: 1,100+
- 테스트 코드 라인: 16,000+
- 커버리지 향상: 14+ 퍼센트 포인트
- 품질 게이트: 95% (달성됨)

## 호환성 변경

없음

## 마이그레이션 가이드

마이그레이션 불필요. 품질 개선 릴리즈입니다.

---

# v0.32.11 - Release Workflow Simplification & Config Enhancement (2025-12-05)

## Summary

This patch release simplifies the release workflow with tag-based deployment, enhances configuration system with section file support, and separates user-facing output from internal agent data formats.

## Changes

### New Features

- **feat**: Separate user-facing output (Markdown) from internal agent data (XML)
  - User-facing responses now consistently use Markdown formatting
  - XML tags reserved exclusively for agent-to-agent data transfer
  - Clarifies output format usage across all agents and documentation

### Bug Fixes

- **fix**: Implement section files support and detached HEAD detection
  - Added support for modular section file configuration loading
  - Enhanced detached HEAD state detection in language config resolver
  - Improves robustness of configuration system
  - Location: `src/moai_adk/core/language_config_resolver.py`

### Refactoring

- **refactor**: Simplify release workflow with tag-based deployment
  - Streamlined release command with focused tag-based approach
  - Removed complex branching and PR creation logic
  - Single workflow: quality gates → review → tag → GitHub Actions deploy
  - Reduced release.md from complex multi-step to simple 6-phase process
  - Location: `.claude/commands/moai/99-release.md`

### Version Management

- **chore**: Bump version to 0.32.11
  - Version synchronization across all files

## Breaking Changes

None

## Migration Guide

No migration required. This is a workflow improvement and bug fix release.

---

# v0.32.11 - 릴리즈 워크플로우 간소화 및 설정 개선 (2025-12-05)

## 요약

이번 패치 릴리즈는 태그 기반 배포로 릴리즈 워크플로우를 단순화하고, 섹션 파일 지원으로 설정 시스템을 개선하며, 사용자 대면 출력과 내부 에이전트 데이터 형식을 분리합니다.

## 변경 사항

### 신규 기능

- **feat**: 사용자 대면 출력(Markdown)과 내부 에이전트 데이터(XML) 분리
  - 사용자 대면 응답이 이제 일관되게 Markdown 형식 사용
  - XML 태그는 에이전트 간 데이터 전송 전용으로 예약
  - 모든 에이전트와 문서에 걸쳐 출력 형식 사용 명확화

### 버그 수정

- **fix**: 섹션 파일 지원 및 detached HEAD 감지 구현
  - 모듈화된 섹션 파일 설정 로딩 지원 추가
  - 언어 설정 리졸버에서 detached HEAD 상태 감지 개선
  - 설정 시스템의 견고성 향상
  - 위치: `src/moai_adk/core/language_config_resolver.py`

### 리팩토링

- **refactor**: 태그 기반 배포로 릴리즈 워크플로우 단순화
  - 집중된 태그 기반 접근 방식으로 릴리즈 명령어 간소화
  - 복잡한 브랜치 및 PR 생성 로직 제거
  - 단일 워크플로우: 품질 게이트 → 리뷰 → 태그 → GitHub Actions 배포
  - release.md를 복잡한 다단계에서 간단한 6단계 프로세스로 축소
  - 위치: `.claude/commands/moai/99-release.md`

### 버전 관리

- **chore**: 버전을 0.32.11로 업데이트
  - 모든 파일에서 버전 동기화

## 호환성 변경

없음

## 마이그레이션 가이드

마이그레이션 불필요. 워크플로우 개선 및 버그 수정 릴리즈입니다.

---

# v0.32.10 - Worktree Registry Validation & CI/CD Improvements (2025-12-05)