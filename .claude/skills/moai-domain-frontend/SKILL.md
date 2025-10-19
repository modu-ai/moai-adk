---
name: moai-domain-frontend
description: React/Vue/Angular development with state management, performance optimization, and accessibility
allowed-tools:
  - Read
  - Bash
tier: 4
auto-load: "false"
---

# Frontend Expert

## What it does

Provides expertise in modern frontend development using React, Vue, or Angular, including state management patterns, performance optimization techniques, and accessibility (a11y) best practices.

## When to use

- "프론트엔드 개발", "React 컴포넌트", "상태 관리", "성능 최적화"
- Automatically invoked when working with frontend projects
- Frontend SPEC implementation (`/alfred:2-build`)

## How it works

**React Development**:
- **Functional components**: Hooks (useState, useEffect, useMemo)
- **State management**: Redux, Zustand, Jotai
- **Performance**: React.memo, useCallback, code splitting
- **Testing**: React Testing Library

**Vue Development**:
- **Composition API**: setup(), reactive(), computed()
- **State management**: Pinia, Vuex
- **Performance**: Virtual scrolling, lazy loading
- **Testing**: Vue Test Utils

**Angular Development**:
- **Components**: TypeScript classes with decorators
- **State management**: NgRx, Akita
- **Performance**: OnPush change detection, lazy loading
- **Testing**: Jasmine, Karma

**Performance Optimization**:
- **Code splitting**: Dynamic imports, route-based splitting
- **Lazy loading**: Images, components
- **Bundle optimization**: Tree shaking, minification
- **Web Vitals**: LCP, FID, CLS optimization

**Accessibility (a11y)**:
- **Semantic HTML**: Proper use of HTML5 elements
- **ARIA attributes**: Roles, labels, descriptions
- **Keyboard navigation**: Focus management
- **Screen reader support**: Alt text, aria-live

## TDD for Frontend Development

### RED: Component Test
```typescript
// @TEST:COMPONENT-001 | SPEC: SPEC-UI-001.md
describe('UserCard', () => {
    it('should render user info', () => {
        const user = { id: 1, name: 'Alice', email: 'alice@test.com' };
        render(<UserCard user={user} />);

        expect(screen.getByText('Alice')).toBeInTheDocument();
        expect(screen.getByText('alice@test.com')).toBeInTheDocument();
    });

    it('should call onEdit when edit button clicked', () => {
        const handleEdit = jest.fn();
        const user = { id: 1, name: 'Alice', email: 'alice@test.com' };
        render(<UserCard user={user} onEdit={handleEdit} />);

        fireEvent.click(screen.getByRole('button', { name: /edit/i }));
        expect(handleEdit).toHaveBeenCalledWith(user);
    });
});
```

### GREEN: Component Implementation
```typescript
// @CODE:COMPONENT-001 | TEST: __tests__/UserCard.test.tsx
export const UserCard: React.FC<UserCardProps> = ({ user, onEdit }) => {
    return (
        <div className="card">
            <h2>{user.name}</h2>
            <p>{user.email}</p>
            <button onClick={() => onEdit(user)}>Edit</button>
        </div>
    );
};
```

### REFACTOR: Performance & Memo
```typescript
// @CODE:COMPONENT-001:REFACTOR | 성능 최적화
export const UserCard = React.memo<UserCardProps>(
    ({ user, onEdit }) => (
        <div className="card">
            <h2>{user.name}</h2>
            <p>{user.email}</p>
            <button onClick={() => onEdit(user)}>Edit</button>
        </div>
    ),
    (prev, next) => prev.user.id === next.user.id
);
```

## Examples

### Example 1: React State Management Comparison

**❌ Before (Prop Drilling)**:
```typescript
// @CODE:STATE-001: Props 전달 지옥
function App() {
    const [user, setUser] = useState(null);

    return <Level1 user={user} setUser={setUser} />;
}

function Level1({ user, setUser }) {
    return <Level2 user={user} setUser={setUser} />;
}

function Level2({ user, setUser }) {
    return <Level3 user={user} setUser={setUser} />;
}

function Level3({ user, setUser }) {
    return <button onClick={() => setUser({ name: 'Bob' })}>Update</button>;
}

// 문제: 3단계 깊이에서만 필요한데 중간 컴포넌트들이 모두 거쳐야 함
```

**✅ After (Context API)**:
```typescript
// @CODE:STATE-001: Context 활용
const UserContext = createContext<UserContextType | null>(null);

export function useUser() {
    const context = useContext(UserContext);
    if (!context) throw new Error('useUser must be used within UserProvider');
    return context;
}

function App() {
    const [user, setUser] = useState(null);

    return (
        <UserContext.Provider value={{ user, setUser }}>
            <Level1 />
        </UserContext.Provider>
    );
}

function Level1() {
    return <Level2 />;
}

function Level2() {
    return <Level3 />;
}

function Level3() {
    const { setUser } = useUser();
    return <button onClick={() => setUser({ name: 'Bob' })}>Update</button>;
}

// 개선: Prop drilling 제거, 필요한 곳에서만 useUser() 사용
```

### Example 2: Performance Optimization

**❌ Before (Unnecessary Re-renders)**:
```typescript
// @CODE:PERF-001: 낭비적인 리렌더
function UserList({ users }) {
    const handleEdit = (id) => {
        // 비싼 작업
        console.log('Editing', id);
    };

    return users.map(user => (
        <UserCard key={user.id} user={user} onEdit={handleEdit} />
    ));
}

// 문제: 매번 handleEdit이 새로 생성되어 모든 UserCard가 리렌더됨
```

**✅ After (Optimized with useCallback)**:
```typescript
// @CODE:PERF-001: useCallback 최적화
function UserList({ users }) {
    const handleEdit = useCallback((id) => {
        console.log('Editing', id);
    }, []);  // 의존성 없음 = 함수 재생성 안 함

    return users.map(user => (
        <UserCard
            key={user.id}
            user={user}
            onEdit={handleEdit}
        />
    ));
}

// 개선: 함수 재생성 방지 → 자식 리렌더 방지 ✅
```

### Example 3: Accessibility (a11y) Improvement

**Checklist**:
```markdown
- [ ] 시맨틱 HTML (header, nav, main, section, footer)
- [ ] ARIA labels (aria-label, aria-describedby)
- [ ] Keyboard navigation (Tab, Enter, Esc)
- [ ] Focus management (autofocus, tabIndex)
- [ ] Screen reader support (role, aria-live)
- [ ] Color contrast (4.5:1 ratio minimum)
- [ ] Alt text for images
```

**Example**:
```typescript
<button
    aria-label="Delete user"
    onClick={handleDelete}
    className="delete-btn"
>
    🗑️
</button>
```

### Example 4: Code Splitting for Performance

**Configuration**:
```typescript
// @CODE:SPLIT-001: 동적 로딩
import { lazy, Suspense } from 'react';

const AdminPanel = lazy(() => import('./pages/AdminPanel'));
const Dashboard = lazy(() => import('./pages/Dashboard'));

export function App() {
    return (
        <Suspense fallback={<Loading />}>
            <Routes>
                <Route path="/admin" element={<AdminPanel />} />
                <Route path="/dashboard" element={<Dashboard />} />
            </Routes>
        </Suspense>
    );
}

// 효과: 초기 번들 크기 감소, 라우트별 로드
```

## Keywords

"프론트엔드 개발", "React", "성능 최적화", "상태 관리", "접근성", "컴포넌트", "Hook", "Code splitting", "TDD", "memorization", "re-render optimization"

## Reference

- Frontend architecture: `.moai/memory/development-guide.md#프론트엔드-아키텍처`
- State management: CLAUDE.md#상태-관리-패턴
- Performance optimization: `.moai/memory/development-guide.md#프론트엔드-성능`

## Works well with

- moai-foundation-trust (프론트엔드 테스트)
- moai-essentials-perf (성능 최적화)
- moai-domain-web-api (백엔드 연동)
