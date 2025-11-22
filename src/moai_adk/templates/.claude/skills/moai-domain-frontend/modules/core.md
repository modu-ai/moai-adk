
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
