# moai-domain-mobile-app - CLI Reference

**Version**: 4.0.0 (Stable 2025-11)  
**Last Updated**: 2025-11-12

---

## React Native CLI Commands

### Project Setup

```bash
# Create new React Native project (Expo)
npx create-expo-app MyApp
cd MyApp

# Or with classic React Native
npx react-native@latest init MyApp --template

# Install Expo SDK
npm install expo@^52 react-native@^0.76

# Verify installation
npm list react-native expo
```

### Development

```bash
# Start development server (Expo)
npm start
npm start -- --clear  # Clear cache

# Run on iOS (Expo)
npm run ios

# Run on Android (Expo)
npm run android

# Run with React Native CLI (native)
npx react-native run-ios
npx react-native run-android
```

### Building

```bash
# Build for preview (Expo)
eas build --platform ios --profile preview
eas build --platform android --profile preview

# Build for production (Expo)
eas build --platform ios --profile production
eas build --platform android --profile production

# Build APK (native Android)
cd android && ./gradlew assembleRelease
cd android && ./gradlew bundleRelease

# Build IPA (native iOS)
xcodebuild -workspace ios/MyApp.xcworkspace \
  -scheme MyApp \
  -configuration Release \
  -destination generic/platform=iOS \
  -archivePath ./build/MyApp.xcarchive archive
```

### Testing

```bash
# Run Jest tests
npm test

# Run tests with coverage
npm test -- --coverage

# Run Detox E2E tests
detox test ios.sim.debug

# Run Detox with build
detox build-ios-framework --release
detox test ios.sim.release

# Run on Android
detox build-android-framework
detox test android.emu.release
```

### Debugging

```bash
# Clear cache
npm start -- --clear
rm -rf node_modules && npm install

# Debug with React Native DevTools
npm install -g @react-native-community/cli-debugger-ui

# Flipper (advanced debugging)
# Download from https://fbflipper.com/

# Metro bundler logs
npm start -- --verbose
```

### Code Quality

```bash
# Lint with ESLint
npm run lint

# Format code
npm run format

# Type check
npm run type-check

# Full audit
npm audit
npm audit fix
```

---

## Flutter CLI Commands

### Project Setup

```bash
# Create new Flutter project
flutter create my_app
cd my_app

# Get Dart packages
flutter pub get

# Upgrade Flutter
flutter upgrade

# Check Flutter setup
flutter doctor
flutter doctor -v
```

### Development

```bash
# Run app
flutter run

# Run in debug mode
flutter run --debug

# Run in release mode
flutter run --release

# Run on specific device
flutter devices
flutter run -d <device_id>

# Run with hot reload
flutter run
# Then press 'r' for hot reload
# Press 'R' for hot restart
```

### Building

```bash
# Build iOS
flutter build ios --release
flutter build ios --profile

# Build iOS app bundle
flutter build ipa --release

# Build Android
flutter build apk --release
flutter build aab --release

# Build web
flutter build web --release

# Build desktop
flutter build windows --release
flutter build macos --release
flutter build linux --release
```

### Testing

```bash
# Run all tests
flutter test

# Run specific test
flutter test test/unit_test.dart

# Generate coverage
flutter test --coverage
lcov --summary coverage/lcov.info

# E2E testing
flutter drive --target=test_driver/app.dart
```

### Code Quality

```bash
# Analyze code
flutter analyze

# Format code
dart format .

# Fix issues
dart fix --apply

# Generate documentation
dart doc

# Audit dependencies
flutter pub outdated
```

---

## Capacitor CLI Commands

### Project Setup

```bash
# Create new Capacitor project
npm init @capacitor/app

# Add Capacitor to existing project
npm install @capacitor/core @capacitor/cli

# Initialize Capacitor
npx cap init

# Add iOS platform
npx cap add ios

# Add Android platform
npx cap add android
```

### Development

```bash
# Sync changes
npx cap sync

# Sync iOS only
npx cap sync ios

# Sync Android only
npx cap sync android

# Open iOS project in Xcode
npx cap open ios

# Open Android project in Android Studio
npx cap open android

# Run app
npx cap run ios
npx cap run android
```

### Building & Deployment

```bash
# Build for iOS
npx cap build ios --scheme=MyApp --configuration=Release

# Build for Android
npx cap build android --buildType=release

# Copy web assets
npx cap copy

# Create plugin
npx cap plugin:generate
```

---

## EAS Build CLI

### Configuration

```bash
# Initialize EAS
eas build:configure

# View configuration
eas build:list

# Show EAS status
eas status
```

### Building

```bash
# Build for iOS (production)
eas build --platform ios --profile production

# Build for Android (production)
eas build --platform android --profile production

# Build for both platforms
eas build --platform all --profile production

# Build with specific workflow
eas build --platform ios --profile preview --local

# Monitor build
eas build:view
```

### Submitting

```bash
# Submit to App Store
eas submit --platform ios --latest

# Submit to Google Play
eas submit --platform android --latest

# Submit to TestFlight
eas submit --platform ios --type testflight

# Submit to Play Console internal testing
eas submit --platform android --type internal
```

---

## fastlane CLI

### Setup

```bash
# Install fastlane
sudo gem install fastlane -NV
fastlane -v

# Setup fastlane for iOS
cd ios
fastlane init

# Setup fastlane for Android
cd android
fastlane init
```

### Building & Deployment

```bash
# Build for iOS
fastlane ios beta

# Build and release iOS
fastlane ios release

# Build for Android
fastlane android beta

# Build and release Android
fastlane android release

# List available lanes
fastlane ios list
fastlane android list
```

### Certificate Management

```bash
# Sync certificates
fastlane match development
fastlane match appstore

# Generate new certificates
fastlane match nuke development
fastlane match nuke appstore

# Update provisioning profiles
fastlane match refresh
```

---

## Package Management

### npm/Node

```bash
# Install dependencies
npm install

# Install specific package
npm install package-name

# Install dev dependency
npm install --save-dev package-name

# Update packages
npm update
npm upgrade

# Audit dependencies
npm audit
npm audit fix

# Check outdated
npm outdated

# Clean cache
npm cache clean --force
```

### Dart/Flutter

```bash
# Get packages
flutter pub get

# Add package
flutter pub add package_name

# Add dev dependency
flutter pub add --dev package_name:version

# Update packages
flutter pub upgrade

# Get outdated packages
flutter pub outdated

# Analyze dependencies
flutter pub global activate pana
pana
```

---

## Version Information

**Tool Versions (Stable 2025-11):**

```bash
# Check Node version
node --version
# Recommended: v20.x or higher

# Check npm version
npm --version
# Recommended: v10.x or higher

# Check React Native version
npx react-native --version
# Stable: 0.76.x

# Check Expo version
npx expo --version
# Stable: 52.x

# Check Flutter version
flutter --version
# Stable: 3.24+

# Check Dart version
dart --version
# Stable: 3.5+

# Check Capacitor version
npm list @capacitor/core
# Stable: 6.x
```

---

**Last Updated**: 2025-11-12  
**Edition**: Stable 2025-11  
**Tools Covered**: React Native, Flutter, Capacitor, EAS, fastlane
