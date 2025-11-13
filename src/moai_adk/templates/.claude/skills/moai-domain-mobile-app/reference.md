# Reference: Advanced Mobile Development Patterns

This document contains detailed reference material for advanced mobile development patterns extracted from the main skill documentation.

## Table of Contents

- [Flutter Advanced Patterns](#flutter-advanced-patterns)
- [React Native Advanced Topics](#react-native-advanced-topics)
- [Testing Frameworks Deep Dive](#testing-frameworks-deep-dive)
- [CI/CD Pipeline Configuration](#cicd-pipeline-configuration)
- [Performance Optimization Techniques](#performance-optimization-techniques)
- [Deployment Strategies](#deployment-strategies)

---

## Flutter Advanced Patterns

### Provider Pattern Implementation

```dart
// Complete Provider implementation with services
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

// Service layer
class ApiService {
  Future<List<Product>> fetchProducts() async {
    // API implementation
    await Future.delayed(Duration(seconds: 1));
    return [
      Product(id: 1, name: "Product 1", price: 99.99),
      Product(id: 2, name: "Product 2", price: 149.99),
    ];
  }
}

// ViewModel/Controller
class ProductViewModel extends ChangeNotifier {
  final ApiService _apiService;
  List<Product> _products = [];
  bool _isLoading = false;
  String? _error;

  ProductViewModel(this._apiService);

  List<Product> get products => _products;
  bool get isLoading => _isLoading;
  String? get error => _error;

  Future<void> loadProducts() async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      _products = await _apiService.fetchProducts();
    } catch (e) {
      _error = e.toString();
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }
}
```

### State Management Comparison

| Pattern | Pros | Cons | Best For |
|---------|------|------|-----------|
| Provider | Simple, good separation | Manual state updates | Small to medium apps |
| GetX | High performance, minimal boilerplate | Less conventional | Performance-critical apps |
| Riverpod | Type-safe, testable | More complex | Large, complex apps |
| Bloc | Strict state management | Verbose | Enterprise apps |

---

## React Native Advanced Topics

### Native Module Integration

```typescript
// Advanced native module with bi-directional communication
import { NativeModules, NativeEventEmitter } from 'react-native';

interface NativeCalendarModule {
  createEvent(options: CalendarEvent): Promise<string>;
  updateEvent(eventId: string, options: Partial<CalendarEvent>): Promise<void>;
  deleteEvent(eventId: string): Promise<void>;
  subscribeToCalendarChanges(callback: (events: CalendarEvent[]) => void): void;
}

interface CalendarEvent {
  id?: string;
  title: string;
  startDate: Date;
  endDate: Date;
  location?: string;
  attendees?: string[];
  reminders?: CalendarReminder[];
}

interface CalendarReminder {
  minutesBefore: number;
  type: 'notification' | 'email';
}

const NativeCalendar = NativeModules.NativeCalendar as NativeCalendarModule;
const calendarEventEmitter = new NativeEventEmitter(NativeCalendar);

// Usage in React component
const CalendarManager: React.FC = () => {
  const [events, setEvents] = useState<CalendarEvent[]>([]);

  useEffect(() => {
    // Subscribe to native calendar changes
    const subscription = calendarEventEmitter.addListener(
      'calendarChanged',
      (newEvents: CalendarEvent[]) => {
        setEvents(newEvents);
      }
    );

    // Request native updates
    NativeCalendar.subscribeToCalendarChanges((updatedEvents) => {
      calendarEventEmitter.emit('calendarChanged', updatedEvents);
    });

    return () => subscription.remove();
  }, []);

  const createEvent = async (eventData: Omit<CalendarEvent, 'id'>) => {
    try {
      const eventId = await NativeCalendar.createEvent(eventData);
      console.log('Event created:', eventId);
    } catch (error) {
      console.error('Failed to create event:', error);
    }
  };

  return (
    <View>
      <Button title="Create Event" onPress={() => createEvent({
        title: 'Meeting',
        startDate: new Date(),
        endDate: new Date(Date.now() + 3600000),
        location: 'Conference Room A',
      })} />
    </View>
  );
};
```

### Advanced Navigation Patterns

```typescript
// Complex navigation with deep linking and authentication
import { 
  NavigationContainer, 
  useNavigationContainerRef,
  createStackNavigator,
  createBottomTabNavigator,
  createMaterialTopTabNavigator
} from '@react-navigation/native';
import { Linking, useLinking } from '@expo-router';

const linking = {
  prefixes: ['myapp://', 'https://myapp.com'],
  config: {
    screens: {
      Main: 'main',
      Profile: 'profile/:userId',
      Settings: 'settings',
      Post: 'post/:postId',
    },
  },
};

// Auth navigator wrapper
const AppNavigator = () => {
  const { user } = useAuth();
  const navigationRef = useNavigationContainerRef();

  const linking = useLinking({
    prefixes: ['myapp://'],
    config: {
      screens: {
        Main: 'main',
        Profile: 'profile/:userId',
      },
    },
  });

  return (
    <NavigationContainer linking={linking} ref={navigationRef}>
      {user ? <MainStack /> : <AuthStack />}
    </NavigationContainer>
  );
};

// Main stack with nested navigators
const MainStack = () => {
  return (
    <Tab.Navigator>
      <Tab.Screen name="Home" component={HomeStack} />
      <Tab.Screen name="Search" component={SearchStack} />
      <Tab.Screen name="Profile" component={ProfileStack} />
    </Tab.Navigator>
  );
};

const HomeStack = () => {
  return (
    <Stack.Navigator>
      <Stack.Screen name="Home" component={HomeScreen} />
      <Stack.Screen 
        name="Post" 
        component={PostScreen}
        options={({ route }) => ({
          title: route.params.title,
          headerRight: () => <ShareButton postId={route.params.postId} />,
        })}
      />
    </Stack.Navigator>
  );
};
```

---

## Testing Frameworks Deep Dive

### Detox Advanced Configuration

```typescript
// detox.config.js - Advanced detox configuration
module.exports = {
  testRunner: 'jest',
  runnerConfig: {
    setupTestFrameworkScriptFile: './jest.setup.js',
    testEnvironment: 'node',
  },
  configurations: {
    'ios.sim.debug': {
      binaryPath: 'ios/build/Build/Products/Debug-iphonesimulator/MyApp.app',
      build: 'xcodebuild -workspace ios/MyApp.xcworkspace -scheme MyApp -configuration Debug -sdk iphonesimulator -destination \'platform=iOS Simulator,name=iPhone 14\'',
      type: 'ios.simulator',
      device: {
        type: 'iPhone 14',
      },
    },
    'ios.sim.release': {
      binaryPath: 'ios/build/Build/Products/Release-iphonesimulator/MyApp.app',
      build: 'xcodebuild -workspace ios/MyApp.xcworkspace -scheme MyApp -configuration Release -sdk iphonesimulator -destination \'platform=iOS Simulator,name=iPhone 14\'',
      type: 'ios.simulator',
      device: {
        type: 'iPhone 14',
      },
    },
    'android.emu.debug': {
      binaryPath: 'android/app/build/outputs/apk/debug/app-debug.apk',
      build: 'cd android && ./gradlew assembleDebug assembleAndroidTest -DtestBuildType=debug && cd ..',
      type: 'android.emulator',
      device: {
        avdName: 'Pixel_3a_API_30_x86',
      },
    },
  },
  testMatch: [
    '<rootDir>/e2e/**/*.test.{js,jsx,ts,tsx}',
  ],
};
```

### Custom Detox Matchers

```typescript
// detox-matchers.js - Custom matchers for complex assertions
import { by, element, expect } from 'detox';

// Custom matcher for accessibility
expect.extend({
  toBeAccessible(received) {
    const pass = received.props.accessible === true && 
                   received.props.accessibilityLabel !== undefined;
    
    if (pass) {
      return {
        message: () => `expected element not to be accessible`,
        pass: true,
      };
    } else {
      return {
        message: () => `expected element to be accessible with accessibilityLabel`,
        pass: false,
      };
    }
  },
});

// Custom matcher for component state
expect.extend({
  toHaveState(received, expectedState) {
    const pass = received.props.testState === expectedState;
    
    if (pass) {
      return {
        message: () => `expected element not to have state "${expectedState}"`,
        pass: true,
      };
    } else {
      return {
        message: () => `expected element to have state "${expectedState}" but got "${received.props.testState}"`,
        pass: false,
      };
    }
  },
});

// Usage in tests
it('should have accessible components', async () => {
  await expect(element(by.id('submit-button'))).toBeAccessible();
  await expect(element(by.id('loading-spinner'))).toHaveState('loading');
});
```

---

## CI/CD Pipeline Configuration

### GitHub Actions for Mobile Apps

```yaml
# .github/workflows/mobile-ci.yml
name: Mobile CI/CD

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]
  release:
    types: [published]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'
          cache: 'npm'
      
      - name: Install dependencies
        run: npm ci
      
      - name: Run linting
        run: npm run lint
      
      - name: Run type checking
        run: npm run type-check
      
      - name: Run unit tests
        run: npm run test:unit
      
      - name: Run E2E tests
        run: |
          npm run build:e2e
          npm run test:e2e:ios
          npm run test:e2e:android

  build-ios:
    runs-on: macos-latest
    needs: test
    if: github.event_name == 'release'
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'
          cache: 'npm'
      
      - name: Install dependencies
        run: npm ci
      
      - name: Setup Xcode
        uses: maxim-lobanov/setup-xcode@v1
        with:
          xcode-version: '14.3'
      
      - name: Install CocoaPods
        run: |
          cd ios
          pod install
      
      - name: Build iOS app
        run: |
          cd ios
          xcodebuild -workspace MyApp.xcworkspace \
            -scheme MyApp \
            -configuration Release \
            -destination 'generic/platform=iOS' \
            -archivePath MyApp.xcarchive \
            archive
      
      - name: Export IPA
        run: |
          cd ios
          xcodebuild -exportArchive \
            -archivePath MyApp.xcarchive \
            -exportOptionsPlist ExportOptions.plist \
            -exportPath ./build
      
      - name: Upload to App Store
        uses: apple-actions/upload-testflight-build@v1
        with:
          app-id: com.myapp.mobile
          issuer-id: ${{ secrets.APPSTORE_ISSUER_ID }}
          api-key-id: ${{ secrets.APPSTORE_API_KEY_ID }}
          api-key: ${{ secrets.APPSTORE_API_KEY }}

  build-android:
    runs-on: ubuntu-latest
    needs: test
    if: github.event_name == 'release'
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'
          cache: 'npm'
      
      - name: Install dependencies
        run: npm ci
      
      - name: Setup Java JDK
        uses: actions/setup-java@v3
        with:
          distribution: 'zulu'
          java-version: '17'
      
      - name: Setup Android SDK
        uses: android-actions/setup-android@v2
      
      - name: Grant execute permission for gradlew
        run: cd android && chmod +x ./gradlew
      
      - name: Build Android app
        run: |
          cd android
          ./gradlew assembleRelease
      
      - name: Upload to Play Store
        uses: r0adkkl/upload-google-play@v1
        with:
          serviceAccountJsonPlainText: ${{ secrets.GOOGLE_PLAY_SERVICE_ACCOUNT_JSON }}
          packageName: com.myapp.mobile
          releaseFiles: android/app/build/outputs/apk/release/app-release.apk
          track: production
          status: completed
```

---

## Performance Optimization Techniques

### React Native Performance Monitoring

```typescript
// Performance monitoring setup
import { Performance } from 'react-native';
import { InteractionManager } from 'react-native';

// Performance monitoring HOC
const withPerformanceMonitoring = <P extends object>(
  WrappedComponent: React.ComponentType<P>
) => {
  return class PerformanceMonitoredComponent extends React.Component<P> {
    private renderStartTime: number = 0;
    private interactionHandle: number | null = null;

    componentDidMount() {
      this.renderStartTime = Performance.now();
      this.interactionHandle = InteractionManager.createInteractionHandle();
    }

    componentWillUnmount() {
      if (this.interactionHandle !== null) {
        InteractionManager.clearInteractionHandle(this.interactionHandle);
      }
    }

    render() {
      const renderTime = Performance.now() - this.renderStartTime;
      
      // Log performance metrics in development
      if (__DEV__) {
        console.log(`[Performance] ${WrappedComponent.name} render time: ${renderTime.toFixed(2)}ms`);
      }

      return <WrappedComponent {...this.props} />;
    }
  };
};

// Usage
const PerformanceOptimizedScreen = withPerformanceMonitoring(MyScreen);

// Memoization strategies
const OptimizedListItem = React.memo<{ item: Item }>({ item }) => {
  return (
    <View style={styles.itemContainer}>
      <Text style={styles.itemTitle}>{item.title}</Text>
      <Text style={styles.itemDescription}>{item.description}</Text>
    </View>
  );
}, (prevProps, nextProps) => {
  return prevProps.item.id === nextProps.item.id && 
         prevProps.item.updated === nextProps.item.updated;
});

// FlatList optimization
const OptimizedFlatList: React.FC<{ data: Item[] }> = ({ data }) => {
  const renderItem = useCallback(({ item }: { item: Item }) => (
    <OptimizedListItem item={item} />
  ), []);

  const keyExtractor = useCallback((item: Item) => item.id, []);

  const getItemLayout = useCallback(
    (_: any, index: number) => ({
      length: ITEM_HEIGHT,
      offset: ITEM_HEIGHT * index,
      index,
    }),
    []
  );

  return (
    <FlatList
      data={data}
      renderItem={renderItem}
      keyExtractor={keyExtractor}
      getItemLayout={getItemLayout}
      removeClippedSubviews={true}
      maxToRenderPerBatch={10}
      updateCellsBatchingPeriod={50}
      initialNumToRender={15}
      windowSize={10}
      showsVerticalScrollIndicator={false}
      contentContainerStyle={styles.listContainer}
    />
  );
};
```

### Flutter Performance Optimization

```dart
// Flutter performance optimization techniques
import 'package:flutter/foundation.dart';

class OptimizedListView extends StatefulWidget {
  final List<Item> items;
  final Widget Function(BuildContext, Item) itemBuilder;

  const OptimizedListView({
    Key? key,
    required this.items,
    required this.itemBuilder,
  }) : super(key: key);

  @override
  _OptimizedListViewState createState() => _OptimizedListViewState();
}

class _OptimizedListViewState extends State<OptimizedListView> {
  final ScrollController _scrollController = ScrollController();
  final List<Item> _visibleItems = [];
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _loadMoreItems();
    _scrollController.addListener(_onScroll);
  }

  void _onScroll() {
    if (_scrollController.position.pixels >= 
        _scrollController.position.maxScrollExtent - 200) {
      _loadMoreItems();
    }
  }

  Future<void> _loadMoreItems() async {
    if (_isLoading) return;
    
    setState(() => _isLoading = true);
    
    // Simulate async data loading
    await Future.delayed(Duration(milliseconds: 500));
    
    setState(() {
      _visibleItems.addAll(widget.items);
      _isLoading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    return NotificationListener<ScrollNotification>(
      onNotification: (notification) {
        if (notification is ScrollEndNotification &&
            notification.metrics.extentAfter == 0) {
          _loadMoreItems();
          return true;
        }
        return false;
      },
      child: ListView.builder(
        controller: _scrollController,
        itemCount: _visibleItems.length + (_isLoading ? 1 : 0),
        itemBuilder: (context, index) {
          if (index == _visibleItems.length && _isLoading) {
            return const Center(
              child: CircularProgressIndicator(),
            );
          }
          
          return widget.itemBuilder(context, _visibleItems[index]);
        },
        cacheExtent: 500.0,
      ),
    );
  }
}

// Image optimization
class OptimizedImage extends StatelessWidget {
  final String imageUrl;
  final double? width;
  final double? height;
  final BoxFit fit;

  const OptimizedImage({
    Key? key,
    required this.imageUrl,
    this.width,
    this.height,
    this.fit = BoxFit.cover,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Image.network(
      imageUrl,
      width: width,
      height: height,
      fit: fit,
      loadingBuilder: (context, child, loadingProgress) {
        if (loadingProgress == null) return child;
        return Center(
          child: CircularProgressIndicator(
            value: loadingProgress.expectedTotalBytes != null
                ? loadingProgress.cumulativeBytesLoaded /
                    loadingProgress.expectedTotalBytes!
                : null,
          ),
        );
      },
      errorBuilder: (context, error, stackTrace) {
        return Container(
          width: width,
          height: height,
          color: Colors.grey[300],
          child: const Icon(Icons.error),
        );
      },
      cacheWidth: width?.toInt(),
      cacheHeight: height?.toInt(),
    );
  }
}
```

---

## Deployment Strategies

### Fastlane Advanced Configuration

```ruby
# Fastfile - Advanced mobile deployment
default_platform(:ios)

platform :ios do
  desc "Run all tests"
  lane :test do
    run_tests(
      scheme: "MyApp",
      device: "iPhone 14"
    )
  end

  desc "Run tests on multiple simulators"
  lane :multi_device_test do
    devices = ["iPhone 14", "iPhone 14 Plus", "iPhone 14 Pro", "iPad Air"]
    
    devices.each do |device|
      run_tests(
        scheme: "MyApp",
        device: device
      )
    end
  end

  private_lane :run_tests(options)
    scan(
      scheme: options[:scheme],
      device: options[:device],
      clean: true,
      output_directory: "./test_output/#{options[:device]}",
      output_types: ["html", "junit"],
      skip_build: true
    )
  end

  desc "Build and upload to TestFlight"
  lane :beta do
    # Increment build number
    increment_build_number(
      xcodeproj: "ios/MyApp.xcodeproj"
    )

    # Update version number if needed
    update_app_identifier(
      xcodeproj: "ios/MyApp.xcodeproj",
      app_identifier: ENV["APP_IDENTIFIER"]
    )

    # Build the app
    build_app(
      workspace: "ios/MyApp.xcworkspace",
      scheme: "MyApp",
      configuration: "Release",
      export_method: "app-store",
      output_directory: "./build",
      output_name: "MyApp.ipa",
      include_symbols: true,
      include_bitcode: false
    )

    # Upload dSYMs to Sentry
    upload_symbols_to_sentry(
      auth_token: ENV["SENTRY_AUTH_TOKEN"],
      org_slug: ENV["SENTRY_ORG_SLUG"],
      project_slug: ENV["SENTRY_PROJECT_SLUG"]
    ) if ENV["SENTRY_AUTH_TOKEN"]

    # Upload to TestFlight
    upload_to_testflight(
      skip_submission: true,
      skip_waiting_for_build_processing: true,
      notify_external_testers: true
    )

    # Post webhook
    slack(
      message: "Successfully deployed new beta version to TestFlight! 🚀",
      channel: "#mobile-deployments",
      payload: {
        "Build Number": get_build_number,
        "Version": get_version_number
      }
    )
  end

  desc "Deploy to App Store"
  lane :release do
    # Ensure we're on the main branch
    ensure_git_branch(branch: "main")

    # Validate that all tests pass
    test

    # Build and upload
    beta

    # Submit to App Store review
    deliver(
      skip_screenshots: false,
      skip_metadata: false,
      force: true,
      submit_for_review: true
    )
  end

  error do |lane, exception|
    # Send error notifications
    slack(
      message: "❌ Lane #{lane} failed with error: #{exception.message}",
      channel: "#mobile-deployments"
    )
    
    # Upload logs to a storage service
    upload_logs_to_storage(exception)
    
    # Re-raise the exception to fail the lane
    raise exception
  end
end

platform :android do
  desc "Build and upload to Play Store internal testing"
  lane :internal do
    # Set up keystore
    setup_keystore

    # Gradle build
    gradle(
      task: "assembleRelease",
      build_type: "Release",
      print_command: false
    )

    # Upload to Play Store internal track
    upload_to_play_store(
      track: "internal",
      aab: "../build/app/outputs/bundle/release/app-release.aab",
      json_key: ENV["GOOGLE_PLAY_JSON_KEY_DATA"],
      release_status: "draft"
    )
  end

  desc "Promote to production"
  lane :promote_to_production do
    upload_to_play_store(
      track: "production",
      track_promote_to: "production",
      json_key: ENV["GOOGLE_PLAY_JSON_KEY_DATA"],
      release_status: "completed",
      changes_not_sent_for_review: true
    )
  end

  private_lane :setup_keystore
    # Validate keystore exists
    unless File.exist?(ENV["KEYSTORE_PATH"])
      UI.error("Keystore file not found at #{ENV['KEYSTORE_PATH']}")
      raise "Keystore file not found"
    end
  end
end

# Helper methods
private_lane :upload_logs_to_storage(exception)
  # Create a zip file with logs and crash reports
  zip(
    path: "./build/logs",
    output_path: "./build/logs_#{Time.now.to_i}.zip"
  )
  
  # Upload to your preferred storage service
  # This is a placeholder - implement your actual upload logic
  puts "Would upload logs to storage service"
end
```

This reference document contains the advanced patterns and configurations that were extracted from the main skill documentation to keep the SKILL.md file focused and concise.
