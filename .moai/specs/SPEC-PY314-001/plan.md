# SPEC-PY314-001 구현 계획

## 우선순위별 마일스톤

### 1차 목표: 프로젝트 구조 생성
- [ ] moai-adk-py/ 디렉토리 생성
- [ ] pyproject.toml 작성
- [ ] src/moai_adk/ 패키지 구조 생성
- [ ] __init__.py, __main__.py 생성

### 2차 목표: 의존성 설정
- [ ] uv.lock 파일 생성
- [ ] core 의존성 설치 (click, rich, gitpython, jinja2, pyyaml)
- [ ] dev 의존성 설치 (pytest, ruff, mypy)

### 3차 목표: 빌드 시스템 검증
- [ ] uv build 명령 테스트
- [ ] wheel 패키지 생성 확인
- [ ] editable 설치 확인 (`uv pip install -e .`)

### 최종 목표: 진입점 검증
- [ ] `moai --version` 명령 실행 확인
- [ ] `moai --help` 출력 확인

---

## 기술적 접근 방법

### 1. src/ 레이아웃 사용 이유
- 패키지 가져오기 충돌 방지
- 테스트와 소스 코드 명확한 분리
- PyPI 배포 시 표준 구조

### 2. uv 선택 이유
- npm/bun과 유사한 빠른 속도 (Rust 기반)
- pip보다 10~100배 빠른 의존성 해결
- lock 파일로 재현 가능한 빌드

### 3. 의존성 최소화 전략
- core: 6개 패키지만 사용
- dev: 테스트/린트/타입 체크만
- 선택적 의존성 없음 (all-in-one 패키지)

---

## 아키텍처 설계 방향

### 모듈 구조
```
moai_adk/
├── cli/          # Click 기반 명령어 (SPEC-CLI-001)
├── core/         # 핵심 비즈니스 로직
│   ├── git/      # Git 워크플로우 (SPEC-CORE-GIT-001)
│   ├── template/ # 템플릿 처리 (SPEC-CORE-TEMPLATE-001)
│   └── project/  # 프로젝트 초기화 (SPEC-CORE-PROJECT-001)
└── hooks/        # Hooks 런타임 (SPEC-HOOKS-001, SPEC-HOOKS-002)
```

### 의존성 그래프
- PY314-001 (Foundation) → 모든 Core 모듈
- Core 모듈들은 서로 독립적 (순환 의존성 없음)

---

## 리스크 및 대응 방안

### 리스크 1: Python 3.14 호환성
- **문제**: Python 3.14는 아직 베타 버전일 수 있음
- **대응**: `requires-python = ">=3.12"`로 낮춰서 호환성 확보
- **검증**: CI에서 3.12, 3.13, 3.14 테스트

### 리스크 2: uv 생태계 성숙도
- **문제**: uv는 비교적 새로운 도구 (2024년 출시)
- **대응**: 기존 pip/poetry로 fallback 가능하도록 문서화
- **검증**: pyproject.toml은 표준 PEP 621 준수

### 리스크 3: TypeScript 기능 동등성
- **문제**: TS → Python 전환 시 기능 손실 가능
- **대응**: 기능 매핑 테이블 작성 (SPEC별로 검증)
- **검증**: 통합 테스트로 CLI 명령어 동등성 확인

### 리스크 4: 패키지명 충돌
- **문제**: PyPI에 이미 moai-adk가 존재할 수 있음
- **대응**: 사전 검색, 필요 시 moai-dev-kit 등 대체명
- **검증**: `pip search moai-adk` 또는 pypi.org 검색
