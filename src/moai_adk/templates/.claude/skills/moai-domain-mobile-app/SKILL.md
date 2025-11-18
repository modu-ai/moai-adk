---
name: moai-domain-mobile-app
version: 4.0.0
created: 2025-11-13
updated: '2025-11-18'
status: stable
tier: domain
description: Enterprise mobile development with React Native 0.76+, Flutter 3.24+,
  Capacitor 6.x, cross-platform patterns, testing, CI/CD
allowed-tools: Read, Bash, WebSearch, WebFetch, mcp__context7__resolve-library-id,
  mcp__context7__get-library-docs
primary-agent: frontend-expert
secondary-agents:
- qa-validator
- alfred
- doc-syncer
tags:
- mobile
- react-native
- flutter
- capacitor
- ionic
- cross-platform
- testing
- deployment
orchestration: null
can_resume: true
typical_chain_position: middle
depends_on: []
stability: stable
---


# moai-domain-mobile-app

**Enterprise Mobile Development Excellence**

> **Version**: 4.0.0 (React Native 0.76+, Flutter 3.24+)
> **Technology Stack**: Modern async mobile development with CI/CD automation
> **Focus**: Cross-platform strategies, performance optimization, production deployment

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

**Purpose**: Enterprise mobile development covering React Native, Flutter, Capacitor with modern patterns.

**Core Stack (2025 Stable)**:
- **React Native**: 0.76+ (New Architecture default)
- **Flutter**: 3.24+ (Dart 3.5+)
- **Expo**: 52.x (EAS Build/Submit)
- **Navigation**: React Navigation 6.x, Flutter Navigator 2.x
- **State**: Provider 6.x, GetX 4.x, Redux Toolkit
- **Testing**: Detox 20.x (E2E), Jest 30.x (Unit)
- **CI/CD**: GitHub Actions, fastlane 2.x, EAS Build

**Quick Setup Patterns**:

**React Native App Structure**:
```typescript
// Main app with modern navigation
import { NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import AsyncStorage from '@react-native-async-storage/async-storage';

export const AppNavigator = () => (
  <NavigationContainer>
    <Stack.Navigator>
      <Stack.Screen name="Home" component={HomeScreen} />
      <Stack.Screen name="Profile" component={ProfileScreen} />
    </Stack.Navigator>
  </NavigationContainer>
);
```

**Flutter App Structure**:
```dart
import 'package:flutter/material.dart';
import 'package:flutter/services/navigation.dart';

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Flutter App',
      theme: ThemeData(primarySwatch: Colors.blue),
      home: MyHomePage(),
      routes: {
        '/': (context) => MyHomePage(),
        '/profile': (context) => ProfilePage(),
      },
    );
  }
}
```

**When to Use**:
- ✅ Native iOS/Android app development
- ✅ Cross-platform apps with shared business logic
- ✅ Web app deployment (Capacitor/Ionic)
- ✅ Performance optimization and battery management
- ✅ E2E automation and CI/CD integration

---

### Level 2: Practical Implementation (Essential Patterns)

#### Pattern 1: React Native Architecture (0.76+)

**Modern Setup**:
```bash
npx react-native@latest init MyApp --template
cd MyApp
npm install @react-navigation/native @react-native-async-storage @react-native-community/netinfo
```

**Async Data Management**:
```typescript
const [products, setProducts] = useState<Product[]>([]);
const [loading, setLoading] = useState(false);
const [isOnline, setIsOnline] = useState(true);

// Load from cache first, then API
const loadProducts = useCallback(async () => {
  try {
    setLoading(true);

    // Try cache first
    const cached = await AsyncStorage.getItem('products');
    if (cached) {
      setProducts(JSON.parse(cached));
    }

    // Fetch fresh data if online
    if (isOnline) {
      const response = await fetch('https://api.example.com/products');
      const data = await response.json();
      setProducts(data);
      await AsyncStorage.setItem('products', JSON.stringify(data));
    }
  } finally {
    setLoading(false);
  }
}, [isOnline]);
```

#### Pattern 2: Testing Strategy

**Detox E2E Testing**:
```typescript
// detox/e2e/app.config.js
module.exports = {
  testEnvironment: 'node',
  testRunner: 'jest',
  setupFilesAfterEnv: ['<rootDir>/jest.setup.js'],
};
```

**Component Testing**:
```typescript
// Component Testing with React Testing Library
import { render, screen, fireEvent } from '@testing-library/react-native';
import { renderHook } from '@testing-library/react-native');

it('should render correctly', () => {
  render(<HomeScreen />);
  expect(screen.getByText('Welcome')).toBeTruthy();
});
```

#### Pattern 3: Performance Optimization

**Image Optimization**:
```typescript
const OptimizedImage = ({ uri, style, resizeMode, width, height }) => {
  return (
    <Image
      source={{ uri }}
      style={style}
      resizeMode={resizeMode}
      width={width}
      height={height}
      onLoad={handleImageLoad}
      onError={() => console.log('Image failed to load')}
    />
  );
};
```

---

### Level 3: Advanced Integration (Complex Scenarios)

#### Pattern 1: Capacitor Plugin Development

**Plugin Structure**:
```typescript
// capacitor-plugin/src/index.ts
import { registerPlugin } from '@capacitor/core';

export interface CustomPluginPlugin {
  echo(options: { value: string }): Promise<{ value: string }>;
  openNativeScreen(): Promise<void>;
}

export const CustomPlugin = registerPlugin<CustomPlugin>('CustomPlugin', {
  web: () => import('./web').then(m => new m.CustomPluginWeb()),
});
```

**React Native Integration**:
```typescript
// capacitor.config.ts
import { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
  plugins: [
    {
      name: 'Camera',
      plugins: ['camera', 'filesystem']
    }
  ]
};

export default config;
```

#### Pattern 2: Cross-Platform Business Logic

**Shared Service Architecture**:
```typescript
// services/user-service.ts
export class UserService {
  async getUserProfile(userId: string): Promise<UserProfile> {
    // Shared implementation across platforms
    return await this.apiCall(`/users/${userId}`);
  }
}
```

**Platform-Specific Implementations**:
```typescript
// services/platform-specific/web-service.ts
export class WebUserService implements UserService {
  async getUserProfile(userId: string): Promise<UserProfile> {
    const response = await fetch(`/api/users/${userId}`);
    return response.json();
  }
}

// services/platform-specific/mobile-service.ts
export class MobileUserService implements UserService {
  async getUserProfile(userId: string): Promise<UserProfile> {
    const user = await this.realm.objects.get(userId);
    return user.profile || null;
  }
}
```

#### Pattern 3: Advanced Production Features

**Sentry Integration**:
```typescript
import * as Sentry from '@sentry/react-native';
import { ReactNode } from 'react';

Sentry.init({
  dsn: 'https://@[key]@[domain].ingest.sentry.io/[projectId]',
  tracesSampleRate: 1.0,
  environment: __DEV__ ? 'development' : 'production',
  integrations: [
    new Sentry.BrowserTracing(),
    new Sentry.ReactNavigationRoutingInstrumentation()
  ]
});
```

**Custom Metrics**:
```typescript
import Analytics from 'analytics-react-native';
import { MetricsCollector } from './metrics';

const AppAnalytics = Analytics.getAnalytics({
  trackAppSessions: true,
  trackScreen: true,
  trackUserProperties: true,
  trackExceptions: true
});

Analytics.track('screen_view', { screen: 'Home', platform: platform });
```

---

## 🎯 Decision Matrix

| Scenario | Best Choice | Why |
|----------|-------------|------|
| Dashboard App | Tabler Icons (5900+) | Optimized for admin UI |
| Tailwind Project | Heroicons (official) | Native integration |
| Multiple Weights | Phosphor (6 weights) | Design flexibility |
| 200K+ Icons | Iconify (universal) | Complete coverage |
| Compact UI | Radix Icons (~5KB) | Minimal bundle |

---

## 📊 Technology Comparison

| Framework | Icons | Bundle Size | Use Case | Type System |
|----------|-------|------------|-----------|------------|
| **React Native** | 1000+ | Native apps | TypeScript | Medium |
| **Flutter** | 800+ | Cross-platform | Dart | High |
| **Capacitor** | 500+ | Hybrid apps | Web + native | Medium |
| **Ionic** | 1300+ | Hybrid apps | Angular/React | Medium |

---

## 🔗 Related Skills

- `Skill("moai-domain-frontend")` – UI/UX integration patterns
- `Skill("moai-domain-testing")` - Mobile testing strategies
- `Skill("moai-domain-devops")` - CI/CD and deployment
- `Skill("moai-domain-backend")` - API integration and data management

---

**Version**: 4.0.0 Enterprise
**Last Updated**: 2025-11-18
**Status**: Production Ready
**Stack**: React Native 0.76+, Flutter 3.24+, Capacitor 6.x