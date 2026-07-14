---
title: CLI リファレンス
weight: 90
draft: false
---

ターミナルで実行する `moai` (Go バイナリ) のすべてのコマンドとフラグを参照します。Claude Code の対話画面で入力する `/moai` (スラッシュサブコマンド) とは完全に別のツールです — このページはターミナル CLI のみを扱います。

## コマンドツリー

```bash
moai --help
```

`moai` CLI は 3 つのグループに分かれます。

| グループ | コマンド | 説明 |
|------|--------|------|
| **Launch** | `moai cc` · `moai cg` · `moai glm` | Claude Code セッションの開始 (バックエンド選択) |
| **Project** | `moai init` · `moai update` · `moai doctor` · `moai status` | プロジェクトの初期化、アップデート、診断、状態照会 |
| **Tools** | `moai profile` · `moai inventory` · `moai hook` · `moai worktree` · `moai spec` · `moai harness` · ... | 設定、インベントリ、フック、ワークツリーなどのツール |

`moai version` で現在インストールされているバージョンを確認します。

```bash
moai version
# 出力例: moai <バージョン> (commit: <ハッシュ>, built: <ビルド日付>)
```

---

## moai init

プロジェクトを初期化します。対話型ウィザードが言語、Git 自動化、モデルポリシー、ハーネスプロファイルなどを設定します。

```bash
moai init [project-name] [OPTIONS]
```

### フラグ

| フラグ | 説明 |
|--------|------|
| `--non-interactive` | 対話型ウィザードをスキップ (フラグとデフォルト値を使用) |
| `--force` | 既存プロジェクトの強制再初期化 (現在の `.moai/` をバックアップ) |
| `--no-hooks` | Git フックのインストールをスキップ |
| `--all` | カタログの全項目を配布 (core + optional packs + harness-generated) |
| `--standard` | Phase 1 質問を表示 (project mode, harness profile, LSP, quality gates, design) |
| `--advanced` | Phase 1 + Phase 2 質問を表示 (`--standard` を含む; Phase 2 は前提条件を満たす場合のみ) |
| `--project-mode <personal\|team>` | プロジェクトモード (デフォルト値: personal) |
| `--harness-profile <profile>` | ハーネス評価者プロファイル (default, strict, lenient, frontend) |
| `--enable-lsp` | LSP 連携の有効化 (デフォルト値: false) |
| `--enforce-quality` | 品質ゲートの強制 (デフォルト値: true) |
| `--enable-design` | デザインワークフローの有効化 (デフォルト値: true) |
| `--model-policy <max\|medium\|low>` | パフォーマンスティア — `llm.yaml` `performance_tier` に保存 |
| `--plan-type <api\|subscription>` | 料金プランタイプ — `llm.yaml` `plan_type` に保存 |
| `--high` | **削除予定** `--model-policy max` の別名 |

### 例

```bash
# 新規プロジェクトの初期化 (対話型ウィザード)
moai init my-project

# 既存フォルダへのインストール
cd my-existing-project
moai init

# 非対話型 (CI/CD)
moai init --non-interactive --project-mode personal --model-policy medium

# Phase 1 質問まで表示
moai init my-project --standard
```

詳しいウィザードステップは [初期設定](./init-wizard) ページを参照してください。

---

## moai update

MoAI-ADK を最新バージョンにアップデートします。フラグなしで実行するとバイナリとテンプレートを一緒に更新し、ユーザーのカスタム資産は自動保存されます。

```bash
moai update [OPTIONS]
```

### フラグ

| フラグ | 説明 |
|--------|------|
| `--check` | 新バージョンがあるかのみ確認 (アップデートしない) |
| `-c, --config` | 設定ウィザードの再実行 (テンプレート同期はしない) |
| `--force` | 強制アップデート (バージョン一致をスキップ、バックアップ+マージを強制、アーカイブドリフトを上書き) |
| `--yes` | すべての確認を自動承認 (CI/CD モード) |
| `--templates-only` | バイナリアップデートをスキップしテンプレートのみ同期 |
| `--binary` | テンプレート同期をスキップしバイナリのみアップデート |
| `--dry-run` | ファイルシステム変更なしで計画された作業のみ表示 |
| `--no-hooks` | Git フックのインストールをスキップ |
| `--verbose` | すべての警告を表示 (診断モード) |
| `--shell-env` | Claude Code 用のシェル環境変数を構成 |
| `--plan-type <api\|subscription>` | 料金プランタイプの上書き (`llm.yaml` `plan_type` とティアプロファイルを再適用) |

### 例

```bash
# 基本アップデート (バイナリ + テンプレート)
moai update

# 新バージョンがあるか確認のみ
moai update --check

# 設定ウィザードの再実行
moai update -c

# テンプレートのみ同期
moai update --templates-only
```

詳しいアップデート手順は [アップデート](./update) ページを参照してください。

---

## moai doctor

システム診断を実行します。Git、プロジェクト構造、設定ファイル、言語別の開発ツールを検査します。

```bash
moai doctor [OPTIONS]
```

### フラグ

| フラグ | 説明 |
|--------|------|
| `-v, --verbose` | 詳細なツールバージョンおよび言語検出結果を表示 |
| `--fix` | 欠落ツールの修正提案 |
| `--export <path>` | 診断結果を JSON ファイルにエクスポート |
| `--check <tool>` | 特定のツールのみ確認 (例: git, go, config) |

### 下位コマンド

| コマンド | 説明 |
|--------|------|
| `moai doctor sandbox` | サンドボックス環境の診断 |
| `moai doctor permission` | 権限設定の診断 |
| `moai doctor hook` | フックロードの問題診断 |
| `moai doctor config dump` | 現在の設定を JSON でダンプ |
| `moai doctor config diff` | ローカル設定とテンプレートのデフォルト値を比較 |

### 例

```bash
# 全体診断
moai doctor

# 詳細診断
moai doctor --verbose

# 診断結果のエクスポート
moai doctor --export diagnostics.json
```

---

## moai status

プロジェクトの状態を一目で照会します。初期化の有無、SPEC 個数、設定ファイル数を表示します。

```bash
moai status
```

フラグを持たない読み取り専用コマンドです。詳しい出力内容は [プロジェクト状態](./status) ページを参照してください。

---

## moai inventory

アクティブセッション、ワークツリー、ハーネスを統合照会する読み取り専用コマンドです。

```bash
moai inventory [OPTIONS]
```

### フラグ

| フラグ | 説明 |
|--------|------|
| `--json` | 構造化された JSON 出力 |
| `--project-root <path>` | プロジェクトルートパス (デフォルト値: 現在のディレクトリ) |

詳しい JSON スキーマと活用例は [moai inventory](./inventory) ページを参照してください。

---

## moai profile

Claude Code の設定プロファイルを管理します。プロファイルごとに独立したモデル、言語、表示設定を維持できます。

```bash
moai profile [COMMAND]
```

### 下位コマンド

| コマンド | 説明 |
|--------|------|
| `moai profile list` | 利用可能なすべてのプロファイルを表示 |
| `moai profile setup` | 対話型設定ウィザードの実行 |
| `moai profile current` | 現在アクティブなプロファイルを表示 |
| `moai profile delete <name>` | 指定されたプロファイルの削除 |

プロファイルの実行は `-p` フラグで指定します:

```bash
moai cc -p work       # work プロファイルで Claude 実行
moai glm -p cost-save # cost-save プロファイルで GLM 実行
moai cg -p team       # team プロファイルで CG モード実行
```

詳しい内容は [プロファイル管理](./profile) ページを参照してください。

---

## moai hook

Claude Code のフックイベントを処理するディスパッチャーです。`settings.json` のフック設定から `moai hook <event>` の形で呼び出されます。

```bash
moai hook <event>
```

### 対応イベント (26 個)

すべてのイベント名は kebab-case です。

| イベント | 説明 |
|-------|------|
| `session-start` | セッション開始 |
| `session-end` | セッション終了 |
| `pre-tool` | ツール実行前 (PreToolUse) |
| `post-tool` | ツール実行後 (PostToolUse) |
| `post-tool-failure` | ツール実行失敗後 |
| `stop` | セッション停止 |
| `stop-failure` | 停止失敗 |
| `compact` | コンテキスト圧縮前 (PreCompact) |
| `post-compact` | コンテキスト圧縮後 |
| `notification` | システム通知 |
| `subagent-start` | サブエージェント開始 |
| `subagent-stop` | サブエージェント終了 |
| `user-prompt-submit` | ユーザープロンプト送信 |
| `permission-request` | 権限リクエスト |
| `permission-denied` | 権限拒否 |
| `teammate-idle` | チームメイトのアイドル状態 |
| `task-completed` | タスク完了 |
| `task-created` | タスク生成 |
| `worktree-create` | ワークツリー生成 |
| `worktree-remove` | ワークツリー削除 |
| `instructions-loaded` | インストラクションロード完了 |
| `config-change` | 設定変更 |
| `cwd-changed` | 作業ディレクトリ変更 |
| `file-changed` | ファイル変更 |
| `elicitation` | MCP elicitation リクエスト |
| `elicitation-result` | MCP elicitation 結果 |

フックはユーザーが直接実行しません — Claude Code の `settings.json` が自動的に呼び出します。

---

## moai worktree

Git worktree を管理して並列 SPEC 開発を行います。

```bash
moai worktree <COMMAND> [ARGS]...
```

### 下位コマンド

| コマンド | 説明 |
|--------|------|
| `moai worktree new <SPEC_ID>` | 新しい worktree の生成 |
| `moai worktree list` | アクティブな worktree 一覧 |
| `moai worktree go <SPEC_ID>` | worktree ディレクトリへ移動 |
| `moai worktree remove <SPEC_ID>` | worktree の削除 |
| `moai worktree clean` | 古い worktree の整理 |
| `moai worktree recover` | 既存ディレクトリからの復旧 |
| `moai worktree status` | worktree の状態照会 |

---

## moai cc / moai cg / moai glm

Claude Code を開始しながらバックエンドを選択するランチコマンドです。3 つのコマンドすべて `-p <profile>` フラグでプロファイルを指定でき、`--` 以降の引数を Claude Code にそのまま渡します。

```bash
moai cc [-p profile] [-- claude-args...]
moai cg [-p profile] [-- claude-args...]
moai glm [-p profile] [-- claude-args...]
```

| コマンド | リーダー | ワーカー | tmux 必須 | 用途 |
|--------|------|------|-----------|------|
| `moai cc` | Claude | Claude | いいえ | 最高品質 (単一バックエンド) |
| `moai glm` | GLM | GLM | 推奨 | コスト最適化 (GLM 単独) |
| `moai cg` | Claude | GLM | 必須 | 品質 + コストのバランス (ハイブリッド) |

`moai cg` は CG モード (Claude リーダー + GLM チームメイト) を有効化します。tmux セッション内で実行する必要があり、GLM 環境変数を tmux セッションに注入してリーダー画面は Claude API を使います。

```bash
# 1. GLM API キーの保存 (最初の 1 回)
moai glm sk-your-glm-api-key

# 2. CG モードの有効化 (tmux 内で実行)
moai cg

# 3. 同じ画面で Claude Code を開始
claude
```

詳しい CG モードの案内は [紹介 — GLM でトークン節約](./introduction#glm-でトークン節約-5070) を参照してください。

---

## moai version

バージョン、コミットハッシュ、ビルド日付を表示します。

```bash
moai version
moai --version    # 同じ
```

---

## モデルポリシー (パフォーマンスティア)

MoAI-ADK はエージェントに最適な AI モデルを割り当てるパフォーマンスティアシステムを提供します — トークノミクスの出発点です。`llm.yaml` の `performance_tier` フィールドで設定し、`--model-policy` フラグまたは初期化ウィザードで選択します。

| ティア | 特徴 |
|------|------|
| **max** | 最高品質 — 計画・監査に Opus 割り当て、最大の推論深度 |
| **medium** (デフォルト値) | 品質とコストのバランス |
| **low** | 経済的 — Sonnet 中心の配分 |

```bash
# 初期化時に設定
moai init my-project --model-policy max

# 既存プロジェクトで再設定
moai update -c
```

料金プランタイプ (`plan_type`: api または subscription) は別途設定し、同じティアでも課金方式に応じてモデル割り当てが変わります。詳しいモデル-ティアマッピングは [モデルポリシー](/ja/multi-llm/model-policy) ページを参照してください。

---

## 参考

- [クイックスタート](./quickstart)
- [インストール](./installation)
- [アップデート](./update)
- [初期設定](./init-wizard)
- [プロファイル管理](./profile)
- [プロジェクト状態](./status)
