# Language: Java and Kotlin

When to read: `internal/quality` sent you here because the project's detected languages include `java` or `kotlin`. The project's own declarations (CI steps, `AGENTS.md`) win over every default below.

## Which build tool

| Evidence | Command form |
| --- | --- |
| `gradlew` | `./gradlew <task>` (the wrapper pins the Gradle version) |
| `build.gradle` or `build.gradle.kts` without a wrapper | `gradle <task>` |
| `mvnw` | `./mvnw <phase>` |
| `pom.xml` without a wrapper | `mvn <phase>` |

## Evidence and defaults

| Kind | Gradle | Maven | Use it only when |
| --- | --- | --- | --- |
| format | `spotlessCheck` | `spotless:check` | the Spotless plugin is applied |
| lint | `ktlintCheck`, `detekt`, `checkstyleMain` | `checkstyle:check` | that plugin is applied |
| test | `test` | `test` | always |
| all checks | `check` | `verify` | always; it runs the tests and every check plugin bound to it |
| build | `build` | `package` | always |

Run with `--offline` when dependencies are already cached and the run must not reach the network.

## Dependency existence

- Maven Central: `https://repo1.maven.org/maven2/<group path>/<artifact>/maven-metadata.xml` answers 404 for an unknown group and artifact (`<group path>` is the group with `.` replaced by `/`). A project with its own repositories names them in `settings.gradle`, `build.gradle` or `pom.xml`; check there.
- The coordinate is `group:artifact:version`; check all three.

## Traps

- Without the wrapper, the machine's Gradle or Maven version may differ from the project's.
- `@Disabled` (JUnit 5), `@Ignore` (JUnit 4) and `-DskipTests` skip tests; a new one is a finding.
- A Gradle daemon keeps state between runs; when a result looks stale, rerun with `--no-daemon` and say so.
- Kotlin and Java share one build: a Kotlin change can break Java callers, so run the whole `check`.
