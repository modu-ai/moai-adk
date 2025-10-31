# moai-domain-frontend

Frontend architecture with state management, performance optimization, accessibility patterns, and component organization.

## Quick Start

Modern frontend development requires careful attention to architecture, performance, and user experience. Use this skill when designing component systems, implementing state management, optimizing Core Web Vitals, or building accessible UIs.

## Core Patterns

### Pattern 1: Component Architecture & Organization

**Pattern**: Structure components using atomic design and composition principles.

```typescript
// components/Button.tsx - Atomic level component
'use client';

import React from 'react';
import { cn } from '@/lib/utils';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  isLoading?: boolean;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'primary', size = 'md', isLoading, ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={cn(
          'inline-flex items-center justify-center rounded-md font-medium transition-colors',
          {
            'bg-blue-600 text-white hover:bg-blue-700': variant === 'primary',
            'bg-gray-200 text-gray-900 hover:bg-gray-300': variant === 'secondary',
            'border border-gray-300 hover:bg-gray-50': variant === 'outline',
            'hover:bg-gray-100': variant === 'ghost',
            'px-2 py-1 text-sm': size === 'sm',
            'px-4 py-2 text-base': size === 'md',
            'px-6 py-3 text-lg': size === 'lg',
          },
          className
        )}
        disabled={isLoading || props.disabled}
        {...props}
      >
        {isLoading ? <LoadingSpinner /> : props.children}
      </button>
    );
  }
);

// components/Card.tsx - Molecule level component
export function Card({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn('rounded-lg border border-gray-200 p-6 bg-white shadow-sm', className)}>
      {children}
    </div>
  );
}

// components/UserCard.tsx - Organism level component (combines atoms + molecules)
export function UserCard({ user }: { user: { name: string; email: string; avatar: string } }) {
  return (
    <Card>
      <div className="flex items-center gap-4">
        <img src={user.avatar} alt={user.name} className="w-12 h-12 rounded-full" />
        <div className="flex-1">
          <h3 className="font-semibold">{user.name}</h3>
          <p className="text-sm text-gray-500">{user.email}</p>
        </div>
        <Button size="sm" variant="outline">
          View Profile
        </Button>
      </div>
    </Card>
  );
}

// pages/dashboard.tsx - Page level component
export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">Dashboard</h1>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <UserCard user={{ name: 'Alice', email: 'alice@example.com', avatar: '...' }} />
        <UserCard user={{ name: 'Bob', email: 'bob@example.com', avatar: '...' }} />
      </div>
    </div>
  );
}
```

**When to use**:
- Building scalable component systems
- Organizing large codebases
- Creating reusable components
- Maintaining consistency

**Key benefits**:
- Clear component hierarchy
- Reusable building blocks
- Scalable architecture
- Easy maintenance

### Pattern 2: State Management Patterns

**Pattern**: Choose appropriate state management based on complexity.

```typescript
// lib/hooks/use-form.ts - Simple component state
export function useForm<T extends Record<string, any>>(initialValues: T) {
  const [values, setValues] = React.useState(initialValues);
  const [errors, setErrors] = React.useState<Partial<T>>({});
  const [isSubmitting, setIsSubmitting] = React.useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setValues((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (onSubmit: (values: T) => Promise<void>) => {
    return async (e: React.FormEvent) => {
      e.preventDefault();
      setIsSubmitting(true);
      try {
        await onSubmit(values);
      } catch (error) {
        setErrors({ error: String(error) } as Partial<T>);
      } finally {
        setIsSubmitting(false);
      }
    };
  };

  return { values, errors, isSubmitting, handleChange, handleSubmit };
}

// context/AuthContext.tsx - Context for app-wide state
'use client';

import React, { createContext, useContext } from 'react';

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  error: Error | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = React.useState<User | null>(null);
  const [isLoading, setIsLoading] = React.useState(true);
  const [error, setError] = React.useState<Error | null>(null);

  // Fetch user on mount
  React.useEffect(() => {
    (async () => {
      try {
        const response = await fetch('/api/me');
        setUser(await response.json());
      } catch (err) {
        setError(err as Error);
      } finally {
        setIsLoading(false);
      }
    })();
  }, []);

  const login = async (email: string, password: string) => {
    const response = await fetch('/api/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
    setUser(await response.json());
  };

  const logout = async () => {
    await fetch('/api/logout', { method: 'POST' });
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, isLoading, error, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}

// Usage in component
function ProfileComponent() {
  const { user, logout } = useAuth();

  return (
    <div>
      <h1>{user?.name}</h1>
      <button onClick={logout}>Logout</button>
    </div>
  );
}
```

**When to use**:
- Managing form state
- Handling app-wide authentication state
- Managing user preferences
- Sharing state across components

**Key benefits**:
- Clear state ownership
- Reduced prop drilling
- Better performance
- Easier testing

### Pattern 3: Performance Optimization & Accessibility

**Pattern**: Optimize Core Web Vitals and ensure WCAG compliance.

```typescript
// lib/hooks/use-lazy-load.ts - Lazy loading optimization
export function useLazyLoad(ref: React.RefObject<Element>) {
  const [isVisible, setIsVisible] = React.useState(false);

  React.useEffect(() => {
    const observer = new IntersectionObserver(([entry]) => {
      if (entry.isIntersecting) {
        setIsVisible(true);
        observer.unobserve(entry.target);
      }
    });

    if (ref.current) {
      observer.observe(ref.current);
    }

    return () => observer.disconnect();
  }, [ref]);

  return isVisible;
}

// components/LazyImage.tsx - Lazy loaded image component
export function LazyImage({
  src,
  alt,
  width,
  height,
}: {
  src: string;
  alt: string;
  width: number;
  height: number;
}) {
  const ref = React.useRef<HTMLImageElement>(null);
  const isVisible = useLazyLoad(ref);

  return (
    <img
      ref={ref}
      src={isVisible ? src : 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg"%3E%3C/svg%3E'}
      alt={alt}
      width={width}
      height={height}
      loading="lazy"
      className="transition-opacity duration-300"
      style={{ opacity: isVisible ? 1 : 0.5 }}
    />
  );
}

// components/AccessibleForm.tsx - WCAG compliant form
export function AccessibleForm() {
  return (
    <form onSubmit={handleSubmit}>
      <div className="space-y-4">
        <div>
          <label htmlFor="email" className="block font-medium">
            Email Address
            <span className="text-red-600 ml-1">*</span>
          </label>
          <input
            id="email"
            name="email"
            type="email"
            required
            aria-describedby="email-hint"
            aria-invalid={errors.email ? 'true' : 'false'}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          {errors.email && (
            <p id="email-hint" className="text-red-600 text-sm mt-1">
              {errors.email}
            </p>
          )}
        </div>

        <button
          type="submit"
          aria-busy={isSubmitting}
          disabled={isSubmitting}
          className="w-full bg-blue-600 text-white py-2 rounded-md hover:bg-blue-700 disabled:opacity-50"
        >
          {isSubmitting ? 'Submitting...' : 'Submit'}
        </button>
      </div>
    </form>
  );
}

// lib/metrics.ts - Core Web Vitals monitoring
export function trackWebVitals() {
  // Largest Contentful Paint (LCP)
  const observer = new PerformanceObserver((list) => {
    const entries = list.getEntries();
    const lastEntry = entries[entries.length - 1];
    console.log('LCP:', lastEntry.renderTime || lastEntry.loadTime);
  });

  observer.observe({ entryTypes: ['largest-contentful-paint'] });

  // Cumulative Layout Shift (CLS)
  let clsValue = 0;
  const clsObserver = new PerformanceObserver((list) => {
    for (const entry of list.getEntries()) {
      if (!('hadRecentInput' in entry) || !entry.hadRecentInput) {
        clsValue += entry.value;
        console.log('CLS:', clsValue);
      }
    }
  });

  clsObserver.observe({ entryTypes: ['layout-shift'] });
}
```

**When to use**:
- Improving page load performance
- Ensuring accessibility compliance
- Optimizing images and assets
- Monitoring performance metrics

**Key benefits**:
- Better search engine ranking
- Improved user experience
- Faster page loads
- Accessible to all users

## Progressive Disclosure

### Level 1: Basic Frontend
- Component structure and organization
- Styling with CSS/Tailwind
- Basic state management with useState
- Simple forms and interactions

### Level 2: Advanced Patterns
- Context API for shared state
- Custom hooks for reusable logic
- Performance optimization techniques
- Accessibility (WCAG) compliance

### Level 3: Expert Architecture
- Scalable component systems
- Advanced state management (Zustand, Redux)
- Complete performance optimization
- Testing strategies and patterns

## Works Well With

- **React 19**: Component framework with hooks
- **Next.js 16**: Meta-framework with Server Components
- **TypeScript**: Type-safe component props
- **Tailwind CSS**: Utility-first styling
- **Vercel**: Deployment and analytics
- **Playwright**: E2E testing

## References

- **React Documentation**: https://react.dev
- **Next.js Docs**: https://nextjs.org/docs
- **Web Vitals**: https://web.dev/vitals/
- **WCAG Guidelines**: https://www.w3.org/WAI/WCAG21/quickref/
- **Atomic Design**: https://bradfrost.com/blog/post/atomic-web-design/
