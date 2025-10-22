# moai-lang-clojure - Working Examples

_Last updated: 2025-10-22_

## Example 1: Basic Clojure Project Setup with Leiningen

### Create New Project

```bash
# Create new application project
lein new app calculator
cd calculator
```

### Project Structure

```
calculator/
├── project.clj
├── README.md
├── src/
│   └── calculator/
│       └── core.clj
├── test/
│   └── calculator/
│       └── core_test.clj
└── resources/
```

### src/calculator/core.clj

```clojure
(ns calculator.core
  (:gen-class))

;; @CODE:CALC-001 | SPEC: SPEC-CALC-001.md | TEST: test/calculator/core_test.clj
(defn add
  "Add two numbers"
  [a b]
  (+ a b))

(defn subtract
  "Subtract b from a"
  [a b]
  (- a b))

(defn multiply
  "Multiply two numbers"
  [a b]
  (* a b))

(defn divide
  "Divide a by b, returns nil if b is zero"
  [a b]
  (when-not (zero? b)
    (/ a b)))

(defn -main
  "Application entry point"
  [& args]
  (println "Calculator ready!")
  (println "2 + 3 =" (add 2 3))
  (println "5 - 2 =" (subtract 5 2))
  (println "4 * 3 =" (multiply 4 3))
  (println "10 / 2 =" (divide 10 2)))
```

### test/calculator/core_test.clj

```clojure
(ns calculator.core-test
  (:require [clojure.test :refer [deftest is testing are]]
            [calculator.core :as calc]))

;; @TEST:CALC-001
(deftest test-add
  (testing "Addition of two positive numbers"
    (is (= 5 (calc/add 2 3))))
  (testing "Addition of negative numbers"
    (is (= -5 (calc/add -2 -3))))
  (testing "Addition with zero"
    (is (= 5 (calc/add 5 0)))))

(deftest test-subtract
  (are [expected a b] (= expected (calc/subtract a b))
    1 3 2
    -1 2 3
    5 5 0
    0 5 5))

(deftest test-multiply
  (testing "Multiplication"
    (is (= 12 (calc/multiply 3 4)))
    (is (= 0 (calc/multiply 5 0)))
    (is (= -12 (calc/multiply -3 4)))))

(deftest test-divide
  (testing "Division"
    (is (= 2 (calc/divide 10 5)))
    (is (= 1/2 (calc/divide 1 2))))
  (testing "Division by zero returns nil"
    (is (nil? (calc/divide 10 0)))))
```

### Run Tests

```bash
# Run all tests
lein test

# Expected output:
# Testing calculator.core-test
# Ran 4 tests containing 11 assertions.
# 0 failures, 0 errors.
```

## Example 2: TDD Workflow (RED-GREEN-REFACTOR)

### RED Phase: Write Failing Test First

```clojure
;; test/calculator/stats_test.clj
(ns calculator.stats-test
  (:require [clojure.test :refer [deftest is testing]]
            [calculator.stats :as stats]))

;; @TEST:STATS-001
(deftest test-average
  (testing "Calculate average of numbers"
    (is (= 3 (stats/average [1 2 3 4 5])))
    (is (= 0 (stats/average [])))  ;; Empty list returns 0
    (is (= 5 (stats/average [5])))))
```

```bash
# Run test - should FAIL (namespace doesn't exist)
lein test

# Output:
# Could not locate calculator/stats__init.class or calculator/stats.clj
# Tests failed.
```

### GREEN Phase: Implement Minimum Code to Pass

```clojure
;; src/calculator/stats.clj
(ns calculator.stats)

;; @CODE:STATS-001 | SPEC: SPEC-STATS-001.md | TEST: test/calculator/stats_test.clj
(defn average
  "Calculate the average of a collection of numbers"
  [numbers]
  (if (empty? numbers)
    0
    (/ (reduce + numbers) (count numbers))))
```

```bash
# Run test again - should PASS
lein test

# Output:
# Testing calculator.stats-test
# Ran 1 tests containing 3 assertions.
# 0 failures, 0 errors.
```

### REFACTOR Phase: Improve Code Quality

```clojure
;; src/calculator/stats.clj (refactored)
(ns calculator.stats)

;; @CODE:STATS-001 | SPEC: SPEC-STATS-001.md | TEST: test/calculator/stats_test.clj
(defn average
  "Calculate the average of a collection of numbers.
   Returns 0 for empty collections."
  [numbers]
  (if (seq numbers)  ; More idiomatic than (empty? numbers)
    (/ (reduce + numbers) (count numbers))
    0))

;; Additional helper functions
(defn median
  "Calculate the median of a collection of numbers"
  [numbers]
  (when (seq numbers)
    (let [sorted (sort numbers)
          cnt (count sorted)
          mid (quot cnt 2)]
      (if (odd? cnt)
        (nth sorted mid)
        (/ (+ (nth sorted mid)
              (nth sorted (dec mid)))
           2)))))

(defn variance
  "Calculate variance of a collection of numbers"
  [numbers]
  (when (seq numbers)
    (let [avg (average numbers)
          squared-diffs (map #(Math/pow (- % avg) 2) numbers)]
      (average squared-diffs))))
```

```bash
# Run tests after refactor - should still PASS
lein test

# Run static analysis
lein kibit
# Output: No suggestions! Your code is idiomatic.
```

## Example 3: Quality Gate with Coverage and Linting

### Setup project.clj with Quality Tools

```clojure
(defproject calculator "0.1.0-SNAPSHOT"
  :description "Calculator with TDD"
  :url "https://github.com/user/calculator"
  :license {:name "EPL-2.0"
            :url "https://www.eclipse.org/legal/epl-2.0/"}
  :dependencies [[org.clojure/clojure "1.12.0"]]
  :main ^:skip-aot calculator.core
  :target-path "target/%s"
  :plugins [[lein-cloverage "1.2.4"]
            [lein-kibit "0.1.8"]
            [jonase/eastwood "1.4.3"]]
  :profiles {:dev {:dependencies [[cloverage "1.2.4"]]}
             :uberjar {:aot :all
                       :jvm-opts ["-Dclojure.compiler.direct-linking=true"]}})
```

### Run Quality Checks

```bash
# 1. Run all tests
lein test
# ✅ Ran 5 tests containing 15 assertions.
# 0 failures, 0 errors.

# 2. Check code coverage (must be ≥85%)
lein cloverage --fail-threshold 85
# ================== Summary ==================
# Lines      : 88.2% (75/85)
# Branches   : 85.7% (12/14)
# Forms      : 87.5% (56/64)
# =============================================
# ✅ Coverage passed threshold

# 3. Run static analysis (Kibit)
lein kibit
# ✅ No suggestions! Your code is idiomatic.

# 4. Run linter (Eastwood)
lein eastwood
# == Eastwood 1.4.3 Clojure 1.12.0 JVM 11.0.11
# == Linting calculator.core ==
# == Linting calculator.stats ==
# ✅ No warnings found
```

### Full Quality Gate Script

```bash
#!/bin/bash
# quality-gate.sh

set -e

echo "=== Running tests ==="
lein test

echo "=== Checking code coverage (≥85%) ==="
lein cloverage --fail-threshold 85

echo "=== Running static analysis (Kibit) ==="
lein kibit

echo "=== Running linter (Eastwood) ==="
lein eastwood

echo "✅ All quality gates passed!"
```

## Example 4: Fixtures and Advanced Testing

### Using Fixtures for Database Testing

```clojure
(ns calculator.db-test
  (:require [clojure.test :refer [deftest is testing use-fixtures]]
            [calculator.db :as db]))

;; Test state
(def test-db (atom {}))

;; Once fixture - runs once before all tests
(defn setup-db-once [f]
  (println "Setting up test database...")
  (reset! test-db {:connection "test-db"})
  (f)
  (println "Tearing down test database...")
  (reset! test-db {}))

;; Each fixture - runs before each test
(defn clean-db-each [f]
  (swap! test-db assoc :data {})
  (f)
  (swap! test-db dissoc :data))

(use-fixtures :once setup-db-once)
(use-fixtures :each clean-db-each)

;; @TEST:DB-001
(deftest test-db-connection
  (testing "Database connection exists"
    (is (= "test-db" (:connection @test-db)))))

(deftest test-db-operations
  (testing "Can store and retrieve data"
    (swap! test-db assoc-in [:data :key1] "value1")
    (is (= "value1" (get-in @test-db [:data :key1])))))
```

### Using `are` Macro for Multiple Assertions

```clojure
(ns calculator.validation-test
  (:require [clojure.test :refer [deftest are]]
            [calculator.validation :as v]))

;; @TEST:VALID-001
(deftest test-valid-email?
  (are [expected email] (= expected (v/valid-email? email))
    true  "user@example.com"
    true  "test.user+tag@example.co.uk"
    false "notanemail"
    false "@example.com"
    false "user@"
    false ""))

(deftest test-in-range?
  (are [expected value min max] (= expected (v/in-range? value min max))
    true  5   0 10
    true  0   0 10
    true  10  0 10
    false -1  0 10
    false 11  0 10))
```

### Testing Exceptions

```clojure
(ns calculator.errors-test
  (:require [clojure.test :refer [deftest is testing]]
            [calculator.errors :as errors]))

;; @TEST:ERROR-001
(deftest test-safe-divide
  (testing "Safe division throws on zero divisor"
    (is (thrown? ArithmeticException
                 (errors/safe-divide 10 0))))

  (testing "Safe division with message check"
    (is (thrown-with-msg? ArithmeticException
                          #"Cannot divide by zero"
                          (errors/safe-divide 10 0)))))
```

## Example 5: REPL-Driven Development

### Interactive Development Workflow

```clojure
;; Start REPL
;; $ lein repl

;; Load namespace
(require '[calculator.core :as calc])

;; Test functions interactively
(calc/add 2 3)
;; => 5

(calc/divide 10 0)
;; => nil

;; Reload namespace after changes
(require '[calculator.core :as calc] :reload)

;; Run tests from REPL
(require '[clojure.test :refer [run-tests]])
(run-tests 'calculator.core-test)
;; => {:test 4, :pass 11, :fail 0, :error 0}

;; Run specific test
(require '[clojure.test :refer [test-var]])
(test-var #'calculator.core-test/test-add)
;; => {:test 1, :pass 3, :fail 0, :error 0}

;; Check documentation
(doc calc/divide)
;; => Divide a by b, returns nil if b is zero

;; Examine source
(source calc/divide)
```

---

_For more details on best practices and TRUST principles, see SKILL.md_
