---
title: 개발 환경 설정
description: MoAI-ADK 로컬 개발 환경 구성 및 기여 가이드
status: stable
---

# 개발 환경 설정

MoAI-ADK에 기여하기 위한 로컬 개발 환경을 구성하는 방법을 설명합니다.

## 사전 요구사항

<span class="material-icons">checklist</span> **필수 설치 항목**

- Python 3.13+
- Git
- UV (Python 패키지 관리자)
- Docker (선택)

## 개발 환경 구성

<span class="material-icons">developer_mode</span> **로컬 환경 설정**

### 1단계: 저장소 클론

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
```

### 2단계: 개발 의존성 설치

```bash
# UV를 사용한 설치 (권장)
uv sync --all-extras

# 또는 pip 사용
pip install -e ".[dev,test,docs]"
```

### 3단계: 사전 커밋 훅 설정

```bash
# Pre-commit 훅 설치
uv run pre-commit install

# 모든 파일에 대해 사전 검사 실행
uv run pre-commit run --all-files
```

## 테스트 실행

<span class="material-icons">science</span> **테스트 수행**

### 전체 테스트 스위트

```bash
# 모든 테스트 실행
uv run pytest

# 커버리지 리포트 포함
uv run pytest --cov=src/moai_adk --cov-report=html
```

### 특정 테스트 실행

```bash
# 특정 파일 테스트
uv run pytest tests/test_core.py

# 특정 함수 테스트
uv run pytest tests/test_core.py::test_function_name

# 마커 기반 실행
uv run pytest -m integration
```

## 코드 스타일 검사

<span class="material-icons">code</span> **코드 품질 관리**

### 린팅

```bash
# Ruff로 린팅
uv run ruff check src/ tests/

# Black으로 포맷팅
uv run black src/ tests/

# mypy로 타입 검사
uv run mypy src/moai_adk
```

### 자동 수정

```bash
# Ruff 자동 수정
uv run ruff check --fix src/ tests/

# Black 자동 포맷팅
uv run black src/ tests/
```

## 문서 빌드

<span class="material-icons">description</span> **문서화 작업**

### 로컬 문서 서버

```bash
cd docs

# 개발 서버 시작
uv run mkdocs serve

# 브라우저에서 http://localhost:8000 방문
```

### 프로덕션 빌드

```bash
# 정적 사이트 생성
uv run mkdocs build

# 출력: site/ 디렉토리
```

## 🔧 개발 워크플로우

### 기능 브랜치 생성

```bash
# 최신 develop 브랜치 동기화
git checkout develop
git pull origin develop

# 기능 브랜치 생성
git checkout -b feature/SPEC-XXX

# 또는 Alfred를 사용
/alfred:1-plan "기능 제목"
```

### 로컬 개발 및 테스트

```bash
# 코드 작성
# ... 수정 작업 ...

# 테스트 실행
uv run pytest

# 코드 스타일 검사
uv run ruff check --fix src/
uv run black src/

# 타입 검사
uv run mypy src/moai_adk
```

### 커밋 및 푸시

```bash
# 변경사항 추가
git add .

# Alfred를 사용한 커밋 (권장)
/alfred:2-run SPEC-XXX

# 또는 수동 커밋
git commit -m "feat: 기능 설명"
git push origin feature/SPEC-XXX
```

## 🔄 Pull Request 프로세스

1. **PR 생성**: 기능 브랜치에서 develop으로 PR 생성
2. **자동 검사**: GitHub Actions 자동 테스트 및 린팅 실행
3. **코드 리뷰**: 유지보수자의 검토 대기
4. **병합**: 승인 후 develop 브랜치로 병합

## 🐛 디버깅

### 로그 레벨 설정

```bash
# 디버그 모드 활성화
export MOAI_DEBUG=true
uv run moai-adk init my-project
```

### VS Code 디버깅

`.vscode/launch.json` 예제:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Python: Current File",
      "type": "python",
      "request": "launch",
      "program": "${file}",
      "console": "integratedTerminal"
    }
  ]
}
```

## 📚 참고 문서

- [코드 스타일 가이드](style.md)
- [릴리즈 프로세스](releases.md)
- [기여자 행동 강령](index.md)

## ❓ 문제 해결

### 의존성 오류

```bash
# 캐시 초기화 및 재설치
uv cache clean
uv sync --all-extras
```

### 테스트 실패

```bash
# 자세한 출력으로 테스트 실행
uv run pytest -vv

# 특정 테스트만 실행
uv run pytest tests/test_xxx.py::test_name -vv
```

### 문서 빌드 오류

```bash
# 캐시 정리
rm -rf docs/site docs/.cache

# 재빌드
cd docs
uv run mkdocs build --strict
```

---

**Questions?** GitHub Issues에서 질문하거나 Discussions에 참여하세요!
