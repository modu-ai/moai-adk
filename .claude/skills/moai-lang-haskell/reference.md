# moai-lang-haskell - CLI Reference

_Last updated: 2025-10-22_

## Official Documentation Links

- **Haskell.org**: https://www.haskell.org/
- **GHC User Guide**: https://downloads.haskell.org/ghc/latest/docs/users_guide/
- **Stack**: https://docs.haskellstack.org/
- **Cabal**: https://cabal.readthedocs.io/
- **HUnit**: https://hackage.haskell.org/package/HUnit
- **Haddock**: https://haskell-haddock.readthedocs.io/

## Tool Versions (2025-10-22)

| Tool | Version | Release Date | Status |
|------|---------|--------------|--------|
| **GHC** | 9.10.3 | 2025-09 | ✅ Latest |
| **Stack** | 3.7.1 | 2025-08 | ✅ Stable |
| **Cabal** | 3.12.0 | 2025-07 | ✅ Latest |
| **HUnit** | 1.6.2 | 2023-03 | ✅ Stable |

## Installation

### Prerequisites (via ghcup)

```bash
# Install ghcup (recommended method)
curl --proto '=https' --tlsv1.2 -sSf https://get-haskell.io | sh

# Check installation
ghcup --version

# Install GHC
ghcup install ghc 9.10.3
ghcup set ghc 9.10.3

# Install Cabal
ghcup install cabal 3.12.0
ghcup set cabal 3.12.0

# Install Stack
ghcup install stack 3.7.1
ghcup set stack 3.7.1

# Verify installations
ghc --version
cabal --version
stack --version
```

### Project Initialization

**With Stack**:
```bash
# Create new project
stack new my-project
cd my-project

# Setup GHC
stack setup

# Build project
stack build

# Install dependencies
stack install HUnit
```

**With Cabal**:
```bash
# Initialize project
cabal init --non-interactive

# Update package list
cabal update

# Install dependencies
cabal install --only-dependencies

# Build project
cabal build
```

## Common Commands

### Stack Commands

```bash
# Build project
stack build

# Run executable
stack exec my-project-exe

# Run tests
stack test

# Run tests with coverage
stack test --coverage

# Clean build artifacts
stack clean

# Install executable globally
stack install

# Run GHCi REPL
stack ghci

# Run specific test suite
stack test --test-arguments="--match '/UserService/'"

# Build with optimizations
stack build --fast

# Build with profiling
stack build --profile

# Generate documentation
stack haddock

# Check for dependency updates
stack list-dependencies
```

### Cabal Commands

```bash
# Configure project
cabal configure --enable-tests

# Build project
cabal build

# Run executable
cabal run my-project

# Run tests
cabal test

# Run tests with coverage
cabal test --enable-coverage

# Clean build artifacts
cabal clean

# Install executable
cabal install

# Run GHCi REPL
cabal repl

# Generate documentation
cabal haddock

# Check for outdated dependencies
cabal outdated

# Format code
cabal-fmt --inplace *.cabal
```

### GHC Compiler Options

```bash
# Compile with warnings
ghc -Wall -Werror Main.hs

# Compile with optimizations
ghc -O2 Main.hs

# Generate profiling executable
ghc -prof -fprof-auto Main.hs

# Enable type holes for development
ghc -fdefer-type-errors Main.hs

# Generate core output (for optimization analysis)
ghc -ddump-simpl Main.hs

# Check type inference
ghc -ftype-directed-code-analysis Main.hs
```

### Testing with HUnit

```bash
# Run single test file
runhaskell test/Spec.hs

# Run with stack
stack test

# Run with cabal
cabal test

# Run with coverage
stack test --coverage
cabal test --enable-coverage

# View coverage report
stack hpc report
```

## Configuration Files

### stack.yaml

```yaml
resolver: lts-22.0  # GHC 9.10.3

packages:
  - .

extra-deps: []

flags: {}

ghc-options:
  "$locals": -Wall -Werror -Wcompat
```

### package.yaml (Stack)

```yaml
name: my-project
version: 0.1.0.0
github: "username/my-project"
license: BSD3
author: "Author Name"
maintainer: "author@example.com"
copyright: "2025 Author Name"

extra-source-files:
  - README.md
  - CHANGELOG.md

synopsis: Short description
category: Web

description: >
  Longer description of the package.

dependencies:
  - base >= 4.7 && < 5
  - text
  - containers
  - mtl

library:
  source-dirs: src
  ghc-options:
    - -Wall
    - -Wcompat
    - -Widentities
    - -Wincomplete-record-updates
    - -Wincomplete-uni-patterns
    - -Wmissing-home-modules
    - -Wpartial-fields
    - -Wredundant-constraints

executables:
  my-project-exe:
    main: Main.hs
    source-dirs: app
    ghc-options:
      - -threaded
      - -rtsopts
      - -with-rtsopts=-N
    dependencies:
      - my-project

tests:
  my-project-test:
    main: Spec.hs
    source-dirs: test
    ghc-options:
      - -threaded
      - -rtsopts
      - -with-rtsopts=-N
    dependencies:
      - my-project
      - HUnit >= 1.6.2
```

### .cabal File

```cabal
cabal-version: 3.0
name: my-project
version: 0.1.0.0
license: BSD-3-Clause
author: Author Name
maintainer: author@example.com
build-type: Simple

common warnings
  ghc-options:
    -Wall
    -Wcompat
    -Widentities
    -Wincomplete-record-updates
    -Wincomplete-uni-patterns
    -Wmissing-home-modules
    -Wpartial-fields
    -Wredundant-constraints

library
  import: warnings
  exposed-modules:
    Calculator
    UserService
  build-depends:
      base >= 4.18 && < 5
    , text
    , containers
  hs-source-dirs: src
  default-language: Haskell2010

executable my-project
  import: warnings
  main-is: Main.hs
  build-depends:
      base
    , my-project
  hs-source-dirs: app
  default-language: Haskell2010

test-suite my-project-test
  import: warnings
  type: exitcode-stdio-1.0
  main-is: Spec.hs
  hs-source-dirs: test
  build-depends:
      base
    , my-project
    , HUnit >= 1.6.2
  default-language: Haskell2010
```

### .ghci Configuration

```haskell
-- .ghci file for REPL customization
:set -Wall
:set -Wcompat
:set prompt "λ> "
:set prompt-cont " | "

-- Load common modules
:module + Data.List Data.Maybe Control.Monad

-- Enable multi-line input
:set +m

-- Show types after evaluation
:set +t
```

## Best Practices

### Code Organization

✅ Use explicit type signatures for all top-level definitions
✅ Prefer pure functions over IO
✅ Use newtype for type safety
✅ Leverage type classes for polymorphism
✅ Keep modules focused and cohesive

### Type Safety

✅ Make illegal states unrepresentable
✅ Use sum types (Either, Maybe) for error handling
✅ Avoid partial functions (head, tail, !!)
✅ Use totality analysis tools
✅ Leverage the type system for compile-time guarantees

### Testing

✅ Write property-based tests with QuickCheck
✅ Use HUnit for unit tests
✅ Test pure functions extensively
✅ Mock IO with type classes
✅ Maintain ≥85% code coverage

### Performance

✅ Profile before optimizing
✅ Use strict evaluation when needed
✅ Leverage GHC optimizations (-O2)
✅ Use appropriate data structures
✅ Avoid space leaks with strictness annotations

### Error Handling

✅ Use Either for recoverable errors
✅ Use Maybe for optional values
✅ Use ExceptT for monadic error handling
✅ Document partial functions clearly
✅ Provide totality analysis

## Common Issues & Solutions

### Issue: Stack resolver too old
**Solution**: Update `resolver` in stack.yaml to latest LTS

### Issue: Cabal hell (dependency conflicts)
**Solution**: Use stack instead, or cabal freeze for reproducible builds

### Issue: Space leak in long-running programs
**Solution**: Use strict evaluation (!), bang patterns, and profiling

### Issue: Slow compilation times
**Solution**: Enable -j flag for parallel builds, use incremental compilation

### Issue: Type errors difficult to read
**Solution**: Enable -fprint-explicit-kinds and -fprint-explicit-foralls

## HUnit Test Functions Quick Reference

### Basic Assertions
- `assertEqual msg expected actual` - Assert equality
- `assertBool msg condition` - Assert boolean condition
- `assertFailure msg` - Fail with message
- `assertString msg` - Fail if string non-empty

### Test Organization
- `TestCase assertion` - Single test case
- `TestList [test1, test2]` - List of tests
- `TestLabel "name" test` - Named test

### Running Tests
- `runTestTT tests` - Run test suite
- `runTestText putTextToHandle tests` - Custom output

## GHC Warning Flags

```haskell
-- Recommended warning flags
-Wall                           -- Enable all warnings
-Wcompat                        -- Future compatibility warnings
-Widentities                    -- Redundant identities
-Wincomplete-record-updates     -- Incomplete record updates
-Wincomplete-uni-patterns       -- Incomplete pattern matches
-Wmissing-home-modules          -- Missing home modules
-Wpartial-fields                -- Partial record fields
-Wredundant-constraints         -- Redundant type constraints
-Werror                         -- Treat warnings as errors
```

## Debugging

```bash
# Run with debugging
stack build --trace

# Profile heap usage
stack build --profile
stack exec -- my-project-exe +RTS -p -h

# View profiling report
stack exec -- hp2ps -c my-project-exe.hp
```

---

_For working examples and use cases, see examples.md_
