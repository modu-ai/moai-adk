# 빠른 시작 가이드

단 10분 만에 완전한 MoAI-ADK 워크플로우를 경험해 보세요. 이 가이드는 SPEC-First TDD 방법론을 사용하여 첫 API를 만드는 과정을 안내합니다.

## 전제 조건

시작하기 전에 다음을 확인하세요:

- ✅ **MoAI-ADK 설치됨**: `uv tool install moai-adk`
- ✅ **Claude Code 설치됨**: 터미널에서 사용 가능
- ✅ **10분의 방해받지 않는 시간**

## 0단계: 프로젝트 초기화 (1분)

새 프로젝트를 만들고 Alfred로 초기화하세요.

```bash
# 새 프로젝트 생성
moai-adk init hello-api
cd hello-api

# Claude Code 시작
claude
```

Claude Code에서 프로젝트 초기화 명령을 실행하세요:

```
/alfred:0-project
```

Alfred가 몇 가지 질문을 할 것입니다:

- **프로젝트 이름**: hello-api
- **프로젝트 목표**: MoAI-ADK 학습
- **주요 언어**: python
- **모드**: personal (로컬 개발용)

**결과**: `.moai/` 구성으로 프로젝트가 초기화되고, 스킬이 로드되며, Alfred가 준비됩니다.

## 1단계: SPEC 생성 (2분)

Alfred의 spec-builder를 사용하여 만들 것을 정의하세요.

```bash
/alfred:1-plan "name 쿼리 매개변수를 받아서 인사말을 반환하는 GET /hello 엔드포인트"
```

### Alfred가 하는 일

1. **요청을 분석**하고 요구사항을 추출

2. **SPEC ID 생성**: `HELLO-001`

3. **EARS 형식 사양 생성**:

   ```yaml
   ---
   id: HELLO-001
   version: 0.0.1
   status: draft
   priority: high
   ---
   # @SPEC:EX-HELLO-001: Hello World API

   ## Ubiquitous Requirements
   - 시스템은 HTTP GET /hello 엔드포인트를 제공해야 함

   ## Event-driven Requirements
   - name 쿼리 매개변수가 제공되면 "Hello, {name}!"을 반환해야 함
   - name이 제공되지 않으면 "Hello, World!"를 반환해야 함

   ## Constraints
   - name은 최대 50자로 제한되어야 함
   - 응답은 JSON 형식이어야 함
   ```

4. **지원 파일 생성**:

   - `.moai/specs/SPEC-HELLO-001/spec.md` - 주요 사양
   - `.moai/specs/SPEC-HELLO-001/plan.md` - 구현 계획
   - 기능 브랜치: `feature/SPEC-HELLO-001` (팀 모드인 경우)

### 확인

```bash
# SPEC이 생성되었는지 확인
cat .moai/specs/SPEC-HELLO-001/spec.md

# TAG 할당 확인
rg '@SPEC:HELLO-001' -n
```

## 2단계: TDD 구현 (5분)

테스트 주도 개발을 사용하여 API를 구현하세요.

```bash
/alfred:2-run HELLO-001
```

### 1단계: 🔴 RED - 실패하는 테스트 작성

Alfred의 `tdd-implementer`가 먼저 포괄적인 테스트를 생성합니다:

```python
# tests/test_hello.py
# @TEST:EX-HELLO-001 | SPEC: SPEC-HELLO-001.md

import pytest
from fastapi.testclient import TestClient
from src.hello.api import app

client = TestClient(app)

def test_hello_with_name_should_return_personalized_greeting():
    """name이 제공되면 'Hello, {name}!'을 반환해야 함"""
    response = client.get("/hello?name=Alice")
    assert response.status_code == 200
    assert response.json() == {"message": "Hello, Alice!"}

def test_hello_without_name_should_return_default_greeting():
    """name이 제공되지 않으면 'Hello, World!'를 반환해야 함"""
    response = client.get("/hello")
    assert response.status_code == 200
    assert response.json() == {"message": "Hello, World!"}

def test_hello_with_long_name_should_return_400():
    """name이 50자를 초과하면 400 에러를 반환해야 함"""
    long_name = "a" * 51
    response = client.get(f"/hello?name={long_name}")
    assert response.status_code == 400
```

**테스트 실행** (실패할 것입니다 - 이것이 예상됨):

```bash
pytest tests/test_hello.py -v
# 결과: FAILED - No module named 'src.hello.api'
```

**RED 단계 커밋**:

```bash
git add tests/test_hello.py
git commit -m "🔴 test(HELLO-001): add failing hello API tests"
```

### 2단계: 🟢 GREEN - 최소 구현

Alfred가 테스트를 통과시키기 위한 최소한의 코드를 생성합니다:

```python
# src/hello/api.py
# @CODE:EX-HELLO-001:API | SPEC: SPEC-HELLO-001.md | TEST: tests/test_hello.py

from fastapi import FastAPI, HTTPException

app = FastAPI()

@app.get("/hello")
def hello(name: str = "World"):
    """@CODE:EX-HELLO-001:API - Hello 엔드포인트"""
    if len(name) > 50:
        raise HTTPException(status_code=400, detail="Name too long (max 50 chars)")
    return {"message": f"Hello, {name}!"}
```

**테스트 실행** (이제 통과해야 함):

```bash
pytest tests/test_hello.py -v
# 결과: PASSED - 모든 3개 테스트 통과
```

**GREEN 단계 커밋**:

```bash
git add src/hello/api.py
git commit -m "🟢 feat(HELLO-001): implement hello API"
```

### 3단계: ♻️ REFACTOR - 코드 품질 개선

Alfred가 TRUST 5 원칙을 적용하여 코드를 개선합니다:

```python
# src/hello/models.py
# @CODE:EX-HELLO-001:MODEL | SPEC: SPEC-HELLO-001.md

from pydantic import BaseModel, Field, validator

class HelloRequest(BaseModel):
    """@CODE:EX-HELLO-001:MODEL - 요청 검증 모델"""
    name: str = Field(default="World", max_length=50, description="인사할 이름")

    @validator('name')
    def validate_name(cls, v):
        if not v.strip():
            raise ValueError('이름은 비워둘 수 없습니다')
        return v.strip()

class HelloResponse(BaseModel):
    """@CODE:EX-HELLO-001:MODEL - 응답 모델"""
    message: str = Field(description="인사말 메시지")
```

```python
# src/hello/api.py (리팩토링됨)
# @CODE:EX-HELLO-001:API | SPEC: SPEC-HELLO-001.md | TEST: tests/test_hello.py

from fastapi import FastAPI, HTTPException, Depends
from .models import HelloRequest, HelloResponse

app = FastAPI(title="Hello API", version="1.0.0")

@app.get("/hello", response_model=HelloResponse)
def hello(params: HelloRequest = Depends()):
    """@CODE:EX-HELLO-001:API - 검증을 포함한 Hello 엔드포인트"""
    try:
        message = f"Hello, {params.name}!"
        return HelloResponse(message=message)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
```

**테스트가 여전히 통과하는지 확인**:

```bash
pytest tests/test_hello.py -v
# 결과: PASSED - 모든 테스트 여전히 통과
```

**REFACTOR 단계 커밋**:

```bash
git add src/hello/models.py src/hello/api.py
git commit -m "♻️ refactor(HELLO-001): add models and improve validation"
```

## 3단계: 문서 동기화 (1분)

모든 문서를 동기화하고 완전한 시스템을 검증하세요.

```bash
/alfred:3-sync
```

### Alfred가 하는 일

1. **API 문서 생성**:

   ````markdown
   # Hello API 문서

   ## GET /hello

   개인화된 인사말 메시지를 반환합니다.

   ### 매개변수
   - `name` (query, 선택적): 인사할 이름 (기본값: "World", 최대 50자)

   ### 응답
   - **200**: 성공
     ```json
     {"message": "Hello, Alice!"}
   ````

   - **400**: 검증 오류

   ### 예제

   ```bash
   curl "http://localhost:8000/hello?name=Alice"
   # → {"message": "Hello, Alice!"}
   ```

   ### 추적성

   - @SPEC:EX-HELLO-001 - 요구사항
   - @TEST:EX-HELLO-001 - 테스트
   - @CODE:EX-HELLO-001 - 구현

   ```

   ```

2. **README.md 업데이트** (API 사용 예제 포함)

3. **CHANGELOG.md 생성** (버전 기록 포함)

4. **TAG 체인 무결성 검증**:

   ```
   ✅ @SPEC:EX-HELLO-001 → .moai/specs/SPEC-HELLO-001/spec.md
   ✅ @TEST:EX-HELLO-001 → tests/test_hello.py
   ✅ @CODE:EX-HELLO-001 → src/hello/ (3개 파일)
   ✅ @DOC:EX-HELLO-001 → docs/api/hello.md (자동 생성)

   TAG 체인 무결성: 100%
   고아 TAG: 없음
   ```

5. **TRUST 5 준수 검증**:

   ```
   ✅ Test First: 100% 커버리지 (3/3 테스트 통과)
   ✅ Readable: 모든 함수 < 50줄
   ✅ Unified: FastAPI 패턴 일관성
   ✅ Secured: 입력 검증 구현됨
   ✅ Trackable: 모든 코드에 @CODE:HELLO-001 태그됨
   ```

## 4단계: 검증 및 축하 (1분)

### 완전한 시스템 검증

```bash
# 1. TAG 체인 무결성 확인
rg '@(SPEC|TEST|CODE|DOC):HELLO-001' -n
# 출력에 모든 4가지 TAG 유형이 표시되어야 함

# 2. 테스트 실행
pytest tests/test_hello.py -v
# 모든 테스트가 통과해야 함

# 3. API 테스트
uvicorn src.hello.api:app --reload &
curl "http://localhost:8000/hello?name=World"
# 반환되어야 함: {"message": "Hello, World!"}

# 4. 생성된 문서 확인
cat docs/api/hello.md
# 완전한 API 문서가 포함되어야 함
```

### 성과 검토

성공적으로 생성한 것들:

```
hello-api/
├── .moai/specs/SPEC-HELLO-001/
│   ├── spec.md              ← 전문적 사양
│   └── plan.md              ← 구현 계획
├── tests/test_hello.py      ← 100% 테스트 커버리지
├── src/hello/
│   ├── api.py               ← 프로덕션 품질 구현
│   ├── models.py            ← 데이터 검증 모델
│   └── __init__.py
├── docs/api/hello.md        ← 자동 생성된 API 문서
├── README.md                ← 사용 예제로 업데이트됨
├── CHANGELOG.md             ← 버전 기록
└── .git/                    ← TDD 커밋과 깨끗한 git 기록
```

### Git 기록

```bash
git log --oneline | head -5
```

예상 출력:

```
a1b2c3d ✅ sync(HELLO-001): update docs and changelog
d4e5f6c ♻️ refactor(HELLO-001): add models and improve validation
b2c3d4e 🟢 feat(HELLO-001): implement hello API
a3b4c5d 🔴 test(HELLO-001): add failing hello API tests
e5f6g7h 🌿 Create feature/SPEC-HELLO-001 branch
```

## 배운 것들

### 경험한 개념

✅ **SPEC-First**: 코딩 전에 명확한 요구사항 생성 ✅ **TDD**: 100% 테스트 커버리지로 RED → GREEN → REFACTOR 사이클 ✅ **@TAG
시스템**: 요구사항부터 문서까지 완전한 추적성 ✅ **TRUST 5**: 검증과 오류 처리를 포함한 프로덕션 품질 코드 ✅ **Alfred 워크플로우**: 자동화된 문서화와 품질
검사

### 얻은 기술

- **EARS 구문**: 구조화된 요구사항 작성
- **테스트 설계**: 포괄적인 테스트 케이스 생성
- **API 개발**: FastAPI 모범 사례
- **문서화**: 자동 생성, 항상 동기화되는 문서
- **Git 워크플로우**: 깨끗하고 추적 가능한 커밋 기록

## 다음 단계

### 계속해서 빌드하기

API에 더 많은 기능 추가:

```bash
# 새 엔드포인트 추가
/alfred:1-plan "JSON 본문을 받는 POST /greet 엔드포인트"

# 또는 기존 기능 향상
/alfred:1-plan "/hello 엔드포인트에 언어 지원 추가"
```

### 고급 주제 탐색

- **[프로젝트 구성](../../guides/project/config.md)**: 프로젝트 설정 사용자정의
- **[SPEC 작성](../../guides/specs/basics.md)**: EARS 구문 마스터하기
- **[TDD 패턴](../../guides/tdd/green.md)**: 고급 테스트 전략 학습
- **[TAG 시스템](../reference/tags/index.md)**: 추적성 깊이 이해하기

### 커뮤니티 참여

- **GitHub Issues**: 버그 리포트 또는 기능 요청
- **Discussions**: 질문하고 경험 공유
- **기여**: MoAI-ADK 개선 도와주기

## 문제 해결

### 일반적인 문제

**가져오기 오류로 테스트 실패**:

```bash
# 의존성 설치
uv add fastapi pytest
uv sync
```

**API가 시작되지 않음**:

```bash
# 포트와 의존성 확인
lsof -i :8000
uvicorn src.hello.api:app --reload --port 8001
```

**문서가 생성되지 않음**:

```bash
# 수동으로 동기화 실행
/alfred:3-sync
```

### 도움 얻기

```bash
# 시스템 진단
moai-adk doctor

# 자동으로 이슈 생성
/alfred:9-feedback
```

## 요약

단 10분 만에 다음을 완료했습니다:

1. ✅ **명확한 요구사항 정의** (SPEC과 EARS 구문 사용)
2. ✅ **TDD로 구현** (100% 테스트 커버리지 달성)
3. ✅ **프로덕션 품질 코드 생성** (검증과 오류 처리 포함)
4. ✅ **완전한 문서 생성** (동기화 유지)
5. ✅ **완전한 추적성 유지** (@TAG 시스템으로)
6. ✅ **모범 사례 준수** (TRUST 5 원칙으로)

이것이 MoAI-ADK의 힘입니다: 전통적인 방법보다 더 빠르게 신뢰할 수 있고, 유지보수가 쉬우며, 잘 문서화된 코드를 생성하세요. 이제 자신감 있게 복잡한 애플리케이션을 빌드할
준비가 되었습니다! 🚀

[Alfred 워크플로우 가이드](../../guides/alfred/index.md)로 여정을 계속하거나 관심 있는 특정 주제를 탐색하세요.
