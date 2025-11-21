---
name: moai-domain-frontend
description: Enterprise Frontend Development with AI-powered modern architecture,
---

## Quick Reference (30 seconds)

# Enterprise Frontend Development Expert 

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

---

## Core Implementation

## What It Does

Enterprise Frontend Development expert with AI-powered modern architecture, Context7 integration, and intelligent component orchestration for scalable user interfaces.

**Revolutionary  capabilities**:
- 🤖 **AI-Powered Frontend Architecture** using Context7 MCP for latest frontend patterns
- 📊 **Intelligent Component Orchestration** with automated design system optimization
- 🚀 **Modern Framework Integration** with AI-driven performance optimization
- 🔗 **Enterprise User Experience** with zero-configuration accessibility and internationalization
- 📈 **Predictive Performance Analytics** with usage forecasting and optimization insights

---

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



---

## Reference & Resources

See [reference.md](reference.md) for detailed API reference and official documentation.

## Context7 Integration

### Related Libraries & Tools
- [React](/facebook/react): JavaScript library for building user interfaces
- [Next.js](/vercel/next.js): The React Framework with App Router and Server Components
- [Zustand](/pmndrs/zustand): Small, fast and scalable state management solution
- [TanStack Query](/tanstack/query): Powerful data synchronization for React
- [shadcn/ui](/shadcn-ui/ui): Re-usable components built with Radix UI and Tailwind CSS

### Official Documentation
- [React 19](https://react.dev/)
- [Next.js 16](https://nextjs.org/docs)
- [Zustand](https://zustand-demo.pmnd.rs/)
- [TanStack Query](https://tanstack.com/query/latest)
- [Tailwind CSS](https://tailwindcss.com/docs)

### Version-Specific Guides
Latest stable version: React 19.x, Next.js 16.x, Zustand 5.x
- [React 19 Upgrade Guide](https://react.dev/blog/2024/12/05/react-19)
- [Next.js 16 Migration](https://nextjs.org/docs/app/building-your-application/upgrading)
- [Zustand v5 Migration](https://github.com/pmndrs/zustand/discussions/2200)
- [TanStack Query v5](https://tanstack.com/query/v5/docs/framework/react/guides/migrating-to-v5)

