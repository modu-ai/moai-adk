# SPEC-PY314-001 수락 기준

## Given-When-Then 테스트 시나리오

### 시나리오 1: 프로젝트 구조 생성
**Given**: moai-adk-py/ 디렉토리가 존재하지 않음
**When**: 프로젝트 구조를 생성함
**Then**:
- [ ] moai-adk-py/src/moai_adk/ 디렉토리가 존재함
- [ ] pyproject.toml 파일이 존재함
- [ ] __init__.py, __main__.py 파일이 존재함

### 시나리오 2: pyproject.toml 검증
**Given**: pyproject.toml 파일이 작성됨
**When**: 파일 내용을 파싱함
**Then**:
- [ ] `name = "moai-adk"`가 포함됨
- [ ] `version = "0.3.0"`가 포함됨
- [ ] `requires-python = ">=3.14"` (또는 3.12)가 포함됨
- [ ] `[project.scripts]` 섹션에 `moai` 진입점이 정의됨
- [ ] 6개 core 의존성이 `dependencies`에 포함됨

### 시나리오 3: uv 빌드 성공
**Given**: pyproject.toml과 소스 코드가 준비됨
**When**: `uv build` 명령을 실행함
**Then**:
- [ ] dist/ 디렉토리가 생성됨
- [ ] .whl 파일이 생성됨
- [ ] .tar.gz 파일이 생성됨
- [ ] 빌드 오류가 없음

### 시나리오 4: Editable 설치
**Given**: 빌드가 성공함
**When**: `uv pip install -e .` 명령을 실행함
**Then**:
- [ ] 설치가 성공함
- [ ] `which moai` 명령이 실행 파일을 찾음
- [ ] `moai --version`이 "0.3.0"을 출력함

### 시나리오 5: 진입점 검증
**Given**: 패키지가 설치됨
**When**: `moai --help` 명령을 실행함
**Then**:
- [ ] 도움말 메시지가 출력됨
- [ ] "MoAI Agentic Development Kit" 문구가 포함됨
- [ ] 사용 가능한 명령어 목록이 표시됨

---

## 품질 게이트 기준

### 1. 구조 검증
- [ ] src/ 레이아웃 준수
- [ ] __init__.py에 `__version__ = "0.3.0"` 정의
- [ ] 순환 의존성 없음 (mypy로 검증)

### 2. 빌드 검증
- [ ] uv build 성공
- [ ] wheel 크기 < 10MB
- [ ] 불필요한 파일 미포함 (__pycache__, .pyc 제외)

### 3. 의존성 검증
- [ ] uv.lock 파일 생성
- [ ] core 의존성 6개만 포함
- [ ] dev 의존성은 선택적으로 설치

### 4. 타입 검증
- [ ] mypy --strict 통과
- [ ] 모든 공개 함수에 타입 힌트 존재

---

## 검증 방법 및 도구

### 자동화 테스트
```bash
# 구조 검증
test -d moai-adk-py/src/moai_adk
test -f pyproject.toml

# 빌드 검증
uv build
test -f dist/moai_adk-0.3.0-py3-none-any.whl

# 설치 검증
uv pip install -e .
moai --version | grep "0.3.0"

# 타입 검증
mypy moai-adk-py/src/ --strict
```

### 수동 검증
1. **pyproject.toml 검토**: 메타데이터 정확성
2. **디렉토리 구조 검토**: src/ 레이아웃 준수
3. **의존성 검토**: 불필요한 패키지 없음
4. **진입점 검토**: `moai` 명령 실행 가능

---

## 완료 조건 (Definition of Done)

### 필수 조건
- [x] pyproject.toml 작성 완료
- [x] src/ 디렉토리 구조 생성
- [ ] uv build 성공
- [ ] moai --version 실행 성공
- [ ] 모든 시나리오 테스트 통과

### 선택 조건
- [ ] README.md에 설치 가이드 추가
- [ ] CI 설정 (GitHub Actions)
- [ ] 로컬 테스트 환경 검증

### 문서화
- [ ] 빌드 명령어 문서화
- [ ] 의존성 매핑 테이블 검증
- [ ] 리스크 대응 방안 확인

---

## 롤백 계획

**실패 시나리오**: uv 빌드 실패 또는 의존성 충돌

**롤백 절차**:
1. moai-adk-py/ 디렉토리 제거
2. 기존 moai-adk-ts 유지
3. SPEC 상태를 `deprecated`로 변경
4. 대체 방안 논의 (poetry 사용, Python 3.12로 다운그레이드 등)
