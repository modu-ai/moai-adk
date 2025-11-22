# Mobile App Advanced Patterns

## React Native Advanced Patterns

### New Architecture and Bridging

```typescript
// Native Module (Kotlin)
package com.myapp

import com.facebook.react.bridge.ReactApplicationContext
import com.facebook.react.bridge.ReactContextBaseJavaModule
import com.facebook.react.bridge.ReactMethod
import com.facebook.react.bridge.Promise

class NativeBridge(reactContext: ReactApplicationContext) : ReactContextBaseJavaModule(reactContext) {
    override fun getName() = "NativeBridge"

    @ReactMethod
    fun getNativeData(promise: Promise) {
        try {
            val data = mapOf(
                "platform" to "android",
                "timestamp" to System.currentTimeMillis()
            )
            promise.resolve(data)
        } catch (e: Exception) {
            promise.reject("ERROR", e)
        }
    }
}

// React Native Usage
import { NativeModules } from 'react-native';

const { NativeBridge } = NativeModules;

export async function getNativeInfo() {
    try {
        const data = await NativeBridge.getNativeData();
        console.log('Native data:', data);
    } catch (error) {
        console.error('Failed to get native data:', error);
    }
}
```

## Flutter Advanced State Management

### GetX Reactive Programming

```dart
// lib/controllers/auth_controller.dart
import 'package:get/get.dart';

class AuthController extends GetxController {
  final email = ''.obs;
  final password = ''.obs;
  final isLoading = false.obs;
  final authToken = Rx<String?>(null);

  // Computed values
  late Rx<bool> isFormValid = Rx(false);

  @override
  void onInit() {
    super.onInit();
    ever(email, (_) => _validateForm());
    ever(password, (_) => _validateForm());
  }

  void _validateForm() {
    isFormValid.value = email.value.isNotEmpty &&
        password.value.length >= 8;
  }

  Future<void> login() async {
    try {
      isLoading.value = true;
      final response = await http.post(
        Uri.parse('/api/login'),
        body: {'email': email.value, 'password': password.value},
      );

      if (response.statusCode == 200) {
        authToken.value = jsonDecode(response.body)['token'];
        Get.offAllNamed('/home');
      } else {
        Get.snackbar('Error', 'Login failed');
      }
    } finally {
      isLoading.value = false;
    }
  }
}

// lib/screens/login_screen.dart
class LoginScreen extends GetView<AuthController> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Form(
        child: Column(
          children: [
            TextField(
              onChanged: (val) => controller.email.value = val,
            ),
            TextField(
              onChanged: (val) => controller.password.value = val,
              obscureText: true,
            ),
            Obx(() => ElevatedButton(
              onPressed: controller.isFormValid.value ? controller.login : null,
              child: controller.isLoading.value
                  ? CircularProgressIndicator()
                  : Text('Login'),
            )),
          ],
        ),
      ),
    );
  }
}
```

## Offline-First Architecture

### Synchronization Strategy

```typescript
// React Native with Realm
import Realm from 'realm';

const UserSchema = {
  name: 'User',
  primaryKey: 'id',
  properties: {
    id: 'string',
    name: 'string',
    email: 'string',
    syncedAt: 'date?',
    isDirty: { type: 'bool', default: false },
  },
};

export class SyncManager {
  private realm: Realm;

  constructor(realm: Realm) {
    this.realm = realm;
  }

  async syncChanges() {
    const dirtyUsers = this.realm.objects('User').filtered('isDirty = true');

    for (const user of dirtyUsers) {
      try {
        await fetch('/api/users', {
          method: 'PUT',
          body: JSON.stringify(user),
        });

        this.realm.write(() => {
          user.isDirty = false;
          user.syncedAt = new Date();
        });
      } catch (error) {
        console.error('Sync failed:', error);
      }
    }
  }
}
```

## Performance Optimization Patterns

### Memory Management

```typescript
// React Native - Memory-efficient list rendering
import { FlatList } from 'react-native';

export function UserList({ users }: { users: User[] }) {
  return (
    <FlatList
      data={users}
      renderItem={({ item }) => <UserCard user={item} />}
      keyExtractor={(item) => item.id}
      maxToRenderPerBatch={10}
      updateCellsBatchingPeriod={50}
      initialNumToRender={20}
      onEndReachedThreshold={0.5}
      onEndReached={loadMore}
      removeClippedSubviews={true}
    />
  );
}
```

### Native Performance Monitoring

```dart
// Flutter with Firebase Performance
import 'package:firebase_performance/firebase_performance.dart';

class PerformanceMonitoring {
  static Future<void> trackScreenLoad(String screenName) async {
    final trace = FirebasePerformance.instance.newTrace(screenName);
    await trace.start();

    // Simulate work
    await Future.delayed(Duration(milliseconds: 500));

    trace.putAttribute('screen_name', screenName);
    trace.putMetric('load_time', 500);

    await trace.stop();
  }
}
```

## Context7 Integration

### Latest Mobile Frameworks

The Context7 MCP provides access to:
- **React Native**: New Architecture, Bridging patterns, Fast Refresh
- **Flutter**: Latest state management (Riverpod, GetX), Native integration
- **SwiftUI**: Modern iOS patterns, Combine framework
- **Kotlin**: Modern Android Jetpack, Coroutines

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
**Status**: Production Ready

## Context7 Integration

### Related Libraries & Tools
- [React Native](/facebook/react-native): Cross-platform mobile framework
- [Flutter](/flutter/flutter): Google's UI framework
- [Expo](/expo/expo): React Native platform
- [Firebase](/firebase/firebase): Backend services
