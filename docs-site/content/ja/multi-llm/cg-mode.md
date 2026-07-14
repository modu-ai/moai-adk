---
title: CG モード (Claude + GLM)
weight: 20
draft: false
---

## CG モードとは?

CG (Claude + GLM) モードはリーダーが **Claude API** を、ワーカーが **GLM API** を使う
ハイブリッドモードです。tmux セッションレベルの環境変数隔離で実装され、「計画は
Claude が深く、実装は GLM が安く」というトークノミクスの配分を 1 つのセッションの中で
実行します。実装中心の作業を基準に約 60-70% のコストが削減されます。

## アーキテクチャ

```
moai cg 実行
    │
    ├── 1. GLM 設定を tmux セッション環境変数に注入
    │      (ANTHROPIC_AUTH_TOKEN, BASE_URL, MODEL_* 変数)
    │
    ├── 2. settings.local.json から GLM 環境変数を除去
    │      → リーダー pane は Claude API を使用
    │
    ├── 3. settings.local.json に teammateMode: "tmux" を設定
    │      → ワーカーは新しい pane で GLM 環境変数を継承
    │
    └── 4. Claude Code 実行 (現在のプロセスを置き換え)
```

```
┌─────────────────────────────────────────────────────────────┐
│  リーダー (現在の tmux pane, Claude API)                     │
│  - ワークフローオーケストレーション                          │
│  - plan, quality, sync ステップの処理                        │
│  - GLM 環境変数なし → Claude API を使用                      │
└──────────────────────┬──────────────────────────────────────┘
                       │ チームメイト spawn (新しい tmux pane)
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  チームメイト (新しい tmux pane, GLM API)                    │
│  - tmux セッション環境変数を継承 → GLM API を使用            │
│  - run ステップで実装作業を実行                              │
│  - SendMessage でリーダーと通信                              │
└─────────────────────────────────────────────────────────────┘
```

## 使用方法

### ステップ 1: GLM API キーの保存 (最初の 1 回)

```bash
moai glm sk-your-glm-api-key
```

キーは `~/.moai/.env.glm` に安全に保存されます。

### ステップ 2: tmux 環境の確認

すでに tmux を使用中なら新しいセッションを作る必要はありません。

```bash
# tmux を使用中でないなら:
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

以降は普段どおりです。オーケストレーター (リーダー、Claude) が計画・品質・同期を
担い、実装の物量が大きい作業は新しい tmux pane の GLM チームメイトに委任されます。

> **参考**: かつての `--team` フラグ (Agent Teams 静的オーケストレーション階層) は
> v3.0 で引退しました。強制的に指定しても sub-agent モードにフォールバックします。CG
> モードのリーダー/ワーカー分離は Claude Code 内蔵の teammate ランタイム (tmux pane) で
> 動作し、このランタイムはそのまま維持されます。

## 重要事項

| 項目 | 説明 |
|------|------|
| **tmux 環境** | すでに tmux を使用中なら新しいセッション不要。VS Code ターミナルのデフォルトを tmux に設定すると便利 |
| **自動実行** | `moai cg` が現在の pane で Claude Code を自動実行。別途 `claude` コマンド不要 |
| **セッション終了** | session_end フックが自動的に tmux セッション環境変数を整理 → 次のセッションは Claude を使用 |
| **チーム通信** | SendMessage ツールでリーダー↔ワーカー間の通信 |
| **モード切替** | `moai glm` から切り替え時に `moai cg` が GLM 設定を自動初期化 — 途中で `moai cc` 不要 |

## tmux 環境変数注入のセキュリティモデル {#tmux-env-security}

v2.20.0-rc1 から `moai cg` が GLM token (`ANTHROPIC_AUTH_TOKEN`) を tmux セッション環境変数に注入するとき、**argv チャネル** (`tmux set-environment <KEY> <VALUE>`) の代わりに **source-file チャネル** (`tmux source-file <tmp>`) を使います。token はもはや `ps auxe`、`/proc/<pid>/cmdline`、auditd ログ、sysmon 追跡、クラッシュダンプに平文で露出しません (CWE-214)。

### 注入フロー

1. `~/.moai/run/` 配下に一時ファイルを `mkstemp` で生成 (mode `0o600` を強制)
2. `set-environment -t <session> <KEY> <VALUE>` の 1 行を記録
3. `tmux source-file <tmp>` で tmux がそのファイルを読んで環境に注入
4. 注入直後に `os.Remove` で unlink

argv には一時ファイルのパスのみが露出し、token 自体は露出しません。

### Non-sensitive な値は argv を維持

`CLAUDE_CONFIG_DIR`、`ANTHROPIC_BASE_URL`、`ANTHROPIC_DEFAULT_*_MODEL` など token でない値は従来の argv 経路を維持します (セキュリティ脅威なし)。

### ユーザーの責任

`~/.moai/.env.glm` の source ファイルはユーザー環境で `0o600` 権限を維持する必要があります。これは `moai glm` コマンドが自動的に設定します:

```bash
stat -c '%a' ~/.moai/.env.glm    # Linux: 600
stat -f '%A' ~/.moai/.env.glm    # macOS: 600
```

### 自己点検

CG モード実行中に token が argv に露出するか確認:

```bash
# moai cg 実行後、新しい tmux セッション内で
ps auxe | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# 期待値: 0 matches (token が argv にない)
```

詳しい脅威モデル、失敗時の動作 (`ErrTmuxSensitiveInjectFailed` sentinel)、追加の点検手順は [セキュリティノート — CWE-214](/ja/advanced/security-notes/#cwe-214) を参照してください。

## ディスプレイモード (teammateMode)

`teammateMode` は Claude Code 内蔵のディスプレイ設定で、`settings.local.json` に
保存されます。MoAI の team-mode (かつての `--team` フラグ、v3.0 引退) とは異なる
概念です — teammate ランタイム自体は Claude Code が提供し、`teammateMode` は
その表示方式のみを制御します。

| 値 | 説明 | リーダー/ワーカー分離 | CG モード |
|------|------|--------------|---------|
| `in-process` | デフォルト値、同じターミナルにインライン | 不可 | 未使用 |
| `auto` | 環境の自動検出 | 未対応 | 未使用 |
| `tmux` | tmux 分割画面 | セッション環境変数の隔離 | {{< icon check ok >}} 使用 |
| `iterm2` | iTerm2 分割画面 | 未対応 | 未使用 |

`moai cg` と `moai glm` は `settings.local.json` の `teammateMode` を `"tmux"` に
設定し、`moai cc` は空の値で解除します。かつての `CLAUDE_CODE_TEAMMATE_DISPLAY`
環境変数は `teammateMode` 設定が優先します。

> **CG モードは `tmux` ディスプレイモードでのみリーダー/ワーカーの API 分離が可能です。**

## モード比較

| コマンド | リーダー | ワーカー | tmux 必要 | コスト削減 | 用途 |
|--------|------|------|----------|----------|------|
| `moai cc` | Claude | Claude | いいえ | - | 複雑な作業、最高品質 |
| `moai glm` | GLM | GLM | 推奨 | ~70% | コスト最適化 |
| `moai cg` | Claude | GLM | **必須** | **~60%** | 品質 + コストのバランス |

### いつ CG モードを使うべきですか?

**CG モードが適する:**
- 実装中心の SPEC 実行 (run ステップ)
- コード生成作業
- テスト作成
- ドキュメント生成

**Claude 専用 (cc) が適する:**
- アーキテクチャ設計/計画 (Opus 推論が必要)
- セキュリティレビュー (Claude のセキュリティトレーニングが必要)
- 複雑なデバッグ (高度な推論が必要)

## 問題解決

| 問題 | 原因 | 解決 |
|------|------|------|
| ワーカーが Claude API を使用 | tmux セッション環境変数が未設定 | tmux 内で `moai cg` を再実行 |
| `moai cg` 後に Claude Code が未起動 | tmux 外部で実行 | `tmux new -s moai` の後に再実行 |
| セッション終了後に GLM 環境変数が残留 | session_end フックの失敗 | `moai cc` で手動整理 |

## 次のステップ

- [モデルポリシー](/ja/multi-llm/model-policy) — エージェント別のモデル割り当て
- [よくある質問](/ja/getting-started/faq) — 実行モード関連の FAQ
- [CLI リファレンス](/ja/getting-started/cli) — moai cc, moai glm, moai cg 詳細
