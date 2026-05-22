---
title: /moai gate
weight: 15
draft: false
---

lint, format, type-check, test를 병렬로 실행하는 경량 pre-commit 품질 게이트 명령어입니다. 30 초 이내에 완료되도록 설계되어 모든 commit 직전 빠른 검증에 사용합니다.

{{< callout type="info" >}}
**슬래시 커맨드**: Claude Code에서 `/moai gate`를 입력하면 이 명령어를 바로 실행할 수 있습니다.
{{< /callout >}}

## 개요

`/moai gate`는 commit 직전 사용하는 경량 품질 게이트입니다. lint + format check + type-check + test 네 가지 검증을 병렬로 실행해 30 초 이내 완료되도록 설계되었습니다. 코드 리뷰(`/moai review`)나 sync Phase 0.5처럼 무거운 분석은 수행하지 않으며, 평소 작업 흐름에서 commit 전 사용하는 빠른 안전망 역할을 합니다.

## 명령어 형식

```bash
/moai gate [--fix] [--staged] [--file PATH]
```

- 인자가 비어 있으면 프로젝트 전체에 대해 4 가지 검증을 병렬 실행합니다.
- `--mode pipeline` 인자는 `MODE_PIPELINE_ONLY_UTILITY` 오류를 유발합니다 (`/moai gate`는 multi-agent class 가 아닙니다).

## 옵션

### `--fix`

lint / format 의 자동 수정 가능 항목을 직접 수정합니다. 기본 동작은 리포트만 출력하며 수정은 하지 않습니다.

- **권장 시점**: 새 코드 작성 직후 한 번 사용해 스타일 결함을 사전 정리.
- **주의**: 자동 수정 후 반드시 `git diff`로 변경 내용을 검토.

### `--staged`

`git diff --staged`로 식별된 stage 된 파일만 검증합니다.

- 대규모 모노레포에서 commit 직전 검증 시간을 더 단축할 수 있습니다.

### `--file PATH`

지정한 단일 파일(또는 glob) 만 검증합니다. 디버깅 시점에 유용합니다.

## 병렬 실행되는 4 단계

`/moai gate`는 다음 네 가지 검증을 **동시에** 실행합니다(완료 시간은 가장 오래 걸리는 검증에 의해 결정).

| Check | 역할 | 주요 도구 (자동 감지) |
|-------|------|-----------------------|
| Lint | 스타일 위반, 미사용 import, dead code 보고 | `golangci-lint`, `ruff`, `eslint`, `clippy`, `rubocop`, `mvn compile`, `php-cs-fixer`, `ktlint`, `swiftlint`, `dotnet build`, `cmake --build`, `mix credo`, `lintr`, `dart analyze`, `sbt compile` |
| Format check | 포맷팅 위반 검출 (자동 수정은 `--fix` 필요) | `gofmt`, `ruff format --check`, `prettier --check`, `cargo fmt --check`, `rubocop`, `php-cs-fixer`, `ktlint`, `swift-format`, `dotnet format --verify-no-changes`, `clang-format`, `mix format --check-formatted` |
| Type check | 정적 타입 검증 | `go vet`, `mypy`, `tsc --noEmit`, `cargo check`, `phpstan`, `dotnet build`, `cmake` |
| Test | 단위/통합 테스트 실행 | `go test -race`, `pytest`, `vitest`/`jest`, `cargo test`, `bundle exec rspec`, `mvn test`, `phpunit`, `gradle test`, `swift test`, `ctest`, `mix test`, `testthat`, `flutter test`, `sbt test` |

## 16-language 자동 감지

`/moai gate`는 프로젝트 루트의 indicator 파일을 우선순위 순서대로 확인해 첫 번째 매칭에 해당하는 toolchain을 사용합니다.

1. Go: `go.mod`
2. Python: `pyproject.toml`
3. TypeScript: `tsconfig.json`
4. JavaScript: `package.json`
5. Rust: `Cargo.toml`
6. Ruby: `Gemfile`
7. Java: `pom.xml`
8. PHP: `composer.json`
9. Kotlin: `build.gradle.kts`
10. Swift: `Package.swift`
11. C#: `.csproj`
12. C++: `CMakeLists.txt`
13. Elixir: `mix.exs`
14. R: `DESCRIPTION`
15. Flutter: `pubspec.yaml`
16. Scala: `build.sbt`

매칭되는 indicator 가 없으면 언어 검사를 건너뛰고 `unknown language` 로 보고합니다.

## /moai gate vs /moai review vs sync Phase 0.5

| Workflow | 범위 | 속도 | 사용 시점 |
|----------|------|------|-----------|
| `/moai gate` | lint + format + type-check + test | 빠름 (<30 초) | 모든 commit 직전 |
| `/moai review` | 4-시점 심층 코드 리뷰 | 중간 (2-5 분) | PR 직전, 디자인 리뷰 |
| sync Phase 0.5 | 전체 품질 + 코드 리뷰 + coverage | 느림 (5-10 분) | `/moai sync` 파이프라인의 일부 |

## 사용 예시

```bash
# 1) commit 직전 빠른 검증
/moai gate

# 2) lint/format 자동 수정 후 재검증
/moai gate --fix

# 3) stage된 파일만 검증 (대규모 monorepo 권장)
/moai gate --staged

# 4) 특정 파일만 검증
/moai gate --file internal/cli/run.go
```

## 관련 자료

- [`.claude/skills/moai/workflows/gate.md`](https://github.com/modu-ai/moai-adk) — workflow body SSOT
- [`/moai review`](/ko/quality-commands/moai-review) — 4-시점 코드 리뷰
- [`/moai sync`](/ko/workflow-commands/moai-sync) — sync Phase 0.5 품질 검증 포함
- [`/moai fix`](/ko/utility-commands/moai-fix) — 자동 수정 파이프라인
