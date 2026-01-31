# v1.12.10 - MoAI-GO Experimental Implementation & UTF-8 Encoding Fixes (2026-01-31)

## Summary

Experimental release introducing MoAI-GO, a high-performance Go-based CLI implementation, alongside critical UTF-8 encoding fixes for Windows compatibility.

**Key Features**:
- **MoAI-GO System**: Experimental Go implementation (Phases 1-4) for performance evaluation
- **UTF-8 Encoding**: Comprehensive UTF-8 encoding fixes for Windows (#314, #316)
- **Cross-Platform**: Improved PATH handling and hook execution

**Note**: MoAI-GO is experimental. Production use still relies on Python implementation.

## Breaking Changes

None. All changes are backward compatible.

## Added

### MoAI-GO Experimental System

- **feat(go)**: Implement MoAI-GO Phase 1 - Foundation and configuration system (b36ddd66)
  - Go-based configuration management
  - Cross-platform compatibility layer
  - Performance-optimized file operations

- **feat(go)**: Implement MoAI-GO Phase 2 - Hook system with Python compatibility (f3e75202)
  - Go-based hook execution engine
  - Backward compatible with Python hooks
  - Improved hook performance

- **feat(go)**: Implement MoAI-GO Phase 3 - CLI commands with direct binary paths (e47534d0)
  - Go-based CLI commands: `rank`, `switch`, `worktree`
  - Direct binary execution (no Python interpreter needed)
  - Cross-platform PATH handling improvements

- **feat(go)**: Implement MoAI-GO Phase 4 - Distribution and migration system (d61f381d)
  - Binary distribution system
  - Migration tools from Python to Go
  - Version compatibility checks

- **feat**: Add MoAI-GO CLI commands (rank, switch, worktree) and improve cross-platform PATH handling (05d7b7be)
  - Integrated CLI commands with improved performance
  - Enhanced PATH resolution for Windows/macOS/Linux
  - Automatic binary fallback to Python implementation

## Fixed

### UTF-8 Encoding Issues

- **fix**: Add UTF-8 encoding to all hook scripts and file operations (33c0a483)
  - Fixed UTF-8 encoding in all `.claude/hooks/moai/*.py` files
  - Added explicit `encoding='utf-8'` to all file operations
  - Ensures correct handling of non-ASCII characters

- **fix**: Resolve cross-platform hook/encoding/credential issues (#314, #316) (bf280eeb)
  - Issue #314: Hook execution failures on Windows due to encoding
  - Issue #316: Credential file corruption with non-ASCII characters
  - Fix: Consistent UTF-8 encoding across all platforms
  - Impact: Windows users can now use hooks without encoding errors

- **fix(cli)**: Resolve Windows UTF-8 encoding error on moai init (894abd3e)
  - Issue: `moai init` failed on Windows with UnicodeEncodeError
  - Root cause: Console output encoding mismatch
  - Fix: Force UTF-8 encoding for all console output
  - Impact: `moai init` now works correctly on Windows

## Changed

### Go Module Restructure

- **refactor**: Restructure Go module path to github.com/modu-ai/moai-adk/go (97be74a7)
  - Aligned Go module path with repository structure
  - Improved import clarity for Go developers
  - Prepared for future Go module distribution

- **revert**: Undo Go module restructure (testing needed) (324097d1)
  - Reverted Go module path change for further testing
  - Ensures stability before finalizing module structure

- **refactor**: Rename moai-adk to moai (2b9bf0e1)
  - Internal renaming for consistency
  - Updated import statements
  - Aligned with unified branding

## Installation & Update

```bash
# Update to the latest version
uv tool update moai-adk

# Update project templates in your folder
moai update

# Verify version
moai --version
```

---

# v1.12.10 - MoAI-GO 실험적 구현 및 UTF-8 인코딩 수정 (2026-01-31)

## 요약

MoAI-GO 고성능 Go 기반 CLI 구현과 Windows 호환성을 위한 UTF-8 인코딩 수정을 포함하는 실험적 릴리스입니다.

**주요 기능**:
- **MoAI-GO 시스템**: 성능 평가를 위한 실험적 Go 구현 (Phase 1-4)
- **UTF-8 인코딩**: Windows용 포괄적 UTF-8 인코딩 수정 (#314, #316)
- **크로스플랫폼**: 개선된 PATH 처리 및 hook 실행

**참고**: MoAI-GO는 실험적입니다. 프로덕션 환경에서는 Python 구현을 사용하세요.

## Breaking Changes

없음. 모든 변경사항은 하위 호환됩니다.

## 추가됨

### MoAI-GO 실험적 시스템

- **feat(go)**: MoAI-GO Phase 1 - Foundation 및 구성 시스템 구현 (b36ddd66)
  - Go 기반 구성 관리
  - 크로스플랫폼 호환성 레이어
  - 성능 최적화된 파일 작업

- **feat(go)**: MoAI-GO Phase 2 - Python 호환성을 갖춘 Hook 시스템 (f3e75202)
  - Go 기반 hook 실행 엔진
  - Python hook과 하위 호환성
  - 향상된 hook 성능

- **feat(go)**: MoAI-GO Phase 3 - 직접 바이너리 경로를 사용하는 CLI 명령 (e47534d0)
  - Go 기반 CLI 명령: `rank`, `switch`, `worktree`
  - 직접 바이너리 실행 (Python 인터프리터 불필요)
  - 크로스플랫폼 PATH 처리 개선

- **feat(go)**: MoAI-GO Phase 4 - 배포 및 마이그레이션 시스템 (d61f381d)
  - 바이너리 배포 시스템
  - Python에서 Go로 마이그레이션 도구
  - 버전 호환성 검사

- **feat**: MoAI-GO CLI 명령 (rank, switch, worktree) 추가 및 크로스플랫폼 PATH 처리 개선 (05d7b7be)
  - 성능이 향상된 통합 CLI 명령
  - Windows/macOS/Linux를 위한 향상된 PATH 해결
  - Python 구현으로의 자동 바이너리 폴백

## 수정됨

### UTF-8 인코딩 이슈

- **fix**: 모든 hook 스크립트 및 파일 작업에 UTF-8 인코딩 추가 (33c0a483)
  - 모든 `.claude/hooks/moai/*.py` 파일의 UTF-8 인코딩 수정
  - 모든 파일 작업에 명시적 `encoding='utf-8'` 추가
  - 비ASCII 문자의 올바른 처리 보장

- **fix**: 크로스플랫폼 hook/encoding/credential 이슈 해결 (#314, #316) (bf280eeb)
  - 이슈 #314: 인코딩으로 인한 Windows hook 실행 실패
  - 이슈 #316: 비ASCII 문자로 인한 credential 파일 손상
  - 해결: 모든 플랫폼에서 일관된 UTF-8 인코딩
  - 영향: Windows 사용자가 인코딩 오류 없이 hook을 사용할 수 있음

- **fix(cli)**: moai init의 Windows UTF-8 인코딩 오류 해결 (894abd3e)
  - 문제: Windows에서 `moai init` 실행 중 UnicodeEncodeError 발생
  - 근본 원인: 콘솔 출력 인코딩 불일치
  - 해결: 모든 콘솔 출력에 UTF-8 인코딩 강제
  - 영향: Windows에서 `moai init`가 올바르게 작동

## 변경됨

### Go 모듈 구조 재구성

- **refactor**: Go 모듈 경로를 github.com/modu-ai/moai-adk/go로 재구성 (97be74a7)
  - Go 모듈 경로를 리포지토리 구조와 정렬
  - Go 개발자를 위한 import 명확성 개선
  - 향후 Go 모듈 배포 준비

- **revert**: Go 모듈 구조 재구성 취소 (테스트 필요) (324097d1)
  - 추가 테스트를 위해 Go 모듈 경로 변경 되돌림
  - 모듈 구조 확정 전 안정성 보장

- **refactor**: moai-adk를 moai로 이름 변경 (2b9bf0e1)
  - 일관성을 위한 내부 이름 변경
  - import 문 업데이트
  - 통합 브랜딩과 정렬

## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
uv tool update moai-adk

# 프로젝트 폴더 템플릿 업데이트
moai update

# 버전 확인
moai --version
```

---

