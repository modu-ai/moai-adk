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


# Reference & Integration (Level 4)


## Core Implementation

## What It Does

Enterprise Frontend Development expert with AI-powered modern architecture, Context7 integration, and intelligent component orchestration for scalable user interfaces.

**Revolutionary  capabilities**:
- 🤖 **AI-Powered Frontend Architecture** using Context7 MCP for latest frontend patterns
- 📊 **Intelligent Component Orchestration** with automated design system optimization
- 🚀 **Modern Framework Integration** with AI-driven performance optimization
- 🔗 **Enterprise User Experience** with zero-configuration accessibility and internationalization
- 📈 **Predictive Performance Analytics** with usage forecasting and optimization insights


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

