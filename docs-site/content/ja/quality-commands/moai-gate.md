---
title: /moai gate
weight: 15
draft: false
---

lint、format、type-check、test を並列実行する軽量 pre-commit 品質ゲートコマンドです。30 秒以内に完了するよう設計されており、すべての commit 直前の素早い検証に使用します。

{{< callout type="info" >}}
**スラッシュコマンド**: Claude Code で `/moai gate` と入力すると、このコマンドをすぐに実行できます。
{{< /callout >}}

## 概要

`/moai gate` は commit 直前に使用する軽量品質ゲートです。lint + format check + type-check + test の 4 つの検証を並列実行し、30 秒以内に完了するよう設計されています。コードレビュー (`/moai review`) や sync Phase 0.5 のような重い分析は行わず、普段のワークフローで commit 前に使う素早いセーフティネットの役割を果たします。

「作業サイズに合った検証の深さ」という v3 トークノミクス原則が、このコマンドの設計理由です — コミットのたびに 5-10 分の完全な品質パイプラインを回す代わりに、30 秒のゲートでほとんどの欠陥をコミット前にふるい落とします。

## コマンド形式

```bash
/moai gate [--fix] [--staged] [--file PATH]
```

- 引数が空の場合、プロジェクト全体に対して 4 つの検証を並列実行します。
- `--mode pipeline` 引数は `MODE_PIPELINE_ONLY_UTILITY` エラーを引き起こします (`/moai gate` は multi-agent class ではありません)。

## オプション

### `--fix`

lint / format の自動修正可能な項目を直接修正します。デフォルトの動作はレポート出力のみで、修正は行いません。

- **推奨タイミング**: 新しいコードを書いた直後に一度使用し、スタイル上の欠陥を事前に整理。
- **注意**: 自動修正後は必ず `git diff` で変更内容をレビュー。

### `--staged`

`git diff --staged` で識別された stage 済みファイルのみを検証します。

- 大規模モノレポで commit 直前の検証時間をさらに短縮できます。

### `--file PATH`

指定した単一ファイル (または glob) のみを検証します。デバッグ時に便利です。

## 並列実行される 4 ステップ

`/moai gate` は次の 4 つの検証を **同時に** 実行します (完了時間は最も時間のかかる検証によって決まります)。

| Check | 役割 | 主なツール (自動検出) |
|-------|------|-----------------------|
| Lint | スタイル違反、未使用 import、dead code の報告 | `golangci-lint`, `ruff`, `eslint`, `clippy`, `rubocop`, `mvn compile`, `php-cs-fixer`, `ktlint`, `swiftlint`, `dotnet build`, `cmake --build`, `mix credo`, `lintr`, `dart analyze`, `sbt compile` |
| Format check | フォーマット違反の検出 (自動修正には `--fix` が必要) | `gofmt`, `ruff format --check`, `prettier --check`, `cargo fmt --check`, `rubocop`, `php-cs-fixer`, `ktlint`, `swift-format`, `dotnet format --verify-no-changes`, `clang-format`, `mix format --check-formatted` |
| Type check | 静的型検証 | `go vet`, `mypy`, `tsc --noEmit`, `cargo check`, `phpstan`, `dotnet build`, `cmake` |
| Test | ユニット/統合テストの実行 | `go test -race`, `pytest`, `vitest`/`jest`, `cargo test`, `bundle exec rspec`, `mvn test`, `phpunit`, `gradle test`, `swift test`, `ctest`, `mix test`, `testthat`, `flutter test`, `sbt test` |

## 16-language 自動検出

`/moai gate` はプロジェクトルートの indicator ファイルを優先順位の順に確認し、最初にマッチした toolchain を使用します。16 言語はすべて同等にサポートされます。

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

マッチする indicator がない場合、言語チェックをスキップして `unknown language` と報告します。

## /moai gate vs /moai review vs sync Phase 0.5

検証ツールは 3 段階の深さに分かれており、状況に合った深さを選ぶことが原則です。

| Workflow | 範囲 | 速度 | 使用タイミング |
|----------|------|------|-----------|
| `/moai gate` | lint + format + type-check + test | 速い (<30 秒) | すべての commit 直前 |
| `/moai review` | 4 視点の詳細コードレビュー | 中間 (2-5 分) | PR 直前、デザインレビュー |
| sync Phase 0.5 | 全体品質 + コードレビュー + coverage | 遅い (5-10 分) | `/moai sync` パイプラインの一部 |

## 使用例

```bash
# 1) commit 直前の素早い検証
/moai gate

# 2) lint/format 自動修正後に再検証
/moai gate --fix

# 3) stage 済みファイルのみ検証 (大規模 monorepo 推奨)
/moai gate --staged

# 4) 特定ファイルのみ検証
/moai gate --file internal/cli/run.go
```

## 関連資料

- [`.claude/skills/moai/workflows/gate.md`](https://github.com/modu-ai/moai-adk) — workflow body SSOT
- [`/moai review`](/ja/quality-commands/moai-review) — 4 視点コードレビュー
- [`/moai sync`](/ja/workflow-commands/moai-sync) — sync Phase 0.5 品質検証を含む
- [`/moai fix`](/ja/utility-commands/moai-fix) — 自動修正パイプライン
