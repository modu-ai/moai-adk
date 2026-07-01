---
title: CG モード (Claude + GLM)
weight: 20
draft: false
---

## CG モードとは?

CG (Claude + GLM) モードは、リーダーが **Claude API** を使用し、ワーカーが **GLM API** を使用するハイブリッドモードです。tmux セッションレベルの環境変数分離により実装されています。

## アーキテクチャ

```
moai cg を実行
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
    └── 4. Claude Code を実行 (現在プロセスを置換)
```

```
┌─────────────────────────────────────────────────────────────┐
│  リーダー (現在の tmux pane, Claude API)                    │
│  - /moai --team を実行時にワークフロー オーケストレーション  │
│  - plan、quality、sync フェーズを処理                       │
│  - GLM 環境変数なし → Claude API を使用                    │
└──────────────────────┬──────────────────────────────────────┘
                       │ Agent Teams (新しい tmux pane)
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  チームメンバー (新しい tmux pane, GLM API)                 │
│  - tmux セッション環境変数を継承 → GLM API を使用           │
│  - run フェーズで実装作業を実行                              │
│  - SendMessage でリーダーと通信                              │
└─────────────────────────────────────────────────────────────┘
```

## 使用方法

### ステップ 1: GLM API キーを保存 (初回のみ)

```bash
moai glm sk-your-glm-api-key
```

キーは `~/.moai/.env.glm` に安全に保存されます。

### ステップ 2: tmux 環境を確認

既に tmux を使用している場合は、新しいセッションを作成する必要はありません。

```bash
# tmux を使用していない場合:
tmux new -s moai
```

> **ヒント**: VS Code ターミナルのデフォルトを tmux に設定すると、このステップを完全にスキップできます。

### ステップ 3: CG モードを実行

```bash
moai cg
```

`moai cg` は現在の pane で自動的に Claude Code を実行します。別途 `claude` を実行する必要はありません。

### ステップ 4: チームワークフローを実行

```bash
/moai --team "ユーザー認証機能を実装"
```

## 重要事項

| 項目 | 説明 |
|------|------|
| **tmux 環境** | 既に tmux を使用している場合、新しいセッション不要。VS Code ターミナルのデフォルトを tmux に設定すると便利 |
| **自動実行** | `moai cg` が現在の pane で Claude Code を自動実行。別途 `claude` コマンド不要 |
| **セッション終了** | session_end フック が自動的に tmux セッション環境変数をクリーンアップ → 次のセッションは Claude を使用 |
| **チーム通信** | SendMessage ツールでリーダー ↔ ワーカー間通信 |
| **モード切り替え** | `moai glm` から切り替え時に `moai cg` が GLM 設定を自動初期化 — 途中で `moai cc` 不要 |

## tmux 環境変数注入セキュリティモデル {#tmux-env-security}

v2.20.0-rc1 から `moai cg` が GLM トークン (`ANTHROPIC_AUTH_TOKEN`) を tmux セッション環境変数に注入する際、**argv チャネル** (`tmux set-environment <KEY> <VALUE>`) の代わりに **source-file チャネル** (`tmux source-file <tmp>`) を使用します。トークンはもはや `ps auxe`、`/proc/<pid>/cmdline`、auditd ログ、sysmon 追跡、クラッシュダンプに平文で公開されません (CWE-214)。

### 注入フロー

1. `~/.moai/run/` の下に `mkstemp` で一時ファイルを作成 (モード `0o600` 強制)
2. `set-environment -t <session> <KEY> <VALUE>` の 1 行を記録
3. `tmux source-file <tmp>` で tmux がそのファイルを読んで環境に注入
4. 注入直後に `os.Remove` で削除

argv にはテンポラリファイルパスのみ公開され、トークン自体は公開されません。

### 非機密値は argv を保持

`CLAUDE_CONFIG_DIR`、`ANTHROPIC_BASE_URL`、`ANTHROPIC_DEFAULT_*_MODEL` など、トークンではない値は既存の argv パスを保持します (セキュリティ上の脅威なし)。

### ユーザー責任

`~/.moai/.env.glm` ソースファイルはユーザー環境で `0o600` パーミッションを保持する必要があります。これは `moai glm` コマンドが自動的に設定します:

```bash
stat -c '%a' ~/.moai/.env.glm    # Linux: 600
stat -f '%A' ~/.moai/.env.glm    # macOS: 600
```

### 自己チェック

CG モード実行中にトークンが argv に公開されるか確認:

```bash
# moai cg 実行後、新しい tmux セッション内で
ps auxe | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# 期待値: 0 マッチ (トークンが argv に無い)
```

詳細な脅威モデル、失敗時の動作 (`ErrTmuxSensitiveInjectFailed` sentinel)、追加チェック手順については [セキュリティノート — CWE-214](/ja/advanced/security-notes/#cwe-214) を参照してください。

## ディスプレイモード

Agent Teams は 2 つのディスプレイモードをサポートしています:

| モード | 説明 | 通信 | リーダー/ワーカー分離 |
|--------|------|------|-----------------|
| `in-process` | デフォルトモード、すべてのターミナル | ✅ SendMessage | ❌ 同じ環境 |
| `tmux` | 分割画面表示 | ✅ SendMessage | ✅ セッション環境変数分離 |

> **CG モードは `tmux` ディスプレイモードでのみリーダー/ワーカー API 分離が可能です。**

## モード比較

| コマンド | リーダー | ワーカー | tmux 必須 | コスト削減 | 用途 |
|--------|--------|--------|----------|----------|------|
| `moai cc` | Claude | Claude | いいえ | - | 複雑なタスク、最高品質 |
| `moai glm` | GLM | GLM | 推奨 | ~70% | コスト最適化 |
| `moai cg` | Claude | GLM | **必須** | **~60%** | 品質とコストのバランス |

### CG モードをいつ使うべきか?

**CG モードに適した用途:**
- 実装中心の SPEC 実行 (run フェーズ)
- コード生成タスク
- テスト作成
- ドキュメント生成

**Claude 専用 (cc) に適した用途:**
- アーキテクチャ設計/計画 (Opus 推論が必要)
- セキュリティレビュー (Claude のセキュリティトレーニングが必要)
- 複雑なデバッグ (高度な推論が必要)

## トラブルシューティング

| 問題 | 原因 | 解決策 |
|------|------|--------|
| ワーカーが Claude API を使用 | tmux セッション環境変数が未設定 | tmux 内で `moai cg` を再実行 |
| `moai cg` 後に Claude Code が実行されない | tmux 外から実行 | `tmux new -s moai` 後に再実行 |
| セッション終了後に GLM 環境変数が残る | session_end フック失敗 | `moai cc` で手動クリーンアップ |

## 次のステップ

- [モデルポリシー](/ja/multi-llm/model-policy) — エージェント別モデル割り当て
- [二重実行モード](/ja/getting-started/faq) — Sub-Agent vs Agent Teams
- [CLI リファレンス](/ja/getting-started/cli) — moai cc、moai glm、moai cg の詳細
