---
title: CG モード (Claude + GLM)
weight: 20
draft: false
---

## CG モードとは?

CG (Claude + GLM) モードは、リーダーが **Claude API** を、ワーカーが **GLM API** を使用する
ハイブリッドモードです。tmux セッションレベルの環境変数分離で実装されており、「計画は
Claude が深く、実装は GLM が安く」というトークノミクス配分を 1 つのセッション内で
実行します。実装中心のタスクで約 60-70% のコストが削減されます。

## アーキテクチャ

```
moai cg 実行
    │
    ├── 1. GLM 設定を tmux セッション環境変数に注入
    │      (ANTHROPIC_AUTH_TOKEN, BASE_URL, MODEL_* 変数)
    │
    ├── 2. settings.local.json から GLM 環境変数を削除
    │      → リーダー pane は Claude API を使用
    │
    ├── 3. CLAUDE_CODE_TEAMMATE_DISPLAY=tmux を設定
    │      → ワーカーは新しい pane で GLM 環境変数を継承
    │
    └── 4. Claude Code を実行 (現在のプロセスを置換)
```

```
┌─────────────────────────────────────────────────────────────┐
│  リーダー (現在の tmux pane, Claude API)                      │
│  - ワークフローのオーケストレーション                          │
│  - plan, quality, sync フェーズの処理                         │
│  - GLM 環境変数なし → Claude API を使用                       │
└──────────────────────┬──────────────────────────────────────┘
                       │ チームメイト spawn (新しい tmux pane)
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  チームメイト (新しい tmux pane, GLM API)                     │
│  - tmux セッション環境変数を継承 → GLM API を使用              │
│  - run フェーズで実装タスクを実行                             │
│  - SendMessage でリーダーと通信                               │
└─────────────────────────────────────────────────────────────┘
```

## 使い方

### ステップ 1: GLM API キーの保存 (初回のみ)

```bash
moai glm sk-your-glm-api-key
```

キーは `~/.moai/.env.glm` に安全に保存されます。

### ステップ 2: tmux 環境の確認

すでに tmux を使用中であれば、新しいセッションを作る必要はありません。

```bash
# tmux を使用していない場合:
tmux new -s moai
```

> **ヒント**: VS Code ターミナルのデフォルトを tmux に設定すると、このステップを完全にスキップできます。

### ステップ 3: CG モードの実行

```bash
moai cg
```

`moai cg` は現在の pane で自動的に Claude Code を実行します。別途 `claude` を実行する必要はありません。

### ステップ 4: ワークフローの実行

```bash
/moai "ユーザー認証機能の実装"
```

以降はいつも通りです。オーケストレーター (リーダー、Claude) が計画・品質・同期を
担当し、実装量の大きいタスクは新しい tmux pane の GLM チームメイトに委任されます。

> **参考**: 従来の `--team` フラグ (Agent Teams 静的オーケストレーション層) は
> v3.0 で引退しました。強制的に指定しても sub-agent モードにフォールバックします。CG
> モードのリーダー/ワーカー分離は Claude Code 内蔵の teammate ランタイム (tmux pane) で
> 動作し、このランタイムはそのまま維持されます。

## 重要事項

| 項目 | 説明 |
|------|------|
| **tmux 環境** | すでに tmux を使用中なら新しいセッションは不要。VS Code ターミナルのデフォルトを tmux に設定すると便利 |
| **自動実行** | `moai cg` が現在の pane で Claude Code を自動実行。別途 `claude` コマンドは不要 |
| **セッション終了** | session_end フックが自動で tmux セッション環境変数をクリーンアップ → 次のセッションは Claude を使用 |
| **チーム通信** | SendMessage ツールでリーダー↔ワーカー間の通信 |
| **モード切り替え** | `moai glm` から切り替える際、`moai cg` が GLM 設定を自動初期化 — 途中で `moai cc` は不要 |

## tmux 環境変数注入のセキュリティモデル {#tmux-env-security}

v2.20.0-rc1 から、`moai cg` が GLM token (`ANTHROPIC_AUTH_TOKEN`) を tmux セッション環境変数に注入する際、**argv チャネル** (`tmux set-environment <KEY> <VALUE>`) の代わりに **source-file チャネル** (`tmux source-file <tmp>`) を使用します。token はもはや `ps auxe`、`/proc/<pid>/cmdline`、auditd ログ、sysmon トレース、クラッシュダンプに平文で露出しません (CWE-214)。

### 注入フロー

1. `~/.moai/run/` 配下に一時ファイルを `mkstemp` で作成 (mode `0o600` を強制)
2. `set-environment -t <session> <KEY> <VALUE>` の 1 行を記録
3. `tmux source-file <tmp>` で tmux がそのファイルを読み込んで環境に注入
4. 注入直後に `os.Remove` で unlink

argv には一時ファイルのパスのみ露出し、token 自体は露出しません。

### Non-sensitive な値は argv を維持

`CLAUDE_CONFIG_DIR`、`ANTHROPIC_BASE_URL`、`ANTHROPIC_DEFAULT_*_MODEL` など token ではない値は、従来の argv 経路を維持します (セキュリティ上の脅威なし)。

### ユーザーの責任

`~/.moai/.env.glm` source ファイルは、ユーザー環境で `0o600` 権限を維持する必要があります。これは `moai glm` コマンドが自動で設定します:

```bash
stat -c '%a' ~/.moai/.env.glm    # Linux: 600
stat -f '%A' ~/.moai/.env.glm    # macOS: 600
```

### セルフチェック

CG モード実行中に token が argv に露出していないか確認:

```bash
# moai cg 実行後、新しい tmux セッション内で
ps auxe | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# 期待値: 0 matches (token が argv にない)
```

詳細な脅威モデル、失敗時の動作 (`ErrTmuxSensitiveInjectFailed` sentinel)、追加のチェック手順は[セキュリティノート — CWE-214](/ja/advanced/security-notes/#cwe-214) を参照してください。

## ディスプレイモード

teammate ランタイムは 2 つのディスプレイモードをサポートします:

| モード | 説明 | 通信 | リーダー/ワーカー分離 |
|------|------|------|--------------|
| `in-process` | デフォルトモード、すべてのターミナル | SendMessage 対応 | 分離なし (同じ環境) |
| `tmux` | 分割画面表示 | SendMessage 対応 | セッション環境変数の分離 |

> **CG モードは `tmux` ディスプレイモードでのみリーダー/ワーカーの API 分離が可能です。**

## モード比較

| コマンド | リーダー | ワーカー | tmux 必要 | コスト削減 | 用途 |
|--------|------|------|----------|----------|------|
| `moai cc` | Claude | Claude | いいえ | - | 複雑なタスク、最高品質 |
| `moai glm` | GLM | GLM | 推奨 | ~70% | コスト最適化 |
| `moai cg` | Claude | GLM | **必須** | **~60%** | 品質とコストのバランス |

### いつ CG モードを使うべきですか?

**CG モードが適する場面:**
- 実装中心の SPEC 実行 (run フェーズ)
- コード生成タスク
- テスト作成
- ドキュメント生成

**Claude 専用 (cc) が適する場面:**
- アーキテクチャ設計/計画 (Opus の推論が必要)
- セキュリティレビュー (Claude のセキュリティトレーニングが必要)
- 複雑なデバッグ (高度な推論が必要)

## トラブルシューティング

| 問題 | 原因 | 解決 |
|------|------|------|
| ワーカーが Claude API を使用 | tmux セッション環境変数が未設定 | tmux 内で `moai cg` を再実行 |
| `moai cg` 後に Claude Code が起動しない | tmux 外部で実行 | `tmux new -s moai` の後に再実行 |
| セッション終了後に GLM 環境変数が残留 | session_end フックの失敗 | `moai cc` で手動クリーンアップ |

## 次のステップ

- [モデルポリシー](/ja/multi-llm/model-policy) — エージェント別モデル割り当て
- [よくある質問](/ja/getting-started/faq) — 実行モード関連の FAQ
- [CLI リファレンス](/ja/getting-started/cli) — moai cc、moai glm、moai cg の詳細
