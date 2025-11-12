# Mobile Languages Validation Patterns

> Referenced from: `moai-validation-quality/SKILL.md`

This document covers validation patterns for mobile languages: **Swift, Kotlin, Dart**.

---

## Swift

### Technology Stack
- **Language**: Swift 6.0+ (Swift evolution)
- **IDE**: Xcode 16+, VS Code
- **Package Manager**: Swift Package Manager (SPM)
- **Test Framework**: XCTest (built-in)
- **Build System**: xcodebuild, swift build

### Validation Tools

#### Build
```bash
# Swift build
swift build

# Xcode build
xcodebuild -scheme MyApp -configuration Release

# With error checking
swift build 2>&1 | grep -i "error" && exit 1 || exit 0
```

#### Testing
```bash
# Swift test
swift test

# Verbose output
swift test --verbose

# With filter
swift test --filter SwiftUITests
```

#### Linting (SwiftLint)
```bash
# Install SwiftLint
brew install swiftlint

# Run linting
swiftlint lint --strict

# Auto-correct
swiftlint autocorrect
```

#### Formatting (SwiftFormat)
```bash
# Install SwiftFormat
brew install swiftformat

# Check format
swiftformat --lint src/

# Auto-format
swiftformat src/
```

### Example Validation Script
```bash
#!/bin/bash
set -e

echo "🍎 Swift Validation Starting..."

echo "▶ Build"
swift build

echo "▶ Testing"
swift test

echo "▶ Linting"
swiftlint lint --strict

echo "▶ Formatting"
swiftformat --lint . || swiftformat .

echo "✅ All validations passed!"
```

### Fallback Chain
```yaml
swift:
  build: [swift-build, xcodebuild]
  test: [swift-test, xctest]
  linter: [swiftlint]
  formatter: [swiftformat]
```

### Context7 Integration
```python
context7_patterns = await context7.get_library_docs(
    context7_library_id="/quality-validation/swift-2025",
    topic="Swift testing swiftlint swiftformat package manager patterns"
)
```

---

## Kotlin

### Technology Stack
- **Language**: Kotlin 1.9.x
- **Build System**: Gradle, Maven
- **Test Framework**: JUnit 5, Kotlin Test
- **Package Manager**: Maven Central, Gradle
- **Linting**: Detekt, ktlint

### Validation Tools

#### Build + Compile
```bash
# Gradle
./gradlew build

# Maven
mvn compile
```

#### Testing
```bash
# Gradle
./gradlew test

# Maven
mvn test

# Specific test
./gradlew test --tests com.example.MyTest
```

#### Linting (Detekt)
```bash
# Install Detekt (usually via Gradle)
./gradlew detekt

# Run with config
./gradlew detekt --configuration-resources=/absolute/path/to/config.yml
```

#### Formatting (Ktlint)
```bash
# Install ktlint
brew install ktlint

# Check format
./gradlew ktlintCheck

# Auto-format
./gradlew ktlintFormat

# Or direct:
ktlint src/
```

### Example Validation Script
```bash
#!/bin/bash
set -e

echo "🟣 Kotlin Validation Starting..."

echo "▶ Build"
./gradlew build

echo "▶ Testing"
./gradlew test

echo "▶ Linting"
./gradlew detekt

echo "▶ Formatting"
./gradlew ktlintCheck || ./gradlew ktlintFormat

echo "✅ All validations passed!"
```

### Fallback Chain
```yaml
kotlin:
  build: [gradle, maven]
  test: [junit5, kotlin-test]
  linter: [detekt, ktlint]
  formatter: [ktlint, kotlinter]
```

### Context7 Integration
```python
context7_patterns = await context7.get_library_docs(
    context7_library_id="/quality-validation/kotlin-2025",
    topic="Kotlin gradle testing detekt ktlint android patterns"
)
```

---

## Dart

### Technology Stack
- **Language**: Dart 3.5.x
- **Package Manager**: Pub (pub.dev)
- **Test Framework**: test, mockito
- **Build System**: Dart CLI
- **IDE Support**: VS Code, Android Studio

### Validation Tools

#### Analysis (dart analyze)
```bash
# Analyze code
dart analyze

# With specific checks
dart analyze --fatal-infos

# Warning as error
dart analyze --fatal-warnings
```

#### Testing
```bash
# Run tests
dart test

# With coverage
dart test --coverage=coverage/

# Specific test file
dart test test/unit/calculator_test.dart
```

#### Formatting (dart format)
```bash
# Check format
dart format --set-exit-if-changed .

# Auto-format
dart format --fix .

# Specific directory
dart format lib/
```

#### Code Metrics (dart_code_metrics)
```bash
# Install globally
dart pub global activate dart_code_metrics

# Analyze metrics
metrics lib/ --report-cli
```

### Example Validation Script
```bash
#!/bin/bash
set -e

echo "🎯 Dart Validation Starting..."

echo "▶ Dependency check"
dart pub get

echo "▶ Analysis"
dart analyze --fatal-warnings

echo "▶ Testing"
dart test

echo "▶ Formatting"
dart format --set-exit-if-changed . || exit 1

echo "▶ Code metrics"
dart pub global activate dart_code_metrics
metrics lib/ --report-cli

echo "✅ All validations passed!"
```

### Fallback Chain
```yaml
dart:
  analyzer: [dart-analyze]
  formatter: [dart-format]
  test: [dart-test]
  coverage: [dart-coverage]
  metrics: [dart-code-metrics]
```

### Context7 Integration
```python
context7_patterns = await context7.get_library_docs(
    context7_library_id="/quality-validation/dart-2025",
    topic="Dart flutter testing formatting pub package manager patterns"
)
```

---

## Installation Commands

### Swift Tools
```bash
# macOS
brew install swiftlint swiftformat

# Check Swift version
swift --version
```

### Kotlin Tools
```bash
# Gradle (usually included in projects)
brew install gradle

# ktlint
brew install ktlint

# Detekt (usually via Gradle plugin)
```

### Dart Tools
```bash
# Install Dart SDK
brew install dart

# Install global packages
dart pub global activate dart_code_metrics

# Check version
dart --version
```

---

**Last Updated**: 2025-11-12
**Related**: [SKILL.md](SKILL.md), [LANGUAGES-INTERPRETED.md](LANGUAGES-INTERPRETED.md), [LANGUAGES-MARKUP.md](LANGUAGES-MARKUP.md)
