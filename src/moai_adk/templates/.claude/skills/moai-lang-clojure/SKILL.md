---

name: moai-lang-clojure
description: Clojure best practices with clojure.test, Leiningen, and immutable data structures. Use when writing or reviewing Clojure code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Clojure Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Clojure code discussions, framework guidance, or file extensions such as .clj/.cljc. |
| Tier | 3 |

## What it does

Provides Clojure-specific expertise for TDD development, including clojure.test framework, Leiningen build tool, and immutable data structures with functional programming.

## When to use

- Engages when the conversation references Clojure work, frameworks, or files like .clj/.cljc.
- "Writing Clojure tests", "How to use clojure.test", "Immutable data structures"
- Automatically invoked when working with Clojure projects
- Clojure SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **clojure.test**: Built-in testing library
- **midje**: BDD-style testing
- **test.check**: Property-based testing
- Test coverage with cloverage

**Build Tools**:
- **Leiningen**: Project automation, dependency management
- **deps.edn**: Official dependency tool
- **Boot**: Alternative build tool

**Code Quality**:
- **clj-kondo**: Linter for Clojure
- **cljfmt**: Code formatting
- **eastwood**: Additional linting

**Clojure Patterns**:
- **Immutable data structures**: Persistent collections
- **Pure functions**: Functional core, imperative shell
- **Threading macros**: -> and ->> for readability
- **Lazy sequences**: Infinite data processing
- **Transducers**: Composable transformations

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Prefer let bindings over def
- Use namespaces for organization
- Destructuring for data access
- Avoid mutable state

## Examples
```bash
lein test && clj-kondo --lint src
```

## Inputs
- 언어별 소스 디렉터리(e.g. `src/`, `app/`).
- 언어별 빌드/테스트 설정 파일(예: `package.json`, `pyproject.toml`, `go.mod`).
- 관련 테스트 스위트 및 샘플 데이터.

## Outputs
- 선택된 언어에 맞춘 테스트/린트 실행 계획.
- 주요 언어 관용구와 리뷰 체크포인트 목록.

## Failure Modes
- 언어 런타임이나 패키지 매니저가 설치되지 않았을 때.
- 다중 언어 프로젝트에서 주 언어를 판별하지 못했을 때.

## Dependencies
- Read/Grep 도구로 프로젝트 파일 접근이 필요합니다.
- `Skill("moai-foundation-langs")`와 함께 사용하면 교차 언어 규약 공유가 용이합니다.

## References
- Clojure.org. "Clojure Documentation." https://clojure.org/guides/getting_started (accessed 2025-03-29).
- clj-kondo. "User Guide." https://github.com/clj-kondo/clj-kondo/blob/master/doc/usage.md (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Clojure-specific review)
- alfred-refactoring-coach (functional refactoring)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
