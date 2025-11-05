# Language Template Variables

This document defines the parameterized variables used by `moai-lang-template` for each supported language.

## Variable Definitions

### Core Variables
```
{{LANGUAGE_NAME}} - Full display name of the language
{{LANGUAGE_ID}} - Technical identifier (lowercase, no spaces)
{{LANGUAGE_VERSION}} - Current stable version number
{{LANGUAGE_RELEASE_DATE}} - Date of current version release
{{LANGUAGE_EXTENSION}} - Primary file extension
{{PACKAGE_NAME}} - Package naming convention
{{CLASS_NAMING}} - Class/type naming convention
{{FUNCTION_NAMING}} - Function/method naming convention
```

### Build & Package Management
```
{{BUILD_TOOL}} - Primary build system (Maven, Gradle, npm, etc.)
{{PACKAGE_MANAGER}} - Package management tool (pip, npm, composer, etc.)
{{DEPENDENCY_FILE}} - File containing dependencies
{{BUILD_COMMAND}} - Command to build the project
{{TEST_COMMAND}} - Command to run tests
{{INSTALL_COMMAND}} - Command to install dependencies
```

### Testing Framework
```
{{TEST_FRAMEWORK}} - Primary testing framework name
{{TEST_FILE_PATTERN}} - Pattern for test files
{{TEST_CLASS_SUFFIX}} - Suffix for test classes
{{ASSERTION_LIBRARY}} - Assertion/assertion library
{{MOCKING_LIBRARY}} - Mocking/stubbing library
{{COVERAGE_TOOL}} - Code coverage tool
```

### Code Quality Tools
```
{{LINTER}} - Code linting tool
{{FORMATTER}} - Code formatting tool
{{STATIC_ANALYZER}} - Static analysis tool
{{TYPE_CHECKER}} - Type checking tool
{{SECURITY_SCANNER}} - Security vulnerability scanner
```

## Language-Specific Variable Values

### Java
```
{{LANGUAGE_NAME}} = "Java"
{{LANGUAGE_ID}} = "java"
{{LANGUAGE_VERSION}} = "21"
{{LANGUAGE_RELEASE_DATE}} = "2023-09"
{{LANGUAGE_EXTENSION}} = ".java"
{{PACKAGE_NAME}} = "com.example.project"
{{CLASS_NAMING}} = "PascalCase"
{{FUNCTION_NAMING}} = "camelCase"

{{BUILD_TOOL}} = "Maven"
{{PACKAGE_MANAGER}} = "Maven"
{{DEPENDENCY_FILE}} = "pom.xml"
{{BUILD_COMMAND}} = "mvn compile"
{{TEST_COMMAND}} = "mvn test"
{{INSTALL_COMMAND}} = "mvn install"

{{TEST_FRAMEWORK}} = "JUnit 5"
{{TEST_FILE_PATTERN}} = "*Test.java"
{{TEST_CLASS_SUFFIX}} = "Test"
{{ASSERTION_LIBRARY}} = "JUnit Assertions"
{{MOCKING_LIBRARY}} = "Mockito"
{{COVERAGE_TOOL}} = "JaCoCo"

{{LINTER}} = "CheckStyle"
{{FORMATTER}} = "Google Java Format"
{{STATIC_ANALYZER}} = "SonarQube"
{{TYPE_CHECKER}} = "Java Compiler"
{{SECURITY_SCANNER}} = "SpotBugs"
```

### Kotlin
```
{{LANGUAGE_NAME}} = "Kotlin"
{{LANGUAGE_ID}} = "kotlin"
{{LANGUAGE_VERSION}} = "1.9.10"
{{LANGUAGE_RELEASE_DATE}} = "2023-10"
{{LANGUAGE_EXTENSION}} = ".kt"
{{PACKAGE_NAME}} = "com.example.project"
{{CLASS_NAMING}} = "PascalCase"
{{FUNCTION_NAMING}} = "camelCase"

{{BUILD_TOOL}} = "Gradle"
{{PACKAGE_MANAGER}} = "Gradle"
{{DEPENDENCY_FILE}} = "build.gradle.kts"
{{BUILD_COMMAND}} = "gradle build"
{{TEST_COMMAND}} = "gradle test"
{{INSTALL_COMMAND}} = "gradle dependencies"

{{TEST_FRAMEWORK}} = "JUnit 5"
{{TEST_FILE_PATTERN}} = "*Test.kt"
{{TEST_CLASS_SUFFIX}} = "Test"
{{ASSERTION_LIBRARY}} = "Kotlin Test"
{{MOCKING_LIBRARY}} = "MockK"
{{COVERAGE_TOOL}} = "JaCoCo"

{{LINTER}} = "Ktlint"
{{FORMATTER}} = "Ktlint"
{{STATIC_ANALYZER}} = "Detekt"
{{TYPE_CHECKER}} = "Kotlin Compiler"
{{SECURITY_SCANNER}} = "Detekt"
```

### Swift
```
{{LANGUAGE_NAME}} = "Swift"
{{LANGUAGE_ID}} = "swift"
{{LANGUAGE_VERSION}} = "5.9"
{{LANGUAGE_RELEASE_DATE}} = "2023-09"
{{LANGUAGE_EXTENSION}} = ".swift"
{{PACKAGE_NAME}} = "com.example.project"
{{CLASS_NAMING}} = "PascalCase"
{{FUNCTION_NAMING}} = "camelCase"

{{BUILD_TOOL}} = "Swift Package Manager"
{{PACKAGE_MANAGER}} = "Swift Package Manager"
{{DEPENDENCY_FILE}} = "Package.swift"
{{BUILD_COMMAND}} = "swift build"
{{TEST_COMMAND}} = "swift test"
{{INSTALL_COMMAND}} = "swift package resolve"

{{TEST_FRAMEWORK}} = "XCTest"
{{TEST_FILE_PATTERN}} = "*Tests.swift"
{{TEST_CLASS_SUFFIX}} = "Tests"
{{ASSERTION_LIBRARY}} = "XCTAssert"
{{MOCKING_LIBRARY}} = "OCMock"
{{COVERAGE_TOOL}} = "Xcode Coverage"

{{LINTER}} = "SwiftLint"
{{FORMATTER}} = "SwiftFormat"
{{STATIC_ANALYZER}} = "SwiftLint"
{{TYPE_CHECKER}} = "Swift Compiler"
{{SECURITY_SCANNER}} = "SwiftLint"
```

### C#
```
{{LANGUAGE_NAME}} = "C#"
{{LANGUAGE_ID}} = "csharp"
{{LANGUAGE_VERSION}} = "12"
{{LANGUAGE_RELEASE_DATE}} = "2023-11"
{{LANGUAGE_EXTENSION}} = ".cs"
{{PACKAGE_NAME}} = "Example.Project"
{{CLASS_NAMING}} = "PascalCase"
{{FUNCTION_NAMING}} = "PascalCase"

{{BUILD_TOOL}} = "MSBuild"
{{PACKAGE_MANAGER}} = "NuGet"
{{DEPENDENCY_FILE}} = "*.csproj"
{{BUILD_COMMAND}} = "dotnet build"
{{TEST_COMMAND}} = "dotnet test"
{{INSTALL_COMMAND}} = "dotnet restore"

{{TEST_FRAMEWORK}} = "xUnit"
{{TEST_FILE_PATTERN}} = "*Tests.cs"
{{TEST_CLASS_SUFFIX}} = "Tests"
{{ASSERTION_LIBRARY}} = "xUnit Assertions"
{{MOCKING_LIBRARY}} = "Moq"
{{COVERAGE_TOOL}} = "dotnet coverlet"

{{LINTER}} = "Roslyn Analyzers"
{{FORMATTER}} = "dotnet format"
{{STATIC_ANALYZER}} = "SonarQube"
{{TYPE_CHECKER}} = "C# Compiler"
{{SECURITY_SCANNER}} = "Roslyn Analyzers"
```

### Dart
```
{{LANGUAGE_NAME}} = "Dart"
{{LANGUAGE_ID}} = "dart"
{{LANGUAGE_VERSION}} = "3.2"
{{LANGUAGE_RELEASE_DATE}} = "2023-10"
{{LANGUAGE_EXTENSION}} = ".dart"
{{PACKAGE_NAME}} = "com.example.project"
{{CLASS_NAMING}} = "PascalCase"
{{FUNCTION_NAMING}} = "camelCase"

{{BUILD_TOOL}} = "Dart Build System"
{{PACKAGE_MANAGER}} = "pub"
{{DEPENDENCY_FILE}} = "pubspec.yaml"
{{BUILD_COMMAND}} = "dart build"
{{TEST_COMMAND}} = "dart test"
{{INSTALL_COMMAND}} = "dart pub get"

{{TEST_FRAMEWORK}} = "Dart Test"
{{TEST_FILE_PATTERN}} = "*_test.dart"
{{TEST_CLASS_SUFFIX}} = "Test"
{{ASSERTION_LIBRARY}} = "expect"
{{MOCKING_LIBRARY}} = "mockito"
{{COVERAGE_TOOL}} = "dart test coverage"

{{LINTER}} = "Dart Analyzer"
{{FORMATTER}} = "dart format"
{{STATIC_ANALYZER}} = "Dart Analyzer"
{{TYPE_CHECKER}} = "Dart Analyzer"
{{SECURITY_SCANNER}} = "Dart Analyzer"
```

### C++
```
{{LANGUAGE_NAME}} = "C++"
{{LANGUAGE_ID}} = "cpp"
{{LANGUAGE_VERSION}} = "20"
{{LANGUAGE_RELEASE_DATE}} = "2020-12"
{{LANGUAGE_EXTENSION}} = ".cpp"
{{PACKAGE_NAME}} = "example_project"
{{CLASS_NAMING}} = "PascalCase"
{{FUNCTION_NAMING}} = "camelCase"

{{BUILD_TOOL}} = "CMake"
{{PACKAGE_MANAGER}} = "Conan"
{{DEPENDENCY_FILE}} = "CMakeLists.txt"
{{BUILD_COMMAND}} = "cmake --build"
{{TEST_COMMAND}} = "ctest"
{{INSTALL_COMMAND}} = "conan install"

{{TEST_FRAMEWORK}} = "Google Test"
{{TEST_FILE_PATTERN}} = "*_test.cpp"
{{TEST_CLASS_SUFFIX}} = "Test"
{{ASSERTION_LIBRARY}} = "gtest"
{{MOCKING_LIBRARY}} = "gmock"
{{COVERAGE_TOOL}} = "gcov"

{{LINTER}} = "clang-tidy"
{{FORMATTER}} = "clang-format"
{{STATIC_ANALYZER}} = "clang-static-analyzer"
{{TYPE_CHECKER}} = "clang-tidy"
{{SECURITY_SCANNER}} = "clang-tidy"
```

### C
```
{{LANGUAGE_NAME}} = "C"
{{LANGUAGE_ID}} = "c"
{{LANGUAGE_VERSION}} = "17"
{{LANGUAGE_RELEASE_DATE}} = "2018-06"
{{LANGUAGE_EXTENSION}} = ".c"
{{PACKAGE_NAME}} = "example_project"
{{CLASS_NAMING}} = "snake_case"
{{FUNCTION_NAMING}} = "snake_case"

{{BUILD_TOOL}} = "Make"
{{PACKAGE_MANAGER}} = "pkg-config"
{{DEPENDENCY_FILE}} = "Makefile"
{{BUILD_COMMAND}} = "make"
{{TEST_COMMAND}} = "make test"
{{INSTALL_COMMAND}} = "make install"

{{TEST_FRAMEWORK}} = "Unity"
{{TEST_FILE_PATTERN}} = "*_test.c"
{{TEST_CLASS_SUFFIX}} = "Test"
{{ASSERTION_LIBRARY}} = "Unity"
{{MOCKING_LIBRARY}} = "CMock"
{{COVERAGE_TOOL}} = "gcov"

{{LINTER}} = "clang-tidy"
{{FORMATTER}} = "clang-format"
{{STATIC_ANALYZER}} = "clang-static-analyzer"
{{TYPE_CHECKER}} = "clang-tidy"
{{SECURITY_SCANNER}} = "clang-tidy"
```

### Ruby
```
{{LANGUAGE_NAME}} = "Ruby"
{{LANGUAGE_ID}} = "ruby"
{{LANGUAGE_VERSION}} = "3.2"
{{LANGUAGE_RELEASE_DATE}} = "2022-12"
{{LANGUAGE_EXTENSION}} = ".rb"
{{PACKAGE_NAME}} = "example_project"
{{CLASS_NAMING}} = "PascalCase"
{{FUNCTION_NAMING}} = "snake_case"

{{BUILD_TOOL}} = "Rake"
{{PACKAGE_MANAGER}} = "Bundler"
{{DEPENDENCY_FILE}} = "Gemfile"
{{BUILD_COMMAND}} = "rake build"
{{TEST_COMMAND}} = "rspec"
{{INSTALL_COMMAND}} = "bundle install"

{{TEST_FRAMEWORK}} = "RSpec"
{{TEST_FILE_PATTERN}} = "*_spec.rb"
{{TEST_CLASS_SUFFIX}} = "Spec"
{{ASSERTION_LIBRARY}} = "RSpec Matchers"
{{MOCKING_LIBRARY}} = "RSpec Mocks"
{{COVERAGE_TOOL}} = "SimpleCov"

{{LINTER}} = "RuboCop"
{{FORMATTER}} = "RuboCop"
{{STATIC_ANALYZER}} = "RuboCop"
{{TYPE_CHECKER}} = "Steep"
{{SECURITY_SCANNER}} = "Brakeman"
```

### PHP
```
{{LANGUAGE_NAME}} = "PHP"
{{LANGUAGE_ID}} = "php"
{{LANGUAGE_VERSION}} = "8.2"
{{LANGUAGE_RELEASE_DATE}} = "2022-12"
{{LANGUAGE_EXTENSION}} = ".php"
{{PACKAGE_NAME}} = "Example\\Project"
{{CLASS_NAMING}} = "PascalCase"
{{FUNCTION_NAMING}} = "camelCase"

{{BUILD_TOOL}} = "Composer"
{{PACKAGE_MANAGER}} = "Composer"
{{DEPENDENCY_FILE}} = "composer.json"
{{BUILD_COMMAND}} = "composer install"
{{TEST_COMMAND}} = "phpunit"
{{INSTALL_COMMAND}} = "composer install"

{{TEST_FRAMEWORK}} = "PHPUnit"
{{TEST_FILE_PATTERN}} = "*Test.php"
{{TEST_CLASS_SUFFIX}} = "Test"
{{ASSERTION_LIBRARY}} = "PHPUnit Assertions"
{{MOCKING_LIBRARY}} = "PHPUnit Mocks"
{{COVERAGE_TOOL}} = "PHPUnit Coverage"

{{LINTER}} = "PHP_CodeSniffer"
{{FORMATTER}} = "PHP-CS-Fixer"
{{STATIC_ANALYZER}} = "PHPStan"
{{TYPE_CHECKER}} = "PHPStan"
{{SECURITY_SCANNER}} = "PHPStan"
```

### Scala
```
{{LANGUAGE_NAME}} = "Scala"
{{LANGUAGE_ID}} = "scala"
{{LANGUAGE_VERSION}} = "3.3"
{{LANGUAGE_RELEASE_DATE}} = "2023-05"
{{LANGUAGE_EXTENSION}} = ".scala"
{{PACKAGE_NAME}} = "com.example.project"
{{CLASS_NAMING}} = "PascalCase"
{{FUNCTION_NAMING}} = "camelCase"

{{BUILD_TOOL}} = "sbt"
{{PACKAGE_MANAGER}} = "sbt"
{{DEPENDENCY_FILE}} = "build.sbt"
{{BUILD_COMMAND}} = "sbt compile"
{{TEST_COMMAND}} = "sbt test"
{{INSTALL_COMMAND}} = "sbt update"

{{TEST_FRAMEWORK}} = "ScalaTest"
{{TEST_FILE_PATTERN}} = "*Spec.scala"
{{TEST_CLASS_SUFFIX}} = "Spec"
{{ASSERTION_LIBRARY}} = "ScalaTest Matchers"
{{MOCKING_LIBRARY}} = "ScalaMock"
{{COVERAGE_TOOL}} = "scoverage"

{{LINTER}} = "Scalafmt"
{{FORMATTER}} = "Scalafmt"
{{STATIC_ANALYZER}} = "Scalafix"
{{TYPE_CHECKER}} = "Scala Compiler"
{{SECURITY_SCANNER}} = "Scalafix"
```

### SQL
```
{{LANGUAGE_NAME}} = "SQL"
{{LANGUAGE_ID}} = "sql"
{{LANGUAGE_VERSION}} = "ANSI SQL:2023"
{{LANGUAGE_RELEASE_DATE}} = "2023"
{{LANGUAGE_EXTENSION}} = ".sql"
{{PACKAGE_NAME}} = "database_schema"
{{CLASS_NAMING}} = "snake_case"
{{FUNCTION_NAMING}} = "snake_case"

{{BUILD_TOOL}} = "Database Migration Tool"
{{PACKAGE_MANAGER}} = "Database Manager"
{{DEPENDENCY_FILE}} = "schema.sql"
{{BUILD_COMMAND}} = "migrate up"
{{TEST_COMMAND}} = "db:test"
{{INSTALL_COMMAND}} = "db:migrate"

{{TEST_FRAMEWORK}} = "Database Tests"
{{TEST_FILE_PATTERN}} = "*_test.sql"
{{TEST_CLASS_SUFFIX}} = "Test"
{{ASSERTION_LIBRARY}} = "SQL Assertions"
{{MOCKING_LIBRARY}} = "Test Containers"
{{COVERAGE_TOOL}} = "Query Coverage"

{{LINTER}} = "SQLFluff"
{{FORMATTER** = "SQLFluff"
{{STATIC_ANALYZER}} = "SQL Linter"
{{TYPE_CHECKER}} = "Database Schema Validator"
{{SECURITY_SCANNER}} = "SQL Security Scanner"
```

### Shell
```
{{LANGUAGE_NAME}} = "Shell"
{{LANGUAGE_ID}} = "shell"
{{LANGUAGE_VERSION}} = "POSIX"
{{LANGUAGE_RELEASE_DATE}} = "2018"
{{LANGUAGE_EXTENSION}} = ".sh"
{{PACKAGE_NAME}} = "project_scripts"
{{CLASS_NAMING}} = "snake_case"
{{FUNCTION_NAMING}} = "snake_case"

{{BUILD_TOOL}} = "Make"
{{PACKAGE_MANAGER}} = "Package Manager"
{{DEPENDENCY_FILE}} = "Makefile"
{{BUILD_COMMAND}} = "make"
{{TEST_COMMAND}} = "bats"
{{INSTALL_COMMAND}} = "make install"

{{TEST_FRAMEWORK}} = "Bats"
{{TEST_FILE_PATTERN}} = "*_test.bats"
{{TEST_CLASS_SUFFIX}} = "Test"
{{ASSERTION_LIBRARY}} = "Bats Assertions"
{{MOCKING_LIBRARY}} = "Test Doubles"
{{COVERAGE_TOOL}} = "Shell Check"

{{LINTER}} = "ShellCheck"
{{FORMATTER}} = "shfmt"
{{STATIC_ANALYZER}} = "ShellCheck"
{{TYPE_CHECKER}} = "ShellCheck"
{{SECURITY_SCANNER}} = "ShellCheck"
```

### R
```
{{LANGUAGE_NAME}} = "R"
{{LANGUAGE_ID}} = "r"
{{LANGUAGE_VERSION}} = "4.4"
{{LANGUAGE_RELEASE_DATE}} = "2024-04"
{{LANGUAGE_EXTENSION}} = ".R"
{{PACKAGE_NAME}} = "project"
{{CLASS_NAMING}} = "snake_case"
{{FUNCTION_NAMING}} = "snake_case"

{{BUILD_TOOL}} = "R CMD"
{{PACKAGE_MANAGER}} = "renv"
{{DEPENDENCY_FILE}} = "DESCRIPTION"
{{BUILD_COMMAND}} = "R CMD build"
{{TEST_COMMAND}} = "testthat"
{{INSTALL_COMMAND}} = "renv::install()"

{{TEST_FRAMEWORK}} = "testthat"
{{TEST_FILE_PATTERN}} = "test-*.R"
{{TEST_CLASS_SUFFIX}} = "Test"
{{ASSERTION_LIBRARY}} = "testthat"
{{MOCKING_LIBRARY}} = "mockery"
{{COVERAGE_TOOL}} = "covr"

{{LINTER}} = "lintr"
{{FORMATTER}} = "styler"
{{STATIC_ANALYZER}} = "lintr"
{{TYPE_CHECKER}} = "R CMD check"
{{SECURITY_SCANNER}} = "R CMD check"
```

## Usage Template

When the skill detects a language context, it:

1. **Identifies Language**: Maps detected patterns to `{{LANGUAGE_ID}}`
2. **Loads Variables**: Sets all variables for that language
3. **Context7 Integration**: Fetches latest documentation for `{{LANGUAGE_NAME}}`
4. **Generates Response**: Uses variables to create language-specific guidance
5. **Applies Best Practices**: Incorporates version-specific patterns

### Example Response Generation
```python
# Template for testing setup
"""
For {{LANGUAGE_NAME}} {{LANGUAGE_VERSION}} testing:

1. Install testing framework:
   {{INSTALL_COMMAND}}

2. Create test files with pattern:
   {{TEST_FILE_PATTERN}}

3. Use {{TEST_FRAMEWORK}} with {{ASSERTION_LIBRARY}}

4. Run tests with:
   {{TEST_COMMAND}}

5. Enable coverage with {{COVERAGE_TOOL}}
"""
```

This parameterized system ensures consistent, up-to-date guidance across all 13 supported languages while maintaining language-specific best practices.