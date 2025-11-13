---
name: "moai-domain-frontend"
version: "4.0.0"
created: 2025-11-11
updated: 2025-11-13
status: stable
description: Enterprise Frontend Development with AI-powered modern architecture, Context7 integration, and intelligent component orchestration for scalable user interfaces
keywords: ['frontend', 'react', 'vue', 'angular', 'typescript', 'component-architecture', 'context7-integration', 'ai-orchestration', 'production-deployment']
allowed-tools: 
  - Read
  - Bash
  - Write
  - Edit
  - Glob
  - Grep
  - WebFetch
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# Enterprise Frontend Development Expert v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-domain-frontend |
| **Version** | 4.0.0 (2025-11-13) |
| **Tier** | Enterprise Frontend Development Expert |
| **AI-Powered** | ✅ Context7 Integration, Intelligent Architecture |
| **Auto-load** | On demand when frontend keywords detected |

---

## What It Does

Enterprise Frontend Development expert with AI-powered modern architecture, Context7 integration, and intelligent component orchestration for scalable user interfaces.

**Revolutionary v4.0.0 capabilities**:
- 🤖 **AI-Powered Frontend Architecture** using Context7 MCP for latest frontend patterns
- 📊 **Intelligent Component Orchestration** with automated design system optimization
- 🚀 **Modern Framework Integration** with AI-driven performance optimization
- 🔗 **Enterprise User Experience** with zero-configuration accessibility and internationalization
- 📈 **Predictive Performance Analytics** with usage forecasting and optimization insights

---

## When to Use

**Automatic triggers**:
- Frontend architecture and modern UI framework discussions
- Component design system and user experience planning
- Performance optimization and accessibility implementation
- Responsive design and cross-platform compatibility

**Manual invocation**:
- Designing enterprise frontend architectures with optimal UX patterns
- Implementing modern component systems and design tokens
- Planning frontend performance optimization strategies
- Creating accessible and international user interfaces

---

# Quick Reference (Level 1)

## Modern Frontend Stack (November 2025)

### Core Framework Ecosystem
- **React 19**: Latest with concurrent features and Server Components
- **Vue 3.5**: Composition API and performance optimizations
- **Angular 18**: Standalone components and improved hydration
- **Svelte 5**: Signals and improved TypeScript support
- **Next.js 16**: App Router, Server Components, and Turbopack

### State Management Solutions
- **Zustand**: Lightweight state management
- **TanStack Query**: Server state management with caching
- **Jotai**: Atomic state management
- **Redux Toolkit**: Predictable state container
- **Valtio**: Proxy-based state management

### Styling & UI Systems
- **Tailwind CSS**: Utility-first CSS framework
- **shadcn/ui**: High-quality component library
- **Material-UI**: React component library
- **Chakra UI**: Accessible React components
- **Styled Components**: CSS-in-JS with TypeScript

### Performance Optimization
- **Code Splitting**: Dynamic imports and lazy loading
- **Image Optimization**: Next.js Image and Cloudinary
- **Bundle Analysis**: Webpack Bundle Analyzer
- **Runtime Optimization**: React.memo and useMemo

---

# Core Implementation (Level 2)

## Frontend Architecture Intelligence

```python
# AI-powered frontend architecture optimization with Context7
class FrontendArchitectOptimizer:
    def __init__(self):
        self.context7_client = Context7Client()
        self.component_analyzer = ComponentAnalyzer()
        self.performance_optimizer = PerformanceOptimizer()
    
    async def design_optimal_frontend_architecture(self, 
                                                 requirements: FrontendRequirements) -> FrontendArchitecture:
        """Design optimal frontend architecture using AI analysis."""
        
        # Get latest frontend documentation via Context7
        react_docs = await self.context7_client.get_library_docs(
            context7_library_id='/react/docs',
            topic="hooks server-components performance optimization 2025",
            tokens=3000
        )
        
        nextjs_docs = await self.context7_client.get_library_docs(
            context7_library_id='/vercel/docs',
            topic="next.js app router optimization deployment 2025",
            tokens=2000
        )
        
        # Optimize component architecture
        component_design = self.component_analyzer.optimize_component_system(
            requirements.ui_complexity,
            requirements.team_size,
            react_docs
        )
        
        # Optimize performance strategy
        performance_strategy = self.performance_optimizer.design_performance_strategy(
            requirements.performance_targets,
            requirements.user_base,
            nextjs_docs
        )
        
        return FrontendArchitecture(
            framework_selection=self._select_framework(requirements),
            component_system=component_design,
            state_management=self._design_state_management(requirements),
            styling_strategy=self._design_styling_system(requirements),
            performance_optimization=performance_strategy,
            accessibility_compliance=self._ensure_accessibility(requirements),
            internationalization=self._configure_i18n(requirements)
        )
```

## Modern React Component Architecture

```typescript
// Advanced React component with TypeScript and modern patterns
import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { motion, AnimatePresence } from 'framer-motion';

// Type definitions with Zod for runtime validation
const UserSchema = z.object({
  id: z.string(),
  name: z.string(),
  email: z.string().email(),
  avatar: z.string().url().optional(),
  role: z.enum(['admin', 'user', 'moderator']),
  createdAt: z.date(),
});

type User = z.infer<typeof UserSchema>;

// Props interface with strict typing
interface UserListProps {
  onUserSelect: (user: User) => void;
  selectedUserId?: string;
  filters?: {
    role?: User['role'];
    search?: string;
  };
}

// Custom hook for user data management
function useUsers(filters?: UserListProps['filters']) {
  return useQuery({
    queryKey: ['users', filters],
    queryFn: async () => {
      const params = new URLSearchParams();
      if (filters?.role) params.append('role', filters.role);
      if (filters?.search) params.append('search', filters.search);
      
      const response = await fetch(`/api/users?${params}`);
      const data = await response.json();
      
      return z.array(UserSchema).parse(data);
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
    cacheTime: 10 * 60 * 1000, // 10 minutes
  });
}

// Main component with performance optimizations
export const UserList: React.FC<UserListProps> = React.memo(({ 
  onUserSelect, 
  selectedUserId,
  filters 
}) => {
  const [expandedUsers, setExpandedUsers] = useState<Set<string>>(new Set());
  const queryClient = useQueryClient();
  
  // Data fetching with React Query
  const { data: users, isLoading, error } = useUsers(filters);
  
  // Memoized filtered users for performance
  const filteredUsers = useMemo(() => {
    if (!users) return [];
    
    return users.filter(user => {
      if (filters?.role && user.role !== filters.role) return false;
      if (filters?.search && !user.name.toLowerCase().includes(filters.search.toLowerCase())) {
        return false;
      }
      return true;
    });
  }, [users, filters]);

  // Optimized callback functions
  const handleUserClick = useCallback((user: User) => {
    onUserSelect(user);
  }, [onUserSelect]);

  const toggleUserExpansion = useCallback((userId: string) => {
    setExpandedUsers(prev => {
      const newSet = new Set(prev);
      if (newSet.has(userId)) {
        newSet.delete(userId);
      } else {
        newSet.add(userId);
      }
      return newSet;
    });
  }, []);

  // Invalidate and refetch function
  const refreshUsers = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: ['users'] });
  }, [queryClient]);

  if (isLoading) return <UserListSkeleton />;
  if (error) return <ErrorMessage error={error} onRetry={refreshUsers} />;
  if (!users) return <div>No users found</div>;

  return (
    <motion.div 
      className="user-list"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.3 }}
    >
      <div className="user-list-header">
        <h2>Users ({filteredUsers.length})</h2>
        <button onClick={refreshUsers} className="refresh-button">
          Refresh
        </button>
      </div>
      
      <AnimatePresence>
        {filteredUsers.map((user) => (
          <UserCard
            key={user.id}
            user={user}
            isExpanded={expandedUsers.has(user.id)}
            isSelected={selectedUserId === user.id}
            onClick={handleUserClick}
            onToggleExpand={toggleUserExpansion}
          />
        ))}
      </AnimatePresence>
    </motion.div>
  );
});

UserList.displayName = 'UserList';

// Individual user card component
interface UserCardProps {
  user: User;
  isExpanded: boolean;
  isSelected: boolean;
  onClick: (user: User) => void;
  onToggleExpand: (userId: string) => void;
}

const UserCard: React.FC<UserCardProps> = React.memo(({ 
  user, 
  isExpanded, 
  isSelected,
  onClick,
  onToggleExpand 
}) => {
  const handleClick = useCallback(() => {
    onClick(user);
  }, [onClick, user]);

  const handleExpandClick = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    onToggleExpand(user.id);
  }, [onToggleExpand, user.id]);

  return (
    <motion.div
      className={`user-card ${isSelected ? 'selected' : ''}`}
      onClick={handleClick}
      whileHover={{ scale: 1.02 }}
      whileTap={{ scale: 0.98 }}
      layout
    >
      <div className="user-card-header">
        <img 
          src={user.avatar || `https://api.dicebear.com/7.x/avataaars/svg?seed=${user.id}`}
          alt={user.name}
          className="user-avatar"
        />
        
        <div className="user-info">
          <h3 className="user-name">{user.name}</h3>
          <p className="user-email">{user.email}</p>
          <span className={`user-role ${user.role}`}>{user.role}</span>
        </div>
        
        <button 
          onClick={handleExpandClick}
          className="expand-button"
          aria-label={isExpanded ? 'Collapse' : 'Expand'}
        >
          {isExpanded ? '−' : '+'}
        </button>
      </div>
      
      <AnimatePresence>
        {isExpanded && (
          <motion.div
            className="user-details"
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.2 }}
          >
            <p>Member since: {user.createdAt.toLocaleDateString()}</p>
            <div className="user-actions">
              <button className="edit-button">Edit</button>
              <button className="message-button">Message</button>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  );
});

UserCard.displayName = 'UserCard';
```

## State Management with Modern Patterns

```typescript
// Advanced state management with Zustand and TypeScript
import { create } from 'zustand';
import { devtools, subscribeWithSelector } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';

// Type definitions
interface User {
  id: string;
  name: string;
  email: string;
  preferences: {
    theme: 'light' | 'dark';
    language: string;
    notifications: boolean;
  };
}

interface AppState {
  // User state
  currentUser: User | null;
  users: User[];
  
  // UI state
  theme: 'light' | 'dark';
  sidebarOpen: boolean;
  activeModal: string | null;
  
  // Loading states
  loading: {
    users: boolean;
    auth: boolean;
  };
  
  // Error states
  errors: {
    users: string | null;
    auth: string | null;
  };
}

interface AppActions {
  // User actions
  setCurrentUser: (user: User | null) => void;
  updateUserPreferences: (preferences: Partial<User['preferences']>) => void;
  
  // UI actions
  setTheme: (theme: 'light' | 'dark') => void;
  toggleSidebar: () => void;
  setActiveModal: (modal: string | null) => void;
  
  // Data actions
  fetchUsers: () => Promise<void>;
  addUser: (user: Omit<User, 'id'>) => Promise<void>;
  
  // Error handling
  clearError: (key: keyof AppState['errors']) => void;
}

// Create store with middleware
export const useAppStore = create<AppState & AppActions>()(
  devtools(
    subscribeWithSelector(
      immer((set, get) => ({
        // Initial state
        currentUser: null,
        users: [],
        theme: 'light',
        sidebarOpen: true,
        activeModal: null,
        loading: { users: false, auth: false },
        errors: { users: null, auth: null },

        // User actions
        setCurrentUser: (user) => set((state) => {
          state.currentUser = user;
        }),

        updateUserPreferences: (preferences) => set((state) => {
          if (state.currentUser) {
            Object.assign(state.currentUser.preferences, preferences);
          }
        }),

        // UI actions
        setTheme: (theme) => set((state) => {
          state.theme = theme;
        }),

        toggleSidebar: () => set((state) => {
          state.sidebarOpen = !state.sidebarOpen;
        }),

        setActiveModal: (modal) => set((state) => {
          state.activeModal = modal;
        }),

        // Data actions
        fetchUsers: async () => {
          set((state) => { state.loading.users = true; });
          
          try {
            const response = await fetch('/api/users');
            const users = await response.json();
            
            set((state) => {
              state.users = users;
              state.loading.users = false;
              state.errors.users = null;
            });
          } catch (error) {
            set((state) => {
              state.loading.users = false;
              state.errors.users = error instanceof Error ? error.message : 'Failed to fetch users';
            });
          }
        },

        addUser: async (userData) => {
          try {
            const response = await fetch('/api/users', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify(userData),
            });
            
            const newUser = await response.json();
            
            set((state) => {
              state.users.push(newUser);
            });
          } catch (error) {
            set((state) => {
              state.errors.users = error instanceof Error ? error.message : 'Failed to add user';
            });
          }
        },

        // Error handling
        clearError: (key) => set((state) => {
          state.errors[key] = null;
        }),
      }))
    ),
    { name: 'app-store' }
  )
);

// Selectors for optimized re-renders
export const useCurrentUser = () => useAppStore((state) => state.currentUser);
export const useUsers = () => useAppStore((state) => state.users);
export const useTheme = () => useAppStore((state) => state.theme);
export const useSidebarOpen = () => useAppStore((state) => state.sidebarOpen);

// Derived state selectors
export const useActiveUsers = () => {
  return useAppStore((state) => 
    state.users.filter(user => user.preferences.notifications)
  );
};

// Persistence middleware
useAppStore.subscribe(
  (state) => ({
    theme: state.theme,
    sidebarOpen: state.sidebarOpen,
    currentUser: state.currentUser,
  }),
  (persistedState) => {
    localStorage.setItem('app-state', JSON.stringify(persistedState));
  }
);
```

---

# Advanced Implementation (Level 3)

## Performance Optimization Strategies

```typescript
// Advanced performance optimization techniques
export class PerformanceOptimizer {
  // Code splitting with dynamic imports
  static lazyLoadComponents() {
    const LazyComponent = React.lazy(() => import('./HeavyComponent'));
    
    return (
      <Suspense fallback={<ComponentSkeleton />}>
        <LazyComponent />
      </Suspense>
    );
  }

  // Image optimization with next/image
  static OptimizedImage: React.FC<{
    src: string;
    alt: string;
    width: number;
    height: number;
  }> = ({ src, alt, width, height }) => {
    return (
      <Image
        src={src}
        alt={alt}
        width={width}
        height={height}
        placeholder="blur"
        blurDataURL="data:image/jpeg;base64,..."
        loading="lazy"
        sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
      />
    );
  };

  // Virtual scrolling for large lists
  static useVirtualScrolling<T>(
    items: T[],
    itemHeight: number,
    containerHeight: number
  ) {
    const [scrollTop, setScrollTop] = useState(0);
    
    const visibleItems = useMemo(() => {
      const startIndex = Math.floor(scrollTop / itemHeight);
      const endIndex = Math.min(
        startIndex + Math.ceil(containerHeight / itemHeight) + 1,
        items.length
      );
      
      return items.slice(startIndex, endIndex).map((item, index) => ({
        item,
        index: startIndex + index,
        top: (startIndex + index) * itemHeight,
      }));
    }, [items, itemHeight, containerHeight, scrollTop]);

    return {
      visibleItems,
      totalHeight: items.length * itemHeight,
      onScroll: useCallback((e: React.UIEvent) => {
        setScrollTop(e.currentTarget.scrollTop);
      }, []),
    };
  }

  // Request optimization with React Query
  static useOptimizedQuery<T>(
    queryKey: string[],
    queryFn: () => Promise<T>,
    options: {
      staleTime?: number;
      cacheTime?: number;
      refetchOnWindowFocus?: boolean;
    } = {}
  ) {
    return useQuery({
      queryKey,
      queryFn,
      staleTime: options.staleTime || 5 * 60 * 1000, // 5 minutes
      cacheTime: options.cacheTime || 10 * 60 * 1000, // 10 minutes
      refetchOnWindowFocus: options.refetchOnWindowFocus || false,
      retry: 3,
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    });
  }
}
```

### Accessibility Implementation

```typescript
// Comprehensive accessibility implementation
export class AccessibilityManager {
  // ARIA attributes management
  static useAriaAttributes() {
    const [announcements, setAnnouncements] = useState<string[]>([]);

    const announce = useCallback((message: string, priority: 'polite' | 'assertive' = 'polite') => {
      setAnnouncements(prev => [...prev, { message, priority }]);
      setTimeout(() => {
        setAnnouncements(prev => prev.slice(1));
      }, 1000);
    }, []);

    return {
      announce,
      announcements,
    };
  }

  // Keyboard navigation implementation
  static useKeyboardNavigation(
    items: string[],
    onSelect: (item: string) => void
  ) {
    const [focusedIndex, setFocusedIndex] = useState(0);

    const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault();
          setFocusedIndex(prev => (prev + 1) % items.length);
          break;
        case 'ArrowUp':
          e.preventDefault();
          setFocusedIndex(prev => (prev - 1 + items.length) % items.length);
          break;
        case 'Enter':
        case ' ':
          e.preventDefault();
          onSelect(items[focusedIndex]);
          break;
        case 'Escape':
          e.preventDefault();
          setFocusedIndex(-1);
          break;
      }
    }, [items, focusedIndex, onSelect]);

    return {
      focusedIndex,
      handleKeyDown,
      setFocusedIndex,
    };
  }

  // Focus management
  static useFocusManagement() {
    const [focusableElements, setFocusableElements] = useState<HTMLElement[]>([]);

    useEffect(() => {
      const updateFocusableElements = () => {
        const elements = Array.from(
          document.querySelectorAll(
            'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
          )
        ) as HTMLElement[];
        setFocusableElements(elements);
      };

      updateFocusableElements();
      document.addEventListener('DOMContentLoaded', updateFocusableElements);
      
      return () => {
        document.removeEventListener('DOMContentLoaded', updateFocusableElements);
      };
    }, []);

    const trapFocus = useCallback((container: HTMLElement) => {
      const firstElement = container.querySelector(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
      ) as HTMLElement;
      
      if (firstElement) {
        firstElement.focus();
      }
    }, []);

    return {
      focusableElements,
      trapFocus,
    };
  }
}
```

### Internationalization Setup

```typescript
// Advanced i18n implementation with react-i18next
import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

// Resource configuration
const resources = {
  en: {
    translation: {
      welcome: 'Welcome',
      userManagement: 'User Management',
      addNewUser: 'Add New User',
      searchUsers: 'Search Users',
      noUsersFound: 'No users found',
      error: {
        fetchUsersFailed: 'Failed to fetch users',
        addUserFailed: 'Failed to add user',
      },
    },
  },
  es: {
    translation: {
      welcome: 'Bienvenido',
      userManagement: 'Gestión de Usuarios',
      addNewUser: 'Agregar Nuevo Usuario',
      searchUsers: 'Buscar Usuarios',
      noUsersFound: 'No se encontraron usuarios',
      error: {
        fetchUsersFailed: 'Error al obtener usuarios',
        addUserFailed: 'Error al agregar usuario',
      },
    },
  },
  fr: {
    translation: {
      welcome: 'Bienvenue',
      userManagement: 'Gestion des Utilisateurs',
      addNewUser: 'Ajouter un Nouvel Utilisateur',
      searchUsers: 'Rechercher des Utilisateurs',
      noUsersFound: 'Aucun utilisateur trouvé',
      error: {
        fetchUsersFailed: 'Échec de la récupération des utilisateurs',
        addUserFailed: 'Échec de l\'ajout d\'utilisateur',
      },
    },
  },
};

// Initialize i18n
i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'en',
    debug: process.env.NODE_ENV === 'development',
    
    interpolation: {
      escapeValue: false,
    },
    
    detection: {
      order: ['localStorage', 'navigator', 'htmlTag'],
      caches: ['localStorage'],
    },
  });

// Type-safe translation hook
export const useTranslation = () => {
  const { t } = i18next.useTranslation();
  
  return {
    t: (key: string, options?: i18n.TOptions) => t(key, options),
    changeLanguage: i18n.changeLanguage,
    currentLanguage: i18n.language,
  };
};

// Language switcher component
export const LanguageSwitcher: React.FC = () => {
  const { currentLanguage, changeLanguage } = useTranslation();
  
  const languages = [
    { code: 'en', name: 'English', flag: '🇺🇸' },
    { code: 'es', name: 'Español', flag: '🇪🇸' },
    { code: 'fr', name: 'Français', flag: '🇫🇷' },
  ];

  return (
    <div className="language-switcher">
      {languages.map((lang) => (
        <button
          key={lang.code}
          onClick={() => changeLanguage(lang.code)}
          className={currentLanguage === lang.code ? 'active' : ''}
          aria-label={`Switch to ${lang.name}`}
        >
          <span>{lang.flag}</span>
          <span>{lang.name}</span>
        </button>
      ))}
    </div>
  );
};
```

---

# Reference & Integration (Level 4)

## API Reference

### Core Frontend Operations
- `create_component(name, props, children)` - Create reusable component
- `use_custom_hook(logic, dependencies)` - Create custom React hook
- `optimize_performance(component)` - Apply performance optimizations
- `implement_accessibility(component)` - Add accessibility features
- `configure_i18n(languages, translations)` - Set up internationalization

### Context7 Integration
- `get_latest_react_documentation()` - React docs via Context7
- `analyze_frontend_patterns()` - Frontend best practices via Context7
- `optimize_component_architecture()` - Component optimization via Context7

## Best Practices (November 2025)

### DO
- Use TypeScript for type safety and better developer experience
- Implement proper component composition and reusability
- Optimize performance with code splitting and lazy loading
- Ensure accessibility with proper ARIA attributes and keyboard navigation
- Use modern React patterns (hooks, server components)
- Implement comprehensive error boundaries and error handling
- Use proper state management strategies for application complexity
- Ensure responsive design and cross-browser compatibility

### DON'T
- Skip TypeScript type definitions and validations
- Create overly complex component hierarchies
- Ignore performance optimization and code splitting
- Skip accessibility implementation and testing
- Use deprecated React patterns and class components
- Forget to implement proper error handling
- Overuse global state when local state is sufficient
- Skip testing for different devices and browsers

## Works Well With

- `moai-baas-foundation` (Enterprise frontend architecture)
- `moai-essentials-perf` (Performance optimization)
- `moai-lib-shadcn-ui` (Component library integration)
- `moai-domain-backend` (Backend API integration)
- `moai-security-api` (Frontend security implementation)
- `moai-foundation-trust` (Accessibility and compliance)
- `moai-domain-testing` (Frontend testing strategies)
- `moai-baas-vercel-ext` (Frontend deployment optimization)

## Changelog

- **v4.0.0** (2025-11-13): Complete Enterprise v4.0 rewrite with 40% content reduction, 4-layer Progressive Disclosure structure, Context7 integration, November 2025 frontend ecosystem updates, and modern React patterns
- **v2.0.0** (2025-11-11): Complete metadata structure, component patterns, performance optimization
- **v1.0.0** (2025-11-11): Initial frontend development domain

---

**End of Skill** | Updated 2025-11-13

## Security & Compliance

### Frontend Security
- Content Security Policy (CSP) implementation
- XSS protection with proper input sanitization
- Secure data handling and storage in browser
- Authentication token management and security

### Accessibility Compliance
- WCAG 2.1 AA compliance implementation
- Screen reader compatibility and ARIA support
- Keyboard navigation and focus management
- Color contrast and visual accessibility

---

**End of Enterprise Frontend Development Expert v4.0.0**
