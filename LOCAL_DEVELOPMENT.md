# MoAI-ADK 로컬 개발 & 배포 가이드

## 🚀 로컬 개발 환경 구성

### 1️⃣ 저장소 클론 & 기본 설정

```bash
# 저장소 클론
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk

# Python 가상환경 생성 (권장)
python -m venv venv
source venv/bin/activate  # macOS/Linux
# venv\Scripts\activate   # Windows
```

### 2️⃣ 개발 모드 설치

```bash
# 개발 의존성과 함께 설치
pip install -e .

# 또는 uv 사용 (10-100x 더 빠름)
uv pip install -e .

# 설치 확인
moai --version
which moai
```

**예상 출력:**
```
MoAI-ADK v0.1.9
/path/to/your/venv/bin/moai
```

### 3️⃣ 의존성 설치

```bash
# 기본 런타임 의존성 (자동으로 설치됨)
# - colorama>=0.4.6 (크로스 플랫폼 콘솔)
# - click>=8.0.0 (CLI 프레임워크)
# - gitpython>=3.1.0 (Git 조작)
# - jinja2>=3.0.0 (템플릿 엔진)
# - jsonschema>=4.0.0 (스키마 검증)
# - pyyaml>=6.0 (YAML 파싱)
# - toml>=0.10.0 (TOML 파싱)
# - watchdog>=3.0.0 (파일 감시)

# 추가 개발 도구 설치 (선택사항)
pip install pytest pytest-cov black isort mypy flake8
```

## 🔧 로컬 빌드 & 패키지 생성

### 1️⃣ 빌드 시스템 실행

```bash
# Python 통합 빌드 도구 사용 (권장)
python scripts/build.py

# 또는 직접 빌드
python -m build --wheel
```

**빌드 결과 확인:**
```bash
ls -la dist/
# moai_adk-0.1.9-py3-none-any.whl (319KB)
# moai_adk-0.1.9.tar.gz (262KB)
```

### 2️⃣ 로컬 설치 테스트

```bash
# 새 가상환경에서 테스트
python -m venv test-env
source test-env/bin/activate

# wheel 파일로 설치
pip install dist/moai_adk-0.1.9-py3-none-any.whl

# 설치 확인
moai --version
moai --help
```

## 🧪 테스트 & 품질 검증

### 1️⃣ 종합 테스트 실행

```bash
# Python 테스트 러너 사용
python scripts/test_runner.py

# 또는 직접 pytest 실행
pytest tests/ -v --cov=src/moai_adk
```

### 2️⃣ 품질 검사

```bash
# TRUST 원칙 준수 검사
python .moai/scripts/check_constitution.py

# TAG 추적성 검사
python .moai/scripts/check-traceability.py --update

# 보안 검사
python scripts/check_secrets.py
```

### 3️⃣ 코드 포매팅

```bash
# 자동 포매팅
black src/ tests/
isort src/ tests/

# 타입 검사
mypy src/moai_adk

# 린터 검사
flake8 src/ tests/
```

## 📦 배포 프로세스

### 1️⃣ 버전 관리

```bash
# 현재 버전 상태 확인
python scripts/version_manager.py status

# 버전 업데이트 (선택사항)
# pyproject.toml에서 version = "0.1.10" 수정 후
python scripts/version_manager.py sync
```

### 2️⃣ 프로덕션 빌드

```bash
# 클린 빌드
rm -rf dist/ build/ src/*.egg-info/
python -m build

# 빌드 검증
python -m twine check dist/*
```

### 3️⃣ 로컬 TestPyPI 업로드 (테스트)

```bash
# TestPyPI 업로드 (테스트용)
python -m twine upload --repository testpypi dist/*

# TestPyPI에서 설치 테스트
pip install -i https://test.pypi.org/simple/ moai-adk==0.1.9
```

### 4️⃣ 실제 PyPI 배포 (주의!)

```bash
# 실제 PyPI 업로드 (신중하게!)
python -m twine upload dist/*

# 설치 확인
pip install moai-adk==0.1.9
```

## 🗂️ 프로젝트 구조 이해

```
MoAI-ADK/
├── src/moai_adk/           # 메인 패키지
│   ├── cli/                # CLI 명령어 (7개 모듈)
│   ├── core/              # 핵심 엔진 (33개 모듈)
│   │   ├── docs/          # 문서 자동화 시스템
│   │   ├── quality/       # 품질 검증 시스템
│   │   └── tag_system/    # TAG 추적 시스템
│   ├── install/           # 설치 관리 (5개 모듈)
│   ├── utils/             # 공통 유틸리티 (3개 모듈)
│   └── resources/         # 템플릿 & 스크립트
├── scripts/               # 빌드/테스트 자동화 도구
├── tests/                 # 테스트 슈트
├── dist/                  # 빌드 결과물
├── pyproject.toml         # 패키지 메타데이터
└── LOCAL_DEVELOPMENT.md   # 이 문서
```

## 🔄 개발 워크플로우

### 1️⃣ 기능 개발 사이클

```bash
# 1. 새 브랜치 생성
git checkout -b feature/new-feature

# 2. 코드 작성 & 테스트
# (개발 작업)

# 3. 품질 검증
python scripts/test_runner.py
python .moai/scripts/check_constitution.py

# 4. 커밋 & 푸시
git add .
git commit -m "feat: add new feature"
git push origin feature/new-feature

# 5. PR 생성 & 리뷰
```

### 2️⃣ 릴리스 프로세스

```bash
# 1. 버전 업데이트
# pyproject.toml에서 version 수정

# 2. 체인지로그 업데이트
# CHANGELOG.md 업데이트

# 3. 빌드 & 테스트
python scripts/build.py
python scripts/test_runner.py

# 4. 태그 생성
git tag v0.1.10
git push origin v0.1.10

# 5. 배포 (GitHub Actions 자동화)
```

## ⚡ 빠른 개발 팁

### 1️⃣ 자주 사용하는 명령어

```bash
# 개발 모드 재설치
pip install -e . --force-reinstall

# 특정 테스트만 실행
pytest tests/test_cli.py::test_version -v

# 커버리지 리포트 생성
pytest --cov=src/moai_adk --cov-report=html
```

### 2️⃣ IDE 설정

**VS Code 설정 (.vscode/settings.json):**
```json
{
    "python.defaultInterpreterPath": "./venv/bin/python",
    "python.formatting.provider": "black",
    "python.linting.enabled": true,
    "python.linting.flake8Enabled": true,
    "python.testing.pytestEnabled": true
}
```

### 3️⃣ 디버깅

```bash
# 상세 로그와 함께 실행
MOAI_DEBUG=1 moai init test-project

# 특정 모듈 디버깅
python -m pdb -c continue -m moai_adk.cli.commands init test-project
```

## 🚨 문제 해결

### 일반적인 이슈들

**1. 설치 실패:**
```bash
# 캐시 클리어 후 재설치
pip cache purge
pip install -e . --no-cache-dir
```

**2. 권한 에러:**
```bash
# 사용자 모드 설치
pip install -e . --user
```

**3. 의존성 충돌:**
```bash
# 새 가상환경 생성
rm -rf venv/
python -m venv venv
source venv/bin/activate
pip install -e .
```

**4. 빌드 실패:**
```bash
# 빌드 도구 업데이트
pip install --upgrade build setuptools wheel
```

## 📊 성능 벤치마크

**빌드 시간:**
- Wheel 생성: ~30초
- 전체 테스트: ~45초
- 품질 검사: ~15초

**패키지 크기:**
- Wheel: 319KB
- Source: 262KB
- 설치 후: ~1.2MB

---

## 🤝 기여 가이드라인

1. **Issue 생성**: 새 기능이나 버그는 먼저 Issue로 논의
2. **브랜치 명명**: `feature/`, `bugfix/`, `docs/` 접두사 사용
3. **커밋 메시지**: Conventional Commits 형식 준수
4. **테스트**: 새 코드는 반드시 테스트 포함
5. **문서**: 공개 API 변경 시 문서 업데이트

**Made with ❤️ for Claude Code Developers**