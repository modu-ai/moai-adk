---
title: /moai gate
weight: 15
draft: false
---

并行执行 lint、format、type-check、test 的轻量级 pre-commit 质量门命令。设计目标是 30 秒内完成,适用于每次 commit 之前的快速检查。

{{< callout type="info" >}}
**斜杠命令**: 在 Claude Code 中输入 `/moai gate` 即可直接执行该命令。
{{< /callout >}}

## 概述

`/moai gate` 是 commit 之前使用的轻量级质量门。它并行运行 lint + format check + type-check + test 四种检查,目标在 30 秒内完成。它有意比 `/moai review` 与 sync Phase 0.5 更轻量,专注于发现机械性缺陷,而非深入的设计评审。

## 命令格式

```bash
/moai gate [--fix] [--staged] [--file PATH]
```

- 参数为空时对整个项目执行 4 项并行检查。
- `--mode pipeline` 参数会触发 `MODE_PIPELINE_ONLY_UTILITY` 错误(`/moai gate` 不属于 multi-agent 类别)。

## 选项

### `--fix`

直接修复 lint / format 中可自动修复的项目。默认行为只输出报告,不做修改。

- **推荐时机**: 编写新代码后一次性运行,清理样式缺陷。
- **注意**: 使用 `--fix` 后务必通过 `git diff` 复核变更。

### `--staged`

仅检查 `git diff --staged` 识别出的暂存文件。

- 在大型 monorepo 中可以进一步缩短 commit 前的检查时间。

### `--file PATH`

仅检查指定的单个文件(或 glob)。调试时非常有用。

## 并行执行的 4 项检查

`/moai gate` **同时**执行以下 4 项检查(整体耗时取决于最慢的那项)。

| 检查 | 作用 | 主要工具(自动检测) |
|------|------|----------------------|
| Lint | 报告样式违规、未使用 import、dead code | `golangci-lint`、`ruff`、`eslint`、`clippy`、`rubocop`、`mvn compile`、`php-cs-fixer`、`ktlint`、`swiftlint`、`dotnet build`、`cmake --build`、`mix credo`、`lintr`、`dart analyze`、`sbt compile` |
| Format check | 检测格式违规(自动修复需要 `--fix`) | `gofmt`、`ruff format --check`、`prettier --check`、`cargo fmt --check`、`rubocop`、`php-cs-fixer`、`ktlint`、`swift-format`、`dotnet format --verify-no-changes`、`clang-format`、`mix format --check-formatted` |
| Type check | 静态类型校验 | `go vet`、`mypy`、`tsc --noEmit`、`cargo check`、`phpstan`、`dotnet build`、`cmake` |
| Test | 单元 / 集成测试 | `go test -race`、`pytest`、`vitest`/`jest`、`cargo test`、`bundle exec rspec`、`mvn test`、`phpunit`、`gradle test`、`swift test`、`ctest`、`mix test`、`testthat`、`flutter test`、`sbt test` |

## 16-language 自动检测

`/moai gate` 按优先级顺序检测项目根目录的 indicator 文件,首个匹配决定使用的 toolchain。

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

若没有任何 indicator 匹配,会跳过特定语言检查并报告 `unknown language`。

## /moai gate vs /moai review vs sync Phase 0.5

| Workflow | 范围 | 速度 | 使用时机 |
|----------|------|------|---------|
| `/moai gate` | lint + format + type-check + test | 快速 (<30 秒) | 每次 commit 之前 |
| `/moai review` | 4 视角深度代码评审 | 中等 (2-5 分钟) | PR 之前、设计评审 |
| sync Phase 0.5 | 完整质量 + 代码评审 + coverage | 较慢 (5-10 分钟) | `/moai sync` 流水线的一部分 |

## 使用示例

```bash
# 1) commit 前快速验证
/moai gate

# 2) 自动修复 lint/format 后再次检查
/moai gate --fix

# 3) 仅检查暂存文件(monorepo 推荐)
/moai gate --staged

# 4) 仅检查单个文件
/moai gate --file internal/cli/run.go
```

## 相关资料

- [`.claude/skills/moai/workflows/gate.md`](https://github.com/modu-ai/moai-adk) — workflow body SSOT
- [`/moai review`](/zh/quality-commands/moai-review) — 4 视角代码评审
- [`/moai sync`](/zh/workflow-commands/moai-sync) — 包含 sync Phase 0.5 质量检查
- [`/moai fix`](/zh/utility-commands/moai-fix) — 自动修复流水线
