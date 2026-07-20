---
title: CLI 概要
weight: 90
draft: false
---

ターミナルで実行する `moai` (Go バイナリ) のすべてのコマンドとフラグを俯瞰します。Claude Code の対話画面で入力する `/moai` (スラッシュサブコマンド) とは完全に別のツールです — このページはターミナル CLI のみを扱います。

> コマンドごとの詳細リファレンス (フラグ・サブコマンド・例) は [CLIリファレンス](/cli-reference) セクションを参照してください。


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
```

```text
╭────────────────────────╮
│                        │
│    moai-adk v3.0.0     │
│                        │
│                        │
╰────────────────────────╯
 v3.0.0   none   built unknown
```

ボックスバナーの下の行は `<バージョン>   <コミットハッシュ>   built <ビルド時刻>` の順で表示されます。`go install` など ldflags なしでビルドした場合、コミットは `none`、ビルド時刻は `unknown` と表示されます。

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
| `--mode <ddd\|tdd>` | 開発方法論 (デフォルト値: tdd) |
| `--language <lang>` | 主要プログラミング言語 |
| `--framework <name>` | フレームワーク名 (デフォルト値: 自動検出または "none") |
| `--name <name>` | プロジェクト名 (デフォルト値: ディレクトリ名) |
| `--root <path>` | プロジェクトルートディレクトリ (デフォルト値: 現在のディレクトリ) |
| `--git-mode <manual\|personal\|team>` | Git ワークフローモード (デフォルト値: manual) |
| `--git-provider <github\|gitlab>` | Git プロバイダー |
| `--project-mode <personal\|team>` | プロジェクトモード (デフォルト値: personal) |
| `--harness-profile <profile>` | ハーネス評価者プロファイル (default, strict, lenient, frontend) |
| `--enable-lsp` | LSP 連携の有効化 (デフォルト値: false) |
| `--enforce-quality` | 品質ゲートの強制 (デフォルト値: true) |
| `--enable-design` | デザインワークフローの有効化 (デフォルト値: true) |
| `--profile <max\|medium\|low>` | モデル+effort プロファイル — `llm.yaml` `profile` に保存 (プロファイルマトリクス列の選択) |
| `--model-policy <max\|medium\|low>` | legacy パフォーマンスティア — `llm.yaml` `performance_tier` に保存 (`profile` 不在時にエイリアス) |
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
| `--profile <max\|medium\|low>` | モデル+effort プロファイルの上書き (`llm.yaml` `profile` に保存) |

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
| `moai doctor sandbox` | サンドボックスバックエンドの可用性診断 |
| `moai doctor permission` | 権限解決の診断 |
| `moai doctor hook` | 27 個のフックイベントカバレッジ表を表示 |
| `moai doctor config dump` | マージ済み設定を provenance とともにダンプ |
| `moai doctor config diff <tier-a> <tier-b>` | 2 つの設定ティアを比較 |

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

### 対応サブコマンド (約 38 個)

`moai hook` ディスパッチャーは、標準の Claude Code フックイベントと MoAI 専用の内部アクションを合わせて約 38 個のサブコマンドを提供します。すべての名前は kebab-case です。以下は代表的なイベントです。

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

MoAI 専用のサブコマンドも含まれます。

| サブコマンド | 説明 |
|-------|------|
| `stop-goal` | ターン終了時にアクティブセッションの goal を評価 |
| `pre-push` | コミットメッセージを規約に沿って検証 |
| `spec-status` | git コミット時に SPEC status を自動更新 |
| `harness-classify` | ハーネス分類器の実行とティア昇格の記録 |
| `harness-observe` · `harness-observe-stop` · `harness-observe-subagent-stop` · `harness-observe-user-prompt-submit` | ハーネス使用ログの記録 |
| `db-schema-sync` | PostToolUse フックで DB スキーマ変更を検出 |

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
| `moai worktree new [branch-name]` | 新しい worktree の生成 |
| `moai worktree list` | アクティブな worktree 一覧 |
| `moai worktree go [branch-name]` | worktree パスを**出力** (シェル移動用) |
| `moai worktree switch [branch-name]` | worktree へ切替 |
| `moai worktree done [branch-name]` | worktree の完了と整理 |
| `moai worktree sync [branch-name]` | base ブランチと worktree を同期 |
| `moai worktree remove [path]` | worktree の削除 |
| `moai worktree config [key] [value]` | worktree 設定の照会/変更 |
| `moai worktree status` | worktree の状態照会 |
| `moai worktree clean` | 古い worktree 参照の整理 |
| `moai worktree recover` | worktree レジストリの復旧 |
| `moai worktree snapshot` | 作業ツリー状態のスナップショットを取得 |
| `moai worktree restore` | スナップショット HEAD 状態へ作業ツリーを復元 |
| `moai worktree verify` | 作業ツリー状態をスナップショットと照合検証 |

`moai worktree go` はディレクトリを変えずにパスだけ出力します。実際の移動はシェルで次のように包んで使います。

```bash
cd "$(moai worktree go my-branch)"
```

---

## moai cc / moai cg / moai glm

Claude Code を開始しながらバックエンドを選択するランチコマンドです。3 つのコマンドすべて `-p <profile>` フラグでプロファイルを指定できます。`--` 以降の引数を Claude Code にそのまま渡すのは `moai cc` と `moai glm` のみ対応します (`moai cg` は非対応)。

```bash
moai cc [-p profile] [-- claude-args...]
moai glm [-p profile] [-- claude-args...]
moai cg [-p profile]
```

| コマンド | リーダー | ワーカー | tmux 必須 | 用途 |
|--------|------|------|-----------|------|
| `moai cc` | Claude | Claude | いいえ | 最高品質 (単一バックエンド) |
| `moai glm` | GLM | GLM | いいえ | コスト最適化 (GLM 単独) |
| `moai cg` | Claude | GLM | 必須 | 品質 + コストのバランス (ハイブリッド) |

`moai cg` は CG モード (Claude リーダー + GLM チームメイト) を有効化します。tmux セッション内で実行する必要があり、GLM 環境変数を tmux セッションに注入してリーダー画面は Claude API を使います。`moai cg` は設定後、現在の画面ですぐに Claude Code を実行するので、別途 `claude` を実行するステップは不要です。

```bash
# 1. GLM API キーの保存 (最初の 1 回)
moai glm setup sk-your-glm-api-key

# 2. CG モードの有効化 (tmux 内で実行 — Claude Code が現在の画面ですぐに開始される)
moai cg
```

詳しい CG モードの案内は [紹介 — GLM でトークン節約](./introduction#glm-でトークン節約-5070) を参照してください。

### ランチフラグ

3 つのランチコマンドが共通で対応するフラグです。

| フラグ | 説明 |
|--------|------|
| `-p, --profile <name>` | 名前付き Claude プロファイルを使用 |
| `--permission-mode <mode>` | 権限モード (default, acceptEdits, plan, auto, bypassPermissions, dontAsk) |
| `-b, --bypass` | `--permission-mode bypassPermissions` の短縮形 |

`moai cc` はさらに次のフラグに対応します。

| フラグ | 説明 |
|--------|------|
| `-c, --continue` | 以前のセッションを継続 |
| `-m, --model <model>` | モデル選択の上書き |
| `--chrome` / `--no-chrome` | Chrome MCP のトグル |

> `auto` 権限モードは GLM (サードパーティプロバイダー) では使えません — `moai cc` または `moai cg` でのみ対応します。

### moai glm 下位コマンド

| コマンド | 説明 |
|--------|------|
| `moai glm setup <api-key>` | GLM API キーの保存 |
| `moai glm status` | 現在の GLM 資格情報状態を表示 |
| `moai glm tools` | Z.AI MCP サーバーのツール管理 (有効/無効) |

---

## moai goal

現在のセッションに条件ベースの自律 goal ループを登録・照会・解除します。条件が満たされるまで毎ターン終了時に評価されます。

```bash
moai goal <COMMAND>
```

| コマンド | 説明 |
|--------|------|
| `moai goal arm <condition>` | アクティブセッションに goal を登録・有効化 |
| `moai goal status` | アクティブセッションの goal 状態を出力 |
| `moai goal clear` | アクティブセッションの goal を解除 |

---

## moai handoff

`/clear` 境界を越えてセッションを継続するための auto-resume ハンドオフ待機レコードを管理します。

```bash
moai handoff <COMMAND>
```

| コマンド | 説明 |
|--------|------|
| `moai handoff save` | 貼り付け用 resume 本文を待機レコードとして保存 |
| `moai handoff clear` | 待機ハンドオフレコードを削除 |

---

## moai session

マルチセッションレース緩和のためのアクティブセッション調整レジストリを管理します。

```bash
moai session <COMMAND>
```

| コマンド | 説明 |
|--------|------|
| `moai session current` | 現在のオーケストレーターセッション UUID を出力 |
| `moai session list` | アクティブセッション一覧 (`--filter-spec` でフィルタリング可) |
| `moai session register <session_id> <spec_id> <phase>` | レジストリにセッションを登録 |
| `moai session deregister <session_id>` | レジストリからセッションを削除 (idempotent) |
| `moai session heartbeat <session_id>` | セッションの last_heartbeat を更新 |
| `moai session purge` | 古い項目を削除 (デフォルト: 最後の heartbeat から 30 分超過) |
| `moai session doctor` | セッションレジストリが空である原因を診断 |

---

## moai web

ブラウザベースの設定エディタ MoAI Web Console を起動します。

```bash
moai web [OPTIONS]
```

| フラグ | 説明 |
|--------|------|
| `--port <N>` | 127.0.0.1 にバインドする TCP ポート (デフォルト: 3041) |
| `--no-open` | ブラウザを自動で開かない |
| `--no-reuse` | 古い moai インスタンスからポートを回収しない |

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

プロファイル (`profile`: max/medium/low) はプロファイルマトリクスのアクティブ列を選択し、各エージェントの model+effort を決定します。詳しいエージェント別マッピングは [プロファイルマトリクス](/ja/advanced/profile-matrix/) ページを参照してください。

---

## 参考

- [クイックスタート](./quickstart)
- [インストール](./installation)
- [アップデート](./update)
