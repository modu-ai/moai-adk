# Part 5: 문서 동기화 및 통합 테스트

> **소요시간**: 약 30분
> **학습 목표**: `/alfred:3-sync`로 Living Document를 생성하고, @TAG 체인을 검증하며, 실제 서버에서 통합 테스트를 수행합니다.

---

## 📋 이전 단계 요약

Part 4까지 완료한 상태:
- ✅ 백엔드: FastAPI + SQLAlchemy 2.0 + Alembic (테스트 커버리지 95%)
- ✅ 프론트엔드: Vite + React 18 + TypeScript + Tailwind CSS
- ✅ Git 커밋: RED-GREEN-REFACTOR 사이클 완료
- ⏳ 문서 동기화: 아직 실행하지 않음

현재 상태:
```bash
git log --oneline -5
# be050e3 ✅ TYPE: mypy strict 모드 타입 힌트 추가
# 099a08a 🔒 SECURITY: .gitignore 추가
# afd88e7 ♻️ REFACTOR: 프로젝트 문서 및 SPEC 추가
# 2ba5664 🟢 GREEN: Todo CRUD API + UI 구현
# ce0278e 🔴 RED: Todo CRUD 테스트 작성
```

---

## Step 1: 문서 동기화 실행 (`/alfred:3-sync`)

### 1.1 동기화 계획 수립

`/alfred:3-sync` 커맨드를 실행하면 먼저 현재 프로젝트 상태를 분석합니다:

```bash
/alfred:3-sync
```

**출력 예시**:
```
📋 STEP 1: 문서 동기화 범위 분석

프로젝트 정보:
- 모드: Personal (개인 작업)
- 프로젝트: my-moai-project (Fullstack - FastAPI + Vite React)
- 현재 브랜치: main

Git 상태:
- 미커밋 변경: 2개 파일 (mypy 타입 힌트 추가)
  - backend/app/database.py (+6, -2)
  - backend/app/main.py (+2, -1)

@TAG 시스템 상태:
- ✅ 26개 파일에서 54개 TAG 발견
- ✅ TAG 체인: @SPEC → @TEST → @CODE 완성
- ⚠️ @DOC TAG 누락 (문서 동기화 후 생성 예정)

---
🎯 동기화 전략

1단계: 미커밋 변경사항 처리
- mypy 타입 힌트 추가 파일 커밋

2단계: SPEC 메타데이터 업데이트 (자동)
- status: draft → completed
- version: 0.0.1 → 0.1.0
- HISTORY 섹션 업데이트

3단계: Living Document 생성
- .moai/specs/SPEC-TODO-001/README.md 생성
- API 문서 자동 생성
- TAG 추적성 테이블

4단계: TAG 시스템 검증
- @SPEC → @TEST → @CODE 체인 무결성
- 고아 TAG, 끊어진 링크 확인
- @DOC TAG 추가

5단계: 동기화 보고서 생성
- .moai/reports/sync-report.md 생성
```

### 1.2 승인 및 실행

사용자 입력:
```bash
진행
```

Alfred가 자동으로 다음 작업을 수행합니다:

**Phase 1: 미커밋 변경사항 처리**
```bash
# 타입 힌트 커밋
git add app/database.py app/main.py
git commit -m "✅ TYPE: mypy strict 모드 타입 힌트 추가

- Generator[Session, None, None] 타입 추가
- dict[str, str] 반환 타입 명시
- mypy --strict 통과 ✅

@TAG:TODO-001-TYPE"
```

**Phase 2: SPEC 메타데이터 업데이트**

`.moai/specs/SPEC-TODO-001/spec.md` 변경 내용:
```diff
 ---
 id: TODO-001
-version: 0.0.1
-status: draft
+version: 0.1.0
+status: completed
 created: 2025-10-11
 updated: 2025-10-11
 author: "@Goos"
 priority: high
 ---

 ## HISTORY

+### v0.1.0 (2025-10-11)
+- **CHANGED**: status → completed (TDD 구현 완료)
+- **ADDED**: 백엔드 구현 완료 (FastAPI + SQLAlchemy 2.0 + Alembic + pytest)
+- **ADDED**: 프론트엔드 구현 완료 (Vite + React 18 + TypeScript 5 + Tailwind CSS)
+- **ADDED**: 테스트 커버리지 95% 달성 (14개 테스트 통과)
+- **ADDED**: mypy strict 모드 타입 힌트 적용
+- **AUTHOR**: @Goos
+- **REVIEW**: N/A (Personal 모드)
+- **RELATED**: Git commits: ce0278e (RED), 2ba5664 (GREEN), afd88e7 (REFACTOR), 099a08a (SECURITY), be050e3 (TYPE)
+
 ### v0.0.1 (2025-10-11)
 - **INITIAL**: Todo 항목 CRUD 기능 명세 작성
```

---

## Step 2: Living Document 생성

Alfred가 자동으로 `.moai/specs/SPEC-TODO-001/README.md`를 생성합니다.

### 2.1 Living Document 주요 섹션

**1. 프로젝트 개요**
```markdown
# Todo CRUD 애플리케이션 - Living Document

> **@DOC:TODO-001** | SPEC: SPEC-TODO-001.md | STATUS: ✅ Completed (v0.1.0)

## 📋 프로젝트 개요

FastAPI + SQLAlchemy 2.0 + Alembic 기반 백엔드와 Vite + React 18 + TypeScript 5 + Tailwind CSS 기반 프론트엔드로 구성된 풀스택 Todo CRUD 애플리케이션입니다.

**핵심 기능**:
- Todo 생성, 조회, 수정, 삭제 (CRUD)
- 완료 상태 토글
- 자동 타임스탬프 관리 (created_at, updated_at)
```

**2. 빠른 시작 가이드**
```markdown
## 🚀 빠른 시작

### 백엔드 실행
cd backend
source .venv/bin/activate
uvicorn app.main:app --reload

# URL: http://localhost:8000
# API 문서: http://localhost:8000/docs

### 프론트엔드 실행
cd frontend
pnpm dev

# URL: http://localhost:5173
```

**3. API 문서 (자동 생성)**
```markdown
## 📡 API 엔드포인트

### 1. Create Todo
**POST** `/api/todos`

**Request Body**:
```json
{
  "title": "장보기",
  "description": "우유, 계란 사기"
}
```

**Response** (201 Created):
```json
{
  "id": 1,
  "title": "장보기",
  "description": "우유, 계란 사기",
  "completed": false,
  "created_at": "2025-10-11T12:00:00",
  "updated_at": "2025-10-11T12:00:00"
}
```

### 2. Get All Todos
**GET** `/api/todos`

**Response** (200 OK):
```json
[
  {
    "id": 1,
    "title": "장보기",
    "description": "우유, 계란 사기",
    "completed": false,
    "created_at": "2025-10-11T12:00:00",
    "updated_at": "2025-10-11T12:00:00"
  }
]
```

### 3. Get Single Todo
**GET** `/api/todos/{id}`

**Response** (200 OK):
```json
{
  "id": 1,
  "title": "장보기",
  "description": "우유, 계란 사기",
  "completed": false,
  "created_at": "2025-10-11T12:00:00",
  "updated_at": "2025-10-11T12:00:00"
}
```

### 4. Update Todo
**PATCH** `/api/todos/{id}`

**Request Body**:
```json
{
  "title": "장보기 (업데이트)",
  "completed": true
}
```

> **참고**: PATCH는 리소스의 부분 업데이트를 의미하며, 제공된 필드만 수정됩니다.

### 5. Delete Todo
**DELETE** `/api/todos/{id}`

**Response**: 204 No Content
```

**4. 테스트 가이드**
```markdown
## 🧪 테스트

### 백엔드 테스트
cd backend
pytest tests/ -v --cov=app

# 결과:
# ============================= test session starts ==============================
# tests/test_todos.py::TestTodoCreate::test_create_todo_success PASSED    [  7%]
# tests/test_todos.py::TestTodoCreate::test_create_todo_missing_title PASSED [ 14%]
# ...
# tests/test_todos.py::TestTodoDelete::test_delete_todo_not_found PASSED  [100%]
#
# ---------- coverage: platform darwin, python 3.13.1-final-0 -----------
# Name                              Stmts   Miss  Cover   Missing
# ---------------------------------------------------------------
# app/__init__.py                       1      0   100%
# app/database.py                      12      0   100%
# app/models/__init__.py                1      0   100%
# app/models/todo.py                   10      0   100%
# app/routers/__init__.py               0      0   100%
# app/routers/todos.py                 48      2    96%   67, 89
# app/schemas/__init__.py               1      0   100%
# app/schemas/todo.py                   6      0   100%
# app/services/__init__.py              1      0   100%
# app/services/todo_service.py         36      0   100%
# ---------------------------------------------------------------
# TOTAL                               116      2    95%
#
# ============================== 14 passed in 0.16s ===============================
```

**5. TAG 추적성**
```markdown
## 🏷️ TAG 추적성

### TAG 체인
@SPEC:TODO-001 → @TEST:TODO-001 → @CODE:TODO-001 → @DOC:TODO-001

### TAG 통계
| TAG 타입        | 파일 수 | TAG 수 |
|----------------|--------|--------|
| @SPEC:TODO-001 | 3      | 3      |
| @TEST:TODO-001 | 5      | 5      |
| @CODE:TODO-001 | 18     | 61     |
| @DOC:TODO-001  | 1      | 1      |
| **총계**       | **27** | **70** |

### 코드 추적성 예시
```python
# backend/app/models/todo.py
# @CODE:TODO-001:DATA | SPEC: SPEC-TODO-001.md
class Todo(Base):
    __tablename__ = "todos"
    id: Mapped[int] = mapped_column(primary_key=True, index=True)
    title: Mapped[str] = mapped_column(String(200), nullable=False)
    ...
```
```

**6. 품질 메트릭 (TRUST 5원칙)**
```markdown
## 📊 품질 메트릭 (TRUST 5원칙)

| 원칙            | 점수   | 상태 | 세부 사항                    |
|----------------|--------|------|----------------------------|
| T (Test First) | 95%    | ✅   | 커버리지 95%, 14개 테스트 통과 |
| R (Readable)   | 100%   | ✅   | 최대 94 LOC (목표 ≤300)    |
| U (Unified)    | 90%    | ⚠️   | TypeScript strict 모드 활성화 |
| S (Secured)    | 85%    | ⚠️   | Pydantic 검증, ORM 사용    |
| T (Trackable)  | 100%   | ✅   | 27개 파일 @TAG 적용         |

**전체 준수율**: 94% (목표 85% 초과 ✅)
```

### 2.2 동기화 보고서 생성

`.moai/reports/sync-report-2025-10-11.md` 생성:

```markdown
# 문서 동기화 보고서

**생성 일시**: 2025-10-11
**프로젝트**: my-moai-project
**모드**: Personal
**실행 커맨드**: `/alfred:3-sync`

---

## 📊 동기화 요약

### SPEC 업데이트
| SPEC ID  | 이전 버전  | 이후 버전  | 이전 상태 | 이후 상태    |
|----------|-----------|-----------|---------|------------|
| TODO-001 | v0.0.1    | v0.1.0    | draft   | completed  |

### TAG 시스템 상태
- 전체 TAG: 70개
- 적용 파일: 27개
- 고아 TAG: 0개 ✅
- TAG 체인 무결성: 100% ✅

### Git 커밋
| 커밋 ID  | 메시지                        | 작업          |
|---------|------------------------------|---------------|
| be050e3 | ✅ TYPE: mypy strict 모드     | 타입 힌트 추가 |
| 173255d | 📝 DOCS: Living Document 동기화 | 문서 생성     |

---

## ✅ 품질 검증 결과

### TRUST 5원칙
- Test First: 95% ✅
- Readable: 100% ✅
- Unified: 90% ⚠️
- Secured: 85% ⚠️
- Trackable: 100% ✅

**전체 준수율**: 94%
```

---

## Step 3: 서버 통합 테스트

문서 동기화가 완료되었으므로 이제 실제 서버에서 통합 테스트를 수행합니다.

### 3.1 백엔드 및 프론트엔드 서버 실행

**터미널 1 (백엔드)**:
```bash
cd backend
source .venv/bin/activate
uvicorn app.main:app --reload --port 8000
```

**출력**:
```
INFO:     Will watch for changes in these directories: ['/Users/goos/MoAI/test/my-moai-project/backend']
INFO:     Uvicorn running on http://127.0.0.1:8000 (Press CTRL+C to quit)
INFO:     Started reloader process [12345] using WatchFiles
INFO:     Started server process [12346]
INFO:     Waiting for application startup.
INFO:     Application startup complete.
```

**터미널 2 (프론트엔드)**:
```bash
cd frontend
pnpm dev --port 5173
```

**출력**:
```
> my-moai-project-frontend@0.1.0 dev
> vite --port 5173

  VITE v7.1.9  ready in 512 ms

  ➜  Local:   http://localhost:5174/
  ➜  Network: use --host to expose
  ➜  press h + enter to show help
```

**주의**: 포트 5173이 사용 중이면 5174로 자동 변경됩니다.

### 3.2 Health Check

**백엔드 상태 확인**:
```bash
curl http://localhost:8000/
```

**응답**:
```json
{
  "status": "healthy",
  "message": "Todo API is running"
}
```

**API 문서 확인**:
브라우저에서 `http://127.0.0.1:8000/docs` 접속 → Swagger UI 확인

**프론트엔드 상태 확인**:
```bash
curl -s http://localhost:5174/ -o /dev/null -w "Frontend Status: %{http_code}\n"
```

**응답**:
```
Frontend Status: 200
```

### 3.3 CRUD 통합 테스트

#### 테스트 1: CREATE (생성)

**테스트 Todo 2개 생성**:
```bash
# Todo 1 생성
curl -X POST http://localhost:8000/api/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "테스트 Todo 1", "description": "백엔드 연결 테스트"}'

# Todo 2 생성
curl -X POST http://localhost:8000/api/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "테스트 Todo 2", "description": "CRUD 동작 확인"}'
```

**응답 (201 Created)**:
```json
{
  "id": 1,
  "title": "테스트 Todo 1",
  "description": "백엔드 연결 테스트",
  "completed": false,
  "created_at": "2025-10-11T12:13:37",
  "updated_at": "2025-10-11T12:13:37"
}
```

#### 테스트 2: READ (전체 조회)

```bash
curl -s http://localhost:8000/api/todos | python3 -m json.tool
```

**응답 (200 OK)**:
```json
[
  {
    "id": 1,
    "title": "테스트 Todo 1",
    "description": "백엔드 연결 테스트",
    "completed": false,
    "created_at": "2025-10-11T12:13:37",
    "updated_at": "2025-10-11T12:13:37"
  },
  {
    "id": 2,
    "title": "테스트 Todo 2",
    "description": "CRUD 동작 확인",
    "completed": false,
    "created_at": "2025-10-11T12:13:41",
    "updated_at": "2025-10-11T12:13:41"
  }
]
```

#### 테스트 3: READ (단일 조회)

```bash
curl -s http://localhost:8000/api/todos/1 | python3 -m json.tool
```

**응답 (200 OK)**:
```json
{
  "id": 1,
  "title": "테스트 Todo 1",
  "description": "백엔드 연결 테스트",
  "completed": false,
  "created_at": "2025-10-11T12:13:37",
  "updated_at": "2025-10-11T12:13:37"
}
```

#### 테스트 4: UPDATE (수정)

```bash
curl -X PATCH http://localhost:8000/api/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"title": "업데이트된 Todo", "completed": true}'
```

**응답 (200 OK)**:
```json
{
  "id": 1,
  "title": "업데이트된 Todo",
  "description": "백엔드 연결 테스트",
  "completed": true,
  "created_at": "2025-10-11T12:13:37",
  "updated_at": "2025-10-11T12:14:03"
}
```

**주목**: `updated_at`이 자동으로 갱신되었습니다! ✅

#### 테스트 5: DELETE (삭제)

```bash
curl -X DELETE http://localhost:8000/api/todos/2 -w "\nHTTP Status: %{http_code}\n"
```

**응답 (204 No Content)**:
```
HTTP Status: 204
```

**삭제 확인**:
```bash
curl -s http://localhost:8000/api/todos | python3 -m json.tool
```

**응답** (Todo 2가 사라짐):
```json
[
  {
    "id": 1,
    "title": "업데이트된 Todo",
    "description": "백엔드 연결 테스트",
    "completed": true,
    "created_at": "2025-10-11T12:13:37",
    "updated_at": "2025-10-11T12:14:03"
  }
]
```

### 3.4 CRUD 테스트 결과 요약

| 작업        | 엔드포인트              | 상태 | 결과                         |
|-------------|------------------------|------|------------------------------|
| CREATE      | POST /api/todos        | 201  | ✅ 2개 Todo 생성 성공         |
| READ (전체) | GET /api/todos         | 200  | ✅ 목록 조회 성공             |
| READ (단일) | GET /api/todos/1       | 200  | ✅ 단일 조회 성공             |
| UPDATE      | PATCH /api/todos/1     | 200  | ✅ 업데이트 성공 (제목 + 완료) |
| DELETE      | DELETE /api/todos/2    | 204  | ✅ 삭제 성공                 |

**모든 CRUD 작업 정상 동작 확인!** 🎉

---

## Step 4: 프론트엔드 UI 테스트

### 4.1 브라우저에서 접속

브라우저에서 `http://localhost:5174/` 접속

**화면 구성**:
1. **상단**: "Todo App - MoAI ADK" 제목
2. **입력 폼**:
   - Title 입력창
   - Description 입력창
   - "Add Todo" 버튼 (파란색)
3. **Todo 목록**:
   - 각 Todo 카드 (흰색 배경, 그림자)
   - 체크박스 (완료 토글)
   - 제목 (굵은 글씨)
   - 설명 (회색 텍스트)
   - 삭제 버튼 (빨간색)

### 4.2 실제 동작 테스트

**시나리오 1: Todo 추가**
1. Title: "프론트엔드 테스트"
2. Description: "UI 동작 확인"
3. "Add Todo" 클릭
4. **결과**: 목록에 새 Todo 카드 즉시 추가 ✅

**시나리오 2: Todo 완료 토글**
1. 첫 번째 Todo의 체크박스 클릭
2. **결과**:
   - 체크박스 체크됨
   - 제목에 취소선 표시
   - 회색 텍스트로 변경
   - 백엔드 API 호출 확인 (개발자 도구 Network 탭) ✅

**시나리오 3: Todo 삭제**
1. 두 번째 Todo의 삭제 버튼 클릭
2. **결과**:
   - 목록에서 즉시 제거
   - 백엔드 API 호출 확인 ✅

**시나리오 4: 데이터 영속성 확인**
1. 브라우저 새로고침 (F5)
2. **결과**:
   - 추가했던 Todo들이 그대로 표시됨
   - SQLite 데이터베이스에 저장된 데이터 로드됨 ✅

---

## Step 5: 최종 Git 상태 확인

### 5.1 Git 로그 확인

```bash
git log --oneline -7
```

**출력 (예시)**:
```
173255d 📝 DOCS: Living Document 동기화
be050e3 ✅ TYPE: mypy strict 모드 타입 힌트 추가
099a08a 🔒 SECURITY: .gitignore 추가
afd88e7 ♻️ REFACTOR: 프로젝트 문서 및 SPEC 추가
2ba5664 🟢 GREEN: Todo CRUD API + UI 구현
ce0278e 🔴 RED: Todo CRUD 테스트 작성
```

**TDD 사이클 커밋 완성** ✅:
- 🔴 RED: 테스트 작성
- 🟢 GREEN: 구현 완료
- ♻️ REFACTOR: 리팩토링
- 🔒 SECURITY: 보안 개선
- ✅ TYPE: 타입 안전성
- 📝 DOCS: 문서화

### 5.2 Working Tree 상태

```bash
git status
```

**출력**:
```
On branch main
nothing to commit, working tree clean
```

**완벽한 상태!** ✅

---

## Step 6: 품질 검증 최종 확인

### 6.1 테스트 커버리지

**백엔드 테스트**:
```bash
cd backend
pytest tests/ -v --cov=app --cov-report=term-missing
```

**결과**:
```
============================== test session starts ==============================
tests/test_todos.py::TestTodoCreate::test_create_todo_success PASSED     [  7%]
tests/test_todos.py::TestTodoCreate::test_create_todo_missing_title PASSED [ 14%]
tests/test_todos.py::TestTodoRead::test_get_todos_empty PASSED           [ 21%]
tests/test_todos.py::TestTodoRead::test_get_todos_multiple PASSED        [ 28%]
tests/test_todos.py::TestTodoRead::test_get_todo_by_id PASSED            [ 35%]
tests/test_todos.py::TestTodoRead::test_get_todo_not_found PASSED        [ 42%]
tests/test_todos.py::TestTodoUpdate::test_update_todo_title PASSED       [ 50%]
tests/test_todos.py::TestTodoUpdate::test_update_todo_completed PASSED   [ 57%]
tests/test_todos.py::TestTodoUpdate::test_update_todo_all_fields PASSED  [ 64%]
tests/test_todos.py::TestTodoUpdate::test_update_todo_not_found PASSED   [ 71%]
tests/test_todos.py::TestTodoUpdate::test_update_todo_partial PASSED     [ 78%]
tests/test_todos.py::TestTodoDelete::test_delete_todo_success PASSED     [ 85%]
tests/test_todos.py::TestTodoDelete::test_delete_todo_not_found PASSED   [ 92%]
tests/test_todos.py::TestTodoDelete::test_delete_todo_verify PASSED      [100%]

---------- coverage: platform darwin, python 3.13.1-final-0 -----------
Name                              Stmts   Miss  Cover   Missing
---------------------------------------------------------------
app/__init__.py                       1      0   100%
app/database.py                      12      0   100%
app/models/__init__.py                1      0   100%
app/models/todo.py                   10      0   100%
app/routers/__init__.py               0      0   100%
app/routers/todos.py                 48      2    96%   67, 89
app/schemas/__init__.py               1      0   100%
app/schemas/todo.py                   6      0   100%
app/services/__init__.py              1      0   100%
app/services/todo_service.py         36      0   100%
---------------------------------------------------------------
TOTAL                               116      2    95%

============================== 14 passed in 0.16s ===============================
```

**커버리지**: 95% ✅ (목표 85% 초과)

### 6.2 @TAG 추적성

```bash
# TAG 통계 확인
rg '@(SPEC|TEST|CODE|DOC):TODO-001' -c .moai/specs/ tests/ backend/ frontend/ | \
  awk -F: '{sum+=$2} END {print "총 TAG 수:", sum}'
```

**출력**:
```
총 TAG 수: 70
```

**TAG 분포**:
- @SPEC:TODO-001: 3개 (spec.md, plan.md, acceptance.md)
- @TEST:TODO-001: 5개 (test_todos.py, conftest.py)
- @CODE:TODO-001: 61개 (18개 파일)
- @DOC:TODO-001: 1개 (README.md)

**TAG 무결성**: 100% ✅ (고아 TAG 없음)

### 6.3 TRUST 5원칙 최종 점수

| 원칙            | 점수   | 상태 | 달성 내역                                  |
|----------------|--------|------|--------------------------------------------|
| T (Test First) | 95%    | ✅   | - 테스트 커버리지 95%<br>- 14개 테스트 통과<br>- pytest + httpx 통합 테스트 |
| R (Readable)   | 100%   | ✅   | - 최대 파일 크기: 94 LOC (목표 ≤300)<br>- 최대 함수 크기: 18 LOC (목표 ≤50)<br>- 의도 드러내는 네이밍 |
| U (Unified)    | 90%    | ⚠️   | - TypeScript strict 모드 활성화<br>- mypy strict 모드 통과<br>- 레이어 분리 (모델/서비스/라우터) |
| S (Secured)    | 85%    | ⚠️   | - Pydantic 검증<br>- SQL 인젝션 방지 (ORM)<br>- CORS 정책 적용 |
| T (Trackable)  | 100%   | ✅   | - 27개 파일 @TAG 적용<br>- TAG 체인 무결성 100%<br>- Git 커밋 추적 가능 |

**전체 준수율**: 94% ✅ (목표 85% 초과)

---

## Step 7: 다음 단계 계획

### 7.1 완료된 SPEC

| SPEC ID  | 제목              | 버전   | 상태        | 완료일     |
|----------|-------------------|--------|-------------|-----------|
| TODO-001 | Todo 항목 CRUD 기능 | v0.1.0 | ✅ completed | 2025-10-11 |

### 7.2 다음 SPEC 후보 (product.md 기준)

**옵션 1: 기능 확장 (권장)**

1. **TODO-004: Todo 필터링 및 검색 기능**
   - 우선순위: High
   - 설명: 완료/미완료 필터, 제목/설명 검색
   - 예상 작업 시간: 3-4시간

2. **TODO-002: Todo 상태 관리 (3단계)**
   - 우선순위: High
   - 설명: pending/in_progress/completed 상태
   - 예상 작업 시간: 4-5시간 (DB 마이그레이션 포함)

3. **TODO-003: Todo 우선순위 설정**
   - 우선순위: Medium
   - 설명: high/medium/low 우선순위
   - 예상 작업 시간: 2-3시간

**옵션 2: 배포 환경 구축 (학습 목표)**

**TODO-006: Docker Compose 배포 환경**
- 우선순위: High
- 설명: 프론트엔드, 백엔드, DB를 Docker로 통합 배포
- 예상 작업 시간: 3-4시간

**옵션 3: 품질 개선**

**프론트엔드 테스트 작성 (Vitest)**
- TRUST 원칙 100% 달성
- 컴포넌트 테스트 (TodoList, TodoForm, TodoItem)
- 예상 작업 시간: 2-3시간

### 7.3 다음 SPEC 시작 방법

**방법 1: 슬래시 커맨드 사용**
```bash
/alfred:1-spec "Todo 필터링 및 검색 기능"
```

**방법 2: 순차 실행**
```bash
/alfred:1-spec "Todo 필터링 및 검색 기능"
# SPEC 작성 완료 후
/alfred:2-build TODO-004
# TDD 구현 완료 후
/alfred:3-sync
```

---

## 🎉 축하합니다!

**Todo CRUD 애플리케이션이 완벽하게 구현되었습니다!**

### 달성 사항

✅ **TDD 방법론 완벽 적용** (RED-GREEN-REFACTOR)
✅ **최신 안정 버전** (2025-10-11 기준)
✅ **높은 테스트 커버리지** (95%)
✅ **완벽한 TAG 추적성** (70개 TAG, 27개 파일)
✅ **Living Document 생성** (자동 API 문서화)
✅ **TRUST 5원칙 94% 준수**
✅ **서버 통합 테스트 통과** (백엔드 + 프론트엔드)

### 학습 포인트

1. **SPEC-First TDD**: 명세 없이는 코드 없다
2. **Alfred SuperAgent**: 9개 전문 에이전트 오케스트레이션
3. **@TAG 시스템**: @SPEC → @TEST → @CODE → @DOC 추적성
4. **Living Document**: 자동 문서 생성 및 동기화
5. **Git 워크플로우**: 의미 있는 커밋 메시지 (TDD 사이클)

### 프로젝트 현황

```
my-moai-project/
├── backend/                    # Python + FastAPI
│   ├── app/
│   │   ├── models/            # SQLAlchemy 모델
│   │   ├── schemas/           # Pydantic 스키마
│   │   ├── services/          # 비즈니스 로직
│   │   ├── routers/           # API 엔드포인트
│   │   └── main.py
│   ├── tests/                 # pytest 테스트 (14개)
│   ├── alembic/               # DB 마이그레이션
│   └── requirements.txt
│
├── frontend/                   # TypeScript + React
│   ├── src/
│   │   ├── components/        # React 컴포넌트
│   │   ├── services/          # API 클라이언트
│   │   ├── types/             # TypeScript 타입
│   │   └── App.tsx
│   └── package.json
│
└── .moai/
    ├── specs/SPEC-TODO-001/   # SPEC 문서
    │   ├── spec.md            # EARS 명세
    │   ├── plan.md            # 구현 계획
    │   ├── acceptance.md      # 수락 기준
    │   └── README.md          # Living Document
    ├── reports/               # 동기화 보고서
    └── project/               # 프로젝트 메타데이터
```

### 서버 실행 방법 (요약)

**백엔드**:
```bash
cd backend
source .venv/bin/activate
uvicorn app.main:app --reload
# http://localhost:8000
# http://localhost:8000/docs (Swagger UI)
```

**프론트엔드**:
```bash
cd frontend
pnpm dev
# http://localhost:5174/
```

**테스트**:
```bash
cd backend
pytest tests/ -v --cov=app
# 14 passed, 95% coverage
```

---

## 📚 다음 학습

### 튜토리얼 시리즈

- ✅ [Part 1: 프로젝트 초기화](./01-project-init.md)
- ✅ [Part 2: SPEC 작성하기](./02-spec-writing.md)
- ✅ [Part 3: Backend TDD 구현](./03-backend-tdd.md)
- ✅ [Part 4: Frontend 구현하기](./04-frontend-impl.md)
- ✅ **Part 5: 문서 동기화 및 통합 테스트** (현재)

### 고급 학습

1. **워크플로우 심화**
   - [alfred:1-spec 가이드](../../workflow/1-spec.md)
   - [alfred:2-build 가이드](../../workflow/2-build.md)
   - [alfred:3-sync 가이드](../../workflow/3-sync.md)

2. **핵심 개념**
   - [SPEC-First TDD](../../concepts/spec-first-tdd.md)
   - [EARS 요구사항 작성법](../../concepts/ears-guide.md)
   - [TAG 시스템](../../concepts/tag-system.md)
   - [TRUST 5원칙](../../concepts/trust-principles.md)

3. **에이전트 활용**
   - [Alfred 에이전트 개요](../../agents/overview.md)
   - [debug-helper 사용법](../../agents/overview.md#debug-helper-트러블슈팅-전문가)
   - [trust-checker 사용법](../../agents/overview.md#trust-checker-품질-보증-리드)

---

## 📖 부록: 문제 해결

### 서버가 실행되지 않을 때

**문제**: `uvicorn: command not found`
```bash
# 해결: 가상 환경 활성화 확인
cd backend
source .venv/bin/activate
which uvicorn  # 경로 확인
```

**문제**: `Port 8000 already in use`
```bash
# 해결: 기존 프로세스 종료
lsof -ti:8000 | xargs kill
```

**문제**: `pnpm: command not found`
```bash
# 해결: pnpm 설치
npm install -g pnpm
```

### 데이터베이스 초기화

```bash
cd backend

# 데이터베이스 삭제 및 재생성
rm todos.db

# 마이그레이션 재실행
.venv/bin/alembic upgrade head
```

### TAG 검증 실패 시

```bash
# TAG 중복 확인
rg "@SPEC:TODO-001" -n .moai/specs/ backend/ frontend/

# TAG 체인 무결성 확인
/alfred:3-sync --check
```

---

## 🎉 튜토리얼 완료

축하합니다! MoAI-ADK의 핵심 워크플로우를 모두 경험했습니다:

1. 프로젝트 초기화 (`/alfred:0-project`)
2. SPEC 작성 (`/alfred:1-spec`)
3. TDD 구현 (`/alfred:2-build`)
4. 문서 동기화 (`/alfred:3-sync`)

다음 프로젝트에서 MoAI-ADK를 활용해보세요! 🚀

**이전**: [Part 4: Frontend 구현하기](./04-frontend-impl.md)

**처음으로**: [튜토리얼 개요](./index.md)
