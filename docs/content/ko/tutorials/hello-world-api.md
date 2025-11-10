# 10분 실습: Hello World API 만들기

MoAI-ADK의 완전한 개발 사이클을 경험하세요. 이 튜토리얼에서는 간단한 REST API를 SPEC-First TDD 방식으로 구현합니다.

## 🎯 학습 목표

- ✅ MoAI-ADK 4단계 워크플로우 이해
- ✅ SPEC 작성 (EARS 문법)
- ✅ TDD 사이클 (RED → GREEN → REFACTOR) 경험
- ✅ Alfred 에이전트와의 상호작용 학습
- ✅ 자동화된 Git 커밋과 문서 동기화 체험

______________________________________________________________________

## 📋 사전 준비 (1분)

### 필수 설치 항목

```bash
# UV 설치 (Python 패키지 매니저)
curl -LsSf https://astral.sh/uv/install.sh | sh

# 프로젝트 폴더 생성
mkdir hello-world-api
cd hello-world-api

# MoAI-ADK 초기화
moai-adk init
```

### 초기 설정

```bash
# 프로젝트 설정
/alfred:0-project
# → 프로젝트 이름, 설명, 언어, 모드 선택
```

______________________________________________________________________

## 🔴 1단계: 계획 (Plan) - 요구사항 정의 (2분)

SPEC을 작성하여 API의 요구사항을 명확히 합니다.

```bash
/alfred:1-plan "간단한 할 일 목록 API 만들기"
```

Alfred가 생성하는 SPEC:

- **SPEC ID**: API-001
- **요구사항** (EARS 형식):
  - 시스템은 할 일 항목을 생성할 수 있어야 함
  - 시스템은 모든 할 일 항목을 조회할 수 있어야 함
  - 시스템은 특정 할 일 항목을 업데이트할 수 있어야 함
  - 시스템은 할 일 항목을 삭제할 수 있어야 함

**확인 사항**:

- ✅ SPEC 파일이 생성되었는가? (`.moai/specs/API-001/spec.md`)
- ✅ 기능 브랜치가 생성되었는가? (`feature/API-001`)

______________________________________________________________________

## 🟢 2단계: 실행 (Run) - TDD 구현 (5분)

### Phase 1: 테스트 작성 (RED)

```bash
/alfred:2-run API-001
```

Alfred의 tdd-implementer가 제시하는 작업:

1. **테스트 파일 생성** (`tests/test_api.py`)
2. **테스트 작성** (SPEC의 각 요구사항별)

```python
# tests/test_api.py
import pytest
from app import app

@pytest.fixture
def client():
    with app.test_client() as client:
        yield client

def test_create_todo(client):
    """시스템은 할 일 항목을 생성할 수 있어야 함"""
    response = client.post('/todos', json={'title': '장보기'})
    assert response.status_code == 201

def test_list_todos(client):
    """시스템은 모든 할 일 항목을 조회할 수 있어야 함"""
    response = client.get('/todos')
    assert response.status_code == 200
    assert isinstance(response.json, list)
```

**확인**: 테스트 실행 결과 모두 **실패** (RED)

### Phase 2: 최소 구현 (GREEN)

```python
# app.py
from flask import Flask, request, jsonify

app = Flask(__name__)
todos = []

@app.route('/todos', methods=['POST'])
def create_todo():
    todo = {'id': len(todos) + 1, **request.json}
    todos.append(todo)
    return jsonify(todo), 201

@app.route('/todos', methods=['GET'])
def list_todos():
    return jsonify(todos)
```

**확인**: 테스트 실행 결과 모두 **통과** (GREEN)

### Phase 3: 코드 개선 (REFACTOR)

- 에러 처리 추가
- 데이터 검증 추가
- 코드 정리 및 최적화

```bash
# 각 단계별 커밋
git add .
git commit -m "feat: 할 일 API 구현"  # GREEN 이후
git commit -m "refactor: 에러 처리 및 검증 추가"
```

______________________________________________________________________

## ♻️ 3단계: 동기화 (Sync) - 문서 및 검증 (2분)

구현 완료 후 문서 동기화 및 품질 검증:

```bash
/alfred:3-sync
```

Alfred가 수행하는 작업:

1. **문서 생성**: API 문서 자동 생성
2. **TAG 검증**: SPEC → TEST → CODE 연결 확인
3. **테스트 커버리지**: 85% 이상 검증
4. **Pull Request 생성**: develop 브랜치 대상

______________________________________________________________________

## 🎉 완료!

축하합니다! 당신은 방금:

- ✅ SPEC-First 개발 방식 체험
- ✅ TDD 사이클 완료 (RED → GREEN → REFACTOR)
- ✅ 자동화된 문서 생성
- ✅ 추적성 있는 개발 (TAG 시스템)

______________________________________________________________________

## 📚 다음 단계

### 기술 심화

- [SPEC 작성 고급](guides/specs/basics.md) - EARS 문법 마스터
- [TDD 패턴](guides/tdd/index.md) - 다양한 TDD 패턴
- [Alfred 에이전트](guides/alfred/index.md) - 19명의 전문가 팀 활용

### 실전 프로젝트

- 인증 시스템 추가
- 데이터베이스 통합
- API 문서화 (OpenAPI/Swagger)
- 배포 자동화

### 팀 협업

- Git 워크플로우 (GitFlow)
- Pull Request 리뷰 프로세스
- 지속적 배포 (CI/CD)

______________________________________________________________________

## 🆘 문제 해결

### Alfred 명령이 인식되지 않음

```bash
# Claude Code 재시작
exit

# 새 세션 시작
claude
```

### 테스트가 실패함

```bash
# 의존성 재설치
uv sync

# 수동 테스트 실행
pytest tests/ -v
```

### SPEC 파일을 찾을 수 없음

```bash
# 프로젝트 상태 확인
moai-adk doctor

# 재초기화
/alfred:0-project
```

______________________________________________________________________

**다음 튜토리얼**: [사용자 인증 시스템 만들기](../coming-soon) (준비 중)
