# moai-lang-clojure - CLI Reference

_Last updated: 2025-10-22_

## Quick Reference

### Installation

```bash
# Install Leiningen (macOS)
brew install leiningen

# Install Leiningen (Linux/Unix)
curl https://raw.githubusercontent.com/technomancy/leiningen/stable/bin/lein > ~/bin/lein
chmod +x ~/bin/lein
lein version  # First run downloads dependencies

# Verify installation
lein version  # Should show Leiningen 2.11.2+ on Java 11+
```

### Common Commands

```bash
# Create new project
lein new app my-project
lein new lib my-library

# Run REPL
lein repl

# Run tests
lein test                    # All tests
lein test :only my.ns/test-fn  # Specific test

# Run application
lein run

# Build uberjar (standalone)
lein uberjar

# Check for outdated dependencies
lein ancient

# Code coverage with Cloverage
lein cloverage

# Clean build artifacts
lein clean
```

## Tool Versions (2025-10-22)

- **Clojure**: 1.12.0 - Functional Lisp for the JVM
- **Leiningen**: 2.11.2 - Build automation and dependency management
- **clojure.test**: 1.12.0 (built-in) - Official testing framework

## Official Documentation Links

- **Clojure Official**: https://clojure.org/
- **Clojure API**: https://clojure.github.io/clojure/
- **clojure.test API**: https://clojure.github.io/clojure/clojure.test-api.html
- **Leiningen**: https://leiningen.org/
- **Leiningen Tutorial**: Run `lein help tutorial`
- **ClojureDocs**: https://clojuredocs.org/
- **Practicalli Clojure**: https://practical.li/clojure/

## clojure.test Testing Framework

### Core Macros

**deftest** - Define a test function
```clojure
(deftest my-test
  (is (= 4 (+ 2 2))))
```

**is** - Core assertion macro
```clojure
(is (= expected actual))
(is (thrown? ExceptionType (risky-fn)))
```

**testing** - Group assertions with description
```clojure
(deftest my-test
  (testing "addition"
    (is (= 4 (+ 2 2))))
  (testing "subtraction"
    (is (= 0 (- 2 2)))))
```

**are** - Multiple assertions with template
```clojure
(are [x y] (= x y)
  2 (+ 1 1)
  4 (* 2 2)
  6 (+ 3 3))
```

### Fixtures

**Once fixtures** - Run once before/after all tests
```clojure
(use-fixtures :once
  (fn [f]
    ;; setup
    (f)
    ;; teardown
    ))
```

**Each fixtures** - Run before/after each test
```clojure
(use-fixtures :each
  (fn [f]
    ;; setup
    (f)
    ;; teardown
    ))
```

### Running Tests

```bash
# From command line
lein test

# Specific namespace
lein test my.app.core-test

# From REPL
(require '[clojure.test :refer [run-tests]])
(run-tests 'my.app.core-test)

# Run single test
(require '[clojure.test :refer [test-var]])
(test-var #'my.app.core-test/my-test)
```

## Leiningen Project Structure

### Standard Directory Layout

```
project-name/
├── project.clj           # Project configuration
├── README.md
├── src/
│   └── project_name/     # Source code (underscores)
│       └── core.clj
├── test/
│   └── project_name/     # Tests mirror src structure
│       └── core_test.clj
├── resources/            # Non-code resources
├── target/               # Build artifacts (gitignore)
└── doc/                  # Documentation
```

### project.clj Configuration

```clojure
(defproject my-project "0.1.0-SNAPSHOT"
  :description "Project description"
  :url "https://github.com/user/my-project"
  :license {:name "EPL-2.0"
            :url "https://www.eclipse.org/legal/epl-2.0/"}
  :dependencies [[org.clojure/clojure "1.12.0"]]
  :main ^:skip-aot my-project.core
  :target-path "target/%s"
  :profiles {:dev {:dependencies [[cloverage "1.2.4"]]}
             :uberjar {:aot :all
                       :jvm-opts ["-Dclojure.compiler.direct-linking=true"]}})
```

### Useful Plugins

```clojure
:plugins [[lein-cloverage "1.2.4"]   ; Code coverage
          [lein-ancient "1.0.0-RC3"]  ; Check outdated deps
          [lein-kibit "0.1.8"]        ; Static analysis
          [jonase/eastwood "1.4.3"]]  ; Clojure lint
```

## Code Coverage with Cloverage

### Installation (User-wide)

Add to `~/.lein/profiles.clj`:
```clojure
{:user {:plugins [[lein-cloverage "1.2.4"]]}}
```

### Usage

```bash
# Generate coverage report
lein cloverage

# Set minimum coverage threshold (85%)
lein cloverage --fail-threshold 85

# Exclude namespaces
lein cloverage --exclude-pattern ".*test.*"

# Output formats
lein cloverage --html  # HTML report (default)
lein covera ge --lcov   # LCOV format
```

### Coverage Output

```
================== Summary ==================
Lines      : 85.5% (205/240)
Branches   : 78.3% (47/60)
Forms      : 82.1% (156/190)
=============================================
```

## Static Analysis Tools

### Kibit - Suggest idiomatic code

```bash
# Install
lein plugin install lein-kibit 0.1.8

# Run analysis
lein kibit

# Example output:
# Consider using:
#   (when test ...)
# instead of:
#   (if test (do ...))
```

### Eastwood - Clojure linter

```bash
# Add to project.clj plugins
[jonase/eastwood "1.4.3"]

# Run linter
lein eastwood

# Exclude specific linters
lein eastwood '{:exclude-linters [:unlimited-use]}'
```

## Testing Best Practices

### Test File Naming Convention

```
src/my_app/core.clj      → test/my_app/core_test.clj
src/my_app/util.clj      → test/my_app/util_test.clj
```

### Test Namespace Convention

```clojure
(ns my-app.core-test
  (:require [clojure.test :refer [deftest is testing]]
            [my-app.core :as core]))
```

### Assertion Patterns

```clojure
;; Equality
(is (= 4 (+ 2 2)))

;; Truthy/Falsy
(is (true? (boolean value)))
(is (false? (nil? value)))

;; Exceptions
(is (thrown? ArithmeticException (/ 1 0)))
(is (thrown-with-msg? ArithmeticException #"Divide by zero" (/ 1 0)))

;; Collections
(is (= #{:a :b} (set [:a :b :a])))
(is (= {:a 1} (select-keys {:a 1 :b 2} [:a])))
```

## Common Leiningen Tasks

### Help System

```bash
lein help                # List all tasks
lein help tutorial       # Interactive tutorial
lein help sample         # Sample project.clj reference
lein help test           # Help for specific task
```

### Dependency Management

```bash
# Show dependency tree
lein deps :tree

# Check for conflicts
lein deps :tree | grep -A 10 conflict

# Update dependencies
lein ancient upgrade
```

---

_For detailed usage and best practices, see SKILL.md_
