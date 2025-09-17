# MoAI-ADK 빌드 시스템 및 버전 관리

> **완전 자동화된 빌드 기반 버전 동기화 시스템**
> 개발자가 실수할 수 없는 완전 자동화된 빌드 프로세스
> **Last Updated**: 2025-09-16 | **Version**: v0.1.17

## 🗿 시스템 개요

MoAI-ADK v0.1.17부터 도입된 새로운 빌드 시스템은 **빌드와 버전 동기화를 완전히 통합**하여 개발자의 실수를 원천적으로 방지합니다. v0.1.17에서는 패키지 구조 재편과 함께 더욱 안정적인 빌드 시스템을 제공합니다.

### 핵심 혁신
- **빌드 = 버전 동기화**: 패키지 빌드 시 자동으로 25개 파일 버전 동기화
- **패키지 구조 검증**: 새로운 cli/, core/, install/ 구조 자동 검증
- **실수 방지**: 개발자가 버전 동기화를 깜빡할 수 없는 구조
- **DevOps 통합**: CI/CD 파이프라인과 자연스러운 연결
- **원클릭 빌드**: `make build` 한 번으로 모든 것 완료

## 🚀 빠른 시작

### 기본 빌드
```bash
# 가장 간단한 방법
make build

# 또는 스크립트 직접 실행
./scripts/build.sh

# 클린 빌드 (이전 아티팩트 제거)
make build-clean
```

### 빌드 결과 확인
```bash
# 빌드 아티팩트 확인
ls -la dist/

# 버전 확인
python -c "from src.moai_adk._version import __version__; print(__version__)"

# 패키지 구조 검증 (v0.1.17 신규)
python -c "import sys; sys.path.insert(0, 'src'); from moai_adk.core import SecurityManager; from moai_adk.cli import CLICommands; print('✅ Package structure verified')"
```

## 📋 빌드 프로세스 상세

### 자동 실행 단계

#### 1️⃣ **빌드 전 훅 실행**
```bash
🗿 MoAI-ADK Pre-Build Hook
============================================================
```
- `build_hooks.py --pre-build` 자동 실행
- 현재 `_version.py`에서 버전 정보 읽기
- 빌드 시작 알림 및 로깅

#### 2️⃣ **자동 버전 동기화**
```bash
🗿 MoAI-ADK Build Hook: Auto-syncing versions for v0.1.17
  ✓ pyproject.toml
  ✓ tests/test_build.py
  ✓ CLAUDE.md
  ✓ docs/MoAI-ADK-Design-Final.md
  ✓ src/moai_adk/cli/banner.py
  ... (25개 파일)
✅ Version sync completed: 25 files updated
```

#### 3️⃣ **패키지 빌드**
```bash
Building packages for MoAI-ADK v0.1.17...
running sdist
running bdist_wheel
```

#### 4️⃣ **패키지 구조 검증** (v0.1.17 신규)
```bash
🗿 MoAI-ADK Package Structure Verification
============================================================
✓ Core modules (9): security, config_manager, template_engine...
✓ CLI modules (4): commands, helpers, wizard, banner
✓ Install modules (2): installer, resource_manager
✓ Legacy code excluded from package
✓ Import paths validated
```

#### 5️⃣ **완료 검증**
```bash
🗿 MoAI-ADK Build Complete
============================================================
✓ Package built successfully
✓ Version synchronization completed
✓ Package structure validated
✓ 25 files updated
```

## 🔧 빌드 명령어 가이드

### Make 명령어 (권장)
```bash
# 기본 빌드 (자동 버전 동기화 포함)
make build

# 강제 빌드 (캐시 무시)
make build-force

# 클린 빌드 (이전 아티팩트 제거 후 빌드)
make build-clean

# 빌드 상태 확인
make status
```

### 직접 스크립트 실행
```bash
# 완전 자동화된 빌드 스크립트
./scripts/build.sh

# 수동 버전 동기화만 실행
python build_hooks.py --sync-only

# 빌드 전 훅만 실행
python build_hooks.py --pre-build
```

### Python 표준 빌드 (권장하지 않음)
```bash
# 버전 동기화 없이 빌드 (권장하지 않음)
python -m build

# 대신 이렇게 사용하세요
make build
```

## 🛠️ 빌드 훅 시스템

### build_hooks.py 구조
```python
def sync_versions_hook():
    """빌드 시 자동 버전 동기화"""
    sync_manager = VersionSyncManager(str(project_root))

    # v0.1.17: 패키지 구조 검증 추가
    if not validate_package_structure():
        raise BuildError("Package structure validation failed")

    result = sync_manager.sync_all_files()
    return result
```

### 지원 옵션
```bash
# 사용 가능한 옵션들
python build_hooks.py --help
python build_hooks.py --sync-only      # 동기화만
python build_hooks.py --verify-only    # 검증만
python build_hooks.py --dry-run        # 시뮬레이션
```

## 📁 동기화 대상 파일 (25개)

### v0.1.17 업데이트된 패키지 구조
```
src/moai_adk/
├── cli/
│   ├── __init__.py
│   ├── commands.py      # ✓ 버전 동기화
│   ├── banner.py        # ✓ 버전 동기화
│   └── helpers.py
├── core/
│   ├── __init__.py
│   ├── template_engine.py  # ✓ 버전 동기화
│   ├── config_manager.py   # ✓ 버전 동기화
│   └── version_sync.py     # ✓ 버전 동기화 (VersionSyncManager)
└── install/
    ├── __init__.py
    └── installer.py     # ✓ 버전 동기화
```

### 자동 업데이트되는 파일들
- **패키지 설정**: `pyproject.toml`, `setup.py`
- **소스 코드**: `src/moai_adk/_version.py`, `src/moai_adk/**/*.py`
- **문서**: `README.md`, `CLAUDE.md`, `docs/*.md`
- **테스트**: `tests/test_*.py`
- **설정**: `.claude/agents/`, `templates/`

### 버전 패턴 자동 교체
```bash
# 자동으로 교체되는 패턴들
__version__ = "<version>"                    # Python 변수
version = "<version>"                        # 설정 파일
MoAI-ADK v<version>                         # 문서
"moai_version": "<version>"                 # JSON 설정
```

## 🎯 개발자 워크플로우

### Before (v0.1.13 이전)
```bash
# 번거로운 수동 과정
1. _version.py 수정
2. 각 파일 개별 버전 업데이트 (25개)
3. 빌드
4. 누락된 파일 재확인 및 수정
```

### After (v0.1.17)
```bash
# 간단한 한 번의 명령어
1. _version.py 수정
2. make build  # 🎯 이것만!
```

### 실제 개발 시나리오
```bash
# 1. 코드 수정 완료 후 버전 변경
echo '__version__ = "<version>"' > src/moai_adk/_version.py

# 2. 빌드 (자동으로 모든 파일 동기화)
make build

# 3. 결과 확인 및 배포
ls -la dist/
python -c "from src.moai_adk import __version__; print(__version__)"
```

## 🔍 검증 및 품질 보증

### 자동 검증 기능
- **버전 일관성**: 25개 파일 버전 자동 검증
- **패키지 구조**: cli/, core/, install/ 구조 검증
- **Import 경로**: 새로운 구조 import 검증
- **빌드 무결성**: 아티팩트 생성 확인

### 수동 검증 방법
```bash
# 현재 패키지 버전 확인
python -c "from src.moai_adk._version import __version__; print(__version__)"

# 패키지 구조 검증
python -c "from src.moai_adk.core import SecurityManager; print('Core module OK')"
python -c "from src.moai_adk.cli import CLICommands; print('CLI module OK')"

# 특정 파일의 버전 확인
grep -r "v0.1.17" docs/ | head -5

# 빌드 아티팩트 확인
ls -la dist/moai_adk-*.whl dist/moai_adk-*.tar.gz
```

## 🚀 CI/CD 통합

### GitHub Actions 예시
```yaml
name: Build and Release
on:
  push:
    tags: ['v*']

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'

      # MoAI-ADK 자동 빌드 시스템 사용
      - name: Build with auto sync
        run: make build

      - name: Upload to PyPI
        run: twine upload dist/*
```

### 자동화 스크립트 예시
```bash
#!/bin/bash
# release.sh

# 빌드 (자동 버전 동기화 포함)
make build

# 버전 정보 추출
VERSION=$(python -c "from src.moai_adk._version import __version__; print(__version__)")

# Git 커밋 및 태그
git add .
git commit -m "🗿 Release v$VERSION - Auto-synced build"
git tag "v$VERSION"

# 배포
twine upload dist/*

# 원격 푸시
git push origin main --tags
```

## 🛡️ 안전장치 및 오류 처리

### 빌드 실패 대응
```bash
# 빌드 실패 시 확인사항
1. Python 버전 확인 (3.11+)
2. 의존성 설치 확인
3. 권한 설정 확인
4. 디스크 공간 확인
```

### 버전 동기화 문제 해결
```bash
# 수동 버전 동기화 재실행
python build_hooks.py --sync-only --force

# 강제 클린 빌드
make build-clean
rm -rf dist/ build/ *.egg-info/
make build
```

### 일반적인 오류와 해결책

#### Makefile 문법 오류
```bash
# 오류: missing separator
# 해결: 탭 문자로 들여쓰기 확인
cat -A Makefile | head -5
```

#### 권한 오류
```bash
# 오류: Permission denied
# 해결: 실행 권한 부여
chmod +x scripts/build.sh
chmod +x build_hooks.py
```

#### 패키지 구조 오류 (v0.1.17 신규)
```bash
# 오류: ModuleNotFoundError
# 해결: 패키지 구조 확인
python -c "import sys; sys.path.insert(0, 'src'); from moai_adk.core import SecurityManager"
```

## 📊 성능 지표

### 빌드 시간 측정
- **전체 빌드**: 평균 15-20초
- **버전 동기화**: 평균 3-5초
- **패키지 구조 검증**: 평균 1-2초
- **아티팩트 생성**: 평균 8-10초

### 동기화 효율성
- **동기화 파일**: 25개 파일 자동 처리
- **정확도**: 100% (정규식 패턴 매칭)
- **백업/복원**: 자동 백업 생성 및 롤백 지원

## 🔮 고급 사용법

### 개발자 전용 기능
```bash
# 드라이 런 모드 (실제 변경 없이 시뮬레이션)
python build_hooks.py --dry-run

# 검증만 실행
python build_hooks.py --verify-only

# 수동 Git 통합
python build_hooks.py --with-git
```

### 커스터마이징
```python
# build_hooks.py 커스터마이징 예시
def custom_build_hook():
    """사용자 정의 빌드 훅"""
    # 패키지 구조 검증
    validate_package_structure()

    # 커스텀 처리
    custom_processing()

    # 표준 버전 동기화
    sync_versions_hook()
```

## 🎉 마이그레이션 가이드

### v0.1.13에서 v0.1.17로
```bash
# 기존 방식 (더 이상 사용하지 않음)
python sync_versions.py
python -m build

# 새로운 방식 (권장)
# 1. _version.py에서 버전 수정
# 2. 빌드 (자동 동기화 포함)
make build
```

## 📚 관련 문서

- **[패키지 구조](package-structure.md)**: cli/, core/, install/ 상세 설명
- **[버전 관리](02-changelog.md)**: v0.1.17 변경사항
- **[설치 가이드](05-installation.md)**: 개발 환경 설정
- **[Constitution](15-constitution.md)**: 품질 게이트 및 검증

---

*📝 빌드 시스템 문의사항이나 이슈가 있다면 GitHub Issues에 등록해 주세요.*
