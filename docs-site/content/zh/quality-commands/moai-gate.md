---
title: /moai gate
weight: 15
draft: false
---

并行执行 lint、format、type-check、test 的轻量 pre-commit 质量门禁命令。设计为在 30 秒内完成,用于每次 commit 之前的快速验证。

{{< callout type="info" >}}
**斜杠命令**: 在 Claude Code 中输入 `/moai gate` 即可直接执行此命令。
{{< /callout >}}

## 概述

`/moai gate` 是在 commit 之前使用的轻量质量门禁。它并行执行 lint + format check + type-check + test 四项验证,设计为在 30 秒内完成。它不执行代码审查 (`/moai review`) 或 sync Phase 0.5 那样的重量级分析,在日常工作流中扮演 commit 前使用的快速安全网角色。

"与任务大小相称的验证深度"这一 v3 令牌经济学原则,正是这条命令的设计理由 — 与其每次提交都运转 5-10 分钟的完整质量流水线,不如用 30 秒的门禁在提交前拦下大多数缺陷。

## 命令格式

```bash
/moai gate [--fix] [--staged] [--file PATH]
```

- 参数为空时,对整个项目并行执行 4 项验证。
- `--mode pipeline` 参数会引发 `MODE_PIPELINE_ONLY_UTILITY` 错误(`/moai gate` 不是 multi-agent class)。

## 选项

### `--fix`

直接修复 lint / format 中可自动修复的项目。默认行为仅输出报告,不做修复。

- **推荐时机**: 写完新代码后使用一次,预先清理风格缺陷。
- **注意**: 自动修复后务必用 `git diff` 审查变更内容。

### `--staged`

仅验证 `git diff --staged` 识别出的已 stage 文件。

- 在大型 monorepo 中可进一步缩短 commit 前的验证时间。

### `--file PATH`

仅验证指定的单个文件(或 glob)。在调试时非常有用。

## 并行执行的 4 个步骤

`/moai gate` **同时** 执行以下四项验证(完成时间取决于耗时最长的验证)。

| Check | 角色 | 主要工具(自动检测) |
|-------|------|-----------------------|
| Lint | 报告风格违规、未使用 import、dead code | `golangci-lint`, `ruff`, `eslint`, `clippy`, `rubocop`, `mvn compile`, `php-cs-fixer`, `ktlint`, `swiftlint`, `dotnet build`, `cmake --build`, `mix credo`, `lintr`, `dart analyze`, `sbt compile` |
| Format check | 检测格式化违规(自动修复需 `--fix`) | `gofmt`, `ruff format --check`, `prettier --check`, `cargo fmt --check`, `rubocop`, `php-cs-fixer`, `ktlint`, `swift-format`, `dotnet format --verify-no-changes`, `clang-format`, `mix format --check-formatted` |
| Type check | 静态类型验证 | `go vet`, `mypy`, `tsc --noEmit`, `cargo check`, `phpstan`, `dotnet build`, `cmake` |
| Test | 运行单元/集成测试 | `go test -race`, `pytest`, `vitest`/`jest`, `cargo test`, `bundle exec rspec`, `mvn test`, `phpunit`, `gradle test`, `swift test`, `ctest`, `mix test`, `testthat`, `flutter test`, `sbt test` |

## 16 语言自动检测

`/moai gate` 按优先级顺序检查项目根目录的 indicator 文件,并使用第一个匹配对应的 toolchain。16 种语言全部得到平等支持。

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

若没有匹配的 indicator,则跳过语言检查并报告为 `unknown language`。

## /moai gate vs /moai review vs sync Phase 0.5

验证工具分为三个深度层级,按情况选择合适的深度是原则。

| Workflow | 范围 | 速度 | 使用时机 |
|----------|------|------|-----------|
| `/moai gate` | lint + format + type-check + test | 快 (<30 秒) | 每次 commit 之前 |
| `/moai review` | 4 视角深度代码审查 | 中等 (2-5 分钟) | PR 之前、设计审查 |
| sync Phase 0.5 | 完整质量 + 代码审查 + coverage | 慢 (5-10 分钟) | `/moai sync` 流水线的一部分 |

## 使用示例

```bash
# 1) commit 前快速验证
/moai gate

# 2) 自动修复 lint/format 后再验证
/moai gate --fix

# 3) 仅验证已 stage 文件(大型 monorepo 推荐)
/moai gate --staged

# 4) 仅验证特定文件
/moai gate --file internal/cli/run.go
```

## 相关资料

- [`.claude/skills/moai/workflows/gate.md`](https://github.com/modu-ai/moai-adk) — workflow body SSOT
- [`/moai review`](/zh/quality-commands/moai-review) — 4 视角代码审查
- [`/moai sync`](/zh/workflow-commands/moai-sync) — 包含 sync Phase 0.5 质量验证
- [`/moai fix`](/zh/utility-commands/moai-fix) — 自动修复流水线
