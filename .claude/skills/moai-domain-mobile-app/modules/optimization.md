# Mobile App Performance Optimization

## App Startup Optimization

### React Native Lazy Loading

```typescript
// index.js - Optimized startup
import { AppRegistry } from 'react-native';
import App from './src/App';

// Defer non-critical imports
setTimeout(() => {
  require('./src/analytics');
  require('./src/crashReporting');
}, 2000);

AppRegistry.registerComponent('MyApp', () => App);
```

### Flutter Initialization

```dart
// lib/main.dart - Lazy initialization
Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Only initialize critical services
  await Firebase.initializeApp();

  // Defer other services
  Future.delayed(Duration(milliseconds: 500), () async {
    await _initializeAnalytics();
    await _initializeCrashReporting();
  });

  runApp(const MyApp());
}
```

## Memory Optimization

### React Native Memory Management

```typescript
// Cleanup subscription in useEffect
import { useEffect } from 'react';

export function MyScreen() {
  useEffect(() => {
    const subscription = eventEmitter.addListener('event', handleEvent);

    return () => {
      subscription.remove(); // Cleanup
    };
  }, []);

  return <View />;
}

// Efficient image loading
import { Image } from 'react-native';

export function UserAvatar({ uri }: { uri: string }) {
  return (
    <Image
      source={{ uri }}
      style={{ width: 100, height: 100 }}
      onLoad={() => console.log('Image loaded')}
      onError={() => console.log('Image failed to load')}
    />
  );
}
```

### Flutter Memory Monitoring

```dart
// Monitoring memory usage
import 'package:flutter/foundation.dart';
import 'dart:developer' as developer;

class MemoryMonitor {
  static Future<void> checkMemory() async {
    final info = await developer.Timeline.instantSync('memory_check');
    final timestamp = DateTime.now();
    debugPrint('Memory check at $timestamp: $info');
  }
}
```

## Battery Optimization

### Reduce Background Activity

```typescript
// React Native - Background task optimization
import BackgroundJob from 'react-native-background-job';

BackgroundJob.register({
  jobKey: 'syncData',
  job: () => {
    // Sync only when battery level > 20%
    return fetch('/api/sync').then(() => {
      console.log('Sync completed');
    });
  },
});

// Schedule with optimization
BackgroundJob.schedule({
  jobKey: 'syncData',
  period: 900000, // 15 minutes
  allowExecutionInForeground: false,
});
```

## Network Optimization

### Efficient API Requests

```typescript
// React Native - Request batching and caching
import axios from 'axios';

const api = axios.create({
  timeout: 10000,
});

// Request batching
const requestQueue: Array<Promise<any>> = [];

export async function batchRequest(url: string) {
  const request = api.get(url);
  requestQueue.push(request);

  if (requestQueue.length > 5) {
    await Promise.all(requestQueue);
    requestQueue.length = 0;
  }

  return request;
}

// Response caching
const cache = new Map<string, any>();

export async function cachedFetch(url: string, ttl = 300000) {
  const cached = cache.get(url);
  if (cached && Date.now() - cached.timestamp < ttl) {
    return cached.data;
  }

  const response = await fetch(url);
  const data = await response.json();
  cache.set(url, { data, timestamp: Date.now() });
  return data;
}
```

## Build Size Optimization

### Android APK Size Reduction

```gradle
// app/build.gradle
android {
    compileSdk 34

    packagingOptions {
        exclude 'META-INF/proguard/androidx-*.pro'
    }

    bundle {
        language.enableSplit = true
        density.enableSplit = true
    }

    buildTypes {
        release {
            minifyEnabled true
            shrinkResources true
            proguardFiles getDefaultProguardFile('proguard-android-optimize.txt'),
                         'proguard-rules.pro'
        }
    }
}
```

### iOS Build Optimization

```bash
# Strip unused assets
find . -name "*.png" -o -name "*.jpg" | \
  xargs du -sh | sort -hr | head -20

# App Thinning configuration (Xcode)
# Build Settings > Product Name > Strip Debug Symbols During Copy
```

## Profiling Tools

### React Native Profiler

```typescript
// Enable Hermes profiling
import { enableDebuggerDelegate } from 'react-native/Libraries/Core/HermesInternal';

if (__DEV__) {
  enableDebuggerDelegate({
    logFunction: (msg) => console.log(msg),
  });
}
```

### Flutter DevTools

```bash
# Connect and profile
flutter pub global run devtools
flutter run --profile

# Analyze performance timeline
# DevTools > Performance tab > Frame rendering
```

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
**Status**: Production Ready

## Context7 Integration

### Performance Monitoring Tools
- [Sentry](/getsentry/sentry): Error and performance monitoring
- [Firebase Performance](/firebase/firebase): Real-time performance monitoring
- [New Relic](/newrelic/newrelic): App performance management
