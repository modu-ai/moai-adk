---
name: "Managing PyPI Authentication and Tokens"
description: "Configure and manage PyPI authentication tokens securely for package publishing. Use when setting up CI/CD environments, configuring local development, or rotating expired tokens. Covers token generation, environment variable setup, .pypirc configuration, and security best practices."
allowed-tools: "Bash(python:*), Read, Write, Edit"
---

# PyPI 인증 토큰 관리 | PyPI Authentication

> **사용 시기**: PyPI 배포 전 인증 설정, 토큰 갱신
> **When to use**: Setup PyPI authentication before `uv publish` or `twine upload`

---

## 🎯 목표 | Overview

안전한 PyPI 인증:
- ✅ PyPI API 토큰 생성
- ✅ 환경 변수 설정 (UV_PUBLISH_TOKEN)
- ✅ .pypirc 파일 설정 및 보안
- ✅ 토큰 만료 감지 및 갱신
- ✅ 다중 환경 관리 (로컬/CI/CD)

---

## 🔑 Step 1: PyPI 토큰 생성

### 1.1 PyPI 계정 접속

1. https://pypi.org/account/login/ 접속
2. 계정 로그인
3. Account Settings → API Tokens 이동

### 1.2 토큰 생성 옵션

**전체 접근 토큰** (모든 프로젝트):
```
Token name: "moai-adk-publish"
Scope: Entire account (모든 프로젝트)
```

**프로젝트 제한 토큰** (권장):
```
Token name: "moai-adk-publish"
Scope: Project-specific (moai-adk 프로젝트만)
```

### 1.3 토큰 저장

```
⚠️ 중요: 토큰을 즉시 안전한 장소에 복사하세요!
          페이지를 벗어나면 다시 확인할 수 없습니다.

형식: pypi-AgEIcHlwaS5vcmcCJ...
```

---

## 💾 Step 2: 인증 방법 선택

### 방법 A: 환경 변수 (권장 - 임시)

```bash
# 일회성 사용 (가장 안전)
UV_PUBLISH_TOKEN="pypi-AgEIcHlwaS5vcmcCJ..." uv publish

# 또는 세션 동안 유지
export UV_PUBLISH_TOKEN="pypi-AgEIcHlwaS5vcmcCJ..."
uv publish
```

**장점**:
- ✅ 파일에 저장하지 않음
- ✅ 임시 사용에 적합
- ✅ CI/CD 환경 안전

**단점**:
- ❌ 매번 설정 필요
- ❌ 장기 저장 어려움

---

### 방법 B: .pypirc 파일 (영구 설정)

#### 2.1 .pypirc 파일 생성

```bash
# $HOME/.pypirc 파일 생성
cat > ~/.pypirc <<'EOF'
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
  password = pypi-test-AgEIcHlwaS5vcmcCJ...
EOF
```

#### 2.2 보안 설정

```bash
# 파일 권한 제한 (매우 중요!)
chmod 600 ~/.pypirc

# 확인
ls -la ~/.pypirc
# 예상: -rw------- (600)
```

#### 2.3 파일 구조 상세

```ini
[distutils]
  # 기본 저장소 목록
  index-servers =
    pypi          # 프로덕션 (https://pypi.org)
    testpypi      # 테스트 (https://test.pypi.org)

[pypi]
  # 프로덕션 저장소
  repository = https://upload.pypi.org/legacy/
  username = __token__          # 고정값 (변경 금지!)
  password = pypi-AgEI...       # 실제 토큰

[testpypi]
  # 테스트 저장소 (선택사항)
  repository = https://test.pypi.org/legacy/
  username = __token__
  password = pypi-test-AgEI...
```

---

## 🔄 Step 3: 다양한 환경 설정

### 로컬 개발 환경

```bash
# Option 1: 환경 변수 추가 (임시)
export UV_PUBLISH_TOKEN="pypi-..."
echo "✅ 현재 세션에만 적용"

# Option 2: .zshrc/.bashrc에 추가 (영구)
echo 'export UV_PUBLISH_TOKEN="pypi-..."' >> ~/.zshrc
source ~/.zshrc
echo "✅ 모든 새 터미널에 적용"
```

### GitHub Actions CI/CD

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Publish to PyPI
        run: uv publish
        env:
          UV_PUBLISH_TOKEN: ${{ secrets.PYPI_TOKEN }}
```

**GitHub 설정**:
1. Repository → Settings → Secrets and variables → Actions
2. "New repository secret" 클릭
3. Name: `PYPI_TOKEN`
4. Value: `pypi-AgEI...` (토큰 값)

### GitLab CI/CD

```yaml
# .gitlab-ci.yml
publish_pypi:
  stage: release
  script:
    - uv publish
  only:
    - tags
  variables:
    UV_PUBLISH_TOKEN: $PYPI_TOKEN
```

---

## ✅ Step 4: 인증 검증

### 인증 상태 확인

```bash
# uv 설정 확인 (토큰 마스킹됨)
uv publish --help | grep -A 5 "authentication"

# 또는 직접 테스트
uv publish --dry-run  # 실제 배포하지 않고 테스트
```

### PyPI 연결 테스트

```bash
# PyPI 서버 연결 확인
curl -I https://upload.pypi.org/legacy/

# HTTP 200 또는 403 (인증 필요)
# 아무 응답 없으면 네트워크 문제
```

---

## 🔄 Step 5: 토큰 갱신

### 토큰 만료 감지

```
❌ 오류: "401 Unauthorized"
   또는 "Invalid or non-existent authentication"
```

### 갱신 절차

1. **PyPI 웹사이트에서 새 토큰 생성**
   ```
   https://pypi.org/manage/account/token/
   ```

2. **환경 변수 갱신**
   ```bash
   # ~/.zshrc 또는 ~/.bashrc 수정
   export UV_PUBLISH_TOKEN="pypi-new-token..."
   source ~/.zshrc
   ```

3. **.pypirc 파일 갱신**
   ```bash
   # ~/.pypirc 편집
   nano ~/.pypirc
   # password = pypi-new-token... 로 변경
   chmod 600 ~/.pypirc
   ```

4. **CI/CD 시크릿 갱신**
   ```
   GitHub/GitLab Settings → Update PYPI_TOKEN
   ```

### 만료된 토큰 삭제 (권장)

```
PyPI 웹사이트 → Account Settings → API Tokens
→ 해당 토큰의 "Revoke" 클릭
```

---

## 🛡️ 보안 모범 사례 | Security Best Practices

| 관행 | 권장 | 이유 |
|------|------|------|
| 토큰 저장 위치 | 환경 변수/CI 시크릿 | 파일 시스템 노출 방지 |
| 파일 권한 | 600 (-rw-------) | 다른 사용자 접근 차단 |
| 토큰 범위 | 프로젝트별 | 피해 최소화 |
| 토큰 갱신 | 분기별 | 보안 위험 감소 |
| 로그 기록 | 마스킹 필수 | 실수로 노출 방지 |

### .pypirc 파일 보안 체크리스트

```bash
#!/bin/bash
echo "🔒 .pypirc 보안 검사..."

# 1. 파일 존재 확인
if [ ! -f ~/.pypirc ]; then
    echo "⚠️  ~/.pypirc 없음"
    exit 0
fi

# 2. 권한 확인 (600이어야 함)
perms=$(stat -f "%A" ~/.pypirc 2>/dev/null || stat -c "%a" ~/.pypirc)
if [ "$perms" != "600" ]; then
    echo "❌ 권한 문제: $perms (예상: 600)"
    echo "→ 해결: chmod 600 ~/.pypirc"
    exit 1
fi

# 3. 토큰 내용 검증
if ! grep -q "^  password = pypi-" ~/.pypirc; then
    echo "⚠️  토큰 형식 이상 (pypi- 접두사 확인)"
    exit 1
fi

echo "✅ .pypirc 보안 검사 완료"
```

---

## 🚀 통합 예제: 완전한 배포 워크플로우

```bash
#!/bin/bash
set -euo pipefail

# 1. 토큰 확인
if [ -z "${UV_PUBLISH_TOKEN:-}" ] && [ ! -f ~/.pypirc ]; then
    echo "❌ PyPI 인증 구성되지 않음"
    echo "→ 해결:"
    echo "   1. export UV_PUBLISH_TOKEN='pypi-...' 또는"
    echo "   2. ~/.pypirc 파일 생성 (chmod 600)"
    exit 1
fi

# 2. 빌드
echo "📦 빌드 중..."
uv build

# 3. 테스트 배포 (선택)
if [ -n "${TEST_PYPI:-}" ]; then
    echo "🧪 Test PyPI에 배포 중..."
    uv publish -r testpypi
fi

# 4. 프로덕션 배포
echo "🚀 PyPI에 배포 중..."
uv publish

echo "✅ 배포 완료!"
```

---

## 📚 참고 자료

- [PyPI Uploading Projects](https://packaging.python.org/guides/publishing-package-distribution-releases-to-pypi/)
- [PyPI API Documentation](https://warehouse.pypa.io/api-reference/index.html)
- [uv Publisher Configuration](https://docs.astral.sh/uv/guides/publish/)
- [Python Keyring (고급)](https://keyring.readthedocs.io/)

---

## ⚠️ 일반적인 오류 | Common Errors

| 오류 | 원인 | 해결 방법 |
|------|------|---------|
| 401 Unauthorized | 토큰 만료/잘못됨 | 새 토큰 생성 |
| 403 Forbidden | 토큰 범위 제한 | 전체 접근 토큰 생성 |
| 400 Bad Request | 패키지 메타데이터 오류 | pyproject.toml 검증 |
| Connection refused | 네트워크/방화벽 | VPN/프록시 확인 |

**응급 문제 해결**: `uv publish --dry-run` 먼저 시도

---

**보안 팁**: 토큰은 절대 버전 컨트롤에 커밋하지 마세요! 🔐
