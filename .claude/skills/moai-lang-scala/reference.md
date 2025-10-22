# Scala 3.7+ CLI Reference — Tool Command Matrix

**Framework**: Scala 3.7+ + ScalaTest 3.2.19+ + sbt 1.10+ + scalafmt 3.8+

---

## Scala Runtime Commands

| Command | Purpose | Example |
|---------|---------|---------|
| `scala --version` | Check Scala version | `scala --version` → `Scala 3.7.3` |
| `scala` | Start Scala REPL | `scala` |
| `scala script.scala` | Run Scala script | `scala script.scala` |
| `scala -e "code"` | Execute Scala code inline | `scala -e "println(1 + 1)"` |
| `scala -cp classpath` | Run with classpath | `scala -cp lib/* Main.scala` |
| `scala --help` | Show Scala help | `scala --help` |
| `scalac file.scala` | Compile Scala file | `scalac Main.scala` |
| `scalac -d output file.scala` | Compile to directory | `scalac -d bin Main.scala` |

---

## Build Tool (sbt 1.10+)

| Command | Purpose | Example |
|---------|---------|---------|
| `sbt` | Start sbt shell | `sbt` |
| `sbt compile` | Compile project | `sbt compile` |
| `sbt run` | Run main class | `sbt run` |
| `sbt test` | Run all tests | `sbt test` |
| `sbt "testOnly TestClass"` | Run specific test | `sbt "testOnly CalculatorTest"` |
| `sbt clean` | Clean build artifacts | `sbt clean` |
| `sbt package` | Package project as JAR | `sbt package` |
| `sbt publishLocal` | Publish to local repository | `sbt publishLocal` |
| `sbt console` | Start Scala REPL with project | `sbt console` |
| `sbt update` | Update dependencies | `sbt update` |
| `sbt reload` | Reload build definition | `sbt reload` |
| `sbt projects` | List all projects | `sbt projects` |
| `sbt new` | Create project from template | `sbt new scala/scala3.g8` |
| `sbt dependencyTree` | Show dependency tree | `sbt dependencyTree` |
| `sbt evicted` | Show evicted dependencies | `sbt evicted` |

**sbt Shell Commands** (inside `sbt`):
```scala
compile          // Compile sources
~compile         // Continuous compilation (watch mode)
test             // Run tests
~test            // Continuous testing
console          // Start Scala REPL
clean            // Clean build
reload           // Reload build.sbt changes
exit             // Exit sbt shell
```

---

## Testing Framework (ScalaTest 3.2.19+)

| Command | Purpose | Example |
|---------|---------|---------|
| `sbt test` | Run all tests | `sbt test` |
| `sbt "testOnly *CalculatorTest"` | Run specific test class | `sbt "testOnly *CalculatorTest"` |
| `sbt "testOnly * -- -z add"` | Run tests matching name | `sbt "testOnly * -- -z add"` |
| `sbt "testOnly * -- -l Slow"` | Exclude tagged tests | `sbt "testOnly * -- -l Slow"` |
| `sbt "testOnly * -- -n Fast"` | Run only tagged tests | `sbt "testOnly * -- -n Fast"` |
| `sbt testQuick` | Run failed tests only | `sbt testQuick` |
| `sbt coverage test` | Run with coverage | `sbt coverage test` |
| `sbt coverageReport` | Generate coverage report | `sbt coverageReport` |

**ScalaTest Styles**:
- **FunSuite**: xUnit-style tests
- **FlatSpec**: BDD-style tests
- **WordSpec**: Specification-style tests
- **FunSpec**: RSpec-style tests
- **FreeSpec**: Free-form nested tests

---

## Code Formatting (scalafmt 3.8+)

| Command | Purpose | Example |
|---------|---------|---------|
| `scalafmt` | Format all files | `scalafmt` |
| `scalafmt --check` | Check if formatting needed | `scalafmt --check` |
| `scalafmt file.scala` | Format specific file | `scalafmt src/Main.scala` |
| `scalafmt --test` | Test formatting (CI mode) | `scalafmt --test` |
| `sbt scalafmt` | Format via sbt | `sbt scalafmt` |
| `sbt scalafmtCheck` | Check formatting via sbt | `sbt scalafmtCheck` |
| `sbt scalafmtAll` | Format all configs | `sbt scalafmtAll` |

**.scalafmt.conf Configuration**:
```conf
version = "3.8.3"
runner.dialect = scala3

maxColumn = 120
indent.defnSite = 2
align.preset = most

rewrite.rules = [
  RedundantBraces,
  RedundantParens,
  SortModifiers,
  PreferCurlyFors
]

rewrite.scala3.convertToNewSyntax = true
rewrite.scala3.removeOptionalBraces = true
```

---

## Linting (Scalafix + WartRemover)

### Scalafix

| Command | Purpose | Example |
|---------|---------|---------|
| `sbt "scalafix RemoveUnused"` | Remove unused imports | `sbt "scalafix RemoveUnused"` |
| `sbt "scalafix --check"` | Check without fixing | `sbt "scalafix --check"` |
| `sbt scalafixAll` | Fix all files | `sbt scalafixAll` |

**.scalafix.conf Configuration**:
```conf
rules = [
  DisableSyntax,
  LeakingImplicitClassVal,
  NoAutoTupling,
  NoValInForComprehension,
  OrganizeImports,
  RemoveUnused
]

OrganizeImports.groupedImports = Merge
OrganizeImports.removeUnused = true
```

---

## Dependency Management (sbt)

**build.sbt Example**:
```scala
name := "my-project"
version := "0.1.0"
scalaVersion := "3.7.3"

libraryDependencies ++= Seq(
  "org.scalatest" %% "scalatest" % "3.2.19" % Test,
  "org.typelevel" %% "cats-core" % "2.13.0",
  "com.typesafe.akka" %% "akka-actor-typed" % "2.9.8"
)

scalacOptions ++= Seq(
  "-encoding", "UTF-8",
  "-deprecation",
  "-feature",
  "-unchecked",
  "-Xfatal-warnings"
)

testOptions in Test += Tests.Argument(TestFrameworks.ScalaTest, "-oD")
```

---

## Combined Workflow (Quality Gate)

**Before Commit** (all must pass):

```bash
#!/bin/bash
set -e

echo "Running Scala quality gate checks..."

# 1. Compile code
echo "1. Compiling..."
sbt compile

# 2. Run tests with coverage
echo "2. Running tests..."
sbt clean coverage test coverageReport

# 3. Check formatting
echo "4. Checking formatting..."
sbt scalafmtCheck scalafmtSbtCheck

# 4. Run linting
echo "5. Running linter..."
sbt "scalafix --check"

echo "✅ All quality gates passed!"
```

---

## TRUST 5 Principles Integration

### T - Test First (ScalaTest 3.2.19+)
```bash
sbt test
```

### R - Readable (scalafmt 3.8+)
```bash
sbt scalafmtAll
sbt scalafmtCheck
```

### U - Unified Types (Scala 3 Type System)
```bash
sbt compile  # Strict compile-time type checking
```

### S - Security
```bash
sbt dependencyCheck
```

### T - Trackable (@TAG)
```bash
rg '@(CODE|TEST|SPEC):' -n src/ --type scala
```

---

**Version**: 0.1.0
**Created**: 2025-10-22
**Framework**: Scala 3.7+ CLI Tools Reference
