# React Native Development

Expo, React Navigation, and cross-platform patterns for React Native development.

## Quick Reference

Key Technologies:
- **Language**: TypeScript (primary), JavaScript
- **Framework**: React Native 0.74+ (New Architecture default), Expo SDK 51+
- **Navigation**: React Navigation v7 (`@react-navigation/native`)
- **State**: Zustand (simple), Redux Toolkit (complex), TanStack Query (server state)
- **Native modules**: Expo SDK first, Turbo Modules + JSI for custom bridges

Expo Workflow:
- **Managed**: `npx create-expo-app` — Expo manages native build, no Xcode/Android Studio required
- **Bare**: When custom native code required; use EAS Build for CI/CD

---

## Core Patterns

### Project Setup (Expo)

```bash
npx create-expo-app@latest MyApp --template blank-typescript
cd MyApp
npx expo install @react-navigation/native @react-navigation/native-stack
npx expo install react-native-screens react-native-safe-area-context
```

### Navigation Setup

```tsx
// app/_layout.tsx (Expo Router file-based routing)
import { Stack } from 'expo-router';

export default function RootLayout() {
  return (
    <Stack>
      <Stack.Screen name="(tabs)" options={{ headerShown: false }} />
      <Stack.Screen name="modal" options={{ presentation: 'modal' }} />
    </Stack>
  );
}
```

### State Management with Zustand

```typescript
import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import AsyncStorage from '@react-native-async-storage/async-storage';

interface UserStore {
  users: User[];
  isLoading: boolean;
  fetchUsers: () => Promise<void>;
}

export const useUserStore = create<UserStore>()(
  persist(
    (set) => ({
      users: [],
      isLoading: false,
      fetchUsers: async () => {
        set({ isLoading: true });
        try {
          const users = await UserApi.getAll();
          set({ users, isLoading: false });
        } catch {
          set({ isLoading: false });
        }
      },
    }),
    {
      name: 'user-store',
      storage: createJSONStorage(() => AsyncStorage),
    }
  )
);
```

### Platform-Specific Code

```typescript
import { Platform, StyleSheet } from 'react-native';

const styles = StyleSheet.create({
  container: {
    paddingTop: Platform.OS === 'ios' ? 44 : 0,
    // Platform.select for multiple values
    ...Platform.select({
      ios: { shadowColor: '#000', shadowOpacity: 0.1 },
      android: { elevation: 4 },
    }),
  },
});

// Platform-specific file: Component.ios.tsx / Component.android.tsx
// RN auto-selects based on platform
```

---

## Works Well With

- expert-mobile (primary agent)
- moai-domain-mobile/modules/strategy-comparison.md (paradigm selection)
- expert-security (certificate pinning with react-native-ssl-pinning, secure storage with expo-secure-store)
