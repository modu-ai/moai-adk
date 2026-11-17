# Language: Swift

When to read: `internal/quality` sent you here because the project's detected languages include `swift`. The project's own declarations (CI steps, `Makefile`, `AGENTS.md`) win over every default below.

## Which build tool

| Evidence | Build | Test |
| --- | --- | --- |
| `Package.swift` | `swift build` | `swift test` |
| `*.xcodeproj` or `*.xcworkspace` | `xcodebuild build -scheme <scheme>` | `xcodebuild test -scheme <scheme> -destination <destination>` |

Take the scheme and destination from the CI file or `xcodebuild -list`; never pick a simulator the project does not name without saying so. `xcodebuild` needs macOS with Xcode: elsewhere it is `blocked`.

## Evidence and defaults

| Kind | Default command | Use it only when |
| --- | --- | --- |
| format | `swift format lint --recursive .` | `.swift-format` exists |
| lint | `swiftlint` | `.swiftlint.yml` exists |
| type check | covered by the build | |

## Dependency existence

- Swift packages are git repositories: `git ls-remote --tags <url>` shows that the repository and the version tag exist. Check that the URL is the one the package's documentation names.
- Add the dependency to `Package.swift` (or through Xcode) and show the `Package.resolved` entry the resolve wrote.

## Traps

- `swift test --filter` that matches nothing runs no test: that is `not_run`.
- `XCTSkip` and disabled tests in a scheme hide tests; a new one is a finding.
- Derived data from another branch can hide a change; when a result looks stale, build into a fresh `-derivedDataPath` and say so.
