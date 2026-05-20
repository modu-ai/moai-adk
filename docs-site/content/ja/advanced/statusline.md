---
title: ステータスラインシステムとPRセグメント
weight: 78
draft: false
---

Claude Code と moai-adk-go を統合するための**カスタムステータスラインシステム**です。v2.1.145 以降では、GitHub PR 情報をステータスラインに表示できます。

> MoAI ワークフローは PR 中心です。すべての SPEC は plan-PR → run-PR → sync-PR を生成するため、ステータスラインに現在の PR 状態を表示することで、開発効率が向上します。

## 概要

### カスタムステータスラインが必要な理由

Claude Code のデフォルトステータスラインは、一般的な使用パターンに最適化されています。しかし MoAI-ADK ユーザーは以下のような特殊な情報が必要です：

- **PR 中心ワークフロー**：現在の PR 番号とレビュー状態（approved/pending/changes_requested）
- **マルチペイン開発**：ワークツリーベースの並列開発時に現在の SPEC ステータスを表示
- **コスト追跡**：GLM 環境使用時のリアルタイムコスト監視
- **コンテキスト管理**：現在のセッションのトークン使用率と累積コスト

カスタムステータスラインは `.moai/status_line.sh` レンダラーを通じてこれらの情報を表示します。

### ステータスラインアーキテクチャ

```
Claude Code stdin (JSON)
    ↓
internal/statusline/types.go (StdinData パース)
    ↓
internal/statusline/builder.go (セグメント構成)
    ↓
internal/statusline/renderer.go (色分けとレンダリング)
    ↓
.moai/status_line.sh (テンプレートベースの最終レンダリング)
```

## 設定

### 基本構造

`.moai/config/sections/statusline.yaml` でステータスラインを設定します：

```yaml
statusline:
  mode: default              # default | compact | verbose
  theme: catppuccin-mocha    # 色のテーマを選択
  preset: full               # full | minimal | custom
  segments:
    model: true              # Claude モデルを表示
    context: true            # コンテキスト使用率を表示
    directory: true          # 作業ディレクトリを表示
    git_status: true         # Git ステータスを表示
    git_branch: true         # Git ブランチを表示
    worktree: false          # ワークツリー情報を表示（オプション）
    effort_thinking: false   # Effort/thinking ステータス（オプション）
    pr: false                # PR 情報を表示（オプション、v2.1.145+）
```

### セグメントオプション

| セグメント | デフォルト | 用途 | 説明 |
|-----------|-----------|------|------|
| `model` | true | 現在のモデル | Claude モデルバージョンを表示 |
| `context` | true | コンテキスト使用率 | 現在のセッションのトークン使用率 |
| `directory` | true | 作業パス | 現在の作業ディレクトリ |
| `git_status` | true | Git ステータス | 変更されたファイル数、stash ステータス |
| `git_branch` | true | 現在のブランチ | ブランチ名とリモートとの差分 |
| `worktree` | false | ワークツリー情報 | 現在のワークツリーを表示（並列開発） |
| `effort_thinking` | false | 思考モード | effort および thinking ステータス |
| `pr` | false | PR 情報 | GitHub PR 番号とレビュー状態（NEW v2.1.145+） |

## 利用可能なセグメント

### 常に有効なセグメント（4個）

**model** — Claude モデル
- 現在のモデル（Claude 3.5 Sonnet、Claude 3.7 Opus など）を表示
- 例：`Claude 3.5 Sonnet`

**context** — コンテキスト使用率
- 現在のセッションのトークン使用率を表示
- 形式：`150K/200K`（使用中 / 全体）
- 75% 以上の場合、警告色で表示

**directory** — 作業ディレクトリ
- 現在の作業ディレクトリの相対パス
- プロジェクトルートからの位置を表示

**git_status** — Git ステータス
- 変更されたファイル数：`M5`（5ファイル変更）
- Stash ステータス：`S2`（2個の stash）
- 例：`M5 S2`

**git_branch** — 現在のブランチ
- ブランチ名
- リモートとのコミット差分表示
- 例：`feat/SPEC-001 +3 -1`

### オプションセグメント（7個）

**worktree** — ワークツリー情報（オプション）
- L2 ワークツリー使用時に表示
- 現在の SPEC 名を表示
- 有効化：`segments.worktree: true`

**effort_thinking** — Effort/thinking ステータス（オプション）
- Claude 4.7 の thinking モード有効化状態
- effort レベル（high/xhigh/max）
- 有効化：`segments.effort_thinking: true`

**output_style** — 出力スタイル（オプション）
- 現在の出力スタイル設定
- 有効化：`segments.output_style: true`

**claude_version** — Claude バージョン（オプション）
- Claude Code バージョン
- 有効化：`segments.claude_version: true`

**moai_version** — moai バージョン（オプション）
- MoAI-ADK バージョン
- 有効化：`segments.moai_version: true`

**session_time** — セッション経過時間（オプション）
- 現在のセッション開始からの経過時間
- 有効化：`segments.session_time: true`

**usage_5h** — 5時間累積コスト（オプション）
- 過去 5 時間のコスト追跡
- GLM 環境で有用
- 有効化：`segments.usage_5h: true`

**usage_7d** — 7日累積コスト（オプション）
- 過去 7 日のコスト追跡
- 有効化：`segments.usage_7d: true`

**task** — アクティブな SPEC ワークフロー情報（オプション）
- 出力形式：`📋 [<command> <SPEC-ID>-<stage>]`（例：`📋 [/moai run SPEC-V3R5-DOCS-SECURITY-001-M3]`）
- データソース：`~/.moai/state/last-session-state.json` の `active_task` フィールド（SessionStart フックが自動設定）
- 非アクティブの場合はセグメント自体が非表示（graceful no-output）
- 有効化：`segments.task: true`（既定値は false — opt-in）

## NEW v2.1.145: PR セグメント

### 概要

Claude Code v2.1.145 以降、ステータスラインの stdin JSON に GitHub PR 情報が含まれます。MoAI-ADK はこれを活用して、ステータスラインに現在の PR のレビュー状態を表示します。

**有効化**：`segments.pr: true` に設定すると有効になります。デフォルトは `false` です。

### PR セグメント表示形式

PR セグメントは以下の形式で表示されます：

```
#1023 ⌥approved
```

- `#1023`：PR 番号
- `⌥`：PR ステータス表示シンボル
- `approved`：レビュー状態（色分け）

### レビュー状態別の色

PR のレビュー状態に応じて異なる色で表示されます：

| 状態 | 色 | 意味 |
|------|-----|------|
| `approved` | 緑 | PR が承認されている |
| `pending` | 黄 | レビュー待ち中 |
| `changes_requested` | 赤 | 変更リクエストされた |
| `draft` | グレー | ドラフト状態 |
| （その他 / 空） | デフォルト | スタイル未適用 |

### 有効化方法

1. `.moai/config/sections/statusline.yaml` ファイルを編集

```yaml
statusline:
  segments:
    pr: true   # PR セグメントを有効化
```

2. Claude Code セッションを再起動

これでステータスラインに現在の PR 番号とレビュー状態が表示されます。

### JSON 入力スキーマ（v2.1.145+）

Claude Code v2.1.145+ は以下の形式の JSON をステータスラインの stdin に渡します：

```json
{
  "pr": {
    "number": 1023,
    "url": "https://github.com/modu-ai/moai-adk/pull/1023",
    "review_state": "pending"
  },
  "workspace": {
    "repo": {
      "host": "github.com",
      "owner": "modu-ai",
      "name": "moai-adk"
    }
  }
}
```

- **pr.number**：PR 番号（必須）
- **pr.url**：PR URL（オプション）
- **pr.review_state**：レビュー状態（オプション、デフォルト：空）
- **workspace.repo.host**：Git ホスト（github.com）
- **workspace.repo.owner**：リポジトリオーナー
- **workspace.repo.name**：リポジトリ名

### 関連情報

- **SPEC リファレンス**：[SPEC-V3R5-STATUSLINE-V2145-001](/ja/advanced/statusline#リファレンス)
- **最小バージョン**：Claude Code v2.1.145 以上が必要
- **オプション機能**：デフォルトは `false` なので、明示的に有効化が必要
- **後方互換性**：以前のバージョンの Claude Code は PR 情報を提供しません（セグメントは表示されません）

## トラブルシューティング：ステータスラインが消える問題

### 症状

- ステータスラインが間欠的に表示されない
- Claude Code UI でステータスラインエリアが空白
- `.moai/cache/statusline_debug.log` ファイルが増え続ける

### 原因分析（v2.1.145 M1 修正前）

ステータスラインレンダラーは Claude Code の**300ms デバウンス契約**を遵守する必要があります。これに違反すると、進行中の実行がキャンセルされます。

以前のコードの問題点：

```bash
# 問題：DEBUG_STATUSLINE のデフォルトが 1（常に有効）
DEBUG_STATUSLINE=${DEBUG_STATUSLINE:-1}

# このため、毎回のレンダリングで：
# 1. python3 -m json.tool プロセスのフォーク（50-250ms）
# 2. ~/.moai/cache/statusline_debug.log への書き込み（~10ms）
# 合計：60-260ms → 300ms デバウンス境界線を超える
# → Claude Code が進行中のステータスラインレンダリングをキャンセル
# → 結果：ステータスラインが表示されない
```

### 解決策（v2.1.145 M1 で修正）

v3.5.0 以降では、`DEBUG_STATUSLINE` のデフォルト値が**0**になっています：

```bash
# 修正：デフォルト 0（無効）
DEBUG_STATUSLINE=${DEBUG_STATUSLINE:-0}

# デバッグが必要な場合は明示的に有効化：
export DEBUG_STATUSLINE=1
```

### パディング調整

以前は `echo ""` を使用してステータスラインの周囲の余白を調整していました。これはもはや推奨されていません。

**代わりに** `.claude/settings.json` で設定してください：

```json
{
  "statusLine": {
    "padding": 1
  }
}
```

- `padding: 0`：余白なし
- `padding: 1`：上下 1 行の余白（デフォルト）
- `padding: 2`：上下 2 行の余白

### チェックリスト

ステータスラインの表示問題を解決する手順：

1. ✓ `DEBUG_STATUSLINE` 環境変数を確認
   ```bash
   echo $DEBUG_STATUSLINE  # デフォルトは unset または 0 であるべき
   ```

2. ✓ `.moai/status_line.sh` ファイルを確認
   ```bash
   grep "DEBUG_STATUSLINE=" ~/.moai/status_line.sh
   # 結果：DEBUG_STATUSLINE=${DEBUG_STATUSLINE:-0} であるべき
   ```

3. ✓ デバッグを明示的に有効化（必要な場合のみ）
   ```bash
   export DEBUG_STATUSLINE=1
   # デバッグ情報が記録されるようになります
   ```

4. ✓ パディングを設定
   ```json
   {
     "statusLine": {
       "padding": 1
     }
   }
   ```

5. ✓ Claude Code セッションを再起動

## リファレンス

### 公式ドキュメント

- [Claude Code ステータスライン公式ドキュメント](https://code.claude.com/docs/en/statusline) — Claude Code のステータスライン契約と JSON スキーマ

### moai-adk-go 内部

- **パッケージ**：`internal/statusline/`
  - `types.go`：StdinData、PRInfo、RepoInfo 構造体定義
  - `builder.go`：セグメント作成ロジック
  - `renderer.go`：色分けと最終レンダリング

- **テンプレート**：`.moai/status_line.sh.tmpl`
  - レンダラー呼び出しと実行ロジック

- **設定**：`.moai/config/sections/statusline.yaml`
  - セグメントの有効化 / 無効化設定

### 関連 SPEC

- **[SPEC-V3R5-STATUSLINE-V2145-001](https://github.com/modu-ai/moai-adk/blob/main/.moai/specs/SPEC-V3R5-STATUSLINE-V2145-001/spec.md)**
  - M1：ステータスラインが消える問題を修正
  - M2：v2.1.145 PR セグメントを追加
  - M3：ドキュメント化（このページ）
