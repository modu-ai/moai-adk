---
title: moai init 初期化
weight: 5
draft: false
---

`moai init` は現在のディレクトリまたは新規フォルダに MoAI プロジェクトを初期化します。Claude Code 連携に必要な `.claude/`、`.moai/` 構造と設定を配置し、必要に応じて対話式ウィザードでプロジェクトモード・言語・品質ゲートを設定します。

## 使い方

```bash
moai init [project-name]
```

| パターン | 動作 |
|----------|------|
| `moai init <name>` | `./<name>/` フォルダを作成しその中に初期化 |
| `moai init .` | 現在のディレクトリに初期化 |
| `moai init` | 現在のディレクトリに初期化 (`moai init .` と同じ) |

引数は最大 1 個まで受け取ります。

## 主なフラグ

### 配置範囲

| フラグ | 説明 |
|--------|------|
| `--all` | カタログ全体を配置 (core + 選択パック + ハーネス生成物)。デフォルトは core-only slim モード |
| `--force` | 既存プロジェクトの再初期化 (現在の `.moai/` をバックアップ) |
| `--no-hooks` | git フックのインストールを省略 |

### プロジェクトデフォルト

| フラグ | 説明 |
|--------|------|
| `--root <dir>` | プロジェクトルート (デフォルト: 現在のディレクトリ) |
| `--name <name>` | プロジェクト名 (デフォルト: ディレクトリ名) |
| `--language <lang>` | 主なプログラミング言語 |
| `--framework <name>` | フレームワーク (デフォルト: 自動検出または `none`) |
| `--mode <ddd\|tdd>` | 開発方法論 (デフォルト: tdd) |
| `--non-interactive` | 対話式ウィザードを省略 — フラグとデフォルトのみ使用 |

### ウィザードステップ

| フラグ | 説明 |
|--------|------|
| `--standard` | Phase 1 の質問を提示 (プロジェクトモード、ハーネスプロファイル、LSP、品質ゲート、デザイン) |
| `--advanced` | Phase 1 + Phase 2 の質問を提示 (`--standard` を含む) |
| `--project-mode <personal\|team>` | プロジェクトモード (デフォルト: personal) |
| `--harness-profile <name>` | ハーネス評価プロファイル: default, strict, lenient, frontend |
| `--enable-lsp` | LSP 統合を有効化 (デフォルト: false) |
| `--enforce-quality` | 品質ゲートを強制 (デフォルト: true) |
| `--enable-design` | デザインワークフローを有効化 (デフォルト: true) |

### Git / モデルポリシー

| フラグ | 説明 |
|--------|------|
| `--git-mode <manual\|personal\|team>` | Git ワークフローモード (デフォルト: manual) |
| `--git-provider <github\|gitlab>` | Git プロバイダ |
| `--github-username <name>` | GitHub ユーザー名 (personal/team モードで必須) |
| `--profile <max\|medium\|low>` | モデル+effort プロファイル — `llm.yaml` の `profile` に保存 (プロファイルマトリクス列の選択) |
| `--model-policy <max\|medium\|low>` | legacy パフォーマンスティア — `llm.yaml` の `performance_tier` に保存 (`profile` 不在時にエイリアスとして読み込み) |

## 例

```bash
# 新規フォルダに初期化
moai init my-app

# 現在のディレクトリに初期化
moai init .

# 方法論を指定
moai init --mode tdd

# カタログ全体を配置 (slim モードを回避)
moai init --all

# 非対話 (CI など)
moai init . --non-interactive --language go
```

## 関連コマンド

| コマンド | 説明 |
|----------|------|
| `moai update` | 初期化済みプロジェクトのテンプレート同期 |
| `moai status` | 初期化状態の確認 |
| `moai doctor` | 初期化後の環境検証 |

## 関連ドキュメント

- [プロジェクト状態](/ja/cli-reference/status)
- [アップデート](/ja/cli-reference/update)
- [CLI 概要](/ja/getting-started/cli)
