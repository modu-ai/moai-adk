# Frontend Advanced Patterns

## Server Components Deep Dive

### Data Fetching in Server Components

```typescript
// app/dashboard/page.tsx
import { Suspense } from 'react';

async function getUserData(userId: string) {
    const res = await fetch(`https://api.example.com/users/${userId}`, {
        next: { revalidate: 3600 } // Revalidate every hour
    });
    return res.json();
}

async function getAnalytics(userId: string) {
    const res = await fetch(`https://api.example.com/analytics/${userId}`, {
        next: { revalidate: 60 }
    });
    return res.json();
}

export default async function Dashboard({ params }) {
    const [user, analytics] = await Promise.all([
        getUserData(params.userId),
        getAnalytics(params.userId)
    ]);

    return (
        <div>
            <h1>{user.name}</h1>
            <Suspense fallback={<LoadingAnalytics />}>
                <AnalyticsChart data={analytics} />
            </Suspense>
        </div>
    );
}

function LoadingAnalytics() {
    return <div>Loading analytics...</div>;
}
```

## State Management Patterns (Zustand)

### Complex State with Middleware

```typescript
import { create } from 'zustand';
import { subscribeWithSelector } from 'zustand/middleware';

interface AppState {
    user: { id: string; name: string } | null;
    theme: 'light' | 'dark';
    notifications: Notification[];
    setUser: (user: AppState['user']) => void;
    setTheme: (theme: AppState['theme']) => void;
    addNotification: (notification: Notification) => void;
    clearNotifications: () => void;
}

export const useApp = create<AppState>()(
    subscribeWithSelector((set) => ({
        user: null,
        theme: 'light',
        notifications: [],
        setUser: (user) => set({ user }),
        setTheme: (theme) => set({ theme }),
        addNotification: (notification) =>
            set((state) => ({
                notifications: [...state.notifications, notification]
            })),
        clearNotifications: () => set({ notifications: [] })
    }))
);

// Subscribe to specific changes
useApp.subscribe(
    (state) => state.theme,
    (theme) => {
        document.documentElement.setAttribute('data-theme', theme);
    }
);
```

## Micro-Frontend Architecture

### Module Federation (Webpack 5)

```javascript
// packages/app1/webpack.config.js
module.exports = {
    plugins: [
        new container.ModuleFederationPlugin({
            name: 'app1',
            filename: 'remoteEntry.js',
            remotes: {
                app2: 'app2@http://localhost:3002/remoteEntry.js'
            },
            exposes: {
                './Button': './src/components/Button',
                './store': './src/store'
            },
            shared: ['react', 'react-dom']
        })
    ]
};

// Usage
import Button from 'app2/Button';

export function AppComponent() {
    return <Button>Click me</Button>;
}
```

## Advanced Hooks and Patterns

### useTransition for Background Updates

```typescript
import { useState, useTransition } from 'react';

export function SearchUsers() {
    const [query, setQuery] = useState('');
    const [results, setResults] = useState([]);
    const [isPending, startTransition] = useTransition();

    const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        setQuery(value);

        startTransition(async () => {
            const res = await fetch(`/api/search?q=${value}`);
            const data = await res.json();
            setResults(data);
        });
    };

    return (
        <div>
            <input
                value={query}
                onChange={handleSearch}
                placeholder="Search..."
            />
            {isPending && <div>Searching...</div>}
            <ul>
                {results.map((result) => (
                    <li key={result.id}>{result.name}</li>
                ))}
            </ul>
        </div>
    );
}
```

### use() Hook for Better Error Boundaries

```typescript
import { use, Suspense } from 'react';

interface UserCardProps {
    userPromise: Promise<User>;
}

function UserCard({ userPromise }: UserCardProps) {
    const user = use(userPromise);

    return (
        <div>
            <h2>{user.name}</h2>
            <p>{user.email}</p>
        </div>
    );
}

export function UserPage() {
    const userPromise = fetch('/api/user').then(r => r.json());

    return (
        <Suspense fallback={<div>Loading user...</div>}>
            <UserCard userPromise={userPromise} />
        </Suspense>
    );
}
```

## Context7 Integration

### Real-time Framework Patterns

The Context7 MCP integration provides:
- **React 19**: Latest Server Components, use() hook patterns
- **Next.js 15**: PPR (Partial Pre-Rendering), Turbopack features
- **Vue 3.5**: Signals reactivity system
- **Vite 5.2**: Modern build optimization patterns

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
**Status**: Production Ready

## Context7 Integration

### Related Libraries & Tools
- [React](/facebook/react): UI library
- [Next.js](/vercel/next.js): React framework
- [Vue](/vuejs/vue): Progressive framework
- [TypeScript](/microsoft/typescript): Type safety
