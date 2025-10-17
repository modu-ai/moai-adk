# Part 4: Frontend 구현

> **소요시간**: 약 1.5시간
> **학습 목표**: Vite + React 18 + TypeScript로 모던 프론트엔드를 구현하고, Tailwind CSS로 깔끔한 UI를 완성합니다.

---

## 개요

백엔드 API가 완성되었으니 이제 사용자가 실제로 사용할 프론트엔드를 구현합니다. Alfred가 자동으로 다음을 수행합니다:

1. **프로젝트 구조 생성**
2. **최신 프론트엔드 스택 설정**
3. **API 클라이언트 작성**
4. **React 컴포넌트 구현**
5. **Tailwind CSS 스타일링**

## Step 1: 프론트엔드 프로젝트 구조 생성

### 1.1 디렉토리 구조

Alfred가 자동으로 생성하는 구조:

```
frontend/
├── src/
│   ├── components/
│   │   ├── TodoList.tsx     # @CODE:TODO-001:UI
│   │   ├── TodoForm.tsx     # @CODE:TODO-001:UI
│   │   └── TodoItem.tsx     # @CODE:TODO-001:UI
│   ├── services/
│   │   └── api.ts           # API 클라이언트
│   ├── types/
│   │   └── todo.ts          # TypeScript 타입
│   ├── App.tsx              # 메인 컴포넌트
│   ├── main.tsx             # 진입점
│   ├── index.css            # Tailwind CSS
│   └── vite-env.d.ts        # Vite 타입 선언
├── tests/
│   └── todos.test.tsx       # 컴포넌트 테스트 (선택)
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
├── vitest.config.ts
└── tailwind.config.js
```

### 1.2 의존성 파일 생성

**`package.json`**:

```json
{
  "name": "my-moai-project-frontend",
  "private": true,
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "test": "vitest",
    "coverage": "vitest run --coverage"
  },
  "dependencies": {
    "react": "^18.3.1",
    "react-dom": "^18.3.1"
  },
  "devDependencies": {
    "@types/react": "^18.3.12",
    "@types/react-dom": "^18.3.1",
    "@vitejs/plugin-react": "^4.3.4",
    "typescript": "^5.9.0",
    "vite": "^7.1.9",
    "vitest": "^3.2.4",
    "@vitest/ui": "^3.2.4",
    "@vitest/coverage-v8": "^3.2.4",
    "tailwindcss": "^4.1.14",
    "postcss": "^8.4.49",
    "autoprefixer": "^10.4.20",
    "eslint": "^9.15.0",
    "prettier": "^3.3.3"
  }
}
```

**핵심 포인트**:
- React 18.3.1: 안정화된 최신 버전
- Vite 7.1.9: 초고속 빌드 도구
- Tailwind CSS 4.1.14: 최신 유틸리티 CSS
- Vitest 3.2.4: Vite 네이티브 테스트 러너

### 1.3 TypeScript 설정

**`tsconfig.json`**:

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["ES2023", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,

    /* Bundler mode */
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",

    /* Linting */
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true
  },
  "include": ["src"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

**핵심 포인트**:
- `strict: true`: 엄격한 타입 검사
- `jsx: "react-jsx"`: React 17+ JSX Transform

### 1.4 Vite 설정

**`vite.config.ts`**:

```typescript
// @CODE:TODO-001 | SPEC: SPEC-TODO-001.md
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8000',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
})
```

**핵심 포인트**:
- `/api` 경로를 백엔드로 프록시 (CORS 우회)
- 개발 서버 포트: 5173 (Vite 기본값)

### 1.5 의존성 설치

```bash
cd frontend
pnpm install
```

출력:

```
Progress: resolved 242, reused 242, downloaded 0, added 242

dependencies:
+ react 18.3.1
+ react-dom 18.3.1

devDependencies:
+ @types/react 18.3.12
+ @vitejs/plugin-react 4.3.4
+ typescript 5.9.0
+ vite 7.1.9
+ tailwindcss 4.1.14
+ vitest 3.2.4

Done in 3.2s
```

---

## Step 2: TypeScript 타입 정의

### 2.1 Todo 타입 정의

**`src/types/todo.ts`**:

```typescript
// @CODE:TODO-001:DATA | SPEC: SPEC-TODO-001.md
/**
 * Todo type definitions matching backend schema
 */

export interface Todo {
  id: number
  title: string
  description: string | null
  completed: boolean
  created_at: string
  updated_at: string
}

export interface TodoCreate {
  title: string
  description?: string | null
}

export interface TodoUpdate {
  title?: string
  description?: string | null
  completed?: boolean
}
```

**핵심 포인트**:
- 백엔드 Pydantic 스키마와 1:1 매칭
- `?`로 선택적 필드 표시
- ISO 8601 문자열로 날짜 표현

### 2.2 Vite 환경 타입 선언

**`src/vite-env.d.ts`**:

```typescript
/// <reference types="vite/client" />
```

이 파일은 CSS import 등의 타입 오류를 방지합니다.

---

## Step 3: API 클라이언트 작성

### 3.1 Todo API 클라이언트

**`src/services/api.ts`**:

```typescript
// @CODE:TODO-001:API | SPEC: SPEC-TODO-001.md
/**
 * API client for Todo backend
 */
import type { Todo, TodoCreate, TodoUpdate } from '../types/todo'

const API_BASE_URL = 'http://localhost:8000/api'

export class TodoAPI {
  static async getAllTodos(): Promise<Todo[]> {
    const response = await fetch(`${API_BASE_URL}/todos`)
    if (!response.ok) {
      throw new Error('Failed to fetch todos')
    }
    return response.json()
  }

  static async getTodoById(id: number): Promise<Todo> {
    const response = await fetch(`${API_BASE_URL}/todos/${id}`)
    if (!response.ok) {
      throw new Error(`Failed to fetch todo ${id}`)
    }
    return response.json()
  }

  static async createTodo(data: TodoCreate): Promise<Todo> {
    const response = await fetch(`${API_BASE_URL}/todos`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    })
    if (!response.ok) {
      throw new Error('Failed to create todo')
    }
    return response.json()
  }

  static async updateTodo(id: number, data: TodoUpdate): Promise<Todo> {
    const response = await fetch(`${API_BASE_URL}/todos/${id}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    })
    if (!response.ok) {
      throw new Error(`Failed to update todo ${id}`)
    }
    return response.json()
  }

  static async deleteTodo(id: number): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/todos/${id}`, {
      method: 'DELETE',
    })
    if (!response.ok) {
      throw new Error(`Failed to delete todo ${id}`)
    }
  }
}
```

**핵심 포인트**:
- `static` 메서드로 상태 없는 API 클라이언트
- `fetch` API 사용 (추가 라이브러리 불필요)
- TypeScript 타입 안전성 보장
- 에러 핸들링 포함

---

## Step 4: React 컴포넌트 구현

### 4.1 TodoItem 컴포넌트 (단일 항목)

**`src/components/TodoItem.tsx`**:

```typescript
// @CODE:TODO-001:UI | SPEC: SPEC-TODO-001.md
/**
 * TodoItem: Component for displaying a single todo
 */
import type { Todo } from '../types/todo'

interface TodoItemProps {
  todo: Todo
  onToggle: (id: number) => Promise<void>
  onDelete: (id: number) => Promise<void>
}

export function TodoItem({ todo, onToggle, onDelete }: TodoItemProps) {
  return (
    <div className="flex items-center gap-3 p-4 bg-white rounded-lg shadow-sm border border-gray-200">
      <input
        type="checkbox"
        checked={todo.completed}
        onChange={() => onToggle(todo.id)}
        className="w-5 h-5 text-blue-600 rounded focus:ring-2 focus:ring-blue-500"
      />
      <div className="flex-1 min-w-0">
        <h3
          className={`text-lg font-medium ${
            todo.completed ? 'line-through text-gray-400' : 'text-gray-900'
          }`}
        >
          {todo.title}
        </h3>
        {todo.description && (
          <p className="text-sm text-gray-500 mt-1">{todo.description}</p>
        )}
      </div>
      <button
        onClick={() => onDelete(todo.id)}
        className="px-3 py-1 text-sm text-red-600 hover:bg-red-50 rounded"
      >
        삭제
      </button>
    </div>
  )
}
```

**핵심 포인트**:
- Props 타입 인터페이스 정의
- 완료 상태에 따른 스타일 변경 (`line-through`)
- Tailwind CSS 유틸리티 클래스 사용

### 4.2 TodoForm 컴포넌트 (생성 폼)

**`src/components/TodoForm.tsx`**:

```typescript
// @CODE:TODO-001:UI | SPEC: SPEC-TODO-001.md
/**
 * TodoForm: Component for creating and editing todos
 */
import { useState } from 'react'
import type { TodoCreate } from '../types/todo'

interface TodoFormProps {
  onSubmit: (data: TodoCreate) => Promise<void>
  initialTitle?: string
  initialDescription?: string
}

export function TodoForm({ onSubmit, initialTitle = '', initialDescription = '' }: TodoFormProps) {
  const [title, setTitle] = useState(initialTitle)
  const [description, setDescription] = useState(initialDescription)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!title.trim()) return

    setIsSubmitting(true)
    try {
      await onSubmit({ title: title.trim(), description: description.trim() || null })
      setTitle('')
      setDescription('')
    } catch (error) {
      console.error('Failed to submit:', error)
      alert('할일 추가에 실패했습니다.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4 p-6 bg-white rounded-lg shadow-md">
      <div>
        <label htmlFor="title" className="block text-sm font-medium text-gray-700 mb-1">
          할일
        </label>
        <input
          id="title"
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="무엇을 해야 하나요?"
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          disabled={isSubmitting}
        />
      </div>
      <div>
        <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-1">
          설명 (선택)
        </label>
        <textarea
          id="description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="상세 설명을 입력하세요"
          rows={3}
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          disabled={isSubmitting}
        />
      </div>
      <button
        type="submit"
        disabled={!title.trim() || isSubmitting}
        className="w-full py-2 px-4 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors"
      >
        {isSubmitting ? '추가 중...' : '추가하기'}
      </button>
    </form>
  )
}
```

**핵심 포인트**:
- `useState`로 폼 상태 관리
- 제출 중 상태 표시 (`isSubmitting`)
- 유효성 검사 (빈 제목 방지)
- 에러 핸들링 및 사용자 피드백

### 4.3 TodoList 컴포넌트 (메인 컨테이너)

**`src/components/TodoList.tsx`**:

```typescript
// @CODE:TODO-001:UI | SPEC: SPEC-TODO-001.md
/**
 * TodoList: Main component for displaying and managing todos
 */
import { useEffect, useState } from 'react'
import { TodoAPI } from '../services/api'
import type { Todo, TodoCreate } from '../types/todo'
import { TodoForm } from './TodoForm'
import { TodoItem } from './TodoItem'

export function TodoList() {
  const [todos, setTodos] = useState<Todo[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // 초기 데이터 로드
  useEffect(() => {
    loadTodos()
  }, [])

  const loadTodos = async () => {
    try {
      setIsLoading(true)
      setError(null)
      const data = await TodoAPI.getAllTodos()
      setTodos(data)
    } catch (err) {
      setError('할일 목록을 불러오는데 실패했습니다.')
      console.error(err)
    } finally {
      setIsLoading(false)
    }
  }

  const handleCreate = async (data: TodoCreate) => {
    const newTodo = await TodoAPI.createTodo(data)
    setTodos([newTodo, ...todos])
  }

  const handleToggle = async (id: number) => {
    const todo = todos.find((t) => t.id === id)
    if (!todo) return

    const updatedTodo = await TodoAPI.updateTodo(id, { completed: !todo.completed })
    setTodos(todos.map((t) => (t.id === id ? updatedTodo : t)))
  }

  const handleDelete = async (id: number) => {
    if (!confirm('정말 삭제하시겠습니까?')) return

    await TodoAPI.deleteTodo(id)
    setTodos(todos.filter((t) => t.id !== id))
  }

  if (isLoading) {
    return (
      <div className="max-w-2xl mx-auto p-6">
        <p className="text-center text-gray-500">로딩 중...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="max-w-2xl mx-auto p-6">
        <p className="text-center text-red-500">{error}</p>
        <button
          onClick={loadTodos}
          className="mt-4 mx-auto block px-4 py-2 bg-blue-600 text-white rounded-lg"
        >
          다시 시도
        </button>
      </div>
    )
  }

  return (
    <div className="max-w-2xl mx-auto p-6 space-y-6">
      <h1 className="text-3xl font-bold text-gray-900">📝 Todo App</h1>

      <TodoForm onSubmit={handleCreate} />

      <div className="space-y-3">
        <h2 className="text-xl font-semibold text-gray-700">
          할일 목록 ({todos.length}개)
        </h2>
        {todos.length === 0 ? (
          <p className="text-center text-gray-400 py-8">할일이 없습니다. 추가해보세요!</p>
        ) : (
          todos.map((todo) => (
            <TodoItem
              key={todo.id}
              todo={todo}
              onToggle={handleToggle}
              onDelete={handleDelete}
            />
          ))
        )}
      </div>
    </div>
  )
}
```

**핵심 포인트**:
- `useEffect`로 컴포넌트 마운트 시 데이터 로드
- 로딩/에러 상태 관리
- 낙관적 UI 업데이트 (즉시 반영)
- 빈 상태 메시지 표시

---

## Step 5: 애플리케이션 진입점

### 5.1 App 컴포넌트

**`src/App.tsx`**:

```typescript
// @CODE:TODO-001:UI | SPEC: SPEC-TODO-001.md
import { TodoList } from './components/TodoList'
import './index.css'

function App() {
  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <TodoList />
    </div>
  )
}

export default App
```

### 5.2 메인 진입점

**`src/main.tsx`**:

```typescript
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App.tsx'
import './index.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
```

### 5.3 HTML 템플릿

**`index.html`**:

```html
<!doctype html>
<html lang="ko">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/vite.svg" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Todo App - MoAI ADK</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

---

## Step 6: Tailwind CSS 스타일링

### 6.1 Tailwind 설정

**`src/index.css`**:

```css
@import "tailwindcss";

:root {
  font-family: Inter, system-ui, Avenir, Helvetica, Arial, sans-serif;
  line-height: 1.5;
  font-weight: 400;

  color-scheme: light dark;
  color: rgba(255, 255, 255, 0.87);
  background-color: #242424;

  font-synthesis: none;
  text-rendering: optimizeLegibility;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

body {
  margin: 0;
  min-width: 320px;
  min-height: 100vh;
}
```

**핵심 포인트**:
- Tailwind CSS 4.x: `@import "tailwindcss"` 한 줄로 완성
- 폰트 및 렌더링 최적화

---

## Step 7: 프론트엔드 실행

### 7.1 개발 서버 실행

```bash
cd frontend
pnpm dev
```

출력:

```
VITE v7.1.9  ready in 234 ms

➜  Local:   http://localhost:5173/
➜  Network: use --host to expose
➜  press h + enter to show help
```

### 7.2 브라우저 확인

브라우저에서 `http://localhost:5173` 접속

**확인 사항**:

1. TodoForm이 표시되는가?
1. "추가하기" 버튼이 작동하는가?
1. 할일 목록이 표시되는가?
1. 체크박스 토글이 작동하는가?
1. 삭제 버튼이 작동하는가?

### 7.3 백엔드 연동 테스트

**전제 조건**: 백엔드 서버가 실행 중이어야 합니다.

```bash
# 백엔드 실행 (별도 터미널)
cd backend
.venv/bin/uvicorn app.main:app --reload
```

**테스트 시나리오**:

1. **할일 추가**:
   - 제목: "프론트엔드 테스트"
   - 설명: "API 연동 확인"
   - "추가하기" 클릭 → 목록에 즉시 표시

1. **체크박스 토글**:
   - 체크박스 클릭 → 취소선 표시
   - 다시 클릭 → 취소선 제거

1. **할일 삭제**:
   - "삭제" 버튼 클릭 → 확인 대화상자
   - "확인" 클릭 → 목록에서 제거

1. **새로고침**:
   - `F5` 또는 `Cmd+R` → 데이터 유지 확인

---

## Step 8: 빌드 및 프로덕션 배포

### 8.1 TypeScript 타입 검사

```bash
pnpm exec tsc --noEmit
```

출력 (정상):

```
(아무 출력 없음 = 타입 오류 없음)
```

### 8.2 프로덕션 빌드

```bash
pnpm build
```

출력:

```
vite v7.1.9 building for production...
✓ 34 modules transformed.
dist/index.html                  0.45 kB │ gzip: 0.30 kB
dist/assets/index-DiwrgTda.css   4.20 kB │ gzip: 1.10 kB
dist/assets/index-BLv8yEiS.js   143.41 kB │ gzip: 46.13 kB
✓ built in 456ms
```

### 8.3 프로덕션 미리보기

```bash
pnpm preview
```

출력:

```
➜  Local:   http://localhost:4173/
➜  Network: use --host to expose
```

---

## 핵심 학습 포인트

### ✅ 최신 React 기술 스택

- **React 18.3.1**: Concurrent Features, 자동 배칭
- **Vite 7.1.9**: 초고속 HMR, ESM 기반
- **TypeScript 5.9.0**: strict 모드, 완벽한 타입 안전성
- **Tailwind CSS 4.1.14**: 유틸리티 우선 스타일링

### ✅ 컴포넌트 설계 패턴

- **TodoList**: 컨테이너 컴포넌트 (상태 관리)
- **TodoForm**: 제어 컴포넌트 (폼 입력)
- **TodoItem**: 프레젠테이셔널 컴포넌트 (순수 표시)

### ✅ 상태 관리 전략

- `useState`: 로컬 상태 관리
- `useEffect`: 데이터 페칭 (마운트 시 1회)
- 낙관적 업데이트: 즉각적인 UI 반응

### ✅ 에러 핸들링

- 로딩 상태 표시
- 에러 메시지 표시
- 재시도 버튼 제공
- 사용자 친화적 피드백

### ✅ TypeScript 활용

- Props 인터페이스 정의
- API 응답 타입 보장
- 컴파일 타임 에러 방지

### ✅ @TAG 추적성

- `@CODE:TODO-001:UI` - React 컴포넌트
- `@CODE:TODO-001:API` - API 클라이언트
- `@CODE:TODO-001:DATA` - TypeScript 타입

---

## 개선 가능한 부분 (선택 사항)

### 1. React Hook Form 추가

```bash
pnpm add react-hook-form
```

더 강력한 폼 검증 및 상태 관리.

### 2. React Query 추가

```bash
pnpm add @tanstack/react-query
```

캐싱, 자동 재시도, 낙관적 업데이트 자동화.

### 3. Zustand 추가

```bash
pnpm add zustand
```

전역 상태 관리 (여러 컴포넌트 간 상태 공유).

### 4. 다크 모드 지원

Tailwind CSS `dark:` 접두사 활용.

### 5. 접근성 개선

- `aria-label` 추가
- 키보드 네비게이션 지원
- 스크린 리더 지원

---

## 🚀 다음

프론트엔드 구현이 완료되었습니다! 이제 Git 커밋, 문서 동기화, 배포 준비를 진행합니다.

**다음**: [Part 5: Sync & Deploy](./05-sync-deploy.md)

**이전**: [Part 3: Backend TDD 구현하기](./03-backend-tdd.md)
