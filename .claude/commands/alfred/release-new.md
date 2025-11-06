---
name: awesome:release-new
description: 패키지 배포 및 GitHub 릴리즈 자동화
argument-hint: "[patch|minor|major] [--dry-run] [--testpypi] - 버전 타입, --dry-run으로 시뮬레이션, --testpypi로 테스트 배포"
tools: Read, Write, Edit, Bash(git:*), Bash(gh:*), Bash(python:*), Bash(uv:*), Task
---

# 🚀 Awesome Release Automation (Python)

**Python 패키지 릴리즈 자동화 커맨드** - pyproject.toml 기반, PyPI 배포

**버전 관리 방식**: SSOT (Single Source of Truth)
- ✅ 버전은 `pyproject.toml` 한 곳에만 존재
- ✅ `__init__.py`는 `importlib.metadata`로 자동 로드
- ✅ 버전 업데이트는 `pyproject.toml`만 수정

---

## 🎯 커맨드 목적

Python 패키지의 릴리즈 프로세스를 완전 자동화:
1. **품질 검증** (pytest, ruff, mypy, 보안 스캔)
2. **버전 업데이트** (pyproject.toml만, SSOT 방식)
3. Git 커밋 및 태그 생성
4. PyPI 배포 (uv publish 또는 twine)
5. GitHub Release 생성 (자동)

**인수**: `$ARGUMENTS` (예: `patch`, `minor`, `major`, `--dry-run`)

---

## ⚙️ Dry-Run 모드 가이드

**Dry-Run 모드**는 릴리즈 프로세스를 **시뮬레이션만 수행**하고 실제 변경은 하지 않습니다.

### 사용 방법

```bash
# 기본 사용
/awesome:release-new [patch|minor|major] --dry-run

# 예시
/awesome:release-new minor --dry-run   # 시뮬레이션: minor 버전 릴리즈
/awesome:release-new patch --dry-run   # 시뮬레이션: patch 버전 릴리즈
```

### Dry-Run 모드에서 수행되는 작업

✅ **실제 실행 (변경 없음)**:
- 품질 검증 (Phase 0): pytest, ruff, mypy, bandit, pip-audit 실행
- 버전 계산 및 분석
- Git 로그 분석
- 릴리즈 계획 보고서 생성
- 변경사항 요약

❌ **실제 실행 안 함 (시뮬레이션만)**:
- ~~파일 수정 (pyproject.toml)~~
- ~~Git 커밋 생성~~
- ~~Git 태그 생성~~
- ~~GitHub PR 생성~~
- ~~GitHub Release 생성~~
- ~~PyPI 배포~~

### Dry-Run 결과 리포트

Dry-Run 모드 완료 시, 다음과 같은 시뮬레이션 리포트를 출력합니다:

```markdown
🔬 Dry-Run 시뮬레이션 완료 (실제 변경 없음)

📊 시뮬레이션 계획:

Phase 1: 버전 분석
- ✅ 현재 버전: v0.13.0
- ✅ 목표 버전: v0.13.1 (patch)
- ✅ 변경사항: 5개 커밋 (3 fix, 2 docs)

Phase 2: GitFlow PR 병합
- ✅ develop → main PR 생성 (Draft)
- ✅ PR Ready for Review로 전환
- ✅ GitHub에서 병합 (대기)

Phase 3: GitHub Actions 자동 릴리즈
- ✅ Git 태그 생성: v0.13.1
- ✅ GitHub Release 생성: v0.13.1
- ✅ PyPI 배포 시작 (자동)

🎯 실제 릴리즈 명령:
/awesome:release-new patch
```

### 실제 릴리즈 실행

Dry-Run 결과가 만족스러우면, `--dry-run` 플래그를 제외하고 실행하세요:

```bash
# Dry-Run 먼저 확인
/awesome:release-new minor --dry-run

# 결과가 좋으면 실제 릴리즈 실행
/awesome:release-new minor
```

---

## 📋 릴리즈 정보 포맷 표준화

릴리즈 정보는 일관된 형식으로 제공되어야 합니다. 모든 릴리즈는 다음 표준을 따릅니다.

### 릴리즈 정보 구성 (영어 & 한국어)

모든 릴리즈는 **영어와 한국어** 두 언어로 작성되어야 합니다.

**구성 순서**:
1. 🚀 Major Features (주요 기능) - 영어/한국어
2. 📊 Release Statistics (릴리즈 통계)
3. 🧪 Quality Assurance (품질 보증)
4. 💻 Installation Guide (설치 가이드) - **uv tool 중심**
5. 🔗 Documentation (문서)
6. 🔄 Migration & Compatibility (호환성)
7. 👏 Credits (크레딧)

### 설치 가이드 표준 형식

#### ✅ 추천 방식: uv tool (CLI 도구)

```markdown
### 🎯 Recommended: CLI Tool Usage

Use `uv tool` to install moai-adk as a standalone command-line tool (recommended):

```bash
# Install as CLI tool
uv tool install moai-adk==X.Y.Z

# Verify installation
moai-adk --version

# Use as CLI command
moai-adk /alfred:1-plan "새 기능"
```

**Advantages**:
- ✅ Works anywhere (global command)
- ✅ Isolated environment (no conflicts)
- ✅ Easy to update: `uv tool upgrade moai-adk`
- ✅ Recommended for most users
```

#### 대체 방식 1: Python 라이브러리 (pip)

```markdown
### 📚 Alternative: Python Library

If you need to use moai-adk as a Python library in your project:

```bash
# Install with pip (standard)
pip install moai-adk==X.Y.Z

# Use in Python code
from moai_adk import Alfred
```

**Use this if**:
- You're building on top of moai-adk
- You need to import moai-adk in your Python code
- You're managing it as a project dependency
```

#### 대체 방식 2: Python 라이브러리 (uv pip - 빠른 설치)

```markdown
# Or install with uv (faster)
uv pip install moai-adk==X.Y.Z
```

### GitHub Release 기본 템플릿

```markdown
## 🚀 Major Features

### 1. [Feature Name]
[Feature description]

**Key Benefits**:
- Benefit 1
- Benefit 2
- Benefit 3

## 📊 Release Statistics

| Metric | Value |
|--------|-------|
| **Commits** | X since vX.Y.Z |
| **Files Changed** | X files |
| **Lines Added** | X |
| **Lines Removed** | X |

## 🧪 Quality Assurance

✅ **Testing**:
- X tests passed (Y% pass rate)
- Z tests skipped
- 0 test failures

✅ **Code Quality**:
- 0 security issues
- 0 type errors
- X minor linting issues (non-blocking)

## 💻 Installation Guide

### 🎯 Recommended: CLI Tool Usage

[See template above]

### 📚 Alternative: Python Library

[See template above]

## 🔄 Migration & Compatibility

**Breaking Changes**: [None/List]
**Deprecations**: [None/List]
**Migration Required**: [No/Yes with details]
**Backward Compatible**: ✅ Yes

## 👏 Credits

Released with Claude Code

Co-Authored-By: 🎩 Alfred@MoAI
```

### 언어 규칙

| 항목 | 영어 | 한국어 | 예시 |
|------|------|--------|------|
| **Feature 제목** | 영어 | 선택 | "Multi-Language Translation" |
| **설명문** | 영어 + 한국어 혼합 | 괄호로 구분 | "Multi-Language Runtime Translation System (다국어 런타임 번역)" |
| **Installation** | 영어 (코드는 동일) | 코드만 표시 | 코드 블록은 언어 중립적 |
| **Benefits** | 영어 | 선택 옵션 | 영어로 주요 내용, 필요시 한국어 추가 |
| **표 헤더** | 영어 | 필요시 이중 제공 | "Metric \| Value" 또는 "항목 \| 값" |

### 예시: v0.16.0 포맷

**GitHub Release 예시**:
```markdown
## 🚀 Major Features

### 1. 🌐 Multi-Language Runtime Translation System
Single English source with unlimited language support via runtime translation.
(단일 영어 소스에서 무제한 언어 지원)

**Key Benefits**:
- Zero code modification for language support
- Unlimited language support (Korean, Japanese, Chinese, Spanish, etc.)
- Dynamic variable mapping for localization
- Consistent terminology across all languages

### 2. 🏗️ Master-Clone Pattern Architecture
...

## 💻 Installation Guide

### 🎯 Recommended: CLI Tool Usage

Use `uv tool` to install moai-adk as a standalone command-line tool (recommended):

```bash
uv tool install moai-adk==0.16.0
moai-adk --version
```

### 📚 Alternative: Python Library

If you need to use moai-adk as a Python library:

```bash
pip install moai-adk==0.16.0
# or
uv pip install moai-adk==0.16.0
```
```

### CHANGELOG.md 포맷

**CHANGELOG.md의 Installation 섹션**:

```markdown
### 💻 Installation

**Using uv tool** (recommended for CLI usage):
```bash
uv tool install moai-adk==X.Y.Z
moai-adk --version
```

**Using pip** (if you need Python library):
```bash
pip install moai-adk==X.Y.Z
```

**Using uv pip** (faster Python library installation):
```bash
uv pip install moai-adk==X.Y.Z
```
```

### 일관성 체크리스트

각 릴리즈 전에 다음 항목을 확인하세요:

- [ ] Feature 섹션: 3개 이상의 주요 기능 기술
- [ ] 각 Feature: 1-2 문장 설명 + Benefits 나열
- [ ] Installation: uv tool **먼저**, pip/uv pip는 **대체 방식**으로
- [ ] Quality: 테스트, 보안, 타입 체크, 커버리지 포함
- [ ] Statistics: 커밋, 파일, 라인 수 포함
- [ ] Migration: Breaking changes 명시
- [ ] Credits: Claude Code + Alfred 크레딧 포함
- [ ] Links: CHANGELOG, 문서 링크 포함

---

## 🧪 TestPyPI 배포 (테스트 배포)

**TestPyPI**는 PyPI의 테스트 환경입니다. 실제 릴리즈 전에 패키지를 테스트 환경에 배포하여 검증할 수 있습니다.

### TestPyPI란?

- **목적**: 실제 사용자에게 영향을 주지 않고 패키지 배포를 테스트
- **URL**: https://test.pypi.org/
- **특징**:
  - 실제 PyPI와 동일한 환경
  - 실제 배포 전 검증용
  - 테스트 패키지는 30일 후 자동 삭제
  - 독립적인 패키지 저장소 (실제 PyPI와 분리)

### 사용 방법

```bash
# TestPyPI로 테스트 배포
/awesome:release-new minor --testpypi

# Dry-Run + TestPyPI 조합
/awesome:release-new minor --dry-run --testpypi
```

### TestPyPI 배포 워크플로우

```
/awesome:release-new [version] --testpypi
    ↓
Phase 0: 품질 검증 (동일)
├─ pytest, ruff, mypy, bandit, pip-audit
    ↓
Phase 1: 버전 분석 (동일)
├─ 버전 계산, Git 로그 분석
    ↓
Phase 1.5: 릴리즈 계획 보고서 (수정됨)
├─ "PyPI 배포" → "TestPyPI 배포" 표시
    ↓
Phase 2: PR 관리 (생략됨)
├─ GitHub PR/Release 생성 안 함
    ↓
Phase 3: TestPyPI 배포 (수정됨)
├─ Git 태그 생성 안 함
├─ GitHub Release 생성 안 함
└─ TestPyPI에만 배포
    ↓
✅ TestPyPI 배포 완료
└─ 설치 테스트 링크 제공
```

### TestPyPI 설정 (초기 설정 한 번만)

#### 1단계: TestPyPI 토큰 생성

https://test.pypi.org/manage/account/token/ 에서:

```bash
# PyPI 토큰 생성
# - Scope: "Entire account (all projects)"
# - 토큰 형식: pypi-AgEIcHlwaS5vcmcCJ...
```

#### 2단계: 로컬 환경 설정

```bash
# .pypirc 파일 생성 (~/.pypirc)
cat > ~/.pypirc << 'EOF'
[distutils]
index-servers =
    pypi
    testpypi

[pypi]
repository = https://upload.pypi.org/legacy/
username = __token__
password = pypi-AgEIcHlwaS5vcmcCJ...

[testpypi]
repository = https://test.pypi.org/legacy/
username = __token__
password = pypi-AgEIcHlwaS5vcmcCJ...
EOF

chmod 600 ~/.pypirc
```

또는 환경 변수 사용:

```bash
# ~/.bashrc 또는 ~/.zshrc에 추가
export UV_PUBLISH_TOKEN_TESTPYPI="pypi-AgEIcHlwaS5vcmcCJ..."
```

### TestPyPI 배포 실행 단계

#### Step 1: 테스트 배포 실행

```bash
# TestPyPI로 배포
/awesome:release-new patch --testpypi

# 출력 예:
# 🧪 TestPyPI 배포 모드 활성화
# 📊 버전 정보: v0.13.1
# ✅ 품질 검증 통과
# 📤 TestPyPI에 배포 중...
# ✅ TestPyPI 배포 완료!
```

#### Step 2: TestPyPI에서 패키지 확인

```bash
# TestPyPI 프로젝트 페이지
https://test.pypi.org/project/moai-adk/0.13.1/

# 명령줄에서 확인
pip index versions moai-adk -i https://test.pypi.org/simple/
```

#### Step 3: TestPyPI에서 설치 테스트

```bash
# 임시 가상환경 생성
python -m venv /tmp/test_moai
source /tmp/test_moai/bin/activate

# TestPyPI에서 설치
pip install --index-url https://test.pypi.org/simple/ moai-adk==0.13.1

# 기본 테스트
moai-adk --version

# 테스트 완료 후 정리
deactivate
rm -rf /tmp/test_moai
```

#### Step 4: 테스트 완료 후 실제 배포

테스트가 만족스러우면 실제 PyPI에 배포:

```bash
# --testpypi 없이 실행
/awesome:release-new patch
```

### TestPyPI 배포 예시

#### 예시 1: 신규 마이너 버전 테스트

```bash
# v0.14.0 테스트 배포
/awesome:release-new minor --testpypi

# TestPyPI에서 확인
pip install --index-url https://test.pypi.org/simple/ moai-adk==0.14.0

# 테스트 완료 후 실제 배포
/awesome:release-new minor
```

#### 예시 2: 긴급 패치 검증

```bash
# 긴급 패치 미리 테스트
/awesome:release-new patch --testpypi

# TestPyPI에서 설치 및 테스트
pip install --index-url https://test.pypi.org/simple/ moai-adk==0.13.1

# 문제 없으면 실제 배포
/awesome:release-new patch
```

#### 예시 3: Dry-Run + TestPyPI 조합

```bash
# 먼저 시뮬레이션
/awesome:release-new minor --dry-run

# 그 다음 TestPyPI로 테스트
/awesome:release-new minor --testpypi

# 최종 실제 배포
/awesome:release-new minor
```

### TestPyPI 배포 시 주의사항

#### ✅ TestPyPI 배포의 장점

- 실제 배포 전 검증 가능
- 다른 사용자에게 영향 없음
- 패키지 메타데이터 확인 가능
- 설치 테스트로 의존성 확인 가능

#### ⚠️ TestPyPI 배포 시 주의사항

- **GitHub PR/Release 생성 안 함**: TestPyPI 배포는 develop 브랜치용 (main 병합 안 함)
- **Git 태그 생성 안 함**: 테스트 버전이므로 정식 태그 생성 안 함
- **PyPI에는 배포 안 함**: TestPyPI에만 배포 (main 배포 안 함)
- **일반 사용자에게 공개 안 함**: TestPyPI 패키지는 비공개 상태
- **토큰 관리**: TestPyPI 토큰은 별도로 관리 필수

### TestPyPI 트러블슈팅

#### Q: TestPyPI 토큰이 작동하지 않습니다

**A**: 토큰 형식 확인:

```bash
# 토큰은 반드시 "pypi-" 접두사로 시작
echo $UV_PUBLISH_TOKEN_TESTPYPI
# 출력: pypi-AgEIcHlwaS5vcmcCJ...

# 만료된 토큰인 경우 새로 생성
# https://test.pypi.org/manage/account/token/
```

#### Q: TestPyPI에 배포했는데 설치가 안 됩니다

**A**: 인덱스 URL 확인:

```bash
# TestPyPI 인덱스 URL 정확하게
pip install --index-url https://test.pypi.org/simple/ moai-adk==VERSION

# 또는 .pip/pip.conf 확인
cat ~/.config/pip/pip.conf
```

#### Q: TestPyPI에 배포된 패키지를 삭제하고 싶습니다

**A**: TestPyPI 웹에서 Yank 수행:

```
https://test.pypi.org/project/moai-adk/0.13.1/
→ "Release History" → "Yank this version"
```

---

## 📋 실행 흐름

## 🔧 파라미터 처리 및 Dry-Run 모드 초기화

### 파라미터 파싱

```bash
# 인수 처리
# $ARGUMENTS에서 version_type, dry_run, testpypi 플래그 분리

# 기본값
VERSION_TYPE="patch"
DRY_RUN=false
TEST_PYPI=false

# 인수 파싱
for arg in $ARGUMENTS; do
    case "$arg" in
        patch|minor|major)
            VERSION_TYPE="$arg"
            ;;
        --dry-run)
            DRY_RUN=true
            ;;
        --testpypi)
            TEST_PYPI=true
            ;;
        *)
            echo "⚠️ 알 수 없는 인수: $arg"
            echo "사용법: /awesome:release-new [patch|minor|major] [--dry-run] [--testpypi]"
            exit 1
            ;;
    esac
done

# Dry-Run과 TestPyPI 동시 지정 확인
if [ "$DRY_RUN" = "true" ] && [ "$TEST_PYPI" = "true" ]; then
    echo "ℹ️  Dry-Run + TestPyPI 모드"
    echo "   버전 분석은 수행하지만, TestPyPI 배포는 시뮬레이션만 합니다"
fi

# 설정 출력
echo "🚀 릴리즈 설정:"
echo "   - 버전 타입: $VERSION_TYPE"

if [ "$DRY_RUN" = "true" ]; then
    echo "   - 모드: 🔬 Dry-Run (시뮬레이션)"
else
    echo "   - 모드: 실제 릴리즈"
fi

if [ "$TEST_PYPI" = "true" ]; then
    echo "   - 배포 대상: 🧪 TestPyPI (테스트 환경)"
else
    echo "   - 배포 대상: PyPI (프로덕션)"
fi

echo ""
```

### Dry-Run 모드 함수 정의

Dry-Run 모드에서는 실제 파일/Git 수정을 방지하기 위해 래퍼 함수를 사용합니다:

```bash
# Git 커밋 래퍼 (dry-run 모드에서는 로깅만)
git_commit_if_needed() {
    local message="$1"

    if [ "$DRY_RUN" = "true" ]; then
        echo "   [DRY-RUN] Git 커밋 예정: $message"
    else
        git add -A
        git commit -m "$message"
    fi
}

# Git 태그 래퍼 (dry-run 모드에서는 로깅만)
git_tag_if_needed() {
    local tag="$1"
    local message="$2"

    if [ "$DRY_RUN" = "true" ]; then
        echo "   [DRY-RUN] Git 태그 예정: $tag - $message"
    else
        git tag -a "$tag" -m "$message"
        git push origin "$tag"
    fi
}

# GitHub PR 생성 래퍼 (dry-run 모드에서는 로깅만)
gh_pr_create_if_needed() {
    local title="$1"
    local body="$2"

    if [ "$DRY_RUN" = "true" ]; then
        echo "   [DRY-RUN] GitHub PR 생성 예정: $title"
    else
        gh pr create --title "$title" --body "$body" --draft
    fi
}

# 파일 수정 래퍼 (dry-run 모드에서는 로깅만)
file_modify_if_needed() {
    local file="$1"
    local new_value="$2"

    if [ "$DRY_RUN" = "true" ]; then
        echo "   [DRY-RUN] 파일 수정 예정: $file"
    else
        # 실제 파일 수정 로직
        # sed, cat 등을 사용하여 파일 수정
        echo "$new_value" > "$file"
    fi
}

# PyPI 배포 래퍼 (TestPyPI vs 실제 PyPI)
pypi_publish_if_needed() {
    local version="$1"

    if [ "$DRY_RUN" = "true" ]; then
        if [ "$TEST_PYPI" = "true" ]; then
            echo "   [DRY-RUN] TestPyPI 배포 예정: moai-adk==$version"
        else
            echo "   [DRY-RUN] PyPI 배포 예정: moai-adk==$version"
        fi
    else
        if [ "$TEST_PYPI" = "true" ]; then
            echo "📤 TestPyPI에 배포 중..."
            uv publish --publish-url https://test.pypi.org/legacy/
            echo "✅ TestPyPI 배포 완료!"
            echo "🔗 TestPyPI 프로젝트: https://test.pypi.org/project/moai-adk/$version/"
            echo "📦 설치 테스트: pip install --index-url https://test.pypi.org/simple/ moai-adk==$version"
        else
            echo "📤 PyPI에 배포 중..."
            uv publish
            echo "✅ PyPI 배포 완료!"
            echo "🔗 PyPI 프로젝트: https://pypi.org/project/moai-adk/$version/"
        fi
    fi
}
```

### Dry-Run 요약 수집

Dry-Run 모드에서 수행될 모든 작업을 수집하여 마지막에 요약 보고서를 생성합니다:

```bash
# Dry-Run 작업 로그 파일
DRY_RUN_ACTIONS="/tmp/dry_run_actions_$$.log"

# Dry-Run 작업 기록
log_dry_run_action() {
    local action="$1"
    echo "[$(date '+%H:%M:%S')] $action" >> "$DRY_RUN_ACTIONS"
}

# Dry-Run 작업 요약 출력
print_dry_run_summary() {
    if [ "$DRY_RUN" = "true" ]; then
        echo ""
        echo "================================"

        if [ "$TEST_PYPI" = "true" ]; then
            echo "🔬 Dry-Run + 🧪 TestPyPI 시뮬레이션 완료"
        else
            echo "🔬 Dry-Run 시뮬레이션 완료"
        fi

        echo "================================"
        echo ""
        echo "예정된 작업 목록:"
        if [ -f "$DRY_RUN_ACTIONS" ]; then
            cat "$DRY_RUN_ACTIONS"
        fi
        echo ""
        echo "⚠️ 위의 작업들은 시뮬레이션만 수행되었으며, 실제로 적용되지 않았습니다."
        echo ""

        if [ "$TEST_PYPI" = "true" ]; then
            echo "TestPyPI 배포를 진행하려면 다음 명령을 실행하세요:"
            echo "/awesome:release-new $VERSION_TYPE --testpypi"
            echo ""
            echo "또는 실제 PyPI 배포를 진행하려면:"
            echo "/awesome:release-new $VERSION_TYPE"
        else
            echo "실제 릴리즈를 진행하려면 다음 명령을 실행하세요:"
            echo "/awesome:release-new $VERSION_TYPE"
        fi

        echo ""

        # 정리
        rm -f "$DRY_RUN_ACTIONS"
        exit 0
    fi
}
```

---

### Phase 0: 품질 검증 (자동, 필수)
1. 테스트 실행 및 커버리지 검증 (`pytest --cov`)
2. 린트 검사 (`ruff check`)
3. 타입 체크 (`mypy`)
4. 보안 스캔 (`bandit`, `pip-audit`)

**검증 실패 시**: 릴리즈 중단, 문제 해결 후 재시도

**🔬 Dry-Run 모드에서**: Phase 0은 **실제 실행**됩니다 (품질 검증은 항상 수행되어야 하므로)
- Dry-Run 모드에서도 테스트, 린트, 타입, 보안 검사를 모두 실행합니다
- 검증 실패 시 Dry-Run도 중단됩니다

### Phase 1: 버전 분석 및 검증
1. 현재 프로젝트 버전 확인 (pyproject.toml, __init__.py)
2. 목표 버전 결정 (인수 또는 자동 증가)
3. Git 상태 확인 (커밋 가능 여부)
4. 변경사항 요약

### Phase 1.5: 사용자 확인
- **릴리즈 계획 보고서** 생성 및 승인 대기
- 사용자 응답: "진행" / "수정 [내용]" / "중단"

### Phase 2: GitFlow PR 병합 (develop → main)
**📋 워크플로우:**
1. develop 브랜치 확인 (release는 develop에서 시작)
2. main 브랜치 최신화 (git fetch origin main:main)
3. GitHub PR 생성 (develop → main, Draft 상태)
4. PR을 Ready for Review로 전환
5. **CodeRabbit AI 자동 리뷰 완료** (품질 80% 이상 자동 승인)
6. **GitHub에서 PR 병합** (로컬이 아닌 GitHub 웹에서 병합)

⚠️ **중요**: PR 병합은 **GitHub 웹 인터페이스에서만** 수행합니다. 로컬에서 직접 push를 하면 안 됩니다.

### Phase 3: GitHub Actions 자동 릴리즈 (CI/CD 자동화)
**⚠️ CRITICAL**: Phase 3은 이제 모두 **GitHub Actions**에서 자동으로 처리됩니다!

PR이 main에 병합되면, GitHub Actions 워크플로우가 자동으로:
1. ✅ 버전 파일 업데이트 (pyproject.toml)
2. ✅ Git 커밋 및 Annotated Tag 생성
3. ✅ PyPI 배포 (uv publish)
4. ✅ GitHub Release 생성 및 공개

**로컬에서 할 작업은 없습니다!** GitHub Actions가 모든 것을 처리합니다.

---

## 🧪 Phase 0: 품질 검증 (필수 통과)

릴리즈 전 모든 품질 기준을 자동으로 검증합니다. **하나라도 실패 시 릴리즈 중단**됩니다.

### 🤖 CodeRabbit AI 자동 리뷰 통합

**MoAI-ADK의 모든 PR은 이미 CodeRabbit으로 자동 리뷰됨:**

```
PR 생성 (feature branch)
    ↓
CodeRabbit 자동 리뷰 (1-2분)
├─ 코드 품질 분석
├─ 보안 이슈 검출
├─ 테스트 커버리지 확인
└─ 자동 승인 (Pro - 80% 이상 품질)
    ↓
개발자가 PR 병합 (이미 승인됨)
    ↓
develop → main 병합 (GitFlow)
```

> **📋 CodeRabbit 설정**: `.coderabbit.yaml` 및 `.github/CODERABBIT_SETUP.md` 참고
> - 자동 리뷰 활성화 (모든 브랜치)
> - Agentic Chat 상호작용 가능
> - 자동 승인 (Pro 기능, 80% 임계값)
> - Auto-fix 제안 (한 클릭 적용)

### Phase 0.0: CodeRabbit 리뷰 결과 확인

**이미 수행된 AI 리뷰 검증:**

```bash
# develop 브랜치의 최근 PR 확인
gh pr list --base develop --state merged --json title,author,createdAt --limit 5

# 또는 마지막 PR의 CodeRabbit 코멘트 확인
gh pr view --json comments --template '{{range .comments}}{{if .author.login | contains "coderabbit"}}{{.body}}{{end}}{{end}}'
```

**CodeRabbit이 검증한 항목:**
- ✅ 코드 품질 (디자인 패턴, 가독성, 유지보수성)
- ✅ 보안 (OWASP Top 10, 취약점 검출)
- ✅ 테스트 (커버리지, 엣지 케이스)
- ✅ 문서화 (Docstring, 주석 품질)
- ✅ 성능 (알고리즘 최적화, 복잡도)

> **Skill 통합**: 자세한 검증 절차는 `Skill("moai-awesome-release-verify")`를 참고하세요.
> - Python 환경 확인 (3.13+)
> - pytest, ruff, mypy, bandit, pip-audit 검증
> - Git 상태 확인

### 0.1 Python 환경 확인

**Python 인터프리터 확인**:
```bash
# Python 버전 확인 (>=3.13 필요)
python_version=$(python --version 2>&1 | awk '{print $2}')
echo "🐍 Python 버전: $python_version"

# pyproject.toml의 requires-python 확인
required_python=$(rg "^requires-python = " pyproject.toml | awk -F'"' '{print $2}')
echo "📋 요구 Python 버전: $required_python"
```

**패키지 매니저 감지**:
```bash
# uv 우선, 없으면 pip 사용
if command -v uv &>/dev/null; then
    PKG_MANAGER="uv"
    echo "📦 패키지 매니저: uv (권장)"
else
    PKG_MANAGER="pip"
    echo "📦 패키지 매니저: pip (기본)"
fi
```

### 0.2 테스트 실행 (필수)

**pytest 실행 및 커버리지 검증**:
```bash
echo "🧪 테스트 실행 중..."

# pytest 실행 (커버리지 포함)
pytest tests/ --cov --cov-report=term-missing

if [ $? -ne 0 ]; then
    echo "❌ 테스트 실패: 모든 테스트가 통과해야 합니다"
    echo "→ pytest tests/ 실행하여 문제를 해결하세요"
    exit 1
fi

# 커버리지 확인 (85% 이상)
coverage=$(pytest tests/ --cov --cov-report=term-missing | grep "TOTAL" | awk '{print $4}' | sed 's/%//')

if [ "$coverage" -lt 85 ]; then
    echo "⚠️ 테스트 커버리지 부족: ${coverage}% (목표: 85%)"
    echo "→ 추가 테스트 작성을 권장하지만 계속 진행합니다"
else
    echo "✅ 테스트 통과 (커버리지: ${coverage}%)"
fi
```

**검증 항목**:
- ✅ 모든 테스트 케이스 통과
- ✅ 커버리지 ≥85% (권장, 경고만)
- ❌ 테스트 실패 시 중단

### 0.3 린트 검사 (필수)

**ruff 린트 실행**:
```bash
echo "🔍 린트 검사 중..."

# ruff check 실행
ruff check src/ tests/

if [ $? -ne 0 ]; then
    echo "❌ 린트 오류: 코드 스타일 위반이 있습니다"
    echo "→ ruff check --fix src/ tests/로 자동 수정 시도"
    exit 1
fi

echo "✅ 린트 통과"
```

**검증 항목**:
- ✅ 린트 규칙 위반 없음
- ✅ 코드 스타일 일관성
- ❌ 린트 오류 시 중단

### 0.4 타입 체크 (권장)

**mypy 타입 체크**:
```bash
echo "🔤 타입 체크 중..."

# mypy 실행 (missing imports 무시)
mypy src/moai_adk --ignore-missing-imports

if [ $? -ne 0 ]; then
    echo "⚠️ 타입 오류: TypeScript와 달리 Python은 경고만 표시"
    echo "→ mypy src/moai_adk 실행하여 확인"
    # 중단하지 않음
else
    echo "✅ 타입 체크 통과"
fi
```

**검증 항목**:
- ✅ 타입 오류 없음
- ⚠️ 타입 오류는 경고만 (중단하지 않음)

### 0.5 보안 스캔 (권장)

**보안 스캔 스크립트 실행**:
```bash
echo "🔒 보안 스캔 중..."

# 보안 스캔 스크립트 실행
python scripts/security-scan.py

if [ $? -ne 0 ]; then
    echo "⚠️ 보안 취약점 발견: 검토 권장"
    echo "→ scripts/security-scan.py 결과 확인"
    # 중단하지 않음
else
    echo "✅ 보안 스캔 통과"
fi
```

**검증 항목**:
- ✅ pip-audit: 의존성 취약점 없음
- ✅ bandit: 코드 보안 이슈 없음
- ⚠️ 취약점 발견 시 경고만 (중단하지 않음)

### 0.5.5 의존성 업데이트 확인 (권장)

**주요 의존성 업데이트 체크**:
```bash
echo "📦 의존성 업데이트 확인 중..."

# uv pip list --outdated 로 업데이트 가능한 패키지 확인
OUTDATED=$(uv pip list --outdated --format=json 2>/dev/null)

if [ -z "$OUTDATED" ] || [ "$OUTDATED" = "[]" ]; then
    echo "✅ 모든 의존성이 최신 버전입니다"
else
    # 주요 버전 업그레이드 (breaking changes 가능) 감지
    echo "$OUTDATED" | python -c "
import json, sys
try:
    outdated = json.load(sys.stdin)

    # 주요 버전 업그레이드 감지
    major_updates = []
    minor_updates = []

    for pkg in outdated:
        current = pkg['version'].split('.')
        latest = pkg['latest_version'].split('.')

        # Major version 비교
        if int(current[0]) < int(latest[0]):
            major_updates.append((pkg['name'], pkg['version'], pkg['latest_version']))
        elif len(current) > 1 and len(latest) > 1 and int(current[1]) < int(latest[1]):
            minor_updates.append((pkg['name'], pkg['version'], pkg['latest_version']))

    if major_updates:
        print('⚠️  주요 버전 업그레이드 가능 (breaking changes 주의):')
        for name, old, new in major_updates:
            print(f'   - {name}: {old} → {new}')
        print('   → 릴리즈 전에 호환성 검증 권장')
        print()

    if minor_updates:
        print('ℹ️  부분 버전 업그레이드 가능:')
        for name, old, new in minor_updates[:5]:  # 최대 5개만 표시
            print(f'   - {name}: {old} → {new}')
        if len(minor_updates) > 5:
            print(f'   ... 그 외 {len(minor_updates) - 5}개')

    if not major_updates and not minor_updates:
        print('✅ 모든 의존성이 최신 버전입니다')
except:
    print('✅ 의존성 버전 확인 완료')
" || echo "✅ 의존성 버전 확인 완료"
fi
```

**주의사항**:
- ⚠️ Major version 업그레이드 발견 시: 릴리즈 전 호환성 테스트 권장
- ℹ️ Minor version 업그레이드는 일반적으로 안전함
- 보안 패치는 가능한 빨리 적용 권장

### 0.6 품질 검증 요약

**성공 시 요약**:
```markdown
✅ 품질 검증 완료

- 🐍 Python: 3.13.x
- 📦 패키지 매니저: uv
- ✅ 테스트: 통과 (커버리지 87%)
- ✅ 린트: 통과 (ruff)
- ✅ 타입: 통과 (mypy)
- ✅ 보안: 통과 (bandit + pip-audit)

→ Phase 1으로 진행합니다...
```

**실패 시 중단**:
```markdown
❌ 품질 검증 실패

릴리즈를 진행할 수 없습니다. 다음 문제를 해결하세요:

- ❌ 테스트: 3개 실패
  → pytest tests/ 실행 결과 확인
  → tests/auth.test.py:45 - AssertionError

- ❌ 린트: 12개 오류
  → ruff check --fix src/ tests/로 자동 수정
  → src/utils.py:23 - Unused variable 'foo'

문제 해결 후 다시 시도하세요:
/awesome:release-new {VERSION_TYPE}
```

---

## 🔍 Phase 1: 버전 분석 및 검증

### 1.1 프로젝트 정보 수집

**버전 정보 읽기 (SSOT 방식)**:
```bash
# pyproject.toml에서 버전 읽기 (SSOT - 유일한 진실의 출처)
current_version=$(rg "^version = " pyproject.toml | awk -F'"' '{print $2}')
echo "📌 현재 버전 (pyproject.toml): $current_version"

# 설치된 패키지 버전 확인 (검증용)
installed_version=$(python -c "from importlib.metadata import version; print(version('moai-adk'))" 2>/dev/null || echo "N/A")
echo "📦 설치된 버전: $installed_version"

# 버전 일치 여부 확인
if [ "$current_version" != "$installed_version" ] && [ "$installed_version" != "N/A" ]; then
    echo "⚠️ 경고: pyproject.toml과 설치된 버전이 다릅니다"
    echo "→ pyproject.toml: $current_version"
    echo "→ 설치된 버전: $installed_version"
    echo "→ 해결: uv pip install -e . --force-reinstall --no-deps"
fi

# __init__.py는 자동 로드 (확인만)
echo "ℹ️ __init__.py는 importlib.metadata로 자동 로드 (수정 불필요)"
```

**Git 상태 확인**:
```bash
# Git 상태
git status --short
git log --oneline -5

# 브랜치 확인
current_branch=$(git branch --show-current)
echo "🌿 현재 브랜치: $current_branch"

# 미커밋 변경사항 확인
if [ -n "$(git status --short)" ]; then
    echo "⚠️ 미커밋 변경사항 있음 (자동 커밋 예정)"
fi
```

### 1.2 목표 버전 결정

**인수 파싱 로직**:
```bash
# $1 = version_type (patch, minor, major)
VERSION_TYPE="${1:-patch}"  # 기본값: patch

echo "🎯 버전 타입: $VERSION_TYPE"
```

**버전 증가 로직** (Python 스크립트):
```python
# semver_bump.py
import sys

def bump_version(current: str, bump_type: str) -> str:
    major, minor, patch = map(int, current.split("."))

    if bump_type == "patch":
        return f"{major}.{minor}.{patch + 1}"
    elif bump_type == "minor":
        return f"{major}.{minor + 1}.0"
    elif bump_type == "major":
        return f"{major + 1}.0.0"
    else:
        raise ValueError(f"Invalid bump type: {bump_type}")

if __name__ == "__main__":
    current = sys.argv[1]
    bump_type = sys.argv[2]
    print(bump_version(current, bump_type))
```

**버전 계산**:
```bash
# Python 스크립트 사용
new_version=$(python -c "
current = '$current_version'
bump_type = '$VERSION_TYPE'

major, minor, patch = map(int, current.split('.'))

if bump_type == 'patch':
    new = f'{major}.{minor}.{patch + 1}'
elif bump_type == 'minor':
    new = f'{major}.{minor + 1}.0'
elif bump_type == 'major':
    new = f'{major + 1}.0.0'
else:
    new = current

print(new)
")

echo "📊 버전 변경: $current_version → $new_version"
```

**🔬 Dry-Run 모드에서**: 버전 계산만 수행되고, 실제 파일 수정은 하지 않습니다.

### 1.3 변경사항 분석

**Git 로그 분석**:
```bash
# 마지막 릴리즈 태그 찾기
last_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

if [ -n "$last_tag" ]; then
    echo "🏷️ 마지막 릴리즈: $last_tag"
    # 마지막 릴리즈 이후 커밋 목록
    commits=$(git log $last_tag..HEAD --oneline --pretty=format:"- %s (%h)")
else
    echo "🏷️ 첫 릴리즈"
    # 전체 커밋 목록 (최근 20개)
    commits=$(git log --oneline --pretty=format:"- %s (%h)" | head -20)
fi

echo "📝 변경사항:"
echo "$commits"
```

**변경 타입 분류** (Git 커밋 메시지 기준):
```bash
# 이모지 기반 분류 (MoAI-ADK 커밋 메시지 표준)
added=$(echo "$commits" | grep -E "^- (✨|🎉|🚀)" || echo "")
fixed=$(echo "$commits" | grep -E "^- (🐛|🔥|🩹)" || echo "")
docs=$(echo "$commits" | grep -E "^- (📝|📚|📖)" || echo "")
refactor=$(echo "$commits" | grep -E "^- (♻️|🔨|🔧)" || echo "")
test=$(echo "$commits" | grep -E "^- (✅|🧪)" || echo "")
```

### 1.4 릴리즈 계획 보고서 생성

```markdown
## 🚀 릴리즈 계획 보고서: v{new_version}

### 🧪 품질 검증 결과 (Phase 0)
- ✅ 테스트: 통과 (커버리지 87%)
- ✅ 린트: 통과 (ruff)
- ✅ 타입: 통과 (mypy)
- ✅ 보안: 통과 (bandit + pip-audit)

### 📊 버전 정보
- **현재 버전**: v{current_version}
- **목표 버전**: v{new_version}
- **버전 타입**: {VERSION_TYPE}
- **릴리즈 날짜**: {YYYY-MM-DD}

### 📁 프로젝트 정보
- **프로젝트**: moai-adk
- **현재 브랜치**: {current_branch}
- **마지막 커밋**: {git log -1 --oneline}

### 📝 변경사항 요약
{마지막 릴리즈 이후 커밋 목록}

#### ✨ Added (N개)
{added 커밋}

#### 🐛 Fixed (N개)
{fixed 커밋}

#### 📝 Documentation (N개)
{docs 커밋}

### 🔄 업데이트할 파일 (SSOT)
- [ ] pyproject.toml: version = "{current_version}" → "{new_version}" (SSOT)
- [ ] src/moai_adk/__init__.py: 수정 불필요 (importlib.metadata로 자동 로드)

### 🚀 릴리즈 작업
- [ ] Git 커밋: `🔖 RELEASE: v{new_version}`
- [ ] Git 태그: `v{new_version}` (Annotated)
- [ ] PyPI 배포: `uv publish`
- [ ] GitHub Release: `gh release create` (Draft)

### ⚠️ 주의사항
- 현재 브랜치: {current_branch} (main 권장)
- 미커밋 변경: {N개 파일} (자동 커밋 예정)

---
**승인 요청**: 위 계획으로 릴리즈를 진행하시겠습니까?
("진행", "수정 [내용]", "중단" 중 선택)
```

**🔬 Dry-Run 모드에서**: 위의 릴리즈 계획 보고서가 출력되며, 사용자 승인을 기다리지 않습니다.
- 대신 "실제 릴리즈 명령"이 표시되어 사용자가 확인 후 실행할 수 있게 합니다.

---

## 🔄 Phase 2: Branch Merge and PR Management

**전제조건**: Phase 1에서 사용자가 "진행"을 선택한 경우만 실행

**🔬 Dry-Run 모드에서**: Phase 2는 **완전히 건너뜁니다**
- GitHub PR 생성 단계를 시뮬레이션하고 로깅만 합니다
- 실제 GitHub API 호출을 하지 않습니다
- 대신 "다음 단계: GitHub에서 PR 병합" 메시지를 표시합니다

**워크플로우 자동 감지**:
- ✅ **GitFlow 모드**: develop 브랜치 존재 시 (feature → develop → main)
- ✅ **Simplified 모드**: develop 브랜치 없을 시 (feature → main)

**핵심 원칙**:
- ✅ **PR은 GitHub CLI(`gh pr create`)로 생성**
- ✅ **CodeRabbit AI가 자동으로 리뷰 및 승인**
- ✅ **PR 병합은 GitHub 웹에서만 수행 (로컬 merge 금지)**
- ✅ **PR 병합 후 GitHub Actions가 자동으로 Phase 3 실행**

**모드 감지**:
- **Personal 모드**: PR 단계 자동 건너뜀 (로컬 개발용)
- **Team 모드**: full GitFlow PR 프로세스 실행

### Step 2.0: 프로젝트 모드 및 워크플로우 감지 (자동)

**Mode 및 Workflow 자동 감지**:
```bash
# 1. 프로젝트 모드 감지 (.moai/config.json)
project_mode=$(rg '"mode":\s*"([^"]+)"' .moai/config.json -o '$1' | head -1)
echo "🎭 프로젝트 모드: $project_mode"

if [ "$project_mode" = "personal" ]; then
    echo "ℹ️  Personal 모드: PR 단계를 자동으로 건너뜁니다 (로컬 개발용)"
    echo "→ Phase 3 (릴리즈 실행)로 직접 진행합니다..."
    # Phase 3으로 자동 점프 (PR 단계 건너뜀)
    return 0
fi

echo "🔀 Team 모드: Branch merge 프로세스 실행"
echo ""

# 2. GitFlow 워크플로우 감지 (develop 브랜치 존재 여부)
if git show-ref --verify --quiet refs/heads/develop; then
    WORKFLOW_MODE="gitflow"
    BASE_BRANCH="develop"
    TARGET_BRANCH="main"
    echo "📋 Workflow: GitFlow (develop → main)"
    echo "   - Feature branches merge to: develop"
    echo "   - Releases merge from: develop → main"
else
    WORKFLOW_MODE="simplified"
    BASE_BRANCH="main"
    TARGET_BRANCH="main"
    echo "📋 Workflow: Simplified (feature → main)"
    echo "   - Feature branches merge to: main"
    echo "   - Releases deploy from: main"
fi

echo ""
```

### Step 2.1: 현재 브랜치 확인 및 검증

**브랜치 검증** (Team 모드만, 워크플로우 감지):
```bash
current_branch=$(git branch --show-current)
echo "🌿 현재 브랜치: $current_branch"

if [ "$WORKFLOW_MODE" = "gitflow" ]; then
    # GitFlow: develop 브랜치에서 시작 권장
    if [ "$current_branch" != "$BASE_BRANCH" ]; then
        echo "⚠️  GitFlow 모드에서는 $BASE_BRANCH 브랜치에서 릴리즈를 시작하는 것을 권장합니다"
        echo "→ 현재 브랜치: $current_branch"
        read -p "계속 진행하시겠습니까? (y/n): " continue_anyway
        if [ "$continue_anyway" != "y" ]; then
            echo "→ git checkout $BASE_BRANCH 실행 후 재시도"
            exit 1
        fi
    else
        echo "✅ $BASE_BRANCH 브랜치 확인 완료"
    fi
else
    # Simplified: feature 브랜치에서 바로 main으로 PR
    echo "✅ Simplified 모드: $current_branch에서 $TARGET_BRANCH로 PR 생성"
fi
```

### Step 2.2: main 브랜치 최신화

**main 브랜치 동기화**:
```bash
echo "🔄 main 브랜치 최신화 중..."

# origin/main 최신 상태 확인
git fetch origin main:main

if [ $? -ne 0 ]; then
    echo "⚠️ main 브랜치 동기화 실패"
    echo "→ git fetch origin main:main 실행 확인"
fi

echo "✅ main 브랜치 최신화 완료"
```

### Step 2.3: GitHub PR 생성 (Draft)

**develop → main PR 생성**:
```bash
echo "📝 Creating GitHub PR (develop → main)..."

# PR title and description (English only)
pr_title="🔖 Release v{new_version} | {VERSION_TYPE} | {Release Description}"

pr_body="## GitFlow Release PR

### 📊 Release Information
- **Version**: v{new_version}
- **Type**: {VERSION_TYPE}
- **Date**: {YYYY-MM-DD}

### 📝 Changelog
{Commits since last release}

### 🧪 Quality Assurance
- ✅ Tests: {TEST_RESULT}
- ✅ Linting: {LINT_RESULT}
- ✅ Type Checking: {TYPE_RESULT}
- ✅ Security Scan: {SECURITY_RESULT}

### 🚀 Release Steps
- [x] PR created from develop
- [ ] CodeRabbit AI review (automatic)
- [ ] Merge to main (manual)
- [ ] Tag creation (automatic)
- [ ] PyPI deployment (automatic)
- [ ] GitHub Release publication (automatic)

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Alfred <alfred@mo.ai.kr>"

# Create PR with gh CLI (Draft)
echo "⏳ Creating GitHub PR..."

# NOTE: --label release은 GitHub Actions moai-release-pipeline.yml에서 감지하여
#       자동으로 Tag 생성 및 GitHub Release 생성을 트리거합니다
#       이는 "🔖 RELEASE:" 커밋 패턴보다 더 신뢰할 수 있습니다 (실패율 <5%)

gh pr create \
  --base main \
  --head develop \
  --title "$pr_title" \
  --body "$pr_body" \
  --label release \
  --draft 2>&1

pr_exit_code=$?

if [ $pr_exit_code -ne 0 ]; then
    echo ""
    echo "⚠️  GitHub PR 생성 실패"
    echo ""
    echo "가능한 원인:"
    echo "1️⃣  인증 오류: gh auth status 확인"
    echo "2️⃣  중복 PR: 이미 존재하는 PR 확인"
    echo "3️⃣  커밋 차이 없음: develop과 main이 같은 상태"
    echo "4️⃣  네트워크 오류: 인터넷 연결 확인"
    echo ""
    echo "다음 중 선택:"
    echo "1. 수동으로 GitHub에서 PR 생성 후 진행"
    echo "2. PR 없이 직접 릴리즈 진행"
    read -p "선택 (1 또는 2): " pr_fallback

    if [ "$pr_fallback" = "1" ]; then
        echo ""
        read -p "PR 생성이 완료되었나요? (yes/no): " pr_complete
        if [ "$pr_complete" != "yes" ]; then
            echo "❌ PR 생성 필요. 먼저 GitHub에서 PR을 생성하세요."
            exit 1
        fi
        pr_number=$(gh pr list --head develop --base main --json number -q '.[0].number' 2>/dev/null || echo "")
        if [ -z "$pr_number" ]; then
            echo "⚠️  PR을 찾을 수 없습니다. 계속 진행하겠습니다..."
            pr_number="unknown"
        fi
    else
        echo "→ PR 없이 릴리즈를 진행합니다 (로컬 모드)"
        pr_number="none"
    fi
else
    echo "✅ GitHub PR이 생성되었습니다"
    # PR 번호 추출
    pr_number=$(gh pr list --head develop --base main --json number -q '.[0].number' 2>/dev/null || echo "")
    if [ -z "$pr_number" ]; then
        pr_number="latest"
    fi
    echo "✅ PR #$pr_number 생성됨 (Draft 상태)"
    echo "→ PR 링크: https://github.com/modu-ai/moai-adk/pull/$pr_number"
fi
```

### Step 2.3.5: 🤖 CodeRabbit 자동 리뷰 & 승인 (자동 실행)

**PR 생성 후 CodeRabbit 자동 실행:**

```bash
echo "🤖 CodeRabbit AI 자동 리뷰 시작 (1-2분 소요)..."

# CodeRabbit이 자동으로:
# 1. 코드 품질 분석
# 2. 보안 이슈 검출
# 3. 테스트 커버리지 확인
# 4. 품질 80% 이상 시 자동 승인 (Pro 기능)

# PR 리뷰 대기
for i in {1..12}; do
    sleep 10

    review_status=$(gh pr view $pr_number --json reviews --jq '.reviews | length')
    comments=$(gh pr view $pr_number --json comments --jq '.comments[] | select(.author.login == "coderabbitai") | .body' 2>/dev/null | head -1)

    if [ -n "$comments" ]; then
        echo "✅ CodeRabbit 리뷰 완료!"
        echo "→ PR #$pr_number: https://github.com/modu-ai/moai-adk/pull/$pr_number"
        break
    fi

    echo "⏳ CodeRabbit 리뷰 중... (${i}/12)"
done

# CodeRabbit 자동 승인 확인
approval_status=$(gh pr view $pr_number --json reviewDecision --jq '.reviewDecision')

if [ "$approval_status" = "APPROVED" ]; then
    echo "✅ CodeRabbit이 자동 승인했습니다! (품질 80% 이상)"
    echo "→ PR이 병합 가능 상태입니다"
else
    echo "ℹ️  CodeRabbit 리뷰 완료 (추가 수정 제안 있음)"
    echo "→ PR 코멘트 확인 후 수정 진행"
fi
```

**CodeRabbit 자동 승인 조건:**
- ✅ 코드 품질 점수: 80% 이상
- ✅ 보안 이슈: 중대 문제 없음
- ✅ 테스트 커버리지: 85% 이상 (권장)
- ✅ Agentic Chat: 개발자가 추가 질문 가능

### Step 2.4: PR Ready for Review로 전환

**Draft → Ready 상태 변경** (PR이 존재할 때만):
```bash
if [ "$pr_number" = "none" ] || [ "$pr_number" = "unknown" ]; then
    echo "ℹ️  PR이 없으므로 Ready 전환 단계를 건너뜁니다"
else
    echo "📋 PR을 Ready for Review로 전환 중..."

    gh pr ready $pr_number 2>/dev/null

    if [ $? -ne 0 ]; then
        echo "⚠️ PR 상태 변경 실패 (이미 Ready 상태일 수 있음)"
        echo "→ GitHub에서 확인: https://github.com/modu-ai/moai-adk/pulls"
    else
        echo "✅ PR이 Ready for Review 상태로 변경되었습니다"
        echo "🤖 CodeRabbit 자동 리뷰가 이미 완료되었습니다"
    fi
fi
```

### Step 2.5: 자동 병합 또는 사용자 승인

**병합 방식 선택**:
```bash
echo "🤔 PR 병합 방식을 선택하세요:"
echo "1. 자동 병합 (gh pr merge로 즉시 병합)"
echo "2. 수동 승인 (GitHub에서 리뷰 후 병합)"
read -p "선택 (1 또는 2): " merge_choice

if [ "$merge_choice" = "1" ]; then
    echo "🔄 자동 병합 중..."

    gh pr merge $pr_number \
      --merge \
      --auto

    if [ $? -eq 0 ]; then
        echo "✅ PR이 자동 병합 대기 상태로 설정되었습니다"
    else
        echo "❌ 자동 병합 설정 실패"
        echo "→ GitHub에서 수동으로 병합하세요"
        exit 1
    fi
else
    echo "⏳ GitHub에서 리뷰 후 병합해주세요"
    echo "→ PR 링크: https://github.com/modu-ai/moai-adk/pull/$pr_number"
    echo "→ 병합 완료 후 다시 릴리즈 명령 실행"
    exit 0
fi
```

### Step 2.6: 병합 완료 확인

**main 브랜치 업데이트**:
```bash
echo "⏳ PR 병합 완료 대기 중..."

# 최대 30초 동안 병합 상태 확인
for i in {1..6}; do
    sleep 5
    merge_status=$(gh pr view $pr_number --json mergeStateStatus -q '.mergeStateStatus')

    if [ "$merge_status" = "MERGED" ]; then
        echo "✅ PR이 성공적으로 병합되었습니다"

        # main 브랜치 로컬 업데이트
        git fetch origin
        git checkout main
        git pull origin main

        echo "✅ main 브랜치가 최신화되었습니다"
        return 0
    fi
done

echo "⚠️ PR 병합 확인 시간 초과"
echo "→ GitHub에서 병합 상태 확인 후 수동으로 계속하세요"
```

### Step 2.7: SPEC 기반 체인지로그 생성 (Feature-Accumulation 지원)

**목적**: 2-4주 개발 주기에서 누적된 기능들을 사용자 친화적으로 요약하기

**체인지로그 생성 스크립트**:
```bash
echo "📝 Phase 2.7: SPEC-기반 체인지로그 생성 중..."

# 변수 설정
LAST_MINOR_TAG=$(git describe --tags --abbrev=0 --match "v*.*.0" 2>/dev/null || echo "")
CURRENT_VERSION="$new_version"

# 빌드 데이터 저장소
CHANGELOG_DATA=".moai/temp/changelog-$CURRENT_VERSION.md"
mkdir -p .moai/temp

echo "## 🎉 What's New in v$CURRENT_VERSION" > "$CHANGELOG_DATA"
echo "" >> "$CHANGELOG_DATA"

if [ -z "$LAST_MINOR_TAG" ]; then
    echo "⚠️ 이전 minor 버전을 찾을 수 없습니다. 전체 기능 목록 생성..."
    LAST_MINOR_TAG=$(git rev-list --max-parents=0 HEAD)
else
    echo "📊 마지막 릴리즈: $LAST_MINOR_TAG"
    echo "현재 버전: v$CURRENT_VERSION"
    echo ""
fi

# 커밋에서 SPEC ID 추출 (예: @SPEC:AUTH-001)
echo "🔍 SPEC 문서 검색 중..."

SPEC_IDS=$(git log $LAST_MINOR_TAG..HEAD --oneline 2>/dev/null | \
           grep -o '@SPEC:[A-Z_][A-Z_0-9]*-[0-9]\{3\}' | \
           sed 's/@SPEC:/SPEC-/' | \
           sort -u)

SPEC_COUNT=$(echo "$SPEC_IDS" | grep -c "SPEC" || echo 0)
echo "✅ 발견된 SPEC: $SPEC_COUNT개"
echo ""

# 카테고리별 기능 수집
declare -A FEATURES_BY_CATEGORY

for SPEC_ID in $SPEC_IDS; do
    SPEC_DIR=".moai/specs/$SPEC_ID"

    if [ -f "$SPEC_DIR/spec.md" ]; then
        # SPEC ID에서 카테고리 추출 (예: SPEC-UPDATE-001 → UPDATE)
        CATEGORY=$(echo "$SPEC_ID" | sed 's/^SPEC-//' | sed 's/-[0-9]*$//')

        # SPEC 제목 추출
        TITLE=$(grep '^# @SPEC:' "$SPEC_DIR/spec.md" 2>/dev/null | sed 's/^# @SPEC:[^ ]* //' || echo "N/A")

        # 짧은 설명 추출 (첫 번째 단락)
        SUMMARY=$(grep -A 2 '## Overview\|## 개요' "$SPEC_DIR/spec.md" 2>/dev/null | tail -1 | head -c 100)

        # Acceptance criteria 개수 세기
        CRITERIA_COUNT=$(grep -c '✅\|- WHEN\|- GIVEN\|### ' "$SPEC_DIR/acceptance.md" 2>/dev/null || echo "0")

        # 카테고리별 저장
        if [ -z "${FEATURES_BY_CATEGORY[$CATEGORY]}" ]; then
            FEATURES_BY_CATEGORY[$CATEGORY]="- **[$SPEC_ID]** $TITLE\n"
        else
            FEATURES_BY_CATEGORY[$CATEGORY]+="- **[$SPEC_ID]** $TITLE\n"
        fi
    fi
done

# 카테고리별로 체인지로그 작성
for CATEGORY in $(echo "${!FEATURES_BY_CATEGORY[@]}" | tr ' ' '\n' | sort); do
    echo "### 🔹 $CATEGORY" >> "$CHANGELOG_DATA"
    echo -e "${FEATURES_BY_CATEGORY[$CATEGORY]}" >> "$CHANGELOG_DATA"
    echo "" >> "$CHANGELOG_DATA"
done

# 버그 수정 요약
echo "### 🐛 Bug Fixes" >> "$CHANGELOG_DATA"
BUG_FIX_COUNT=$(git log $LAST_MINOR_TAG..HEAD --oneline 2>/dev/null | \
                grep -c 'fix\|Fix\|FIX' || echo "0")
echo "- Fixed $BUG_FIX_COUNT bugs and improvements" >> "$CHANGELOG_DATA"
echo "" >> "$CHANGELOG_DATA"

# 성능 개선
echo "### ⚡ Performance" >> "$CHANGELOG_DATA"
PERF_COUNT=$(git log $LAST_MINOR_TAG..HEAD --oneline 2>/dev/null | \
             grep -c 'perf\|optimi\|Optimi' || echo "0")
echo "- $PERF_COUNT performance optimizations" >> "$CHANGELOG_DATA"
echo "" >> "$CHANGELOG_DATA"

# 테스트 커버리지
echo "### 🧪 Quality Metrics" >> "$CHANGELOG_DATA"
if [ -f "pyproject.toml" ]; then
    COVERAGE=$(grep -i 'coverage\|test' pyproject.toml | head -1 || echo "✅ Comprehensive test coverage")
    echo "- $COVERAGE" >> "$CHANGELOG_DATA"
fi
echo "" >> "$CHANGELOG_DATA"

# 모니터링 정보 출력
echo "📄 생성된 체인지로그:"
cat "$CHANGELOG_DATA"
echo ""

# GitHub Release에서 사용하기 위해 환경변수 저장
echo "SPEC_CHANGELOG=$CHANGELOG_DATA" >> "$GITHUB_OUTPUT" 2>/dev/null || true

echo "✅ 체인지로그 생성 완료"
echo "→ 파일: $CHANGELOG_DATA"
```

**주의사항**:
- ✅ SPEC 문서가 있어야 기능 정보 추출 가능
- ✅ `.moai/specs/SPEC-*/` 디렉토리 구조 필요
- ✅ 생성된 파일은 GitHub Release 생성 시 자동으로 포함됨

---

## 🚀 Phase 3: GitHub Actions 자동 릴리즈 실행

**⚠️ 주의**: Phase 3은 부분적으로 **자동화**되어 있습니다. 몇 가지 수동 단계가 필요할 수 있습니다.

### 🤖 GitHub Actions 자동 워크플로우

PR이 main 브랜치에 병합되면, GitHub Actions의 여러 워크플로우가 실행됩니다:

**자동 실행 워크플로우**:
1. ✅ **moai-gitflow.yml** (PR 병합 시 자동 트리거)
   - Release commit 감지 (🔖 RELEASE: 패턴)
   - Git Tag 생성
   - GitHub Release 생성 (Draft)

2. ✅ **release.yml** (GitHub Release published 시 자동 트리거)
   - 패키지 빌드 (uv build)
   - PyPI 배포 (uv publish with PYPI_API_TOKEN)

**⚠️ 주의**: Release Pipeline이 merge commit을 감지하지 못할 수 있습니다.
- GitHub merge commit 형식: "Merge pull request #XX..."
- Release Pattern: "🔖 RELEASE: v..." (PR 설명에 있어야 감지됨)

### 모니터링 방법

**Step 1: Release Pipeline 실행 확인** (merge 후 1-2초)
```bash
# GitHub Actions 실행 상태 확인
gh run list --branch main --limit 5 --json name,status,conclusion

# 최신 Release Pipeline 상세 정보
gh run view $(gh run list --branch main --limit 1 --json databaseId -q '.[0].databaseId') --json jobs
```

**Step 2: Release 생성 확인** (5-10초)
```bash
# GitHub Release Draft 확인
gh release list --limit 3

# 특정 버전 Release 상세 정보
gh release view v{new_version}
```

**Step 3: PyPI 배포 확인** (30-60초)
```bash
# PyPI API로 패키지 버전 확인
curl -s https://pypi.org/pypi/moai-adk/json | python3 -c "import sys, json; data=json.load(sys.stdin); print('Latest:', data['info']['version'])"

# 또는 PyPI 페이지 직접 확인
open https://pypi.org/project/moai-adk/
```

### ✨ 완료 확인

릴리즈가 성공적으로 완료되면 다음 모두 확인:

- ✅ GitHub Release 페이지: `https://github.com/modu-ai/moai-adk/releases/tag/v{new_version}`
- ✅ PyPI 패키지: `https://pypi.org/project/moai-adk/{new_version}`
- ✅ Git 태그: `git tag -l "v{new_version}"`
- ✅ GitHub Actions: moai-gitflow.yml + release.yml 모두 success

**설치 테스트**:
```bash
# uv로 설치 테스트
uv tool install moai-adk=={new_version}
moai-adk --version
```

---

## 🚀 Step 3.1 (참고): 버전 파일 업데이트 - GitHub Actions가 자동 수행

**⚠️ 중요**: MoAI-ADK는 **SSOT (Single Source of Truth)** 버전 관리를 사용합니다.

**버전 관리 구조**:
```python
# pyproject.toml (SSOT - 유일한 진실의 출처)
version = "0.4.0"

# src/moai_adk/__init__.py (동적 로드)
from importlib.metadata import version, PackageNotFoundError

try:
    __version__ = version("moai-adk")  # pyproject.toml에서 자동 로드
except PackageNotFoundError:
    __version__ = "0.4.0-dev"
```

**업데이트 방법**:

**1. pyproject.toml만 업데이트** (Edit 도구 사용):
```bash
# OLD: version = "0.3.0"
# NEW: version = "0.3.1"
```

**2. __init__.py는 수정하지 않음** (자동으로 새 버전 읽기)
```bash
# ❌ 수정 금지: __init__.py는 importlib.metadata로 자동 로드
# ✅ pyproject.toml만 수정하면 자동으로 반영됨
```

**3. editable install 재설치** (버전 메타데이터 업데이트):
```bash
uv pip install -e . --force-reinstall --no-deps
```

### Step 3.1.5: 패키지 템플릿 동기화 (Post-Release Sync)

**커밋 전에 템플릿 파일 업데이트** (`Step 3.1 직후, Step 3.2 커밋 전`):

```bash
echo "🔄 패키지 템플릿 동기화 중..."
echo ""

# 프로젝트 정보 읽기 (.moai/config.json)
PROJECT_NAME=$(rg '"name":\s*"([^"]+)"' .moai/config.json -o '$1' | head -1)
PROJECT_OWNER=$(rg '"nickname":\s*"([^"]+)"' .moai/config.json -o '$1' | head -1)
PROJECT_LOCALE=$(rg '"locale":\s*"([^"]+)"' .moai/config.json -o '$1' | head -1)
PROJECT_LANGUAGE=$(rg '"language":\s*"([^"]+)"' .moai/config.json -o '$1' | head -1)

echo "📌 프로젝트 정보:"
echo "  - 이름: $PROJECT_NAME"
echo "  - Owner: $PROJECT_OWNER"
echo "  - Locale: $PROJECT_LOCALE"
echo "  - Language: $PROJECT_LANGUAGE"
echo ""

# 템플릿 경로
TEMPLATE_DIR="src/moai_adk/templates"
TEMPLATE_CLAUDE="$TEMPLATE_DIR/.claude"
TEMPLATE_MOAI="$TEMPLATE_DIR/.moai"

# 1️⃣ .claude/ 동기화
echo "1️⃣  .claude/ 디렉토리 동기화 중..."

if [ ! -d "$TEMPLATE_CLAUDE" ]; then
    echo "⚠️ 템플릿 .claude/ 디렉토리를 찾을 수 없습니다"
    echo "→ 경로: $TEMPLATE_CLAUDE"
else
    # .claude/ 디렉토리 전체 복사 (덮어쓰기)
    cp -rv "$TEMPLATE_CLAUDE/" ".claude/" 2>&1 | grep -E "^'[^']+' -> " | wc -l | xargs echo "  ✅ 파일 동기화:"

    # 백업 생성
    echo "  📦 백업 생성 중..."
    mkdir -p .moai-backups/pre-sync/
    cp -r .claude/.claude-backup .moai-backups/pre-sync/claude-$(date +%Y%m%d-%H%M%S) 2>/dev/null || true
fi

echo ""

# 2️⃣ .moai/ 동기화 (config, project, memory 만)
echo "2️⃣  .moai/ 디렉토리 동기화 중 (config, project, memory)..."

if [ ! -d "$TEMPLATE_MOAI" ]; then
    echo "⚠️ 템플릿 .moai/ 디렉토리를 찾을 수 없습니다"
    echo "→ 경로: $TEMPLATE_MOAI"
else
    # 선택적 디렉토리만 동기화 (specs 제외)
    for subdir in config project memory; do
        if [ -d "$TEMPLATE_MOAI/$subdir" ]; then
            cp -rv "$TEMPLATE_MOAI/$subdir/" ".moai/$subdir/" 2>&1 | grep -E "^'[^']+' -> " | wc -l | xargs echo "  ✅ .moai/$subdir - 파일 동기화:"
        fi
    done

    # 백업 생성
    echo "  📦 백업 생성 중..."
    mkdir -p .moai-backups/pre-sync/
    cp -r .moai/.moai-backup .moai-backups/pre-sync/moai-$(date +%Y%m%d-%H%M%S) 2>/dev/null || true
fi

echo ""

# 3️⃣ CLAUDE.md 템플릿 변수 치환
echo "3️⃣  CLAUDE.md 변수 치환 중..."

TEMPLATE_CLAUDE_MD="$TEMPLATE_DIR/CLAUDE.md"

if [ -f "$TEMPLATE_CLAUDE_MD" ]; then
    # CLAUDE.md 복사
    cp "$TEMPLATE_CLAUDE_MD" CLAUDE.md

    # .moai/config.json에서 추가 정보 추출
    CONVERSATION_LANGUAGE=$(rg '"conversation_language":\s*"([^"]+)"' .moai/config.json -o '$1' | head -1)
    CONVERSATION_LANGUAGE_NAME=$(rg '"conversation_language_name":\s*"([^"]+)"' .moai/config.json -o '$1' | head -1)
    CODEBASE_LANGUAGE=$(rg '"language":\s*"([^"]+)"' .moai/config.json -o '$1' | head -1)

    # 변수 치환 (sed 사용, 플레이스홀더 기반)
    # {{project_name}} → $PROJECT_NAME
    sed -i '' "s|{{project_name}}|$PROJECT_NAME|g" CLAUDE.md
    sed -i '' "s|{{project_owner}}|$PROJECT_OWNER|g" CLAUDE.md
    sed -i '' "s|{{locale}}|$PROJECT_LOCALE|g" CLAUDE.md
    sed -i '' "s|{{conversation_language}}|$CONVERSATION_LANGUAGE|g" CLAUDE.md
    sed -i '' "s|{{conversation_language_name}}|$CONVERSATION_LANGUAGE_NAME|g" CLAUDE.md
    sed -i '' "s|{{codebase_language}}|$CODEBASE_LANGUAGE|g" CLAUDE.md

    # 대문자 플레이스홀더도 처리
    sed -i '' "s|{{PROJECT_NAME}}|$PROJECT_NAME|g" CLAUDE.md
    sed -i '' "s|{{PROJECT_OWNER}}|$PROJECT_OWNER|g" CLAUDE.md
    sed -i '' "s|{{PROJECT_LOCALE}}|$PROJECT_LOCALE|g" CLAUDE.md

    echo "  ✅ CLAUDE.md 생성 완료"
    echo "     - project_name: $PROJECT_NAME"
    echo "     - project_owner: $PROJECT_OWNER"
    echo "     - locale: $PROJECT_LOCALE"
    echo "     - conversation_language: $CONVERSATION_LANGUAGE"
    echo "     - conversation_language_name: $CONVERSATION_LANGUAGE_NAME"
    echo "     - codebase_language: $CODEBASE_LANGUAGE"
else
    echo "⚠️ 템플릿 CLAUDE.md를 찾을 수 없습니다"
fi

echo ""
echo "📢 동기화 완료!"
echo "→ 다음 단계: Git 커밋 (버전 + 템플릿 파일 포함)"
```

**동기화 검증 체크리스트**:
- ✅ `.claude/` 디렉토리 업데이트됨 (agents, commands, hooks, output-styles, skills, settings.json)
- ✅ `.moai/project/` 업데이트됨 (product.md, structure.md, tech.md)
- ✅ `.moai/memory/` 업데이트됨 (개발 가이드, SPEC 메타데이터)
- ✅ `CLAUDE.md` 생성되고 변수 치환됨 (프로젝트명, Owner, 언어)
- ✅ 백업 생성됨 (`.moai-backups/pre-sync/`)

### Step 3.1.6: 템플릿 동기화 검증 (Integrity Check)

**목적**: 템플릿이 제대로 동기화되었는지 검증하고 불일치 항목 보고

**검증 스크립트**:
```bash
echo "🔍 Step 3.1.6: 템플릿 동기화 검증 중..."
echo ""

TEMPLATE_DIR="src/moai_adk/templates"
TEMPLATE_CLAUDE="$TEMPLATE_DIR/.claude"
TEMPLATE_MOAI="$TEMPLATE_DIR/.moai"

# 1️⃣ .claude/ 동기화 검증
echo "1️⃣  .claude/ 디렉토리 검증..."

if [ -d "$TEMPLATE_CLAUDE" ] && [ -d ".claude" ]; then
    # 디렉토리 간 diff 수행 (바이너리 제외)
    DIFF_COUNT=$(diff -r "$TEMPLATE_CLAUDE" ".claude" \
                --exclude='*.pyc' --exclude='__pycache__' \
                --exclude='.DS_Store' 2>/dev/null | wc -l)

    if [ "$DIFF_COUNT" -eq 0 ]; then
        echo "  ✅ .claude/ 완벽하게 동기화됨"
    else
        echo "  ⚠️  .claude/ 차이 감지됨 ($DIFF_COUNT 라인)"
        echo "     → 첫 5개 차이:"
        diff -r "$TEMPLATE_CLAUDE" ".claude" \
            --exclude='*.pyc' --exclude='__pycache__' \
            --exclude='.DS_Store' 2>/dev/null | head -5
    fi
else
    echo "  ⚠️ .claude/ 디렉토리 구조 문제"
fi

echo ""

# 2️⃣ .moai/ 구성 파일 검증
echo "2️⃣  .moai/ 구성 파일 검증..."

for FILE in config.json; do
    if [ -f ".moai/$FILE" ] && [ -f "$TEMPLATE_MOAI/$FILE" ]; then
        if diff ".moai/$FILE" "$TEMPLATE_MOAI/$FILE" > /dev/null 2>&1; then
            echo "  ✅ .moai/$FILE 최신 버전"
        else
            echo "  ⚠️  .moai/$FILE 버전 차이 있음"
            echo "     → 수동 검토 권장"
        fi
    fi
done

echo ""

# 3️⃣ CLAUDE.md 검증
echo "3️⃣  CLAUDE.md 검증..."

if [ -f "CLAUDE.md" ]; then
    # 템플릿 변수 치환 확인
    UNREPLACED_VARS=$(grep -c '{{' CLAUDE.md 2>/dev/null || echo "0")

    if [ "$UNREPLACED_VARS" -eq 0 ]; then
        echo "  ✅ CLAUDE.md 변수 치환 완료"
    else
        echo "  ❌ CLAUDE.md에 미치환 변수 $UNREPLACED_VARS개 발견"
        grep '{{' CLAUDE.md | head -3
        exit 1
    fi

    # 파일 크기 검증
    CLAUDE_SIZE=$(wc -c < CLAUDE.md)
    if [ "$CLAUDE_SIZE" -gt 500 ]; then
        echo "  ✅ CLAUDE.md 파일 크기 정상 ($CLAUDE_SIZE bytes)"
    else
        echo "  ⚠️  CLAUDE.md 파일이 너무 작음 ($CLAUDE_SIZE bytes)"
    fi
else
    echo "  ⚠️ CLAUDE.md가 생성되지 않음"
fi

echo ""

# 4️⃣ 설정 스키마 버전 검증
echo "4️⃣  설정 스키마 버전 검증..."

if [ -f ".moai/config.json" ] && [ -f "$TEMPLATE_MOAI/config.json" ]; then
    LOCAL_VERSION=$(grep '"version"' .moai/config.json | head -1 | grep -o '[0-9.]*' | head -1 || echo "unknown")
    TEMPLATE_VERSION=$(grep '"version"' "$TEMPLATE_MOAI/config.json" | head -1 | grep -o '[0-9.]*' | head -1 || echo "unknown")

    if [ "$LOCAL_VERSION" = "$TEMPLATE_VERSION" ]; then
        echo "  ✅ 설정 스키마 버전 일치: $LOCAL_VERSION"
    else
        echo "  ⚠️  설정 스키마 버전 불일치"
        echo "     - Local: $LOCAL_VERSION"
        echo "     - Template: $TEMPLATE_VERSION"
        echo "     → 설정 마이그레이션 확인 필요"
    fi
fi

echo ""

# 5️⃣ 안전성 검증 (백업 확인)
echo "5️⃣  백업 파일 검증..."

if [ -d ".moai-backups/pre-sync" ]; then
    BACKUP_COUNT=$(ls .moai-backups/pre-sync/ | wc -l)
    echo "  ✅ 백업 디렉토리 생성됨 ($BACKUP_COUNT개 항목)"
else
    echo "  ⚠️  백업 디렉토리가 생성되지 않음"
fi

echo ""

# 최종 검증 결과
echo "✅ 템플릿 동기화 검증 완료"
echo ""
echo "📝 다음 단계:"
echo "  1. 로컬 .claude/, .moai/ 파일 검토"
echo "  2. CLAUDE.md에 프로젝트 지침 추가 (선택사항)"
echo "  3. git add -A && git commit"
echo "  4. GitHub에 푸시"
```

**검증 항목**:
- ✅ Package template (.claude/)과 Local (.claude/) 동기화 상태
- ✅ 설정 파일 (.moai/config.json) 버전 호환성
- ✅ CLAUDE.md 템플릿 변수 완전 치환
- ✅ 파일 백업 생성 확인
- ⚠️ 불일치 항목 발견 시 경고

---

## 🤖 GitHub Actions의 자동화 (부분적)

### ⚠️ 현실적인 자동화 한계

GitHub Actions는 **이상적으로는** 다음 모든 단계를 자동으로 처리해야 하지만, **실제로는** 몇 가지 수동 개입이 필요할 수 있습니다:

```
이상적인 흐름:
PR 병합 → moai-gitflow.yml 감지 → Tag 생성 → Release 생성 → release.yml 트리거 → PyPI 배포

현실적인 흐름:
PR 병합 → moai-gitflow.yml (merge commit 감지 실패) → 수동 개입 필요
         또는 Release Pipeline 수정 필요
         또는 수동 배포 필요
```

**자동화가 작동하는 조건**:
1. ✅ Release commit 감지: PR 본문에 "🔖 RELEASE: v..." 포함
2. ✅ Release Pipeline 실행: moai-release-pipeline.yml이 패턴 감지
3. ✅ PyPI 배포: release.yml이 GitHub Release (published) 감지

**자동화가 실패하는 경우**:
1. ❌ merge commit 형식이 Release Pattern을 감지하지 못함 (가장 흔함)
2. ❌ GitHub Release가 Draft 상태로 남아있음
3. ❌ PYPI_API_TOKEN 미설정

### ✅ 자동화가 정상 작동 시 처리 단계

PR이 main에 병합되고 Release Pipeline이 정상 작동하면:

### ✅ Step 3.2a: SPEC Issue Auto-Detection (실행 필수)

**⚠️ CRITICAL**: 이 단계를 반드시 실행하여 SPEC 이슈가 자동으로 닫히도록 해야 합니다.

**SPEC 이슈 자동 감지 및 Closes 참조 생성:**
```bash
echo "🔍 Detecting SPEC issues for auto-close..."

# 1. .moai/specs 디렉토리에서 SPEC ID 찾기
SPEC_ID=$(find .moai/specs -maxdepth 2 -name "spec.md" -exec rg '@SPEC:[A-Z]+-[A-Z]+-\d+' --only-matching {} \; 2>/dev/null | head -1 | sed 's/@SPEC://')

if [ -z "$SPEC_ID" ]; then
    echo "ℹ️  No SPEC detected for this release"
    CLOSE_ISSUE_LINE=""
else
    echo "✅ SPEC detected: $SPEC_ID"

    # 2. GitHub에서 해당 SPEC 이슈 찾기 (title에 SPEC ID 포함된 것)
    SPEC_ISSUE=$(gh issue list --search "$SPEC_ID in:title" --state open --json number -q '.[0].number' 2>/dev/null)

    if [ -z "$SPEC_ISSUE" ]; then
        echo "ℹ️  No open GitHub issue found for $SPEC_ID"
        echo "→ Skipping auto-close (issue may already be closed or not exist)"
        CLOSE_ISSUE_LINE=""
    else
        echo "✅ Found open issue: #$SPEC_ISSUE ($SPEC_ID)"
        echo "→ Will add 'Closes #$SPEC_ISSUE' to commit message"
        CLOSE_ISSUE_LINE="\n\nCloses #${SPEC_ISSUE}"
    fi
fi

echo ""
echo "📝 Commit message will include:"
if [ -n "$CLOSE_ISSUE_LINE" ]; then
    echo "  - @SPEC:$SPEC_ID (TAG reference for traceability)"
    echo "  - Closes #$SPEC_ISSUE (GitHub auto-close trigger)"
else
    echo "  - Standard release message (no SPEC issue to close)"
fi
```

**검증 체크리스트:**
- ✅ SPEC ID 감지 확인 (e.g., SPEC-DOCS-004)
- ✅ GitHub issue 번호 확인 (e.g., #116)
- ✅ "Closes #XX" 문구 생성 확인
- ⚠️ 실패 시: 수동으로 issue 번호 확인 후 CLOSE_ISSUE_LINE 설정

### ✅ Step 3.2b: Git 커밋 생성 (SPEC 이슈 참조 포함)

**릴리즈 커밋 메시지 생성 및 커밋:**
```bash
echo "📝 Creating release commit..."

# 커밋 메시지 생성 (CLOSE_ISSUE_LINE 포함)
COMMIT_MSG="🔖 RELEASE: v${new_version}

Release v${new_version}

**변경사항**:
- 버전 관리 (SSOT): pyproject.toml ${current_version} → ${new_version}
- 템플릿 동기화 (자동)
- CLAUDE.md: 프로젝트 변수 적용

**SPEC Reference**:
@SPEC:${SPEC_ID}${CLOSE_ISSUE_LINE}

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Alfred <alfred@mo.ai.kr>"

# Git commit 실행
git add pyproject.toml uv.lock
git commit -m "$COMMIT_MSG"

if [ $? -eq 0 ]; then
    echo "✅ Release commit created: $(git rev-parse --short HEAD)"
    echo "📌 Commit message preview:"
    git log -1 --pretty=format:"%B"
else
    echo "❌ Failed to create release commit"
    exit 1
fi
```

**효과:**
- ✅ GitHub가 merge 시 자동으로 issue #XX를 close
- ✅ SPEC 문서와 GitHub Issue의 라이프사이클 자동 동기화
- ✅ 개발자의 수동 close 작업 제거
- ✅ @TAG와 Closes 참조 모두 포함으로 완벽한 traceability

### ✅ Step 3.3: Git Annotated Tag 생성 (자동)
GitHub Actions가 자동으로 생성:
```
v{new_version}

Release v{new_version}
Released: {YYYY-MM-DD}
```

### ✅ Step 3.4: 패키지 빌드 (자동)
GitHub Actions가 `uv build` 실행하여 wheel과 tar.gz 생성

### ✅ Step 3.5: PyPI 배포 (자동)
GitHub Actions가 `uv publish` 실행하여 PyPI에 배포

**🔒 보안**: PyPI 토큰은 GitHub Secrets에서 관리되므로, 로컬에서 설정할 필요 없음

### ✅ Step 3.6: GitHub Release 생성 (자동)

#### 📋 Release Notes Format Guide

**Title Format** (English only):
```
🔖 v[VERSION] | [TYPE] | [Release Title]

Examples:
🔖 v0.5.4 | patch | Feature Enhancement & uv Standardization
🔖 v0.6.0 | minor | New Skill System Implementation
🔖 v1.0.0 | major | First Stable Release
```

⚠️ **Important**: All release information must be in **English only** for international consistency.

**Release Notes Structure** (English only):

```
# 🎉 Release v[VERSION] | [TYPE] | [Release Title]

**Version**: v[VERSION]
**Type**: [patch | minor | major]
**Release Date**: YYYY-MM-DD

## What's Changed

### ✨ New Features (N)
- Feature 1 description
- Feature 2 description

### 🐛 Bug Fixes (N)
- Bug fix 1 description
- Bug fix 2 description

### ♻️ Improvements (N)
- Improvement 1 description
- Improvement 2 description

### 📚 Documentation (N)
- Documentation 1 update
- Documentation 2 update

## Quality Assurance Results

| Metric | Result | Status |
|--------|--------|--------|
| Tests Passed | X/X (Y%) | ✅ Passed |
| Code Coverage | Y% | ✅ Target Met |
| Linting | 0 errors | ✅ Passed |
| Type Checking | 0 issues | ✅ Passed |
| Security Scan | 0 vulnerabilities | ✅ Passed |

## Installation

### Using uv (Recommended)
\`\`\`bash
uv tool install moai-adk==[VERSION]
\`\`\`

### Using pip (Legacy)
\`\`\`bash
pip install moai-adk==[VERSION]
\`\`\`

## Full Changelog

Compare all changes: [v[PREV]...v[VERSION]](https://github.com/modu-ai/moai-adk/compare/v[PREV]...v[VERSION])

## Contributors

Thanks to all contributors who made this release possible.

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Alfred <alfred@mo.ai.kr>
```

#### 작성 팁

- ✅ **Use clear English throughout**
- ✅ **Highlight changes compared to previous version**
- ✅ **Use consistent terminology**
- ✅ **Include accurate metrics** (test pass rate, coverage)
- ✅ **Verify links** (v[PREV] and v[VERSION] accurately)

**Create Release with gh CLI** (English only):
```bash
# Generate release notes (use template above)
release_title="🔖 v{new_version} | {VERSION_TYPE} | {Release Title}"

release_notes="# 🎉 Release v{new_version} | {VERSION_TYPE}

**Version**: v{new_version}
**Type**: {VERSION_TYPE}
**Release Date**: $(date +%Y-%m-%d)

## What's Changed

### ✨ New Features (N)
- Feature 1 description
- Feature 2 description

### 🐛 Bug Fixes (N)
- Bug fix 1 description
- Bug fix 2 description

### ♻️ Improvements (N)
- Improvement 1 description
- Improvement 2 description

### 📚 Documentation (N)
- Documentation 1 update
- Documentation 2 update

## Quality Assurance Results

| Metric | Result | Status |
|--------|--------|--------|
| Tests Passed | X/X (Y%) | ✅ Passed |
| Code Coverage | Y% | ✅ Target Met |
| Linting | 0 errors | ✅ Passed |
| Type Checking | 0 issues | ✅ Passed |
| Security Scan | 0 vulnerabilities | ✅ Passed |

## Installation

### Using uv (Recommended)
\`\`\`bash
uv tool install moai-adk=={new_version}
\`\`\`

### Using pip (Legacy)
\`\`\`bash
pip install moai-adk=={new_version}
\`\`\`

## Full Changelog

Compare all changes: [v{current_version}...v{new_version}](https://github.com/modu-ai/moai-adk/compare/v{current_version}...v{new_version})

## Contributors

Thanks to all contributors who made this release possible.

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Alfred <alfred@mo.ai.kr>"

# Create GitHub Release (Draft, English only)
gh release create "v{new_version}" \
  --title "$release_title" \
  --notes "$release_notes" \
  --draft

echo "ℹ️ GitHub Release created as Draft"
echo "→ https://github.com/modu-ai/moai-adk/releases/tag/v{new_version}"
echo "→ Verify content and publish the release..."
```

### Step 3.7: Publish GitHub Release (Draft → Published)

**Convert Draft Release to Published**:
```bash
# Change GitHub Release Draft to Published
echo "📢 Publishing GitHub Release..."
gh release edit "v{new_version}" --draft=false

if [ $? -ne 0 ]; then
    echo "❌ Failed to publish GitHub Release"
    echo "→ Check: gh CLI authentication status"
    echo "→ Solution: gh auth login or gh auth refresh"
    exit 1
fi

echo "✅ GitHub Release Published!"
echo "→ Latest releases: https://github.com/modu-ai/moai-adk/releases"
echo "→ Release page: https://github.com/modu-ai/moai-adk/releases/tag/v{new_version}"
```

**Verification Checklist**:
- ✅ Verify Draft status before publishing
- ✅ Confirm "Latest" release is updated
- ✅ Verify GitHub Release page is public

---

### Step 3.8: Final Report

```markdown
# ✅ Release Complete: v{new_version}

## Release Results
✅ Version updated (pyproject.toml)
✅ Git tag created and pushed: v{new_version}
✅ Package built (dist/)
✅ Deployed to PyPI (https://pypi.org/project/moai-adk/{new_version}/)
✅ GitHub Release published (https://github.com/modu-ai/moai-adk/releases/tag/v{new_version})

## Next Steps
1. Verify GitHub Release page
2. Execute Step 3.9: Post-Release Cleanup
3. Start planning next feature with /alfred:1-plan

## Installation Test
```bash
# Test installation with uv tool (Recommended)
uv tool install moai-adk=={new_version}
moai-adk --version

# Or install with pip (Legacy)
# pip install moai-adk=={new_version}
# moai-adk --version
```
```

### Step 3.10: 패키지 템플릿 동기화 및 변수 최적화 (필수)

**⚠️ CRITICAL**: PyPI 배포 완료 후 반드시 실행. 로컬 프로젝트에 최신 패키지 템플릿을 동기화합니다.

**목적**:
- 패키지 템플릿 (`src/moai_adk/templates/`)의 최신 변경사항을 로컬 프로젝트에 반영
- `config.json` 버전 및 메타데이터 최적화
- CLAUDE.md 프로젝트 지침 업데이트
- 패키지 템플릿이 source of truth 원칙 유지

**패키지 템플릿 동기화 스크립트:**

```bash
echo "🔄 패키지 템플릿 동기화 및 최적화 중..."
echo ""

# 1️⃣ 버전 정보 수집
CURRENT_VERSION=$(rg "^version = " pyproject.toml -A0 -o '$1' | head -1 | sed 's/version = "\(.*\)"/\1/')
echo "📌 배포 완료 버전: v$CURRENT_VERSION"
echo ""

# 2️⃣ 패키지 템플릿 디렉토리 경로 설정
TEMPLATE_DIR="src/moai_adk/templates"
TEMPLATE_CONFIG="$TEMPLATE_DIR/.moai/config.json"

if [ ! -f "$TEMPLATE_CONFIG" ]; then
    echo "❌ 패키지 템플릿을 찾을 수 없습니다: $TEMPLATE_CONFIG"
    exit 1
fi

echo "📦 패키지 템플릿 위치: $TEMPLATE_DIR"
echo ""

# 3️⃣ .claude/ 동기화 (agents, commands, hooks, output-styles)
echo "1️⃣  .claude/ 디렉토리 동기화 중..."

for subdir in agents commands hooks output-styles; do
    if [ -d "$TEMPLATE_DIR/.claude/$subdir" ]; then
        cp -r "$TEMPLATE_DIR/.claude/$subdir" .claude/ 2>/dev/null
        file_count=$(find .claude/$subdir -type f | wc -l)
        echo "  ✅ .claude/$subdir - $file_count 파일 동기화됨"
    fi
done

# settings.json 동기화
if [ -f "$TEMPLATE_DIR/.claude/settings.json" ]; then
    cp "$TEMPLATE_DIR/.claude/settings.json" .claude/
    echo "  ✅ .claude/settings.json 동기화됨"
fi

echo ""

# 4️⃣ .moai/memory/ 동기화 (개발 가이드, 규칙, 실습 문서)
echo "2️⃣  .moai/memory/ 디렉토리 동기화 중..."

if [ -d "$TEMPLATE_DIR/.moai/memory" ]; then
    cp -r "$TEMPLATE_DIR/.moai/memory" .moai/
    file_count=$(find .moai/memory -type f | wc -l)
    echo "  ✅ .moai/memory - $file_count 파일 동기화됨"
fi

echo ""

# 5️⃣ 버전 최적화 (.moai/config.json)
echo "3️⃣  config.json 버전 및 메타데이터 최적화 중..."

# 패키지 템플릿의 config.json에서 최신 구조 읽기
TEMPLATE_VERSION=$(rg '"version":\s*"([^"]+)"' "$TEMPLATE_CONFIG" -o '$1' | head -1)

# 로컬 config.json 업데이트
# - moai.version 업데이트 (배포된 실제 버전으로)
# - language section 구조화
# - project section 정규화

python3 << 'EOF'
import json
import sys
from pathlib import Path

config_path = Path(".moai/config.json")
if not config_path.exists():
    print("❌ .moai/config.json을 찾을 수 없습니다")
    sys.exit(1)

with open(config_path) as f:
    config = json.load(f)

# 버전 업데이트 (배포된 버전)
config["moai"]["version"] = "CURRENT_VERSION"

# version_check 섹션 추가
if "version_check" not in config["moai"]:
    config["moai"]["update_check_frequency"] = "daily"
    config["moai"]["version_check"] = {
        "enabled": True,
        "cache_ttl_hours": 24
    }

# language section 구조화 (있으면 유지, 없으면 추가)
if "language" not in config:
    config["language"] = {
        "conversation_language": config["project"].get("conversation_language", "en"),
        "conversation_language_name": config["project"].get("conversation_language_name", "English")
    }

# project section 정규화 (중복 제거)
if "conversation_language" in config["project"]:
    del config["project"]["conversation_language"]
if "conversation_language_name" in config["project"]:
    del config["project"]["conversation_language_name"]
if "user_nickname" in config["project"]:
    del config["project"]["user_nickname"]
if "template_version" in config["project"]:
    del config["project"]["template_version"]

# 저장
with open(config_path, "w") as f:
    json.dump(config, f, indent=2, ensure_ascii=False)
    f.write("\n")

print("  ✅ config.json 최적화 완료")
print(f"     - moai.version: {config['moai']['version']}")
print(f"     - language section 정규화됨")
print(f"     - project section 정규화됨 (중복 제거)")
EOF

# 변수 치환 (CURRENT_VERSION → 실제 버전)
sed -i '' "s/CURRENT_VERSION/$CURRENT_VERSION/g" .moai/config.json

echo ""

# 6️⃣ CLAUDE.md 동기화 및 변수 치환
echo "4️⃣  CLAUDE.md 프로젝트 지침 업데이트 중..."

TEMPLATE_CLAUDE_MD="$TEMPLATE_DIR/CLAUDE.md"

if [ -f "$TEMPLATE_CLAUDE_MD" ]; then
    # CLAUDE.md 복사
    cp "$TEMPLATE_CLAUDE_MD" CLAUDE.md

    # .moai/config.json에서 프로젝트 정보 추출
    PROJECT_NAME=$(rg '"name":\s*"([^"]+)"' .moai/config.json -o '$1' | head -1)
    PROJECT_LOCALE=$(rg '"locale":\s*"([^"]+)"' .moai/config.json -o '$1' | head -1)
    CONVERSATION_LANGUAGE=$(rg '"conversation_language":\s*"([^"]+)"' .moai/config.json -o '$1' | head -1)
    CONVERSATION_LANGUAGE_NAME=$(rg '"conversation_language_name":\s*"([^"]+)"' .moai/config.json -o '$1' | head -1)
    CODEBASE_LANGUAGE=$(rg '"language":\s*"([^"]+)"' .moai/config.json -o '$1' | head -1)

    # 변수 치환 (소문자, 대문자, {{}} 형식 모두 처리)
    sed -i '' "s|{{project_name}}|$PROJECT_NAME|g" CLAUDE.md
    sed -i '' "s|{{PROJECT_NAME}}|$PROJECT_NAME|g" CLAUDE.md
    sed -i '' "s|{{locale}}|$PROJECT_LOCALE|g" CLAUDE.md
    sed -i '' "s|{{LOCALE}}|$PROJECT_LOCALE|g" CLAUDE.md
    sed -i '' "s|{{conversation_language}}|$CONVERSATION_LANGUAGE|g" CLAUDE.md
    sed -i '' "s|{{conversation_language_name}}|$CONVERSATION_LANGUAGE_NAME|g" CLAUDE.md
    sed -i '' "s|{{codebase_language}}|$CODEBASE_LANGUAGE|g" CLAUDE.md
    sed -i '' "s|{{CODEBASE_LANGUAGE}}|$CODEBASE_LANGUAGE|g" CLAUDE.md

    echo "  ✅ CLAUDE.md 업데이트 완료"
    echo "     - project_name: $PROJECT_NAME"
    echo "     - locale: $PROJECT_LOCALE"
    echo "     - conversation_language: $CONVERSATION_LANGUAGE"
    echo "     - codebase_language: $CODEBASE_LANGUAGE"
fi

echo ""
echo "✅ 패키지 템플릿 동기화 및 최적화 완료!"
echo ""
echo "📋 동기화 체크리스트:"
echo "  ✅ .claude/ (agents, commands, hooks, output-styles, settings.json)"
echo "  ✅ .moai/memory/ (CLAUDE-RULES.md, DEVELOPMENT-GUIDE.md, etc.)"
echo "  ✅ .moai/config.json (버전, 언어, 메타데이터 최적화)"
echo "  ✅ CLAUDE.md (프로젝트 변수 치환)"
echo ""
echo "→ 다음 단계: Git 커밋 (Step 3.11)"
```

**검증 체크리스트:**
- ✅ `.claude/agents/`, `.claude/commands/`, `.claude/hooks/` 최신 파일
- ✅ `.moai/memory/` 개발 가이드 최신화
- ✅ `config.json` 버전: v{new_version}
- ✅ `config.json` 구조: language section 분리, project section 정규화
- ✅ `CLAUDE.md` 프로젝트 변수 자동 치환
- ✅ 모든 변경사항이 staged 상태 준비

---

### Step 3.11: 패키지 템플릿 동기화 커밋 (필수)

**목적**:
- Step 3.10에서 동기화한 패키지 템플릿을 Git에 커밋
- 로컬 프로젝트가 패키지 템플릿과 동일한 상태로 유지
- 버전 번호와 메타데이터 최적화 기록

**패키지 템플릿 동기화 커밋 스크립트:**

```bash
echo "📝 패키지 템플릿 동기화 커밋 생성 중..."
echo ""

# 1️⃣ 변경사항 확인
echo "📊 변경된 파일 목록:"
git status --short | grep -E "\.claude/|\.moai/|CLAUDE.md" | head -20

echo ""

# 2️⃣ 파일 staging
echo "📦 변경사항 staging 중..."

git add .claude/
git add .moai/config.json
git add .moai/memory/
git add CLAUDE.md

staged_count=$(git diff --cached --name-only | wc -l)
echo "  ✅ $staged_count개 파일 staged"

echo ""

# 3️⃣ 커밋 메시지 생성 (영문만)
COMMIT_MSG="chore: Synchronize package templates to local project after release

- Sync .claude/ (agents, commands, hooks, output-styles, settings.json)
- Sync .moai/memory/ (development guides, rules, practices)
- Update .moai/config.json: version v${CURRENT_VERSION}, optimize structure
- Update CLAUDE.md: project variables substitution
- Maintain package template as source of truth (src/moai_adk/templates/)

**File Changes**:
- .claude/: Alfred agents, commands, hooks, output styles
- .moai/: Project configuration, development documentation
- CLAUDE.md: Project directives with substituted variables

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Alfred <alfred@mo.ai.kr>"

# 4️⃣ 커밋 실행
echo "💾 커밋 생성 중..."
git commit -m "$COMMIT_MSG"

if [ $? -eq 0 ]; then
    COMMIT_HASH=$(git rev-parse --short HEAD)
    echo "  ✅ 커밋 생성 완료: $COMMIT_HASH"
    echo ""
    echo "📝 커밋 메시지:"
    git log -1 --pretty=format:"%B"
    echo ""
else
    echo "⚠️ 커밋 실패 (변경사항 없을 수 있음)"
fi

echo ""
echo "✅ 패키지 템플릿 동기화 커밋 완료!"
echo "→ 다음 단계: Push to remote (선택 사항) 또는 Step 3.12 (Post-Release Cleanup)"
```

**검증 체크리스트:**
- ✅ 변경된 파일이 모두 staged 상태
- ✅ 커밋 메시지: 영문으로 작성, Alfred 공동저자 포함
- ✅ 커밋 해시 기록됨

**참고사항:**
- ℹ️ 이 커밋은 develop/main 브랜치에 추가됨
- ℹ️ PyPI 배포는 이미 완료된 상태
- ℹ️ 패키지 사용자에게는 영향 없음 (코드 변경 없음)

---

### Step 3.12: Post-Release Cleanup (필수)
- 저장소 상태 정리
- 다음 개발 사이클 준비

**Cleanup 스크립트:**
```bash
echo "🧹 Starting post-release cleanup..."
echo ""

# 1. 현재 브랜치 확인
CURRENT_BRANCH=$(git branch --show-current)
echo "📍 Currently on: $CURRENT_BRANCH"

# 2. develop 브랜치 존재 여부 확인
if git show-ref --verify --quiet refs/heads/develop; then
    TARGET_BRANCH="develop"
    echo "🎯 Target branch: develop (GitFlow mode)"
else
    TARGET_BRANCH="main"
    echo "🎯 Target branch: main (Simplified mode)"
fi

# 3. target 브랜치로 전환 및 최신화
echo "🔄 Switching to $TARGET_BRANCH..."
git checkout $TARGET_BRANCH

if [ $? -ne 0 ]; then
    echo "❌ Failed to checkout $TARGET_BRANCH"
    exit 1
fi

echo "⬇️  Pulling latest changes from origin/$TARGET_BRANCH..."
git pull origin $TARGET_BRANCH

# 4. 병합된 로컬 브랜치 삭제
echo ""
echo "🗑️  Identifying merged local branches..."
MERGED_BRANCHES=$(git branch --merged | grep -v "^\*" | grep -v "\bmain\b" | grep -v "\bdevelop\b" | xargs)

if [ -n "$MERGED_BRANCHES" ]; then
    echo "Found merged branches:"
    echo "$MERGED_BRANCHES" | tr ' ' '\n' | sed 's/^/  - /'
    echo ""

    for branch in $MERGED_BRANCHES; do
        git branch -d "$branch" 2>/dev/null
        if [ $? -eq 0 ]; then
            echo "  ✅ Deleted local branch: $branch"
        fi
    done
else
    echo "  ℹ️  No merged local branches to clean up"
fi

# 5. 병합된 원격 브랜치 삭제
echo ""
echo "🌐 Identifying merged remote branches..."
REMOTE_MERGED=$(git branch -r --merged origin/$TARGET_BRANCH | grep -v "HEAD" | grep -v "\bmain\b" | grep -v "\bdevelop\b" | sed 's|origin/||' | xargs)

if [ -n "$REMOTE_MERGED" ]; then
    echo "Found merged remote branches:"
    echo "$REMOTE_MERGED" | tr ' ' '\n' | sed 's/^/  - origin\//'
    echo ""

    for branch in $REMOTE_MERGED; do
        # Skip if branch doesn't exist on remote
        if git ls-remote --heads origin "$branch" | grep -q "$branch"; then
            git push origin --delete "$branch" 2>/dev/null
            if [ $? -eq 0 ]; then
                echo "  ✅ Deleted remote branch: origin/$branch"
            else
                echo "  ⚠️  Failed to delete remote branch: origin/$branch (may require permissions)"
            fi
        fi
    done
else
    echo "  ℹ️  No merged remote branches to clean up"
fi

# 6. 최종 상태 확인
echo ""
echo "📊 Final repository status:"
echo ""
echo "📍 Current branch: $(git branch --show-current)"
echo "🌲 Local branches:"
git branch | sed 's/^/  /'
echo ""
echo "🌐 Remote branches:"
git branch -r | grep -v "HEAD" | sed 's/^/  /'
echo ""

# 7. 선택적: dist/ 디렉토리 정리
if [ -d "dist" ]; then
    echo "🗂️  dist/ directory found (build artifacts)"
    read -p "Delete dist/ directory? (y/n): " delete_dist
    if [ "$delete_dist" = "y" ]; then
        rm -rf dist/
        echo "  ✅ Deleted dist/"
    else
        echo "  ℹ️  Keeping dist/"
    fi
fi

echo ""
echo "✅ Post-release cleanup complete!"
echo ""
echo "🚀 Repository is ready for next development cycle!"
echo "→ Start planning next feature: /alfred:1-plan \"Feature name\""
```

**검증 체크리스트:**
- ✅ develop 또는 main 브랜치로 복귀 확인
- ✅ 병합된 feature 브랜치 삭제 확인 (local)
- ✅ 병합된 feature 브랜치 삭제 확인 (remote)
- ✅ 로컬 저장소 최신 상태 확인
- ✅ dist/ 디렉토리 정리 (선택 사항)

**참고사항:**
- ⚠️ 원격 브랜치 삭제는 권한이 필요할 수 있습니다
- ⚠️ 삭제 전 브랜치가 실제로 병합되었는지 자동 확인됩니다
- ℹ️  main/develop 브랜치는 자동으로 보호됩니다 (삭제 안 됨)

### ✨ Release Complete!

All deployment and cleanup tasks have been successfully completed.

**Next Steps**:
```bash
# You are now on develop/main branch
# Start planning next feature
/alfred:1-plan "New feature name"
```

---

## 🛡️ 안전 장치

### 사전 검증

**필수 조건 체크**:
```bash
# Git 저장소 확인
[ -d .git ] || { echo "❌ Git 저장소가 아닙니다"; exit 1; }

# pyproject.toml 존재 확인 (SSOT)
[ -f pyproject.toml ] || { echo "❌ pyproject.toml이 없습니다"; exit 1; }

# __init__.py 존재 확인 (importlib.metadata 사용 여부 확인)
[ -f src/moai_adk/__init__.py ] || { echo "❌ __init__.py가 없습니다"; exit 1; }

# editable install 확인 (SSOT 방식)
python -c "from importlib.metadata import version; version('moai-adk')" 2>/dev/null || {
    echo "⚠️ 패키지가 설치되지 않았습니다"
    echo "→ uv pip install -e . 실행 필요"
}

# gh CLI 인증 확인
gh auth status 2>/dev/null || echo "⚠️ gh CLI 미인증 (GitHub Release 생략)"
```

### 🔄 GitHub Actions 실패 시 롤백

GitHub Actions에서 릴리즈가 실패한 경우:

**1️⃣ 원인 파악**:
```bash
# GitHub Actions 로그 확인
gh run view <RUN_ID> --log

# 또는 특정 워크플로우 상태 확인
gh run list --branch main --limit 5 --json name,status,conclusion

# 최신 실행 상세 정보
gh run view $(gh run list --branch main --limit 1 --json databaseId -q '.[0].databaseId') --json jobs
```

**2️⃣ Release Pipeline 미감지 문제**:
```bash
# 증상: moai-gitflow.yml이 "skipped" 상태
# 원인: merge commit이 Release 패턴을 감지하지 못함

# 해결책 1: .github/workflows/moai-release-pipeline.yml 수정
# 라인 37-38: grep -q "^🔖 RELEASE:" → grep -q "🔖 RELEASE:"
# (^ 제거로 패턴 위치 제약 해제)

# 해결책 2: 수동 배포 (긴급)
git tag -a v{new_version} -m "Release v{new_version}" <COMMIT_SHA>
git push origin v{new_version}
gh release create v{new_version} --title "Release v{new_version}" --notes "[노트]"
```

**3️⃣ PyPI 배포 실패**:
```bash
# 증상: release.yml이 failed 또는 skipped
# 원인: PYPI_API_TOKEN 미설정 또는 GitHub Release 미생성

# 해결책 1: GitHub Release 공개 (draft → published)
gh release edit v{new_version} --draft=false

# 해결책 2: 수동 PyPI 배포
python3 -m build dist/
uv publish dist/moai_adk-{new_version}*.whl dist/moai_adk-{new_version}.tar.gz
```

**4️⃣ PR 복귀 (모든 자동화 실패 시)**:
```bash
# PR #XX를 revert (GitHub 웹에서 수동)
# 또는 CLI로
git revert <COMMIT_HASH>
git push upstream develop
```

**5️⃣ 태그/Release 정리** (GitHub Actions가 이미 생성한 경우):
```bash
# 로컬 태그 삭제
git tag -d v{new_version}
git fetch origin

# GitHub Release 삭제
gh release delete v{new_version} --yes

# 원격 태그 삭제
git push origin :refs/tags/v{new_version}
```

**6️⃣ 문제 해결 후 재시도**:
```bash
# develop에서 문제 수정
git checkout develop
# ... 문제 해결 ...
git commit -m "fix: Release issue"
git push upstream develop

# Phase 2부터 재시작
/awesome:release-new patch
```

⚠️ **PyPI 배포는 롤백 불가**:
- 이미 배포된 버전은 PyPI에서 삭제할 수 없음
- 버전 번호 변경 후 새로 배포해야 함
- 필요 시 yanked 버전으로 표시 (PyPI 웹 대시보드)

### 에러 처리

**주요 에러 케이스**:

1. **Release Pipeline 미감지** (가장 흔함)
   - 증상: moai-gitflow.yml이 "skipped" 상태
   - 원인: merge commit의 첫 줄이 "Merge pull request..."
   - 해결: 라인 37에서 grep -q "^🔖" → grep -q "🔖" 변경
   - 대체: 수동으로 git tag + gh release create 실행

2. **PyPI 배포 실패**
   - 증상: release.yml이 failed 또는 skipped
   - 원인: PYPI_API_TOKEN 미설정 또는 GitHub Release Draft
   - 해결: gh release edit v{VERSION} --draft=false
   - 대체: uv publish 또는 twine upload 로컬 실행

3. **버전 충돌**
   - 증상: git tag -a 실패 (tag already exists)
   - 확인: git tag -l "v{VERSION}"
   - 해결: 버전 번호 증가 후 재시도

4. **gh CLI 실패**
   - 증상: GitHub Release 생성 불가
   - 확인: gh auth status
   - 해결: gh auth login 또는 gh auth refresh

5. **PyPI API 토큰 오류**
   - 증상: uv publish 실패 (authentication failed)
   - 확인: echo $UV_PUBLISH_TOKEN
   - 해결: PyPI에서 새 토큰 생성 후 GitHub Secrets 업데이트

---

## 📊 최종 보고서 형식

```markdown
🎉 릴리즈 v{new_version} 완료!

✅ 완료된 작업

1. 품질 검증 (CodeRabbit AI) ✅
   - 자동 코드 리뷰: 완료 (모든 PR)
   - 자동 승인: ✅ (품질 80% 이상)
   - 보안 검사: ✅ (취약점 0개)
   - 테스트 커버리지: {COVERAGE}%

2. GitFlow PR 병합 ✅
   - 브랜치: develop → main
   - PR #: {PR_NUMBER}
   - CodeRabbit 자동 승인: ✅
   - 병합 타입: Merge Commit

3. 버전 업데이트 (SSOT) ✅
   - pyproject.toml: {old} → {new} (SSOT)
   - __init__.py: 자동 로드 (importlib.metadata)

4. Git 작업 ✅
   - 커밋: {COMMIT_HASH}
   - 태그: v{new_version}
   - 푸시: origin/main ✅

5. 배포 ✅
   - PyPI: moai-adk@{new_version} ✅
   - GitHub Release (Draft): https://github.com/modu-ai/moai-adk/releases/tag/v{new_version} ✅

---

📦 릴리즈 정보

- **버전**: v{new_version}
- **타입**: {VERSION_TYPE}
- **날짜**: {YYYY-MM-DD}
- **커밋**: {COMMIT_HASH}
- **브랜치**: main (develop에서 병합)
- **변경사항**: {N}개 커밋

🔗 링크

- GitHub PR: https://github.com/modu-ai/moai-adk/pull/{PR_NUMBER}
- PyPI: https://pypi.org/project/moai-adk/{new_version}
- GitHub Release: https://github.com/modu-ai/moai-adk/releases/tag/v{new_version}

---

🚀 다음 단계

- [ ] GitHub Release Draft 검토 및 게시
- [ ] 사용자 공지 (필요 시)
- [ ] develop 브랜치 정리 (선택 사항)
- [ ] 다음 개발 사이클 시작

모든 작업이 성공적으로 완료되었습니다! 🎉

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Alfred <alfred@mo.ai.kr>
```

---

## 🔧 고급 옵션

### 환경 변수

```bash
# PyPI 토큰 설정 (uv publish) - pypi- 접두사 필수!
export UV_PUBLISH_TOKEN="pypi-AgEIcHlwaS5vcmcCJ..."

# 또는 .pypirc 파일 사용 (~/.pypirc)
# [pypi]
#   username = __token__
#   password = pypi-AgEIcHlwaS5vcmcCJ...

# GitHub Enterprise 사용 시
export GH_HOST="github.company.com"

# dry-run 모드
export DRY_RUN=true
```

### 플래그 지원

```bash
# PyPI 배포 건너뛰기
/awesome:release-new patch --skip-pypi

# GitHub Release 건너뛰기
/awesome:release-new minor --skip-github

# 자동 승인 (Phase 1 건너뜀, 위험!)
/awesome:release-new major --auto-confirm

# 품질 검증 건너뛰기 (매우 위험!)
/awesome:release-new patch --skip-quality
```

---

## 💡 사용 예시

### 기본 사용

```bash
# 패치 버전 증가 (기본값)
/awesome:release-new
/awesome:release-new patch

# 마이너 버전 증가
/awesome:release-new minor

# 메이저 버전 증가
/awesome:release-new major
```

### 고급 사용

```bash
# PyPI 배포 없이 GitHub Release만
/awesome:release-new patch --skip-pypi

# 빠른 패치 릴리즈 (자동 승인)
/awesome:release-new patch --auto-confirm
```

---

## 📋 체크리스트

**릴리즈 전 확인사항**:
- [ ] 모든 테스트 통과 (`pytest tests/`)
- [ ] 린트 검사 통과 (`ruff check .`)
- [ ] Git 브랜치: main 권장
- [ ] 미커밋 변경사항 정리
- [ ] pyproject.toml 필드 확인

**릴리즈 후 확인사항**:
- [ ] PyPI 패키지 설치 테스트
- [ ] GitHub Release 노트 검토
- [ ] Git 태그 확인 (`git tag -l "v*"`)
- [ ] 다음 버전 계획

---

## 🛠️ 기술 스택

- **Git**: 커밋, 태그, 푸시
- **uv**: 빌드 및 PyPI 배포 (권장)
- **gh CLI**: GitHub Release 자동화
- **pytest**: 테스트 및 커버리지
- **ruff**: 린트
- **mypy**: 타입 체크
- **bandit/pip-audit**: 보안 스캔

---

---

## 🔬 Dry-Run 모드 상세 설명

**Dry-Run 모드**는 릴리즈 프로세스를 **완전하게 시뮬레이션**하여, 실제 릴리즈 전에 모든 단계를 미리 확인할 수 있게 해줍니다.

### 사용 시나리오

#### Scenario 1: 처음 릴리즈하는 경우

```bash
# Dry-Run으로 먼저 확인
/awesome:release-new minor --dry-run

# 출력 예:
# 🔬 Dry-Run 모드 활성화
# 📊 시뮬레이션 계획:
# ...
# 실제 릴리즈를 진행하려면:
# /awesome:release-new minor

# 결과가 만족스러우면 실제 실행
/awesome:release-new minor
```

#### Scenario 2: 자동화된 파이프라인 전에 검증

```bash
# CI/CD 파이프라인이 있는 경우, 수동으로 한 번 Dry-Run으로 검증
/awesome:release-new patch --dry-run

# 검증 후 CI/CD 트리거
gh workflow run release.yml
```

#### Scenario 3: 버전 계획 검토

```bash
# 세 가지 버전 타입을 모두 시뮬레이션해서 비교
/awesome:release-new patch --dry-run   # v0.13.1
/awesome:release-new minor --dry-run   # v0.14.0
/awesome:release-new major --dry-run   # v1.0.0
```

### Dry-Run 모드 실행 순서

```
/awesome:release-new [version] --dry-run
    ↓
Phase 0: 품질 검증 (실제 실행)
├─ pytest 실행
├─ ruff 린트
├─ mypy 타입 체크
└─ bandit + pip-audit 보안 스캔
    ↓
Phase 1: 버전 분석 (시뮬레이션)
├─ 현재/목표 버전 계산
├─ Git 로그 분석
└─ 변경사항 요약
    ↓
Phase 1.5: 릴리즈 계획 보고서 (출력, 승인 대기 없음)
├─ 버전 변경사항
├─ 예정된 파일 수정
└─ 예정된 Git 작업
    ↓
Phase 2: PR 관리 (시뮬레이션만)
├─ [DRY-RUN] GitHub PR 생성 예정
└─ [DRY-RUN] GitHub 웹에서 병합 필요
    ↓
Phase 3: 자동 릴리즈 (실행 안 함)
├─ [DRY-RUN] Git 태그 생성 예정
├─ [DRY-RUN] GitHub Release 생성 예정
└─ [DRY-RUN] PyPI 배포 예정
    ↓
🔬 Dry-Run 시뮬레이션 완료
└─ 실제 릴리즈 명령 제시
```

### Dry-Run 모드에서 출력되는 정보

#### Phase 0 출력 (실제)

```bash
🐍 Python 버전: 3.13.0
📦 패키지 매니저: uv
🧪 테스트 실행 중...
✅ 테스트 통과 (커버리지 87%)
🔍 린트 검사 중...
✅ 린트 통과
🔤 타입 체크 중...
✅ 타입 체크 통과
🔒 보안 스캔 중...
✅ 보안 스캔 통과
```

#### Phase 1 출력 (시뮬레이션)

```bash
📌 현재 버전 (pyproject.toml): 0.13.0
📊 버전 변경: 0.13.0 → 0.14.0
🏷️ 마지막 릴리즈: v0.13.0
📝 변경사항: 5개 커밋
   - feat(spec): Add new features (#155)
   - fix: Resolve issue (#152)
   - docs: Update documentation
   - refactor: Code optimization
   - test: Add test coverage
```

#### Phase 1.5-3 출력 (시뮬레이션 로그)

```
🔬 릴리즈 계획 보고서
📊 버전 정보
- 현재 버전: v0.13.0
- 목표 버전: v0.14.0
- 버전 타입: minor

   [DRY-RUN] 파일 수정 예정: pyproject.toml
   [DRY-RUN] Git 커밋 예정: 🔖 RELEASE: v0.14.0
   [DRY-RUN] Git 태그 예정: v0.14.0
   [DRY-RUN] GitHub PR 생성 예정: Release v0.14.0
   [DRY-RUN] GitHub Release 생성 예정: v0.14.0
```

#### 최종 요약 (Dry-Run만)

```
================================
🔬 Dry-Run 시뮬레이션 완료
================================

예정된 작업 목록:
[12:34:56] 파일 수정 예정: pyproject.toml
[12:34:56] Git 커밋 예정: 🔖 RELEASE: v0.14.0
[12:34:56] Git 태그 예정: v0.14.0 - Release v0.14.0
[12:34:56] GitHub PR 생성 예정: Release v0.14.0

⚠️ 위의 작업들은 시뮬레이션만 수행되었으며, 실제로 적용되지 않았습니다.

실제 릴리즈를 진행하려면 다음 명령을 실행하세요:
/awesome:release-new minor
```

### Dry-Run 모드 주의사항

#### ✅ Dry-Run에서 안전한 작업

- Phase 0 (품질 검증)은 **실제 실행**됩니다
  - 테스트 실행 중 코드 수정 필요할 수 있음 주의
  - 테스트 실패 시 Dry-Run도 중단됨
- 모든 분석 및 계획은 부작용 없음
- Git 로그 조회는 읽기만 함

#### ⚠️ Dry-Run에서 실행 안 되는 작업

- **파일 수정**: pyproject.toml 버전 변경 안 함
- **Git 커밋**: 커밋 생성 안 함
- **Git 태그**: 태그 생성 안 함
- **GitHub API**: PR, Release 생성 안 함
- **PyPI 배포**: 패키지 배포 안 함

### 트러블슈팅

#### Q: Dry-Run에서 테스트가 실패했습니다

**A**: 실제 Dry-Run 전에 로컬에서 테스트를 수정하세요:
```bash
# 먼저 테스트 실행
pytest tests/

# 문제 해결 후
/awesome:release-new [version] --dry-run
```

#### Q: Dry-Run 결과가 이상합니다

**A**: 최신 코드와 의존성을 확인하세요:
```bash
# 최신 코드 pull
git pull origin main

# 의존성 업데이트
uv pip install -e .
uv pip install --no-deps -e .

# 다시 시도
/awesome:release-new [version] --dry-run
```

#### Q: 실제 릴리즈를 진행했는데 뭔가 빠진 것 같습니다

**A**: 다음번에는 꼭 Dry-Run을 먼저 실행하세요:
```bash
# 실패한 버전 다시 테스트
/awesome:release-new patch --dry-run

# 확인 후
/awesome:release-new patch
```

---

## 🔄 로컬 자동 업데이트 및 테스트

**로컬 자동 업데이트** 기능은 새로운 릴리즈가 PyPI에 배포되면, 로컬 개발 환경의 moai-adk 패키지를 자동으로 최신 버전으로 업데이트하고 테스트합니다.

### 언제 사용하는가?

1. **새로운 릴리즈 완료 후**: 실제 배포된 패키지가 설치 가능한지 검증
2. **자동화된 릴리즈 파이프라인**: GitHub Actions 완료 후 자동으로 로컬 테스트
3. **멀티 환경 테스트**: 다양한 Python 버전에서 패키지 호환성 확인

### 설정 방법

로컬 자동 업데이트는 다음 환경 변수로 제어됩니다:

```bash
# ~/.bashrc 또는 ~/.zshrc에 추가
export MOAI_AUTO_UPDATE=true          # 자동 업데이트 활성화
export MOAI_AUTO_UPDATE_TIMEOUT=300   # 업데이트 대기 시간 (초, 기본값 300)
export MOAI_AUTO_UPDATE_RETRY=10      # 재시도 횟수 (기본값 10)
export MOAI_PYTHON_VERSIONS="3.11 3.12 3.13"  # 테스트할 Python 버전
```

### 자동 업데이트 워크플로우

```bash
# 릴리즈 완료 (자동 업데이트 활성화 상태)
/awesome:release-new minor

→ Phase 0: 품질 검증
→ Phase 1: 버전 분석
→ Phase 2: PR 관리
→ Phase 3: GitHub Actions 자동 릴리즈

[PyPI 배포 완료 후]

→ Phase 4: 로컬 자동 업데이트 (신규!)
  ├─ PyPI에서 신규 버전 감지
  ├─ 로컬 가상환경에서 최신 버전 설치
  ├─ 기본 호환성 테스트 실행
  ├─ CLI 버전 검증
  └─ 테스트 결과 리포트

✅ 로컬 자동 업데이트 완료
```

### 자동 업데이트 실행 단계

#### Step 1: 로컬 버전 확인

```bash
# 현재 설치된 moai-adk 버전 확인
moai-adk --version

# 출력 예:
# moai-adk version 0.13.0
```

#### Step 2: PyPI에서 신규 버전 감지

```bash
# PyPI에서 최신 버전 확인 (최대 5분 대기)
pip index versions moai-adk

# 만약 최신 버전이 아직 감지되지 않으면 대기
# 기본값: 10회 재시도, 30초 간격
```

#### Step 3: 임시 가상환경에서 테스트 설치

```bash
# 임시 테스트 환경 생성
python -m venv /tmp/moai_test_$VERSION

# 최신 버전 설치 (캐시 제외)
source /tmp/moai_test_$VERSION/bin/activate
pip install moai-adk==$VERSION --no-cache-dir

# 설치 확인
moai-adk --version
```

#### Step 4: 호환성 검증

```bash
# moai-adk 기본 명령 테스트
moai-adk --help
moai-adk --version

# 프로젝트 초기화 시뮬레이션
moai-adk init --dry-run

# 임시 환경 정리
deactivate
rm -rf /tmp/moai_test_$VERSION
```

### 자동 업데이트 옵션

#### 옵션 1: 활성화/비활성화

```bash
# 자동 업데이트 활성화
export MOAI_AUTO_UPDATE=true
/awesome:release-new minor

# 자동 업데이트 비활성화 (기본값)
export MOAI_AUTO_UPDATE=false
/awesome:release-new minor
```

#### 옵션 2: 대기 시간 조정

```bash
# 5분(300초) 대기 (기본값)
export MOAI_AUTO_UPDATE_TIMEOUT=300

# 10분(600초) 대기
export MOAI_AUTO_UPDATE_TIMEOUT=600

# 즉시 테스트 (권장하지 않음, 보통 실패)
export MOAI_AUTO_UPDATE_TIMEOUT=0
```

#### 옵션 3: 재시도 횟수 조정

```bash
# 10회 재시도, 30초 간격 (기본값, 총 5분)
export MOAI_AUTO_UPDATE_RETRY=10

# 20회 재시도, 30초 간격 (총 10분)
export MOAI_AUTO_UPDATE_RETRY=20

# 재시도 없음 (1회만 시도)
export MOAI_AUTO_UPDATE_RETRY=1
```

#### 옵션 4: 다중 Python 버전 테스트

```bash
# Python 3.11, 3.12, 3.13에서 모두 테스트
export MOAI_PYTHON_VERSIONS="3.11 3.12 3.13"

# Python 3.13만 테스트
export MOAI_PYTHON_VERSIONS="3.13"

# 특정 경로의 Python 버전 테스트
export MOAI_PYTHON_VERSIONS="/usr/bin/python3.11 /usr/bin/python3.12"
```

### 자동 업데이트 실행 예시

#### 예시 1: 기본 설정으로 자동 업데이트

```bash
# ~/.bashrc에 설정
export MOAI_AUTO_UPDATE=true

# 릴리즈 실행
/awesome:release-new minor

# 출력 예:
# ...
# 🚀 PyPI 배포 완료!
# 🔄 로컬 자동 업데이트 시작...
# ⏳ PyPI CDN 대기 중... (1/10)
# ⏳ PyPI CDN 대기 중... (2/10)
# ✅ moai-adk 0.14.0 감지됨
# 📦 테스트 환경에서 설치 중...
# ✅ 설치 성공: moai-adk==0.14.0
# 🧪 호환성 테스트 중...
# ✅ 모든 테스트 통과
# 🎉 로컬 자동 업데이트 완료!
```

#### 예시 2: 멀티 버전 테스트

```bash
# ~/.bashrc에 설정
export MOAI_AUTO_UPDATE=true
export MOAI_PYTHON_VERSIONS="3.11 3.12 3.13"

# 릴리즈 실행
/awesome:release-new patch

# 출력 예:
# ...
# 🔄 Python 3.11에서 테스트 중...
# ✅ Python 3.11: 테스트 통과
# 🔄 Python 3.12에서 테스트 중...
# ✅ Python 3.12: 테스트 통과
# 🔄 Python 3.13에서 테스트 중...
# ✅ Python 3.13: 테스트 통과
# 🎉 모든 버전에서 테스트 통과!
```

### 자동 업데이트 트러블슈팅

#### Q: PyPI에서 새 버전이 감지되지 않습니다

**A**: PyPI CDN 전파 대기 시간 증가:

```bash
# 대기 시간을 10분으로 증가
export MOAI_AUTO_UPDATE_TIMEOUT=600
/awesome:release-new patch
```

#### Q: 호환성 테스트가 실패했습니다

**A**: 테스트 환경 로그 확인:

```bash
# 마지막 테스트 로그 확인
cat /tmp/moai_auto_update_*.log

# 문제 해결 후 수동 테스트
python -m venv /tmp/test_manual
source /tmp/test_manual/bin/activate
pip install moai-adk==[VERSION]
moai-adk --version
deactivate
rm -rf /tmp/test_manual
```

#### Q: 특정 Python 버전에서만 테스트하고 싶습니다

**A**: 환경 변수로 Python 버전 지정:

```bash
# Python 3.13만 테스트
export MOAI_PYTHON_VERSIONS="3.13"
/awesome:release-new minor

# 또는 절대 경로 사용
export MOAI_PYTHON_VERSIONS="/usr/local/bin/python3.13"
/awesome:release-new minor
```

#### Q: 자동 업데이트를 완전히 비활성화하려면?

**A**: 환경 변수 비활성화:

```bash
# 현재 세션에서만 비활성화
export MOAI_AUTO_UPDATE=false

# 또는 bashrc/zshrc에서 주석 처리
# export MOAI_AUTO_UPDATE=true
```

---

## 📚 참고 자료

- [Semantic Versioning](https://semver.org/)
- [PEP 621 (pyproject.toml)](https://peps.python.org/pep-0621/)
- [uv Documentation](https://github.com/astral-sh/uv)
- [gh CLI Docs](https://cli.github.com/manual/)

---

**cc-manager 에이전트를 통해 이 커맨드를 자동 실행하세요!**
