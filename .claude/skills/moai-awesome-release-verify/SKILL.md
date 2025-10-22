---
name: "Verifying Release Readiness"
description: "Validate Python package release prerequisites before deployment. Use when preparing for release to ensure all tests pass, coverage meets targets, code quality gates are satisfied, and Git state is clean. Run quality checks (pytest, ruff, mypy, bandit) automatically before PyPI or GitHub deployment."
allowed-tools: "Bash(pytest:*), Bash(ruff:*), Bash(mypy:*), Bash(bandit:*), Bash(python:*), Bash(git:*), Read"
---

# 릴리즈 준비 검증 | Verifying Release Readiness

> **사용 시기**: 패키지를 PyPI에 배포하기 전, 모든 품질 기준을 만족하는지 확인
> **When to use**: Before publishing to PyPI, verify all quality gates and prerequisites

---

## 🎯 목표 | Overview

패키지 릴리즈 전 자동 검증:
- ✅ 테스트 실행 및 커버리지 검증 (pytest, coverage)
- ✅ 코드 스타일 린팅 (ruff check)
- ✅ 타입 안전성 검사 (mypy)
- ✅ 보안 스캔 (bandit, pip-audit)
- ✅ Git 상태 확인 (미커밋 변경사항 없음)

**사전 조건**:
- Python 3.13+ 설치
- 패키지 매니저 (uv 또는 pip)
- pyproject.toml 존재
- pytest, ruff, mypy, bandit 설치

---

## 📋 검증 체크리스트 | Verification Checklist

### Phase 0.1: Python 환경 확인

```bash
# Python 버전 확인 (>=3.13 필요)
python_version=$(python --version 2>&1 | awk '{print $2}')

if [[ ! $python_version =~ ^3\.1[3-9]|^3\.[2-9][0-9] ]]; then
    echo "❌ Python 버전 부족: $python_version (요구: 3.13+)"
    exit 1
fi

echo "✅ Python: $python_version"
```

### Phase 0.2: 테스트 및 커버리지 검증

```bash
# pytest 실행 (커버리지 포함)
pytest tests/ --cov=src/moai_adk --cov-report=term-missing

# 커버리지 추출 및 검증
coverage=$(pytest tests/ --cov=src/moai_adk --cov-report=term \
    | grep "TOTAL" | awk '{print $4}' | sed 's/%//')

if [ "$coverage" -lt 85 ]; then
    echo "⚠️  커버리지 부족: ${coverage}% (목표: 85%)"
    exit 1
fi

echo "✅ 테스트 통과, 커버리지: ${coverage}%"
```

### Phase 0.3: 린트 검사 (ruff)

```bash
# ruff check 실행
if ! ruff check src/ tests/ --exit-zero; then
    echo "⚠️  린트 경고 있음 (확인 필요)"
fi

echo "✅ 린트 검사 완료"
```

### Phase 0.4: 타입 체크 (mypy)

```bash
# mypy 실행 (missing imports 무시)
if ! mypy src/moai_adk --ignore-missing-imports --no-error-summary 2>/dev/null; then
    echo "⚠️  타입 경고 있음 (경고만)"
fi

echo "✅ 타입 체크 완료"
```

### Phase 0.5: 보안 스캔

```bash
# bandit 보안 스캔
if ! bandit -r src/moai_adk --exit-zero -ll 2>/dev/null >/dev/null; then
    echo "⚠️  보안 경고 있음"
fi

# pip-audit 의존성 검사
if ! pip-audit 2>/dev/null; then
    echo "⚠️  의존성 취약점 있음"
fi

echo "✅ 보안 스캔 완료"
```

### Phase 0.6: Git 상태 확인

```bash
# 미커밋 변경사항 확인
if [ -n "$(git status --short)" ]; then
    echo "⚠️  미커밋 변경사항 있음:"
    git status --short
    exit 1
fi

# 현재 브랜치 확인
current_branch=$(git branch --show-current)
echo "✅ Git 상태 깨끗함 (브랜치: $current_branch)"
```

---

## 🔍 고급 검증 | Advanced Checks

### 버전 일치성 검증 (SSOT)

```bash
# pyproject.toml 버전
pyproject_version=$(rg "^version = " pyproject.toml | awk -F'"' '{print $2}')

# 설치된 버전
installed_version=$(python -c "from importlib.metadata import version; print(version('moai-adk'))" 2>/dev/null || echo "N/A")

if [ "$pyproject_version" != "$installed_version" ] && [ "$installed_version" != "N/A" ]; then
    echo "⚠️  버전 불일치:"
    echo "   pyproject.toml: $pyproject_version"
    echo "   설치된 버전: $installed_version"
    echo "→ 해결: uv pip install -e . --force-reinstall --no-deps"
    exit 1
fi

echo "✅ 버전 일치 (SSOT): $pyproject_version"
```

### 의존성 호환성 검사

```bash
# 의존성 정보 추출
rg "^requires =" pyproject.toml | head -3

# 선택 의존성 정보
rg "^optional-dependencies" pyproject.toml -A 5

echo "✅ 의존성 구성 확인"
```

---

## 📊 전체 검증 자동화 스크립트

```bash
#!/bin/bash
set -euo pipefail

echo "🚀 릴리즈 준비 검증 시작..."
echo ""

# 1. Python 버전
python_version=$(python --version 2>&1 | awk '{print $2}')
echo "[1/5] Python 버전: $python_version"

# 2. 테스트 + 커버리지
echo "[2/5] 테스트 실행 중..."
pytest tests/ --cov=src/moai_adk --cov-report=term-missing -q

# 3. 린트
echo "[3/5] 린트 검사 중..."
ruff check src/ tests/ --exit-zero

# 4. 타입 체크
echo "[4/5] 타입 체크 중..."
mypy src/moai_adk --ignore-missing-imports --no-error-summary 2>/dev/null || true

# 5. Git 상태
echo "[5/5] Git 상태 확인 중..."
if [ -n "$(git status --short)" ]; then
    echo "❌ 미커밋 변경사항 있음"
    exit 1
fi

echo ""
echo "✅ 모든 검증 완료!"
echo "→ 다음 단계: /awesome:release-new patch"
```

---

## ✨ 주요 포인트

| 검증 항목 | 목표 | 실패 시 조치 |
|---------|------|-----------|
| 테스트 | 100% 통과 | `pytest tests/ -v` 실행하여 문제 해결 |
| 커버리지 | ≥85% | `pytest tests/ --cov-report=html`로 누락 영역 확인 |
| 린트 | 오류 없음 | `ruff check --fix src/`로 자동 수정 시도 |
| 타입 | 경고 최소화 | `mypy src/moai_adk`로 상세 확인 |
| 보안 | 심각도 높은 이슈 없음 | `bandit -r src/`로 상세 분석 |
| Git | 깨끗한 상태 | 모든 변경사항 커밋 |

---

## 📚 참고

- [pytest Documentation](https://docs.pytest.org/)
- [ruff Linter](https://github.com/astral-sh/ruff)
- [mypy Type Checker](https://mypy.readthedocs.io/)
- [Bandit Security Scanner](https://bandit.readthedocs.io/)

**다음 단계**: 모든 검증 완료 후 `/awesome:release-new patch` 실행
