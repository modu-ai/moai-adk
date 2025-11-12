# ✅ 릴리즈 검증 체크리스트

> **로컬 전용 문서** - 패키지 배포에 포함되지 않음

`/moai:release` 명령 실행 전에 이 체크리스트를 완료하세요.

---

## 📋 Phase 1: 코드 검증 (30분)

### 1.1 Git 상태 확인

- [ ] `git status` 실행 - 커밋되지 않은 변경사항 없음
- [ ] `git log main -5` 확인 - develop에서 merge 완료
- [ ] `git branch -a` 확인 - 불필요한 feature 브랜치 없음
- [ ] `.gitignore` 검토 - 민감한 파일 보호됨

```bash
# 실행 명령어
git status
git log main -5
git branch -a | grep feature
cat .gitignore | grep -E "\.env|\.vercel|secrets"
```

### 1.2 테스트 검증

- [ ] `pytest tests/` 실행 - 모든 테스트 통과
- [ ] 테스트 커버리지 85% 이상 확인
- [ ] 통합 테스트 통과
- [ ] 회귀 테스트 통과

```bash
# 실행 명령어
pytest tests/ -v
pytest --cov=src/moai_adk tests/ --cov-report=term-missing | tail -10
pytest tests/integration/ -v
```

### 1.3 타입 및 스타일 검증

- [ ] `mypy` 검사 완료 (0 오류)
- [ ] `ruff check` 검사 완료 (0 오류)
- [ ] `black` 포맷팅 검사 완료 (0 변경)
- [ ] 불필요한 import 제거됨

```bash
# 실행 명령어
mypy src/moai_adk --strict
ruff check src/ tests/ --select=E,W,F
black --check src/ tests/
python -m py_compile src/moai_adk/**/*.py
```

### 1.4 보안 검증

- [ ] `bandit` 보안 스캔 완료 (CRITICAL 없음)
- [ ] `pip-audit` 취약점 검사 완료
- [ ] 의존성 버전 검토 (최신 안정 버전)
- [ ] 민감한 정보 하드코딩 검사 (grep 결과 0)

```bash
# 실행 명령어
bandit -r src/moai_adk -f json | jq '.metrics | select(._totals.CRITICAL > 0)'
pip-audit -r pyproject.toml
cat pyproject.toml | grep -E "dependencies|optional-dependencies"
grep -r "password\|api_key\|secret" src/ --exclude-dir=__pycache__
```

---

## 📦 Phase 2: 버전 및 문서 (30분)

### 2.1 버전 확인

- [ ] `pyproject.toml` 버전 업데이트됨 (올바른 semver)
- [ ] `src/moai_adk/__init__.py` 버전 동기화됨
- [ ] CHANGELOG.md 버전 항목 추가됨
- [ ] git 태그 미리 생성 안 됨 (자동 생성 예정)

```bash
# 실행 명령어
grep 'version =' pyproject.toml
grep '__version__' src/moai_adk/__init__.py
head -20 CHANGELOG.md
git tag | tail -5
```

### 2.2 문서 검증

- [ ] README.md 최신 상태 (예: 설치 명령어 정확)
- [ ] CONTRIBUTING.md 개발 가이드 최신화
- [ ] API 문서 생성됨 (해당하는 경우)
- [ ] 예제 코드 동작 검증

```bash
# 실행 명령어
head -50 README.md
ls -la docs/ 2>/dev/null || echo "No docs/ directory"
grep -A 5 "## Installation" README.md
```

### 2.3 CHANGELOG 검증

- [ ] 이번 릴리스 모든 변경사항 포함
- [ ] 시간순 정렬됨 (최신이 맨 위)
- [ ] 포맷 일관성 (링크, 강조 등)
- [ ] Breaking Changes 명확히 표시

```bash
# 실행 명령어
head -50 CHANGELOG.md
diff -u CHANGELOG.md.backup CHANGELOG.md 2>/dev/null || echo "First check"
```

### 2.4 배포 문서 검증

- [ ] RELEASE_SETUP.md 현재 상태 유지
- [ ] ROLLBACK_GUIDE.md 접근 가능
- [ ] 이 체크리스트 최신 상태

```bash
# 실행 명령어
ls -la .moai/release/
cat .moai/release/RELEASE_SETUP.md | wc -l
```

---

## 🔍 Phase 3: 의존성 검증 (20분)

### 3.1 설치 가능성 확인

- [ ] 새 환경에서 `pip install .` 성공
- [ ] 모든 필수 의존성 설치됨
- [ ] 선택 의존성 표시 명확

```bash
# 실행 명령어 (별도 venv)
python -m venv /tmp/test_install
source /tmp/test_install/bin/activate
pip install .
moai-adk --version
deactivate
rm -rf /tmp/test_install
```

### 3.2 의존성 호환성

- [ ] Python 최소 버전 충족 (3.9+)
- [ ] 주요 의존성 버전 호환성 검증
- [ ] 선택 의존성 설치 테스트

```bash
# 실행 명령어
grep 'python =' pyproject.toml
pip list | grep -E "^(FastAPI|SQLAlchemy|Pydantic|Django)"
python --version
```

### 3.3 보안 업데이트 확인

- [ ] 의존성 보안 패치 적용됨
- [ ] pip-audit 취약점 0개

```bash
# 실행 명령어
pip-audit
pip index versions moai-adk 2>/dev/null | head -5
```

---

## 🚀 Phase 4: 배포 사전 검증 (30분)

### 4.1 TestPyPI 배포 테스트

- [ ] TestPyPI에 배포 성공
- [ ] TestPyPI에서 설치 성공
- [ ] 설치 후 기본 기능 동작 확인

```bash
# 실행 명령어
python -m build  # sdist + wheel 생성
twine check dist/*
twine upload dist/* -r testpypi  # 또는 /moai:release patch --dry-run

# 검증
pip install --index-url https://test.pypi.org/simple/ moai-adk==<VERSION>
moai-adk --version
moai-adk --help
```

### 4.2 빌드 산출물 검증

- [ ] wheel 파일 생성됨 (*.whl)
- [ ] sdist 파일 생성됨 (*.tar.gz)
- [ ] 파일 크기 합리적 (> 1MB 경고)
- [ ] 파일 내용 검증

```bash
# 실행 명령어
ls -lh dist/
unzip -l dist/moai_adk*.whl | head -20
tar -tzf dist/moai_adk*.tar.gz | head -20
```

### 4.3 메타데이터 검증

- [ ] `twine check` 통과 (warnings 0)
- [ ] PyPI 렌더링 정상 (테스트 배포에서 확인)
- [ ] 라이선스 정보 포함
- [ ] 프로젝트 URL 유효

```bash
# 실행 명령어
twine check dist/*
cat pyproject.toml | grep -A 5 '\[project\]'
```

---

## 🔐 Phase 5: 배포 환경 검증 (20분)

### 5.1 PyPI 인증 검증

- [ ] PyPI 토큰 설정됨 (`$UV_PUBLISH_TOKEN` 확인)
- [ ] GitHub Secrets 설정됨 (`PYPI_API_TOKEN`)
- [ ] 토큰 유효 기한 확인 (6개월 이상)

```bash
# 실행 명령어
echo "Token set: $([ -z "$UV_PUBLISH_TOKEN" ] && echo 'NO' || echo 'YES')"
gh secret list | grep PYPI_API_TOKEN
# 토큰 생성 날짜는 PyPI 계정 페이지에서 확인
```

### 5.2 Git 구성 검증

- [ ] Git author 설정됨 (`git config user.name`)
- [ ] Git email 설정됨 (`git config user.email`)
- [ ] GPG signing 필요한 경우 설정됨

```bash
# 실행 명령어
git config user.name
git config user.email
git config user.signingkey || echo "GPG signing not configured"
```

### 5.3 GitHub 인증 검증

- [ ] GitHub CLI 로그인됨 (`gh auth status`)
- [ ] 저장소 접근 권한 있음
- [ ] 릴리스 권한 있음 (maintainer 이상)

```bash
# 실행 명령어
gh auth status
gh repo view modu-ai/moai-adk --json nameWithOwner
gh release list -R modu-ai/moai-adk -L 1
```

---

## 📊 Phase 6: 최종 검증 (10분)

### 6.1 배포 명령어 확인

- [ ] 배포 타입 선택됨 (patch/minor/major)
- [ ] --dry-run으로 한 번 검증 완료
- [ ] 실제 배포 준비 완료

```bash
# 실행 명령어
/moai:release patch --dry-run  # 검증
# 출력 검토 후 진행

/moai:release patch  # 실제 배포
```

### 6.2 배포 후 검증

- [ ] PyPI에서 새 버전 확인 (5분 소요)
- [ ] 설치 가능 확인 (`pip install moai-adk`)
- [ ] GitHub Release 생성됨
- [ ] git 태그 생성됨

```bash
# 실행 명령어 (배포 후)
pip search moai-adk 2>/dev/null || curl https://pypi.org/pypi/moai-adk/json | jq '.info.version'
gh release list -R modu-ai/moai-adk -L 1
git tag | grep "^v" | sort -V | tail -1
```

### 6.3 사용자 공지

- [ ] 릴리스 노트 작성 완료
- [ ] GitHub Releases 페이지 업데이트
- [ ] 커뮤니티 공지 (슬랙, 이메일 등)
- [ ] 문서 사이트 업데이트 (있는 경우)

```bash
# 실행 명령어
gh release view $(git tag | grep "^v" | sort -V | tail -1)
```

---

## 🎯 빠른 검증 스크립트

### 전체 검증 한 번에 실행

```bash
#!/bin/bash
# validate-release.sh - 모든 검증을 한 번에 실행

set -e

echo "📋 릴리즈 검증 시작..."

echo "1️⃣  Git 상태 확인"
git status
git log main -1 --oneline

echo "2️⃣  테스트 실행"
pytest tests/ -q
pytest --cov=src/moai_adk tests/ --cov-report=term-missing | grep -E "^(TOTAL|moai_adk)"

echo "3️⃣  코드 품질 검사"
mypy src/moai_adk --strict --no-error-summary 2>&1 | tail -1
ruff check src/ --select=E,W,F --quiet

echo "4️⃣  보안 검사"
bandit -r src/moai_adk -q -ll

echo "5️⃣  버전 확인"
VERSION=$(grep 'version =' pyproject.toml | cut -d'"' -f2)
echo "배포 버전: $VERSION"

echo "6️⃣  빌드 검증"
python -m build
twine check dist/* -q

echo ""
echo "✅ 모든 검증 완료! 배포 준비됨"
echo ""
echo "다음 단계:"
echo "  /moai:release patch --dry-run  # 건조 실행 (확인)"
echo "  /moai:release patch             # 실제 배포"
```

---

## 📋 배포 전 최종 확인

### 배포 직전 체크리스트 (5분)

```
최종 확인 전 이 항목들을 다시 한 번 점검하세요:

[ ] Git: main 브랜치에 모든 변경사항 merge됨
[ ] 테스트: 모든 테스트 통과 (커버리지 85%+)
[ ] 보안: bandit 실행 (CRITICAL 0개)
[ ] 문서: CHANGELOG.md 업데이트됨
[ ] 버전: pyproject.toml 버전 업데이트됨
[ ] TestPyPI: 테스트 배포 성공
[ ] 토큰: PyPI 토큰 설정됨
[ ] GitHub: 로그인 상태 확인

모든 항목 확인 후 배포 시작하세요!
```

---

## 📞 문제 발생 시

| 문제 | 해결 방법 |
|------|----------|
| **테스트 실패** | `pytest tests/ -v` 실행, 실패 테스트 수정 |
| **타입 오류** | `mypy src/moai_adk --show-error-codes` 확인, 타입 주석 추가 |
| **PyPI 인증 실패** | 토큰 유효 기한 확인, 새 토큰 발급 |
| **빌드 실패** | `python -m build --verbose` 실행, 오류 메시지 확인 |
| **배포 후 설치 실패** | TestPyPI에서 재검증, 문제 수정 후 롤백 |

자세한 내용: [ROLLBACK_GUIDE.md](./ROLLBACK_GUIDE.md)

---

**마지막 업데이트**: 2025-11-12
**작성자**: Alfred Release Manager
