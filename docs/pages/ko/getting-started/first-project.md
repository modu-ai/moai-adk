---
title: "첫 프로젝트 상세 가이드"
description: "MoAI-ADK로 첫 프로젝트를 완전히 설정하는 상세 가이드 - SPEC 작성부터 배포까지 전체 과정"
---

# 첫 프로젝트 상세 가이드

이 가이드는 MoAI-ADK로 첫 프로젝트를 처음부터 끝까지 완성하는 과정을 단계별로 안내합니다.

## 🎯 프로젝트 개요

만들 프로젝트: **간단한 To-Do 애플리케이션**

- 기능: 할 일 추가, 완료, 삭제, 목록 조회
- 기술 스택: Python + SQLite + REST API
- 학습 목표: SPEC-First 개발, TDD, 자동화된 문서화

## 1단계: 프로젝트 생성 및 설정

### 1.1 새 프로젝트 생성

```bash
# 1. 프로젝트 생성
moai-adk init todo-app
cd todo-app

# 2. Claude Code 실행
claude-code .
```

### 1.2 프로젝트 설정

Claude Code에서 다음 명령을 실행:

```bash
/alfred:0-project
```

Alfred가 다음 설정을 안내합니다:

**프로젝트 정보:**
- 이름: todo-app
- 설명: 간단한 To-Do 관리 애플리케이션
- 모드: personal
- 언어: python
- 로케일: ko

**Git 전략:**
- 브랜치 전략: GitFlow
- feature 접두사: feature/SPEC-
- develop 브랜치: develop
- main 브랜치: main

**품질 설정:**
- 테스트 커버리지 목표: 85%
- TDD 강제: 활성화
- @TAG 시스템: 활성화

## 2단계: 핵심 기능 SPEC 작성

### 2.1 사용자 관리 기능

```bash
/alfred:1-plan "사용자 인증 및 회원가입 시스템"
```

Alfred는 다음과 같은 구조화된 SPEC을 생성합니다:

**요구사항 분석:**
- 회원가입: 이메일, 비밀번호, 사용자명
- 로그인: JWT 토큰 기반 인증
- 비밀번호 복구: 이메일 인증
- 프로필 관리: 정보 수정, 삭제

**생성된 파일:**
```
.moai/specs/SPEC-AUTH-001/
├── spec.md          # 상세 요구사항
├── plan.md          # 구현 계획
├── acceptance.md    # 인수 테스트 기준
└── research/        # 관련 연구 자료
```

### 2.2 To-Do 핵심 기능

```bash
/alfred:1-plan "To-Do CRUD 기능 구현"
```

**요구사항 정의:**
- 할 일 생성: 제목, 설명, 우선순위, 마감일
- 할 일 조회: 전체 목록, 필터링, 검색
- 할 일 수정: 상태 변경, 내용 수정
- 할 일 삭제: 소프트 삭제, 영구 삭제

## 3단계: TDD 기반 구현

### 3.1 사용자 인증 구현

```bash
/alfred:2-run AUTH-001
```

Alfred가 자동으로 TDD 사이클을 실행합니다:

**1. RED 단계 - 실패하는 테스트 작성:**
```python
# tests/test_auth_service.py
import pytest
from src.auth_service import AuthService

class TestUserService:
    def test_user_registration_success(self):
        service = AuthService()
        result = service.register_user(
            email="test@example.com",
            password="SecurePass123!",
            username="testuser"
        )
        assert result.success is True
        assert result.user_id is not None

    def test_duplicate_email_registration(self):
        service = AuthService()
        # 첫 사용자 등록
        service.register_user("test@example.com", "pass123", "user1")
        # 중복 이메일 등록 시도
        result = service.register_user("test@example.com", "pass456", "user2")
        assert result.success is False
        assert "already exists" in result.error_message
```

**2. GREEN 단계 - 최소 구현:**
```python
# src/auth_service.py
import hashlib
import re
from typing import Optional
from dataclasses import dataclass

@dataclass
class RegistrationResult:
    success: bool
    user_id: Optional[int] = None
    error_message: Optional[str] = None

class AuthService:
    def __init__(self):
        self.users = {}  # 임시 저장소 (실제로는 데이터베이스)
        self.next_id = 1

    def register_user(self, email: str, password: str, username: str) -> RegistrationResult:
        # 이메일 유효성 검사
        email_pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
        if not re.match(email_pattern, email):
            return RegistrationResult(success=False, error_message="Invalid email format")

        # 중복 이메일 확인
        if email in self.users:
            return RegistrationResult(success=False, error_message="Email already exists")

        # 비밀번호 강도 검사
        if len(password) < 8:
            return RegistrationResult(success=False, error_message="Password too short")

        # 사용자 생성
        user_id = self.next_id
        self.next_id += 1

        self.users[email] = {
            'id': user_id,
            'email': email,
            'username': username,
            'password_hash': hashlib.sha256(password.encode()).hexdigest(),
            'created_at': '2025-01-01'
        }

        return RegistrationResult(success=True, user_id=user_id)
```

**3. REFACTOR 단계 - 코드 개선:**
```python
# src/auth_service.py (개선된 버전)
import hashlib
import re
import uuid
from datetime import datetime
from typing import Optional, Dict, Any
from dataclasses import dataclass
from abc import ABC, abstractmethod

@dataclass
class RegistrationResult:
    success: bool
    user_id: Optional[str] = None
    error_message: Optional[str] = None
    user_data: Optional[Dict[str, Any]] = None

class PasswordValidator:
    @staticmethod
    def validate(password: str) -> tuple[bool, str]:
        """비밀번호 강도 검사"""
        if len(password) < 8:
            return False, "Password must be at least 8 characters"

        if not re.search(r'[A-Z]', password):
            return False, "Password must contain uppercase letter"

        if not re.search(r'[a-z]', password):
            return False, "Password must contain lowercase letter"

        if not re.search(r'\d', password):
            return False, "Password must contain number"

        return True, "Password is valid"

class UserRepository:
    def __init__(self):
        self.users: Dict[str, Dict[str, Any]] = {}

    def email_exists(self, email: str) -> bool:
        return email in self.users

    def save_user(self, user_data: Dict[str, Any]) -> str:
        user_id = str(uuid.uuid4())
        user_data['id'] = user_id
        user_data['created_at'] = datetime.utcnow().isoformat()
        self.users[user_data['email']] = user_data
        return user_id

class AuthService:
    def __init__(self, user_repository: UserRepository = None):
        self.user_repo = user_repository or UserRepository()
        self.password_validator = PasswordValidator()

    def register_user(self, email: str, password: str, username: str) -> RegistrationResult:
        # 이메일 유효성 검사
        if not self._validate_email(email):
            return RegistrationResult(success=False, error_message="Invalid email format")

        # 중복 확인
        if self.user_repo.email_exists(email):
            return RegistrationResult(success=False, error_message="Email already exists")

        # 비밀번호 검사
        is_valid, error_msg = self.password_validator.validate(password)
        if not is_valid:
            return RegistrationResult(success=False, error_message=error_msg)

        # 사용자 생성
        user_data = {
            'email': email,
            'username': username,
            'password_hash': hashlib.sha256(password.encode()).hexdigest(),
        }

        user_id = self.user_repo.save_user(user_data)

        return RegistrationResult(
            success=True,
            user_id=user_id,
            user_data=user_data
        )

    def _validate_email(self, email: str) -> bool:
        pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
        return re.match(pattern, email) is not None
```

### 3.2 To-Do 기능 구현

```bash
/alfred:2-run TODO-001
```

Alfred가 자동으로 생성하는 구조:

```
src/
├── auth_service.py       # 인증 서비스
├── todo_service.py       # To-Do 서비스
├── models/
│   ├── user.py          # 사용자 모델
│   └── todo.py          # To-Do 모델
├── database/
│   ├── connection.py    # 데이터베이스 연결
│   └── migrations/      # 마이그레이션
└── api/
    ├── auth.py          # 인증 API
    └── todos.py         # To-Do API

tests/
├── test_auth_service.py
├── test_todo_service.py
├── test_models/
└── test_api/
```

## 4단계: API 엔드포인트 개발

### 4.1 FastAPI 설정

```bash
/alfred:1-plan "FastAPI 기반 REST API 개발"
/alfred:2-run API-001
```

Alfred가 자동으로 생성하는 API 구조:

```python
# src/main.py
from fastapi import FastAPI, Depends, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from src.api.auth import auth_router
from src.api.todos import todo_router
from src.auth_service import AuthService

app = FastAPI(title="Todo App API", version="1.0.0")

# CORS 설정
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 라우터 등록
app.include_router(auth_router, prefix="/auth", tags=["authentication"])
app.include_router(todo_router, prefix="/todos", tags=["todos"])

@app.get("/")
async def root():
    return {"message": "Todo App API", "version": "1.0.0"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
```

### 4.2 API 엔드포인트 구현

```python
# src/api/todos.py
from fastapi import APIRouter, Depends, HTTPException
from typing import List, Optional
from src.todo_service import TodoService, TodoItem, CreateTodoRequest
from src.auth_service import get_current_user

todo_router = APIRouter()

@todo_router.get("/", response_model=List[TodoItem])
async def get_todos(
    user_id: str = Depends(get_current_user),
    status: Optional[str] = None,
    priority: Optional[str] = None
):
    todo_service = TodoService()
    return todo_service.get_user_todos(
        user_id=user_id,
        status=status,
        priority=priority
    )

@todo_router.post("/", response_model=TodoItem)
async def create_todo(
    request: CreateTodoRequest,
    user_id: str = Depends(get_current_user)
):
    todo_service = TodoService()
    result = todo_service.create_todo(
        user_id=user_id,
        title=request.title,
        description=request.description,
        priority=request.priority,
        due_date=request.due_date
    )

    if not result.success:
        raise HTTPException(status_code=400, detail=result.error_message)

    return result.todo
```

## 5단계: 프론트엔드 통합

### 5.1 React 프론트엔드 설정

```bash
/alfred:1-plan "React 기반 프론트엔드 개발"
/alfred:2-run FRONTEND-001
```

Alfred가 생성하는 프론트엔드 구조:

```
frontend/
├── public/
├── src/
│   ├── components/
│   │   ├── Auth/
│   │   │   ├── Login.jsx
│   │   │   └── Register.jsx
│   │   ├── Todo/
│   │   │   ├── TodoList.jsx
│   │   │   ├── TodoItem.jsx
│   │   │   └── TodoForm.jsx
│   │   └── Common/
│   ├── services/
│   │   ├── api.js
│   │   └── auth.js
│   ├── hooks/
│   │   ├── useAuth.js
│   │   └── useTodos.js
│   └── utils/
├── package.json
└── README.md
```

### 5.2 주요 컴포넌트 구현

```jsx
// frontend/src/components/Todo/TodoList.jsx
import React, { useState, useEffect } from 'react';
import { useTodos } from '../../hooks/useTodos';
import TodoItem from './TodoItem';
import TodoForm from './TodoForm';

const TodoList = () => {
    const { todos, loading, error, createTodo, updateTodo, deleteTodo } = useTodos();
    const [filter, setFilter] = useState('all');

    const filteredTodos = todos.filter(todo => {
        switch (filter) {
            case 'active':
                return !todo.completed;
            case 'completed':
                return todo.completed;
            default:
                return true;
        }
    });

    if (loading) return <div>로딩 중...</div>;
    if (error) return <div>오류: {error}</div>;

    return (
        <div className="todo-list">
            <h2>할 일 목록</h2>

            <TodoForm onSubmit={createTodo} />

            <div className="filters">
                <button
                    className={filter === 'all' ? 'active' : ''}
                    onClick={() => setFilter('all')}
                >
                    전체
                </button>
                <button
                    className={filter === 'active' ? 'active' : ''}
                    onClick={() => setFilter('active')}
                >
                    진행 중
                </button>
                <button
                    className={filter === 'completed' ? 'active' : ''}
                    onClick={() => setFilter('completed')}
                >
                    완료됨
                </button>
            </div>

            <div className="todo-items">
                {filteredTodos.map(todo => (
                    <TodoItem
                        key={todo.id}
                        todo={todo}
                        onUpdate={updateTodo}
                        onDelete={deleteTodo}
                    />
                ))}
            </div>
        </div>
    );
};

export default TodoList;
```

## 6단계: 테스트 및 품질 보증

### 6.1 통합 테스트

```bash
/alfred:1-plan "통합 테스트 스위트 개발"
/alfred:2-run INTEGRATION-001
```

Alfred가 생성하는 테스트:

```python
# tests/integration/test_api_integration.py
import pytest
from fastapi.testclient import TestClient
from src.main import app

client = TestClient(app)

class TestTodoAPIIntegration:
    def test_complete_todo_workflow(self):
        # 1. 사용자 등록
        register_response = client.post("/auth/register", json={
            "email": "test@example.com",
            "password": "SecurePass123!",
            "username": "testuser"
        })
        assert register_response.status_code == 201

        # 2. 로그인
        login_response = client.post("/auth/login", json={
            "email": "test@example.com",
            "password": "SecurePass123!"
        })
        assert login_response.status_code == 200
        token = login_response.json()["access_token"]

        # 3. 인증 헤더 설정
        headers = {"Authorization": f"Bearer {token}"}

        # 4. To-Do 생성
        todo_response = client.post("/todos/", json={
            "title": "테스트 할 일",
            "description": "테스트 설명",
            "priority": "high"
        }, headers=headers)
        assert todo_response.status_code == 201
        todo_id = todo_response.json()["id"]

        # 5. To-Do 목록 조회
        list_response = client.get("/todos/", headers=headers)
        assert list_response.status_code == 200
        todos = list_response.json()
        assert len(todos) == 1
        assert todos[0]["title"] == "테스트 할 일"

        # 6. To-Do 완료 처리
        complete_response = client.patch(
            f"/todos/{todo_id}",
            json={"status": "completed"},
            headers=headers
        )
        assert complete_response.status_code == 200

        # 7. To-Do 삭제
        delete_response = client.delete(f"/todos/{todo_id}", headers=headers)
        assert delete_response.status_code == 204
```

### 6.2 성능 테스트

```bash
/alfred:1-plan "성능 테스트 및 최적화"
/alfred:2-run PERF-001
```

## 7단계: 배포 준비

### 7.1 Docker 컨테이너화

```bash
/alfred:1-plan "Docker 컨테이너화 설정"
/alfred:2-run DOCKER-001
```

Alfred가 생성하는 Docker 파일들:

```dockerfile
# Dockerfile
FROM python:3.11-slim

WORKDIR /app

# 의존성 설치
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# 소스 코드 복사
COPY src/ ./src/
COPY tests/ ./tests/

# 애플리케이션 실행
EXPOSE 8000
CMD ["uvicorn", "src.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

### 7.2 CI/CD 파이프라인

```bash
/alfred:1-plan "GitHub Actions CI/CD 파이프라인"
/alfred:2-run CICD-001
```

생성된 워크플로우:

```yaml
# .github/workflows/ci-cd.yml
name: CI/CD Pipeline

on:
  push:
    branches: [develop, main]
  pull_request:
    branches: [develop]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        python-version: [3.9, 3.10, 3.11]

    steps:
    - uses: actions/checkout@v4

    - name: Set up Python ${{ matrix.python-version }}
      uses: actions/setup-python@v4
      with:
        python-version: ${{ matrix.python-version }}

    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install -r requirements.txt

    - name: Run tests with coverage
      run: |
        pytest --cov=src --cov-report=xml

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.xml

  security:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Run security scan
      run: |
        pip install safety bandit
        safety check
        bandit -r src/
```

## 8단계: 문서 자동 생성

### 8.1 API 문서 생성

```bash
/alfred:3-sync
```

Alfred가 자동으로 생성하는 문서들:

```
docs/
├── api/
│   ├── authentication.md
│   ├── todos.md
│   └── openapi.json
├── guides/
│   ├── getting-started.md
│   ├── deployment.md
│   └── troubleshooting.md
└── architecture/
    ├── system-design.md
    ├── database-schema.md
    └── api-design.md
```

### 8.2 OpenAPI 스펙

```json
{
  "openapi": "3.0.0",
  "info": {
    "title": "Todo App API",
    "version": "1.0.0",
    "description": "간단한 To-Do 관리 애플리케이션 API"
  },
  "paths": {
    "/todos": {
      "get": {
        "summary": "할 일 목록 조회",
        "parameters": [
          {
            "name": "status",
            "in": "query",
            "schema": {"type": "string", "enum": ["active", "completed"]}
          }
        ],
        "responses": {
          "200": {
            "description": "성공",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {"$ref": "#/components/schemas/TodoItem"}
                }
              }
            }
          }
        }
      }
    }
  }
}
```

## 최종 결과물

프로젝트 완료 시 다음을 얻게 됩니다:

### ✅ 생성된 파일 구조

```
todo-app/
├── .claude/                  # Alfred 에이전트 및 스킬
├── .moai/                    # MoAI-ADK 설정 및 SPEC
│   ├── config.json           # 프로젝트 설정
│   ├── specs/                # 모든 SPEC 문서
│   └── reports/              # 자동 생성 리포트
├── src/                      # 소스 코드
├── tests/                    # 테스트 스위트
├── docs/                     # 자동 생성 문서
├── frontend/                 # React 프론트엔드
├── docker-compose.yml        # Docker 설정
├── .github/workflows/        # CI/CD 파이프라인
└── README.md                 # 프로젝트 문서
```

### ✅ 품질 지표

- **테스트 커버리지**: 92.3%
- **코드 품질**: TRUST 5 원칙 준수
- **문서화**: 100% 자동 생성 및 동기화
- **보안**: 자동 보안 스캔 및 취약점 검출
- **성능**: 자동 성능 테스트 및 최적화

### ✅ @TAG 추적성

```
@SPEC-AUTH-001 (사용자 인증 시스템)
    ↓
@TEST-AUTH-001 (인증 테스트 스위트)
    ↓
@CODE-AUTH-001:SERVICE (인증 서비스 구현)
    ↓
@DOC-AUTH-001 (API 문서)

@SPEC-TODO-001 (To-Do CRUD 기능)
    ↓
@TEST-TODO-001 (To-Do 테스트 스위트)
    ↓
@CODE-TODO-001:SERVICE (To-Do 서비스 구현)
    ↓
@DOC-TODO-001 (To-Do API 문서)
```

## 다음 단계

프로젝트를 완료했습니다! 이제 다음을 할 수 있습니다:

1. **배포**: Vercel, Railway, AWS 등에 배포
2. **확장**: 새 기능 추가 및 개선
3. **최적화**: 성능 최적화 및 모니터링 설정
4. **팀 협업**: 팀원 초대 및 협업 워크플로우 설정

## 추가 리소스

- **[배포 가이드](../guides/deployment)**: 다양한 배플랫폼 배포 방법
- **[Alfred 고급 활용](../guides/alfred)**: Alfred 슈퍼에이전트 심화 기능
- **[문제 해결](../troubleshooting)**: 일반적인 문제 해결

---

🎉 **축하합니다!** MoAI-ADK로 첫 번째 완전한 프로젝트를 성공적으로 완성했습니다. 이제 신뢰할 수 있고 문서화된 소프트웨어를 생산적으로 개발할 준비가 되었습니다.