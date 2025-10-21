---

name: moai-lang-javascript
description: JavaScript best practices with Jest, ESLint, Prettier, and npm package management. Use when writing or reviewing JavaScript code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# JavaScript Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | JavaScript code discussions, framework guidance, or file extensions such as .js. |
| Tier | 3 |

## What it does

Provides JavaScript-specific expertise for TDD development, including Jest testing, ESLint linting, Prettier formatting, and npm package management.

## When to use

- Engages when the conversation references JavaScript work, frameworks, or files like .js.
- "Writing JavaScript tests", "How to use Jest", "ES6+ grammar"
- Automatically invoked when working with JavaScript projects
- JavaScript SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **Jest**: Unit testing with mocking, snapshots
- **@testing-library**: DOM/React testing
- Test coverage ≥85% enforcement

**Code Quality**:
- **ESLint**: JavaScript linting with recommended rules
- **Prettier**: Code formatting (opinionated)
- **JSDoc**: Type hints via comments (for type safety)

**Package Management**:
- **npm**: Standard package manager
- **package.json** for dependencies and scripts
- Semantic versioning

**Modern JavaScript**:
- ES6+ features (arrow functions, destructuring, spread/rest)
- Async/await over callbacks
- Module imports (ESM) over CommonJS

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Prefer `const` over `let`, avoid `var`
- Guard clauses for early returns
- Meaningful names, avoid abbreviations

## Examples
```bash
npm run test && npm run lint
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
- MDN Web Docs. "JavaScript Guide." https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide (accessed 2025-03-29).
- Jest. "Getting Started." https://jestjs.io/docs/getting-started (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (JavaScript-specific review)
- alfred-debugger-pro (JavaScript debugging)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
