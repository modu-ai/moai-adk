# 🚀 릴리즈 환경 설정 가이드

> **로컬 전용 문서** - 패키지 배포에 포함되지 않음

MoAI-ADK 릴리즈 자동화를 위한 환경 설정 절차입니다.

## 📋 사전 요구사항

- Python 3.13+
- `uv` 패키지 매니저
- GitHub CLI (`gh`) 설치 및 인증
- PyPI 계정 (선택사항이지만 권장)
- TestPyPI 계정 (테스트 배포용)

## 🔐 1단계: PyPI 토큰 설정

### 1.1 PyPI 토큰 생성

**공식 PyPI** (프로덕션):
1. https://pypi.org/manage/account/token/ 방문
2. "Add Token" 클릭
3. 토큰 이름: `moai-adk-ci`
4. Scope: "Entire account (all projects)" 선택
5. 토큰 생성 (예: `pypi-AgEIcHlwaS5vcmcCJ...`)

**TestPyPI** (테스트용):
1. https://test.pypi.org/manage/account/token/ 방문
2. 동일한 절차로 토큰 생성
3. 토큰 복사 (예: `pypi-AgEIcHlwaS5wcmdjLmNvbQIkN...`)

### 1.2 환경 변수 설정

**방법 1: 임시 설정 (권장하지 않음)**
```bash
export UV_PUBLISH_TOKEN="pypi-AgEIcHlwaS5vcmcCJ..."
export UV_PUBLISH_TOKEN_TESTPYPI="pypi-AgEIcHlwaS5wcmdjLmNvbQIkN..."
```

**방법 2: 영구 설정 (권장)**

`~/.bashrc` 또는 `~/.zshrc`에 추가:
```bash
# PyPI 토큰 (프로덕션)
export UV_PUBLISH_TOKEN="pypi-AgEIcHlwaS5vcmcCJ..."

# TestPyPI 토큰 (테스트)
export UV_PUBLISH_TOKEN_TESTPYPI="pypi-AgEIcHlwaS5wcmdjLmNvbQIkN..."
```

그 후 셀 재시작:
```bash
source ~/.bashrc  # bash의 경우
source ~/.zshrc   # zsh의 경우
```

**방법 3: .pypirc 파일 (대체 방법)**

`~/.pypirc` 생성:
```ini
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
password = pypi-AgEIcHlwaS5wcmdjLmNvbQIkN...
```

파일 권한 설정:
```bash
chmod 600 ~/.pypirc
```

## 🔑 2단계: GitHub Secrets 설정 (CI/CD용)

GitHub Actions에서 PyPI 배포를 자동화하려면 GitHub Secrets에 토큰을 저장해야 합니다.

### 2.1 GitHub Secrets 추가

**저장소 설정 페이지**:
1. GitHub 저장소 → Settings → Secrets and variables → Actions
2. "New repository secret" 클릭
3. 다음 Secrets 생성:

| 이름 | 값 |
|------|-----|
| `PYPI_API_TOKEN` | PyPI 토큰 (pypi-...) |
| `TESTPYPI_API_TOKEN` | TestPyPI 토큰 (선택) |

### 2.2 Secrets 검증

```bash
# Secrets 확인 (로컬에서)
gh secret list -R modu-ai/moai-adk

# 출력 예:
# PYPI_API_TOKEN              Updated 2025-11-10
# TESTPYPI_API_TOKEN          Updated 2025-11-10
```

## ✅ 3단계: 설정 검증

### 3.1 로컬 토큰 검증

```bash
# 환경 변수 확인
echo $UV_PUBLISH_TOKEN
# 출력: pypi-AgEIcHlwaS5vcmcCJ...

echo $UV_PUBLISH_TOKEN_TESTPYPI
# 출력: pypi-AgEIcHlwaS5wcmdjLmNvbQIkN...
```

### 3.2 PyPI 연결 테스트

```bash
# TestPyPI에 테스트 빌드 업로드
python -m build
uv publish --publish-url https://test.pypi.org/legacy/ dist/moai_adk-*.whl --token $UV_PUBLISH_TOKEN_TESTPYPI

# 결과 확인
curl https://test.pypi.org/pypi/moai-adk/json | jq '.releases' | head -5
```

### 3.3 설치 테스트

```bash
# TestPyPI에서 설치
pip install --index-url https://test.pypi.org/simple/ moai-adk==0.22.5

# 버전 확인
moai-adk --version
```

## 🚀 4단계: 릴리즈 준비 확인

### 4.1 체크리스트

- [ ] PyPI 토큰 생성됨 (`pypi-...` 형식)
- [ ] 환경 변수 설정됨 (`$UV_PUBLISH_TOKEN` 존재)
- [ ] GitHub Secrets 설정됨 (`PYPI_API_TOKEN` 저장)
- [ ] TestPyPI 토큰 설정됨 (선택)
- [ ] 로컬 테스트 빌드 성공
- [ ] TestPyPI 업로드 성공

### 4.2 준비 완료

모든 설정이 완료되었습니다! 이제 다음 명령으로 릴리즈를 시작할 수 있습니다:

```bash
# 패치 버전 릴리즈
/moai:release patch

# 미리보기
/moai:release patch --dry-run
```

## ⚠️ 보안 주의사항

**절대 금지**:
- ❌ PyPI 토큰을 Git에 커밋하지 마세요
- ❌ 토큰을 채팅이나 이메일로 공유하지 마세요
- ❌ 토큰을 로그 파일에 기록하지 마세요
- ❌ 공개 저장소의 `.env` 파일에 저장하지 마세요

**권장 방법**:
- ✅ GitHub Secrets 사용 (CI/CD)
- ✅ `~/.bashrc` / `~/.zshrc` 사용 (로컬)
- ✅ 문자 외 토큰은 즉시 삭제

## 🆘 문제 해결

### 토큰 인증 실패

```bash
# 증상: "Authentication failed"
# 해결: 토큰 확인 및 재생성
echo $UV_PUBLISH_TOKEN | head -c 10  # "pypi-AgE..."로 시작하는지 확인

# 토큰 만료 시: PyPI에서 새 토큰 생성
```

### 권한 부족

```bash
# 증상: "You don't have permission to upload"
# 해결: PyPI 계정에 프로젝트 소유권 확인
# PyPI → Project → Collaborators → Invite user
```

### TestPyPI vs PyPI 혼동

```bash
# TestPyPI (테스트용)
pip install --index-url https://test.pypi.org/simple/ moai-adk

# PyPI (프로덕션)
pip install moai-adk
```

## 📚 참고 자료

- [PyPI Token Management](https://pypi.org/help/#apitoken)
- [TestPyPI](https://test.pypi.org/)
- [uv Publish Documentation](https://docs.astral.sh/uv/guides/publish/)
- [GitHub Secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets)

---

**다음 단계**: `/moai:release patch`를 실행하여 릴리즈를 시작하세요!
