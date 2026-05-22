---
title: /moai gate
weight: 15
draft: false
---

lint、format、type-check、test を並列実行する軽量な pre-commit 品質ゲートコマンドです。30 秒以内の完了を目標に設計され、すべての commit 直前の高速チェックに利用します。

{{< callout type="info" >}}
**スラッシュコマンド**: Claude Code で `/moai gate` を入力するとこのコマンドを直接実行できます。
{{< /callout >}}

## 概要

`/moai gate` は commit 直前に使用する軽量品質ゲートです。lint + format check + type-check + test の 4 種類の検証を並列に実行し、30 秒以内に完了するよう設計されています。コードレビュー(`/moai review`)や sync Phase 0.5 のような重い分析は行わず、通常の作業フローにおける高速な安全ネットの役割を果たします。

## コマンド形式

```bash
/moai gate [--fix] [--staged] [--file PATH]
```

- 引数が空の場合はプロジェクト全体に対して 4 種類の検証を並列実行します。
- `--mode pipeline` 引数は `MODE_PIPELINE_ONLY_UTILITY` エラーを発生させます(`/moai gate` は multi-agent クラスではありません)。

## オプション

### `--fix`

lint / format の自動修正可能項目を直接修正します。デフォルトはレポート出力のみ。

- **推奨タイミング**: 新規コード作成直後に一度実行してスタイル違反を事前整理。
- **注意**: 自動修正後は必ず `git diff` で変更内容を確認。

### `--staged`

`git diff --staged` で識別された stage 済みファイルのみを検証します。

- 大規模モノレポで commit 直前の検証時間を更に短縮できます。

### `--file PATH`

指定した単一ファイル(または glob)のみを検証します。デバッグ時に有用です。

## 並列実行される 4 段階

`/moai gate` は以下の 4 種類の検証を**同時に**実行します(完了時間は最も時間のかかる検証で決定されます)。

| Check | 役割 | 主要ツール(自動検出) |
|-------|------|----------------------|
| Lint | スタイル違反、未使用 import、dead code を報告 | `golangci-lint`、`ruff`、`eslint`、`clippy`、`rubocop`、`mvn compile`、`php-cs-fixer`、`ktlint`、`swiftlint`、`dotnet build`、`cmake --build`、`mix credo`、`lintr`、`dart analyze`、`sbt compile` |
| Format check | フォーマット違反検出(自動修正は `--fix` が必要) | `gofmt`、`ruff format --check`、`prettier --check`、`cargo fmt --check`、`rubocop`、`php-cs-fixer`、`ktlint`、`swift-format`、`dotnet format --verify-no-changes`、`clang-format`、`mix format --check-formatted` |
| Type check | 静的型検証 | `go vet`、`mypy`、`tsc --noEmit`、`cargo check`、`phpstan`、`dotnet build`、`cmake` |
| Test | 単体・統合テスト | `go test -race`、`pytest`、`vitest`/`jest`、`cargo test`、`bundle exec rspec`、`mvn test`、`phpunit`、`gradle test`、`swift test`、`ctest`、`mix test`、`testthat`、`flutter test`、`sbt test` |

## 16-language 自動検出

`/moai gate` はプロジェクトルートの indicator ファイルを優先順位順に確認し、最初にマッチした toolchain を使用します。

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

マッチする indicator がない場合は言語検査をスキップし、`unknown language` として報告します。

## /moai gate vs /moai review vs sync Phase 0.5

| Workflow | 範囲 | 速度 | 使用タイミング |
|----------|------|------|----------------|
| `/moai gate` | lint + format + type-check + test | 速い (<30 秒) | すべての commit 直前 |
| `/moai review` | 4 視点の深層コードレビュー | 中 (2-5 分) | PR 直前、デザインレビュー |
| sync Phase 0.5 | 全体品質 + コードレビュー + coverage | 遅い (5-10 分) | `/moai sync` パイプラインの一部 |

## 使用例

```bash
# 1) commit 直前の高速検証
/moai gate

# 2) lint/format を自動修正後に再検証
/moai gate --fix

# 3) stage 済みファイルのみ検証(大規模 monorepo 推奨)
/moai gate --staged

# 4) 特定ファイルのみ検証
/moai gate --file internal/cli/run.go
```

## 関連資料

- [`.claude/skills/moai/workflows/gate.md`](https://github.com/modu-ai/moai-adk) — workflow body SSOT
- [`/moai review`](/ja/quality-commands/moai-review) — 4 視点のコードレビュー
- [`/moai sync`](/ja/workflow-commands/moai-sync) — sync Phase 0.5 品質検証を含む
- [`/moai fix`](/ja/utility-commands/moai-fix) — 自動修正パイプライン
