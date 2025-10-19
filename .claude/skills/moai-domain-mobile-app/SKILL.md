---
name: moai-domain-mobile-app
description: Mobile app development with Flutter and React Native, state management, and native integration
allowed-tools:
  - Read
  - Bash
tier: 2
auto-load: "true"
---

# Mobile App Expert

## What it does

Provides expertise in cross-platform mobile app development using Flutter (Dart) and React Native (TypeScript), including state management patterns and native module integration.

## When to use

- "모바일 앱 개발", "Flutter 위젯", "React Native 컴포넌트", "상태 관리", "BLoC", "네이티브 모듈", "반응형 레이아웃", "크로스 플랫폼"
- "Mobile app development", "Flutter", "React Native", "State management", "Native integration", "Cross-platform"
- Automatically invoked when working with mobile app projects
- Mobile app SPEC implementation (`/alfred:2-run`)

- "모바일 앱 개발", "Flutter 위젯", "React Native 컴포넌트", "상태 관리"
- Automatically invoked when working with mobile app projects
- Mobile app SPEC implementation (`/alfred:2-run`)

## How it works

**Flutter Development**:
- **Widget tree**: StatelessWidget, StatefulWidget
- **State management**: Provider, Riverpod, BLoC
- **Navigation**: Navigator 2.0, go_router
- **Platform-specific code**: MethodChannel

**React Native Development**:
- **Components**: Functional components with hooks
- **State management**: Redux, MobX, Zustand
- **Navigation**: React Navigation
- **Native modules**: Turbo modules, JSI

**Cross-Platform Patterns**:
- **Responsive design**: Adaptive layouts for phone/tablet
- **Performance optimization**: Lazy loading, memoization
- **Offline support**: Local storage, sync strategies
- **Testing**: Widget tests (Flutter), component tests (RN)

**Native Integration**:
- **Plugins**: Platform channels, native modules
- **Permissions**: Camera, location, notifications
- **Deep linking**: Universal links, app links
- **Push notifications**: FCM, APNs

## TDD Workflow for Mobile Development

### Phase 1: RED (Widget Test - Component Contract)
```dart
// @TEST:MOBILE-001 | SPEC: SPEC-MOBILE-001.md
import 'package:flutter_test/flutter_test.dart';
import 'package:myapp/features/auth/presentation/pages/login_page.dart';

void main() {
  testWidgets('LoginPage should show username and password fields', (tester) async {
    // RED: 로그인 페이지가 필요한 위젯을 제공해야 함
    await tester.pumpWidget(const MaterialApp(
      home: LoginPage(),
    ));

    expect(find.byType(TextField), findsWidgets);  // 2개 찾아야 함 (username, password)
    expect(find.byType(ElevatedButton), findsOneWidget);  // 로그인 버튼
  });
}
```

### Phase 2: GREEN (Widget Implementation)
```dart
// @CODE:MOBILE-001 | SPEC: SPEC-MOBILE-001.md | TEST: test/login_page_test.dart
class LoginPage extends StatelessWidget {
  const LoginPage({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Login')),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          children: [
            TextField(
              decoration: const InputDecoration(labelText: 'Username'),
            ),
            TextField(
              decoration: const InputDecoration(labelText: 'Password'),
              obscureText: true,
            ),
            ElevatedButton(
              onPressed: () {},
              child: const Text('Login'),
            ),
          ],
        ),
      ),
    );
  }
}
```

### Phase 3: REFACTOR (BLoC Pattern Integration)
```dart
// @CODE:MOBILE-001:REFACTOR | BLoC 패턴 적용
class LoginPage extends StatelessWidget {
  const LoginPage({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: BlocListener<LoginBloc, LoginState>(
        listener: (context, state) {
          if (state is LoginSuccess) {
            Navigator.pushReplacementNamed(context, '/home');
          }
        },
        child: BlocBuilder<LoginBloc, LoginState>(
          builder: (context, state) => Padding(
            padding: const EdgeInsets.all(16.0),
            child: Column(
              children: [
                // username/password 입력
                if (state is LoginLoading)
                  const CircularProgressIndicator()
                else
                  ElevatedButton(
                    onPressed: () => context.read<LoginBloc>().add(
                      LoginButtonPressed(username: '', password: ''),
                    ),
                    child: const Text('Login'),
                  )
              ],
            ),
          ),
        ),
      ),
    );
  }
}

// 효과: 상태 관리 명확화 + 테스트 용이성 ✅
```

## Examples

### Example 1: Flutter App with BLoC State Management

**RED (Widget Test)**:
```dart
// @TEST:MOBILE-001 | SPEC: SPEC-MOBILE-001.md
testWidgets('Counter increments when button is pressed', (tester) async {
  await tester.pumpWidget(
    MaterialApp(
      home: BlocProvider(
        create: (context) => CounterBloc(),
        child: const CounterPage(),
      ),
    ),
  );

  expect(find.text('0'), findsOneWidget);  // 초기값

  await tester.tap(find.byIcon(Icons.add));
  await tester.pump();

  expect(find.text('1'), findsOneWidget);  // 증가함
});
```

**GREEN + REFACTOR (BLoC Implementation)**:
```dart
// @CODE:MOBILE-001 | BLoC 구현
// lib/features/counter/bloc/counter_event.dart
abstract class CounterEvent extends Equatable {}

class CounterIncrementPressed extends CounterEvent {
  @override
  List<Object?> get props => [];
}

// lib/features/counter/bloc/counter_state.dart
class CounterState extends Equatable {
  final int counter;
  const CounterState(this.counter);

  @override
  List<Object?> get props => [counter];
}

// lib/features/counter/bloc/counter_bloc.dart
class CounterBloc extends Bloc<CounterEvent, CounterState> {
  CounterBloc() : super(const CounterState(0)) {
    on<CounterIncrementPressed>((event, emit) {
      emit(CounterState(state.counter + 1));
    });
  }
}

// lib/features/counter/presentation/pages/counter_page.dart
class CounterPage extends StatelessWidget {
  const CounterPage({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Counter')),
      body: Center(
        child: BlocBuilder<CounterBloc, CounterState>(
          builder: (context, state) => Text(
            '${state.counter}',
            style: const TextStyle(fontSize: 48),
          ),
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => context.read<CounterBloc>().add(CounterIncrementPressed()),
        child: const Icon(Icons.add),
      ),
    );
  }
}
```

### Example 2: React Native with Redux State Management

**Redux + TypeScript**:
```typescript
// @CODE:MOBILE-002 | React Native Redux (TypeScript)
import { createSlice, PayloadAction } from '@reduxjs/toolkit';

interface AuthState {
  isLoggedIn: boolean;
  user: { id: string; name: string } | null;
  loading: boolean;
  error: string | null;
}

const initialState: AuthState = {
  isLoggedIn: false,
  user: null,
  loading: false,
  error: null,
};

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    loginStart: (state) => {
      state.loading = true;
      state.error = null;
    },
    loginSuccess: (state, action: PayloadAction<AuthState['user']>) => {
      state.loading = false;
      state.isLoggedIn = true;
      state.user = action.payload;
    },
    loginFailure: (state, action: PayloadAction<string>) => {
      state.loading = false;
      state.error = action.payload;
    },
  },
});

// API 호출 (비동기)
export const login = (email: string, password: string) => async (dispatch: any) => {
  dispatch(authSlice.actions.loginStart());
  try {
    const response = await fetch('https://api.example.com/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });
    const user = await response.json();
    dispatch(authSlice.actions.loginSuccess(user));
  } catch (error: any) {
    dispatch(authSlice.actions.loginFailure(error.message));
  }
};
```

**Component Usage**:
```typescript
// @CODE:MOBILE-002:UI | React Native 컴포넌트
import { useDispatch, useSelector } from 'react-redux';
import { login } from './authSlice';

export const LoginScreen: React.FC = () => {
  const dispatch = useDispatch();
  const { loading, error } = useSelector((state: RootState) => state.auth);

  const handleLogin = () => {
    dispatch(login('user@example.com', 'password123') as any);
  };

  return (
    <View style={styles.container}>
      {loading ? (
        <ActivityIndicator size="large" />
      ) : (
        <>
          {error && <Text style={styles.error}>{error}</Text>}
          <TouchableOpacity onPress={handleLogin} style={styles.button}>
            <Text>Login</Text>
          </TouchableOpacity>
        </>
      )}
    </View>
  );
};
```

### Example 3: Cross-Platform Responsive Design

**Before (Platform-specific hardcoding)**:
```dart
// ❌ 비효율: 각 플랫폼마다 다른 레이아웃
class HomePage extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    if (Platform.isIOS) {
      return iOSLayout();
    } else if (Platform.isAndroid) {
      return androidLayout();
    }
    return webLayout();
  }
}
```

**After (Adaptive layout with responsive breakpoints)**:
```dart
// ✅ 개선: 하나의 적응형 레이아웃
class HomePage extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    final screenWidth = MediaQuery.of(context).size.width;
    final isSmall = screenWidth < 600;
    final isMedium = screenWidth >= 600 && screenWidth < 1024;
    final isLarge = screenWidth >= 1024;

    return Scaffold(
      body: isLarge
          ? Row(
              children: [
                SizedBox(width: 300, child: Sidebar()),
                Expanded(child: MainContent()),
              ],
            )
          : isMedium
              ? Column(
                  children: [
                    Expanded(child: MainContent()),
                    SizedBox(height: 80, child: BottomNav()),
                  ],
                )
              : SingleChildScrollView(
                  child: Column(
                    children: [
                      MainContent(),
                      BottomNav(),
                    ],
                  ),
                ),
    );
  }
}

// 효과:
// - 휴대폰/태블릿/웹 모두 동일 코드로 지원
// - 유지보수 시간 70% 감소 ✅
```

### Example 4: Native Module Integration

**Flutter + Native Method Channel (Kotlin for Android)**:
```dart
// @CODE:MOBILE-004 | Flutter ↔ Native 통신
// lib/services/platform_service.dart
class PlatformService {
  static const platform = MethodChannel('com.example.app/platform');

  static Future<String> getBatteryLevel() async {
    try {
      final int result = await platform.invokeMethod<int>('getBatteryLevel') ?? 0;
      return 'Battery level at $result%';
    } catch (e) {
      return 'Failed to get battery level: $e';
    }
  }

  static Future<void> shareFile(String filePath) async {
    try {
      await platform.invokeMethod('shareFile', {'path': filePath});
    } catch (e) {
      print('Failed to share: $e');
    }
  }
}

// 사용
onPressed: () async {
  final batteryLevel = await PlatformService.getBatteryLevel();
  print(batteryLevel);
}
```

**Native Implementation (Kotlin)**:
```kotlin
// android/app/src/main/kotlin/com/example/app/MainActivity.kt
import androidx.annotation.NonNull
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.embedding.engine.dart.DartExecutor
import io.flutter.plugin.common.MethodChannel

class MainActivity: FlutterActivity() {
    private val CHANNEL = "com.example.app/platform"

    override fun configureFlutterEngine(@NonNull flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)

        MethodChannel(flutterEngine.dartExecutor.binaryMessenger, CHANNEL)
            .setMethodCallHandler { call, result ->
                when (call.method) {
                    "getBatteryLevel" -> {
                        val batteryLevel = getBatteryLevel()
                        result.success(batteryLevel)
                    }
                    "shareFile" -> {
                        val filePath = call.argument<String>("path")
                        shareFile(filePath!!)
                        result.success(null)
                    }
                    else -> result.notImplemented()
                }
            }
    }

    private fun getBatteryLevel(): Int {
        // Native Android API 호출
        val batteryManager = getSystemService(Context.BATTERY_SERVICE) as BatteryManager
        return batteryManager.getIntProperty(BatteryManager.BATTERY_PROPERTY_CAPACITY)
    }

    private fun shareFile(filePath: String) {
        // Native Android 파일 공유
        val intent = Intent(Intent.ACTION_SEND).apply {
            type = "application/octet-stream"
            putExtra(Intent.EXTRA_STREAM, Uri.parse("file://$filePath"))
        }
        startActivity(Intent.createChooser(intent, "Share file"))
    }
}

// 효과: Flutter 앱이 네이티브 기능 (배터리, 파일 공유 등) 직접 사용 가능 ✅
```

## Keywords

"모바일 앱 개발", "Flutter", "React Native", "Dart", "TypeScript", "상태 관리", "BLoC", "Redux", "위젯 테스트", "반응형 레이아웃", "네이티브 모듈 통합", "크로스 플랫폼", "성능 최적화"

## Reference

- Flutter architecture: `.moai/memory/development-guide.md#Flutter-아키텍처`
- React Native patterns: CLAUDE.md#React-Native-패턴
- Mobile app testing: `.moai/memory/development-guide.md#모바일-테스팅`

## Works well with

- moai-foundation-trust (모바일 테스트)
- moai-lang-dart (Flutter 개발)
- moai-lang-typescript (React Native 개발)
- moai-domain-backend (모바일 API 연동)
- moai-essentials-perf (모바일 성능 최적화)
