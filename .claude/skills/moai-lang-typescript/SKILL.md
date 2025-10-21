---

name: moai-lang-typescript
description: TypeScript best practices with Vitest, Biome, strict typing, and npm/pnpm package management. Use when writing or reviewing TypeScript code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# TypeScript Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | TypeScript code discussions, framework guidance, or file extensions such as .ts/.tsx. |
| Tier | 3 |

## What it does

Provides TypeScript-specific expertise for TDD development, including Vitest testing, Biome linting/formatting, strict type checking, and modern npm/pnpm package management.

## When to use

- Engages when the conversation references TypeScript work, frameworks, or files like .ts/.tsx.
- "Writing TypeScript tests", "How to use Vitest", "Type safety"
- Automatically invoked when working with TypeScript projects
- TypeScript SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **Vitest**: Fast unit testing (Jest-compatible API)
- **@testing-library**: Component testing for React/Vue
- Test coverage ≥85% with c8/istanbul

**Type Safety**:
- **strict: true** in tsconfig.json
- **noImplicitAny**, **strictNullChecks**, **strictFunctionTypes**
- Interface definitions, Generics, Type guards

**Code Quality**:
- **Biome**: Fast linter + formatter (replaces ESLint + Prettier)
- Type-safe configurations
- Import organization, unused variable detection

**Package Management**:
- **pnpm**: Fast, disk-efficient package manager (preferred)
- **npm**: Fallback option
- `package.json` + `tsconfig.json` configuration

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Prefer interfaces over types for public APIs
- Use const assertions for literal types
- Avoid `any`, prefer `unknown` or proper types

## Examples
```bash
npm run lint && npm test
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
- Microsoft. "TypeScript Handbook." https://www.typescriptlang.org/docs/ (accessed 2025-03-29).
- OpenJS Foundation. "ESLint User Guide." https://eslint.org/docs/latest/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (TypeScript-specific review)
- alfred-refactoring-coach (type-safe refactoring)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
