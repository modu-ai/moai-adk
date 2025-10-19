---
name: moai-lang-haskell
description: Haskell best practices with HUnit, Stack/Cabal, and pure functional programming
allowed-tools:
  - Read
  - Bash
tier: 3
auto-load: "true"
---

# Haskell Expert

## What it does

Provides Haskell-specific expertise for TDD development, including HUnit testing, Stack/Cabal build tools, and pure functional programming with strong type safety.

## When to use

- "Haskell 테스트 작성", "HUnit 사용법", "순수 함수형 프로그래밍", "타입 시스템", "컴파일러 개발", "암호화"
- "Yesod", "Servant", "Pandoc", "Hakyll", "Parsec"
- "타입 안전성", "수학적 정확성", "도메인 모델링"
- Automatically invoked when working with Haskell projects
- Haskell SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **HUnit**: Unit testing framework
- **QuickCheck**: Property-based testing
- **Hspec**: BDD-style testing
- Test coverage with hpc

**Build Tools**:
- **Stack**: Reproducible builds, dependency resolution
- **Cabal**: Haskell package system
- **hpack**: Alternative package description

**Code Quality**:
- **hlint**: Haskell linter
- **stylish-haskell**: Code formatting
- **GHC warnings**: Compiler-level checks

**Functional Programming**:
- **Pure functions**: No side effects
- **Monads**: IO, Maybe, Either, State
- **Functors/Applicatives**: Abstraction patterns
- **Type classes**: Polymorphism
- **Lazy evaluation**: Infinite data structures

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Prefer total functions (avoid partial)
- Type-driven development
- Point-free style (when readable)
- Avoid do-notation overuse

## Modern Haskell (GHC 9.x)

**Recommended Version**: GHC 9.6+ for production, GHC 9.0+ for modern features

**Modern Features (GHC Extensions)**:
- **OverloadedRecordDot** (9.2+): `person.name` syntax
- **RecordWildCards**: Pattern matching with `{..}`
- **DuplicateRecordFields**: Same field names in different types
- **TypeApplications**: Explicit type arguments `@Int`
- **DataKinds**: Type-level programming
- **GADTs**: Generalized algebraic data types
- **DerivingStrategies**: Control deriving mechanism

**Version Check**:
```bash
ghc --version
stack --version
cabal --version
```

## Build System Commands

### Using Stack (Recommended)
```bash
# Create new project
stack new my-app
stack new my-lib simple

# Build
stack build
stack build --fast  # Without optimization

# Test
stack test
stack test --coverage

# Run
stack run
stack exec my-app

# REPL
stack ghci

# Dependencies
stack ls dependencies
stack update

# Clean
stack clean
```

**stack.yaml Configuration**:
```yaml
resolver: lts-22.0  # GHC 9.6.3

packages:
  - .

extra-deps: []

flags: {}
```

**package.yaml (hpack)**:
```yaml
name: my-app
version: 0.1.0.0
github: user/my-app

dependencies:
  - base >= 4.7 && < 5
  - text
  - bytestring
  - containers

library:
  source-dirs: src

executables:
  my-app:
    main: Main.hs
    source-dirs: app
    dependencies:
      - my-app

tests:
  my-app-test:
    main: Spec.hs
    source-dirs: test
    dependencies:
      - my-app
      - hspec
      - QuickCheck
```

### Using Cabal
```bash
# Initialize project
cabal init --interactive

# Build
cabal build
cabal build --enable-optimization

# Test
cabal test
cabal test --test-show-details=streaming

# Run
cabal run

# REPL
cabal repl

# Dependencies
cabal update
cabal outdated

# Install
cabal install
```

**my-app.cabal**:
```haskell
cabal-version: 3.0
name: my-app
version: 0.1.0.0

library
  exposed-modules: MyLib
  build-depends:
    base ^>=4.18,
    text,
    containers
  hs-source-dirs: src
  default-language: Haskell2010

executable my-app
  main-is: Main.hs
  build-depends:
    base,
    my-app
  hs-source-dirs: app
  default-language: Haskell2010

test-suite my-app-test
  type: exitcode-stdio-1.0
  main-is: Spec.hs
  build-depends:
    base,
    my-app,
    hspec,
    QuickCheck
  hs-source-dirs: test
  default-language: Haskell2010
```

### Common Development Commands
```bash
# Format code
stylish-haskell -i src/**/*.hs

# Lint
hlint src/
hlint --refactor --refactor-options="-i" src/

# Type holes (for type-driven development)
# Use _ in code, GHC will suggest types

# Generate documentation
stack haddock
cabal haddock

# Profile performance
stack build --profile
stack exec -- my-app +RTS -p
```

## Examples

### Example 1: TDD with HUnit
User: "/alfred:2-run PARSE-001"
Claude: (creates RED test with HUnit, GREEN implementation with pure functions, REFACTOR)

### Example 2: Property testing
User: "QuickCheck 속성 테스트"
Claude: (creates property-based tests for invariants)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Haskell-specific review)
- alfred-refactoring-coach (functional refactoring)
