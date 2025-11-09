Translate the following Korean markdown document to Japanese.

**CRITICAL RULES:**
1. Preserve ALL markdown structure (headers, code blocks, links, tables, diagrams)
2. Keep ALL code blocks and technical terms UNCHANGED
3. Maintain the EXACT same file structure and formatting
4. Translate ONLY Korean text content
5. Keep ALL @TAG references unchanged (e.g., @SPEC:AUTH-001)
6. Preserve ALL file paths and URLs
7. Keep ALL emoji and icons as-is
8. Maintain ALL frontmatter (YAML) structure

**Source File:** /Users/goos/MoAI/MoAI-ADK/docs/src/ko/reference/agents/experts.md
**Target Language:** Japanese
**Target File:** /Users/goos/MoAI/MoAI-ADK/docs/src/ja/reference/agents/experts.md

**Content to Translate:**

# 전문가 Agents 상세 가이드

Alfred의 6명 도메인 전문가에 대한 완전한 참고서입니다.

## 개요

| #   | 전문가          | 도메인                | 활성화 키워드                       | 스킬 수 |
| --- | --------------- | --------------------- | ----------------------------------- | ------- |
| 1   | backend-expert  | API, 서버, DB         | server, api, database, microservice | 12개    |
| 2   | frontend-expert | UI, 상태관리, 성능    | frontend, ui, component, state      | 10개    |
| 3   | devops-expert   | 배포, CI/CD, 인프라   | deploy, docker, kubernetes, ci/cd   | 14개    |
| 4   | ui-ux-expert    | 디자인 시스템, 접근성 | design, ux, accessibility, figma    | 8개     |
| 5   | security-expert | 보안, 인증            | security, auth, encryption, owasp   | 11개    |
| 6   | database-expert | DB 설계, 최적화       | database, schema, query, index      | 9개     |

______________________________________________________________________

## 1. backend-expert

**도메인**: API, 서버, 데이터베이스 아키텍처

### 활성화 조건

다음 키워드가 SPEC에 포함되면 자동 활성화:

- `server`, `api`, `endpoint`, `microservice`
- `authentication`, `authorization`
- `database`, `ORM`

### 전문 분야

| 영역               | 기술 스택               | 책임                         |
| ------------------ | ----------------------- | ---------------------------- |
| **API 설계**       | REST, GraphQL           | OpenAPI 3.1 명세 작성        |
| **프레임워크**     | FastAPI, Flask, Django  | 프레임워크 선택 및 구조 설계 |
| **인증**           | JWT, OAuth 2.0, Session | 안전한 인증 시스템 구현      |
| **마이크로서비스** | Celery, RabbitMQ        | 비동기 작업 처리             |
| **캐싱**           | Redis, Memcached        | 성능 최적화                  |

### 주요 책임

1. **API 설계**

   - RESTful 원칙 준수
   - 엔드포인트 구조 설계
   - 요청/응답 스키마 정의
   - 에러 처리 전략

2. **데이터 모델링**

   - Entity-Relationship 다이어그램
   - ORM 모델 설계
   - 관계 설정 (1:1, 1:N, N:N)
   - 인덱싱 전략

3. **성능 최적화**

   - 쿼리 최적화
   - 데이터베이스 인덱싱
   - 캐싱 전략
   - 로드 밸런싱

### 예시: REST API 설계

```python
# @CODE:SPEC-002:backend-design
from fastapi import FastAPI, HTTPException, Depends
from sqlalchemy.orm import Session

app = FastAPI(title="Todo API v1.0")

# 엔드포인트 설계
@app.post("/api/v1/todos", status_code=201)
async def create_todo(
    title: str,
    description: str,
    db: Session = Depends(get_db)
):
    """할 일 생성"""
    todo = Todo(title=title, description=description)
    db.add(todo)
    db.commit()
    return todo

@app.get("/api/v1/todos/{todo_id}")
async def get_todo(todo_id: int, db: Session = Depends(get_db)):
    """할 일 조회"""
    todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404)
    return todo

@app.put("/api/v1/todos/{todo_id}")
async def update_todo(
    todo_id: int,
    title: str,
    description: str,
    db: Session = Depends(get_db)
):
    """할 일 수정"""
    todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404)
    todo.title = title
    todo.description = description
    db.commit()
    return todo

@app.delete("/api/v1/todos/{todo_id}")
async def delete_todo(todo_id: int, db: Session = Depends(get_db)):
    """할 일 삭제"""
    todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404)
    db.delete(todo)
    db.commit()
    return {"status": "deleted"}
```

### 생성 산출물

- OpenAPI 3.1 명세
- API 엔드포인트 리스트
- 요청/응답 스키마
- 에러 코드 문서
- 인증 플로우

______________________________________________________________________

## 2. frontend-expert

**도메인**: UI 컴포넌트, 상태 관리, 성능 최적화

### 활성화 조건

다음 키워드가 SPEC에 포함되면 자동 활성화:

- `frontend`, `ui`, `component`, `page`
- `state`, `store`, `context`
- `performance`, `optimization`

### 전문 분야

| 영역           | 기술 스택                     | 책임                    |
| -------------- | ----------------------------- | ----------------------- |
| **프레임워크** | React 19, Vue 3.5, Angular 19 | 프레임워크 선택 및 구조 |
| **상태관리**   | Redux, Zustand, Pinia         | 전역 상태 설계          |
| **컴포넌트**   | Composition, Hooks            | 재사용 가능한 컴포넌트  |
| **성능**       | 번들 최적화, 지연 로딩        | 렌더링 성능 개선        |
| **접근성**     | WCAG 2.2, ARIA                | 모든 사용자 지원        |

### 주요 책임

1. **컴포넌트 설계**

   - 재사용 가능한 컴포넌트 구조
   - Props 인터페이스 정의
   - 스타일링 전략 (CSS-in-JS, Tailwind)

2. **상태 관리**

   - 전역 상태 구조
   - 상태 업데이트 로직
   - 성능 최적화 (메모이제이션)

3. **성능 최적화**

   - 번들 크기 최소화
   - 렌더링 최적화
   - 이미지 최적화
   - 캐싱 전략

### 예시: React 컴포넌트 설계

```typescript
// @CODE:SPEC-003:frontend-component
import React, { useState, useCallback } from 'react';
import { useTodoStore } from './store';

// 상태 관리 (Zustand)
const useTodoStore = create((set) => ({
  todos: [],
  addTodo: (todo) => set((state) => ({
    todos: [...state.todos, todo]
  })),
  removeTodo: (id) => set((state) => ({
    todos: state.todos.filter(t => t.id !== id)
  }))
}));

// 재사용 가능한 컴포넌트
const TodoItem = React.memo(({ todo, onRemove }) => (
  <div className="todo-item">
    <h3>{todo.title}</h3>
    <p>{todo.description}</p>
    <button onClick={() => onRemove(todo.id)}>
      삭제
    </button>
  </div>
));

// 메인 컴포넌트
export const TodoList = () => {
  const [input, setInput] = useState('');
  const { todos, addTodo } = useTodoStore();

  const handleAdd = useCallback(() => {
    if (input.trim()) {
      addTodo({ id: Date.now(), title: input });
      setInput('');
    }
  }, [input]);

  return (
    <div>
      <input
        value={input}
        onChange={(e) => setInput(e.target.value)}
        placeholder="새 할 일 입력"
      />
      <button onClick={handleAdd}>추가</button>

      <div className="todo-list">
        {todos.map(todo => (
          <TodoItem
            key={todo.id}
            todo={todo}
            onRemove={() => /* remove */}
          />
        ))}
      </div>
    </div>
  );
};
```

### 생성 산출물

- 컴포넌트 트리 다이어그램
- Props 인터페이스 정의서
- 상태 관리 다이어그램
- 성능 최적화 보고서
- 접근성 검사 결과

______________________________________________________________________

## 3. devops-expert

**도메인**: 배포, CI/CD, 클라우드 인프라

### 활성화 조건

다음 키워드가 SPEC에 포함되면 자동 활성화:

- `deploy`, `deployment`, `ci/cd`
- `docker`, `kubernetes`
- `infrastructure`, `cloud`

### 전문 분야

| 영역               | 기술 스택                 | 책임                      |
| ------------------ | ------------------------- | ------------------------- |
| **컨테이너**       | Docker, Docker Compose    | Dockerfile 및 이미지 관리 |
| **오케스트레이션** | Kubernetes, Helm          | 배포 및 스케일링          |
| **CI/CD**          | GitHub Actions, GitLab CI | 자동화 파이프라인         |
| **클라우드**       | AWS, GCP, Azure           | 인프라 코드 작성          |
| **모니터링**       | Prometheus, Grafana       | 성능 모니터링             |

### 주요 책임

1. **배포 파이프라인 설계**

   - 테스트 → 빌드 → 배포 자동화
   - 카나리 배포 및 롤백 전략
   - 무중단 배포

2. **인프라 구성**

   - 프로덕션 환경 설정
   - 로드 밸런싱
   - 데이터베이스 백업/복구

3. **모니터링 & 로깅**

   - 애플리케이션 성능 모니터링
   - 로그 수집 및 분석
   - 알림 설정

### 예시: GitHub Actions CI/CD

```yaml
# @CODE:SPEC-004:devops-pipeline
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-python@v4
        with:
          python-version: '3.13'

      # 테스트
      - name: Install dependencies
        run: |
          python -m pip install --upgrade pip
          pip install -r requirements.txt

      - name: Run tests
        run: pytest --cov=src tests/

      - name: Check coverage
        run: |
          coverage report --fail-under=85

  deploy:
    needs: test
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      # 배포
      - name: Build Docker image
        run: docker build -t app:latest .

      - name: Deploy to production
        run: |
          docker tag app:latest app:${{ github.sha }}
          # 배포 스크립트 실행
          ./scripts/deploy.sh
```

### 생성 산출물

- Dockerfile
- docker-compose.yml
- Kubernetes manifests
- CI/CD 파이프라인
- 배포 가이드
- 모니터링 구성

______________________________________________________________________

## 4. ui-ux-expert

**도메인**: 디자인 시스템, 접근성, 사용자 경험

### 활성화 조건

다음 키워드가 SPEC에 포함되면 자동 활성화:

- `design`, `ui`, `ux`
- `accessibility`, `a11y`
- `figma`, `design-system`

### 전문 분야

| 영역              | 기술 스택         | 책임                |
| ----------------- | ----------------- | ------------------- |
| **디자인 시스템** | Figma, Storybook  | 컴포넌트 라이브러리 |
| **접근성**        | WCAG 2.2, ARIA    | 모든 사용자 포용    |
| **사용자 연구**   | 유저 테스트, 분석 | UX 개선             |
| **성능**          | 로딩 시간, 반응성 | 사용자 만족도       |

### 주요 책임

1. **디자인 시스템 구축**

   - 색상, 타이포그래피, 간격 정의
   - 컴포넌트 라이브러리
   - 디자인 토큰

2. **접근성 보장**

   - 스크린 리더 지원
   - 키보드 네비게이션
   - 색상 명도 대비

3. **사용자 경험 개선**

   - 사용자 테스트
   - 피드백 수집
   - 지속적 개선

### 예시: 접근성 체크리스트

```markdown
# WCAG 2.2 접근성 검사

## 인식성 (Perceivable)
- [ ] 이미지에 대체 텍스트 제공
- [ ] 색상만으로 정보 전달하지 않음
- [ ] 명도 대비 4.5:1 이상

## 운영성 (Operable)
- [ ] 모든 기능을 키보드로 작동
- [ ] 포커스 순서가 논리적
- [ ] 깜빡이는 콘텐츠 없음

## 이해성 (Understandable)
- [ ] 텍스트 읽기 가능성 높음
- [ ] 예측 가능한 네비게이션
- [ ] 오류 메시지 명확

## 견고성 (Robust)
- [ ] 유효한 HTML/CSS
- [ ] ARIA 올바른 사용
- [ ] 호환성 테스트 통과
```

### 생성 산출물

- 디자인 시스템 가이드
- Figma 컴포넌트 라이브러리
- Storybook 문서
- 접근성 감사 보고서
- 사용자 테스트 결과

______________________________________________________________________

## 5. security-expert

**도메인**: 보안, 인증, 암호화

### 활성화 조건

다음 키워드가 SPEC에 포함되면 자동 활성화:

- `security`, `auth`, `encryption`
- `vulnerability`, `owasp`
- `compliance`, `privacy`

### 전문 분야

| 영역          | 기술 스택            | 책임             |
| ------------- | -------------------- | ---------------- |
| **인증**      | JWT, OAuth 2.0, SAML | 보안 인증 시스템 |
| **암호화**    | AES-256, RSA, HTTPS  | 데이터 보호      |
| **OWASP**     | Top 10, SAST/DAST    | 취약점 방지      |
| **접근 제어** | RBAC, ABAC           | 권한 관리        |
| **감사**      | 로깅, 모니터링       | 보안 이벤트 추적 |

### 주요 책임

1. **보안 설계**

   - 위협 모델링
   - 보안 아키텍처
   - 침입 방지 전략

2. **취약점 방지**

   - SQL injection 방지
   - XSS 방지
   - CSRF 토큰
   - 입력 검증

3. **보안 감시**

   - 로그 분석
   - 침입 탐지
   - 사고 대응

### 예시: 보안 구현

```python
# @CODE:SPEC-005:security-implementation
from flask import Flask, request
from werkzeug.security import generate_password_hash, check_password_hash
from cryptography.fernet import Fernet
import jwt

app = Flask(__name__)
SECRET_KEY = "your-secret-key"

# 비밀번호 해싱
def hash_password(password: str) -> str:
    """비밀번호 안전하게 해싱"""
    return generate_password_hash(password, method='pbkdf2:sha256')

def verify_password(hashed, password: str) -> bool:
    """비밀번호 검증"""
    return check_password_hash(hashed, password)

# JWT 토큰
def create_token(user_id: int) -> str:
    """JWT 토큰 생성"""
    return jwt.encode(
        {'user_id': user_id},
        SECRET_KEY,
        algorithm='HS256'
    )

# 입력 검증
@app.before_request
def validate_input():
    """모든 입력 검증"""
    if request.method == 'POST':
        # CSRF 토큰 검증
        token = request.headers.get('X-CSRF-Token')
        if not verify_csrf_token(token):
            return {'error': 'Invalid CSRF token'}, 403

        # SQL injection 방지 (parameterized queries)
        # - SQLAlchemy ORM 사용

        # XSS 방지 (HTML escape)
        # - Jinja2 자동 이스케이프

# HTTPS 강제
@app.after_request
def secure_headers(response):
    """보안 헤더 설정"""
    response.headers['X-Content-Type-Options'] = 'nosniff'
    response.headers['X-Frame-Options'] = 'DENY'
    response.headers['X-XSS-Protection'] = '1; mode=block'
    response.headers['Strict-Transport-Security'] = 'max-age=31536000'
    return response
```

### 생성 산출물

- 보안 정책 문서
- 위협 모델링 다이어그램
- 보안 감사 체크리스트
- 침입 테스트 보고서
- 컴플라이언스 문서 (GDPR, HIPAA)

______________________________________________________________________

## 6. database-expert

**도메인**: 데이터베이스 설계, 최적화, 마이그레이션

### 활성화 조건

다음 키워드가 SPEC에 포함되면 자동 활성화:

- `database`, `db`, `schema`
- `query`, `index`, `migration`
- `optimization`, `performance`

### 전문 분야

| 영역             | 기술 스택                  | 책임               |
| ---------------- | -------------------------- | ------------------ |
| **설계**         | PostgreSQL, MySQL, MongoDB | 스키마 설계        |
| **최적화**       | 인덱싱, 쿼리 튜닝          | 성능 개선          |
| **마이그레이션** | Alembic, Flyway            | 버전 관리          |
| **확장성**       | 파티셔닝, 샤딩             | 대규모 데이터 처리 |
| **백업**         | PITR, 복제                 | 데이터 안전성      |

### 주요 책임

1. **데이터베이스 설계**

   - Entity-Relationship 다이어그램
   - 정규화 (1NF ~ 3NF)
   - 제약조건 설정

2. **성능 최적화**

   - 적절한 인덱스 생성
   - 쿼리 최적화
   - 실행 계획 분석

3. **마이그레이션 관리**

   - 버전 제어
   - 롤백 전략
   - 무중단 마이그레이션

### 예시: 데이터베이스 설계

```sql
-- @CODE:SPEC-006:database-schema
-- 사용자 테이블
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_email (email)
);

-- 할 일 테이블
CREATE TABLE todos (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    completed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_completed (user_id, completed)
);

-- 관계도
-- users (1) --< (N) todos
```

### 생성 산출물

- ERD (Entity-Relationship Diagram)
- DDL (Data Definition Language) 스크립트
- 마이그레이션 스크립트
- 성능 튜닝 보고서
- 백업/복구 계획

______________________________________________________________________

## 전문가 활성화 매트릭스

| SPEC 키워드 | backend | frontend | devops | ui-ux | security | database |
| ----------- | ------- | -------- | ------ | ----- | -------- | -------- |
| API         | ✅      |          |        |       |          |          |
| Frontend    |         | ✅       |        | ✅    |          |          |
| Database    | ✅      |          |        |       |          | ✅       |
| Deploy      |         |          | ✅     |       |          |          |
| Security    |         |          |        |       | ✅       |          |
| Performance | ✅      | ✅       |        | ✅    |          | ✅       |

______________________________________________________________________

**다음**: [핵심 Sub-agents](core.md) 또는 [Agents 개요](index.md)


**Instructions:**
- Translate the content above to Japanese
- Output ONLY the translated markdown content
- Do NOT include any explanations or comments
- Maintain EXACT markdown formatting
- Preserve ALL code blocks exactly as-is
