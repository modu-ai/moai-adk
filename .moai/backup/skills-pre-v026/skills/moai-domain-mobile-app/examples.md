# moai-domain-mobile-app - Code Examples

**Version**: 4.0.0  
**Updated**: 2025-11-12

---

## Example 1: React Native 0.76+ with TypeScript

Complete app with modern patterns, navigation, and error handling.

```typescript
// App.tsx
import React from 'react';
import { SafeAreaView, StatusBar } from 'react-native';
import { NavigationContainer, NavigationProp } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import Sentry from '@sentry/react-native';

// Initialize Sentry for error tracking
Sentry.init({
  dsn: 'YOUR_SENTRY_DSN',
  environment: 'production',
  tracesSampleRate: 1.0,
});

const Stack = createNativeStackNavigator();

const AppNavigator = () => {
  return (
    <Stack.Navigator
      screenOptions={{
        headerShown: true,
        animationEnabled: true,
      }}
    >
      <Stack.Screen name="Home" component={HomeScreen} />
      <Stack.Screen name="Profile" component={ProfileScreen} />
    </Stack.Navigator>
  );
};

const App = () => {
  return (
    <SafeAreaView style={{ flex: 1 }}>
      <StatusBar barStyle="dark-content" />
      <NavigationContainer
        ref={navigationRef}
        onReady={() => {
          // Set up navigation state for Sentry
          Sentry.captureReactNavigationBreadcrumbs(navigationRef);
        }}
      >
        <AppNavigator />
      </NavigationContainer>
    </SafeAreaView>
  );
};

export default Sentry.wrap(App);
```

---

## Example 2: Flutter State Management with Provider

```dart
// models/user_model.dart
class User {
  final String id;
  final String name;
  final String email;

  User({
    required this.id,
    required this.name,
    required this.email,
  });

  User copyWith({String? id, String? name, String? email}) {
    return User(
      id: id ?? this.id,
      name: name ?? this.name,
      email: email ?? this.email,
    );
  }
}

// providers/user_provider.dart
import 'package:flutter/foundation.dart';
import 'package:provider/provider.dart';

class UserProvider with ChangeNotifier {
  User? _user;
  bool _isLoading = false;
  String? _error;

  User? get user => _user;
  bool get isLoading => _isLoading;
  String? get error => _error;

  Future<void> loadUser(String userId) async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      // Simulate API call
      await Future.delayed(Duration(seconds: 1));
      _user = User(id: userId, name: 'John Doe', email: 'john@example.com');
    } catch (e) {
      _error = e.toString();
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<void> updateUser(User updatedUser) async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      // Simulate API call
      await Future.delayed(Duration(seconds: 1));
      _user = updatedUser;
    } catch (e) {
      _error = e.toString();
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }
}

// screens/user_screen.dart
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

class UserScreen extends StatefulWidget {
  final String userId;

  const UserScreen({required this.userId});

  @override
  State<UserScreen> createState() => _UserScreenState();
}

class _UserScreenState extends State<UserScreen> {
  @override
  void initState() {
    super.initState();
    // Load user when screen initializes
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<UserProvider>().loadUser(widget.userId);
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('User Profile')),
      body: Consumer<UserProvider>(
        builder: (context, userProvider, child) {
          if (userProvider.isLoading) {
            return Center(child: CircularProgressIndicator());
          }

          if (userProvider.error != null) {
            return Center(child: Text('Error: ${userProvider.error}'));
          }

          final user = userProvider.user;
          if (user == null) {
            return Center(child: Text('No user found'));
          }

          return Padding(
            padding: EdgeInsets.all(16),
            child: Column(
              children: [
                Text(user.name, style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold)),
                SizedBox(height: 8),
                Text(user.email, style: TextStyle(fontSize: 16, color: Colors.grey)),
                SizedBox(height: 24),
                ElevatedButton(
                  onPressed: () {
                    // Show edit dialog
                    _showEditDialog(context, user);
                  },
                  child: Text('Edit Profile'),
                ),
              ],
            ),
          );
        },
      ),
    );
  }

  void _showEditDialog(BuildContext context, User user) {
    final nameController = TextEditingController(text: user.name);
    final emailController = TextEditingController(text: user.email);

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('Edit Profile'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(controller: nameController, decoration: InputDecoration(labelText: 'Name')),
            SizedBox(height: 8),
            TextField(controller: emailController, decoration: InputDecoration(labelText: 'Email')),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: Text('Cancel')),
          TextButton(
            onPressed: () {
              final updatedUser = user.copyWith(
                name: nameController.text,
                email: emailController.text,
              );
              context.read<UserProvider>().updateUser(updatedUser);
              Navigator.pop(context);
            },
            child: Text('Save'),
          ),
        ],
      ),
    );
  }
}
```

---

## Example 3: Detox E2E Testing

```typescript
// e2e/firstTest.e2e.ts
describe('Product List E2E Test', () => {
  beforeAll(async () => {
    await device.launchApp();
  });

  beforeEach(async () => {
    await device.reloadReactNative();
  });

  afterEach(async () => {
    await device.sendToHome();
  });

  it('should display and interact with product list', async () => {
    // Wait for list header to appear
    await waitFor(element(by.text('Products')))
      .toBeVisible()
      .withTimeout(5000);

    // Check if first product is visible
    await expect(element(by.id('product-0'))).toBeVisible();

    // Scroll down to load more products
    await element(by.id('productList')).multiTap();
    await element(by.id('productList')).swipe('up');

    // Tap on a product
    await element(by.id('product-1')).tap();

    // Verify navigation to detail screen
    await expect(element(by.text('Product Details'))).toBeVisible();

    // Check if back button works
    await element(by.text('Back')).tap();

    // Verify back to list
    await expect(element(by.id('productList'))).toBeVisible();
  });

  it('should handle offline scenario', async () => {
    // Enable airplane mode
    await device.setAirplaneMode(true);

    // Navigate to products
    await element(by.id('productsTab')).tap();

    // Should show offline banner
    await expect(element(by.id('offlineBanner'))).toBeVisible();

    // Disable airplane mode
    await device.setAirplaneMode(false);

    // Should load fresh data
    await waitFor(element(by.id('productList')))
      .toBeVisible()
      .withTimeout(10000);
  });

  it('should handle errors gracefully', async () => {
    // Mock server error by blocking network
    await device.reverseTcpPort(8000);

    // Try to load products
    await element(by.id('loadButton')).tap();

    // Should show error message
    await expect(element(by.text('Error loading products'))).toBeVisible();

    // Restore network
    await device.reverseTcpPort(8000);

    // Try again
    await element(by.id('retryButton')).tap();

    // Should load successfully
    await expect(element(by.id('productList'))).toBeVisible();
  });
});
```

---

## Example 4: Capacitor Native Plugin Integration

```typescript
// capacitor/battery-plugin/BatteryPlugin.ts
import { registerPlugin } from '@capacitor/core';

export interface BatteryPlugin {
  getBatteryLevel(): Promise<{ level: number }>;
  getBatteryStatus(): Promise<{ 
    level: number; 
    isCharging: boolean; 
    chargingTime: number 
  }>;
  monitorBattery(callback: (status: BatteryStatus) => void): Promise<void>;
}

export interface BatteryStatus {
  level: number;
  isCharging: boolean;
  chargingTime: number;
  timestamp: number;
}

export const Battery = registerPlugin<BatteryPlugin>('Battery', {
  web: () => import('./web').then(m => new m.BatteryWeb()),
});

// Usage in React
import { useEffect, useState } from 'react';
import { Battery, BatteryStatus } from './capacitor/battery-plugin/BatteryPlugin';

export const BatteryMonitor: React.FC = () => {
  const [batteryStatus, setBatteryStatus] = useState<BatteryStatus | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const initBatteryMonitoring = async () => {
      try {
        const result = await Battery.getBatteryStatus();
        setBatteryStatus({
          ...result,
          timestamp: Date.now(),
        });

        // Set up continuous monitoring
        await Battery.monitorBattery((status) => {
          setBatteryStatus(status);
        });
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
      }
    };

    initBatteryMonitoring();
  }, []);

  if (error) {
    return <Text>Error: {error}</Text>;
  }

  if (!batteryStatus) {
    return <Text>Loading battery status...</Text>;
  }

  return (
    <View>
      <Text>Battery Level: {batteryStatus.level}%</Text>
      <Text>Charging: {batteryStatus.isCharging ? 'Yes' : 'No'}</Text>
      {batteryStatus.chargingTime > 0 && (
        <Text>Time to Full: {batteryStatus.chargingTime} minutes</Text>
      )}
    </View>
  );
};
```

---

## Example 5: CI/CD with EAS and fastlane

```bash
# eas.json configuration
{
  "cli": {
    "version": ">= 3.0.0"
  },
  "build": {
    "preview": {
      "ios": {
        "resourceClass": "default"
      },
      "android": {
        "resourceClass": "default"
      }
    },
    "production": {
      "ios": {
        "resourceClass": "m1-medium",
        "scheme": "Release",
        "buildConfiguration": "Release"
      },
      "android": {
        "resourceClass": "large",
        "gradleCommand": ":app:bundleRelease"
      }
    }
  },
  "submit": {
    "production": {
      "ios": {
        "ascAppId": "1234567890"
      },
      "android": {
        "googlePlayAccount": "my-service-account"
      }
    }
  }
}

# Build and submit
eas build --platform ios --profile production
eas submit --platform ios --latest

eas build --platform android --profile production
eas submit --platform android --latest
```

```ruby
# fastlane/Fastfile for iOS
default_platform(:ios)

platform :ios do
  desc "Build and upload to TestFlight"
  lane :beta do |options|
    # Increment build number
    increment_build_number(xcodeproj: "ios/MyApp.xcodeproj")

    # Build the app
    build_app(
      workspace: "ios/MyApp.xcworkspace",
      scheme: "MyApp",
      configuration: "Release",
      destination: "generic/platform=iOS",
      export_method: "app-store",
      output_name: "MyApp.ipa",
      clean: true
    )

    # Upload dSYMs to Sentry
    upload_symbols_to_sentry(
      auth_token: ENV['SENTRY_AUTH_TOKEN'],
      org_slug: 'my-org',
      project_slug: 'my-app'
    )

    # Upload to TestFlight
    upload_to_testflight(skip_submission: true)

    # Notify team
    notify_slack(message: "iOS build uploaded to TestFlight")
  end

  desc "Build and upload to App Store"
  lane :release do
    match(type: "appstore")
    beta
    deliver(submit_for_review: true)
  end
end
```

---

## Example 6: Error Tracking with Sentry

```typescript
// sentry-setup.ts
import * as Sentry from '@sentry/react-native';
import { NavigationContainer, NavigationContainerRef } from '@react-navigation/native';
import React from 'react';

export const navigationRef = React.createRef<NavigationContainerRef<any>>();

Sentry.init({
  dsn: 'YOUR_SENTRY_DSN',
  environment: __DEV__ ? 'development' : 'production',
  tracesSampleRate: 1.0,
  integrations: [
    new Sentry.ReactNativeTracing({
      routingInstrumentation: Sentry.registerNavigationContainer(navigationRef),
      tracingOrigins: [
        'localhost',
        /^\//,
        /^https:\/\/api\.example\.com\/api/,
      ],
    }),
  ],
  beforeSend(event, hint) {
    // Filter sensitive data
    if (event.request?.url?.includes('/auth')) {
      return null;
    }
    return event;
  },
});

// Set user context for error tracking
Sentry.setUser({
  id: 'user-123',
  email: 'user@example.com',
  username: 'username',
});

// Capture custom exceptions
export const captureError = (error: Error, context?: Record<string, any>) => {
  Sentry.captureException(error, {
    level: 'error',
    tags: {
      section: context?.section || 'unknown',
    },
    contexts: {
      custom: context || {},
    },
  });
};

// Measure performance
export const measureScreenLoad = (screenName: string) => {
  const transaction = Sentry.startTransaction({
    name: `Screen: ${screenName}`,
    op: 'navigation',
  });

  return () => {
    transaction.finish();
  };
};

// Usage
const handleScreenLoad = () => {
  const endMeasurement = measureScreenLoad('HomeScreen');
  
  setTimeout(() => {
    endMeasurement();
  }, 2000);
};
```

---

## Example 7: Performance Optimization Patterns

```typescript
// optimized-list.tsx
import React, { useCallback, useMemo } from 'react';
import { FlatList, View, Text, Image, StyleSheet } from 'react-native';

interface ListItem {
  id: string;
  title: string;
  imageUrl: string;
  description: string;
}

interface OptimizedListProps {
  data: ListItem[];
  onEndReached: () => void;
  isLoading: boolean;
}

const ListItemComponent = React.memo<{ item: ListItem }>(({ item }) => (
  <View style={styles.item}>
    <Image 
      source={{ uri: item.imageUrl }} 
      style={styles.image}
      cache="force-cache"
    />
    <View style={styles.content}>
      <Text style={styles.title}>{item.title}</Text>
      <Text style={styles.description}>{item.description}</Text>
    </View>
  </View>
));

export const OptimizedList: React.FC<OptimizedListProps> = ({
  data,
  onEndReached,
  isLoading,
}) => {
  const renderItem = useCallback(
    ({ item }: { item: ListItem }) => <ListItemComponent item={item} />,
    []
  );

  const keyExtractor = useCallback((item: ListItem) => item.id, []);

  // Memoize the list to prevent unnecessary re-renders
  const memoizedData = useMemo(() => data, [data]);

  return (
    <FlatList
      data={memoizedData}
      renderItem={renderItem}
      keyExtractor={keyExtractor}
      // Performance optimizations
      removeClippedSubviews={true}
      maxToRenderPerBatch={20}
      updateCellsBatchingPeriod={50}
      initialNumToRender={20}
      windowSize={10}
      scrollEventThrottle={16}
      // Handle pagination
      onEndReached={() => !isLoading && onEndReached()}
      onEndReachedThreshold={0.5}
      // Loading indicator
      ListFooterComponent={
        isLoading ? (
          <View style={styles.loader}>
            <Text>Loading...</Text>
          </View>
        ) : null
      }
    />
  );
};

const styles = StyleSheet.create({
  item: {
    flexDirection: 'row',
    padding: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#ddd',
  },
  image: {
    width: 80,
    height: 80,
    borderRadius: 4,
    marginRight: 12,
  },
  content: {
    flex: 1,
    justifyContent: 'center',
  },
  title: {
    fontSize: 16,
    fontWeight: '600',
    marginBottom: 4,
  },
  description: {
    fontSize: 14,
    color: '#666',
  },
  loader: {
    padding: 20,
    alignItems: 'center',
  },
});
```

---

**Version**: 4.0.0  
**Last Updated**: 2025-11-12  
**Example Count**: 7 Complete Examples  
**Code Quality**: Production-Ready
