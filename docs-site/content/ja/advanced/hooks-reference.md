---
title: Hooks イベントリファレンス
weight: 60
draft: false
---

Claude Code のフックシステムは **30 個のイベントタイプ**、**5 種類のフックタイプ**、**イベント別マッチャー**、**スマート動作** をサポートします。フックはエージェンティックハーネスで唯一「必ず実行される」ことが保証される決定論的 (deterministic) な制御ポイントです — プロンプトは無視されうるが、フックは無視されません。

> フックの基本概念と設定方法は [Hooks ガイド](/ja/advanced/hooks-guide) を参照してください。このページはイベントの全リファレンスです。

## フックタイプ

利用可能なフックタイプは 5 種類です。

| タイプ | 説明 | 例 |
|------|------|------|
| **command** | シェルスクリプト実行 | `".claude/hooks/moai/handle-session-start.sh"` |
| **prompt** | LLM 評価 | プロンプトテキストを LLM が実行して結果を返す |
| **agent** | サブエージェント検証 | エージェントが作業を検証して結果を返す |
| **http** | Webhook エンドポイント | HTTP POST リクエストでイベントを伝達 |
| **mcp_tool** | MCP ツール実行 | MCP サーバーのツールをリモート呼び出し |

## イベント全リファレンス (30 個)

### ライフサイクルイベント

| イベント | 説明 | マッチャー |
|--------|------|------|
| `SessionStart` | セッション開始 | — |
| `SessionEnd` | セッション終了 | — |
| `Stop` | エージェント停止 | — |
| `SubagentStop` | サブエージェント停止 | — |
| `SubagentStart` | サブエージェント開始 | — |
| `StopFailure` | 停止失敗 | `errorType` |
| `Setup` | 初期設定 | — |

### ツールイベント

| イベント | 説明 | マッチャー |
|--------|------|------|
| `PreToolUse` | ツール実行前 | `toolName` |
| `PostToolUse` | ツール実行後 | `toolName` |
| `PostToolUseFailure` | ツール実行失敗 | `toolName`, `errorType` |
| `PostToolBatch` | 並列ツールバッチ実行後 (v2.1.89+) | — |

### コンテキストイベント

| イベント | 説明 | マッチャー |
|--------|------|------|
| `PreCompact` | コンテキスト圧縮前 | — |
| `PostCompact` | コンテキスト圧縮後 | — |
| `InstructionsLoaded` | インストラクションロード完了 | — |

### 入力イベント

| イベント | 説明 | マッチャー |
|--------|------|------|
| `UserPromptSubmit` | ユーザープロンプト送信 | — |
| `UserPromptExpansion` | スラッシュコマンドプロンプト展開 (v2.1.90+) | — |
| `Elicitation` | Elicitation 開始 | — |
| `ElicitationResult` | Elicitation 完了 | — |

### セキュリティイベント

| イベント | 説明 | マッチャー |
|--------|------|------|
| `PermissionRequest` | 権限リクエスト | `toolName` |
| `PermissionDenied` | 権限拒否 | `toolName` |

### チームイベント

| イベント | 説明 | マッチャー |
|--------|------|------|
| `TeammateIdle` | チームメイトのアイドル状態への移行 | — |
| `TaskCompleted` | タスク完了マーク | — |
| `TaskCreated` | タスク生成 | — |

### ワークツリーイベント

| イベント | 説明 | マッチャー |
|--------|------|------|
| `WorktreeCreate` | ワークツリー生成 | — |
| `WorktreeRemove` | ワークツリー削除 | — |

### 環境イベント

| イベント | 説明 | マッチャー |
|--------|------|------|
| `ConfigChange` | 設定変更 | `configSource` |
| `CwdChanged` | 作業ディレクトリ変更 | — |
| `FileChanged` | ファイル変更 | — |

### UI イベント

| イベント | 説明 | マッチャー |
|--------|------|------|
| `Notification` | ユーザー通知 | — |
| `MessageDisplay` | アシスタントメッセージテキスト表示中 (ストリーミング中の発話) | — |

## スマート動作 (Smart Behaviors)

MoAI-ADK フックは単純なイベント処理を超えて知的な動作を行います。

### PermissionDenied 自動リトライ

読み取り専用ツール (Read, Grep, Glob) の権限が拒否されると、フックが自動的にリトライをトリガーします。これはバックグラウンドエージェントで権限プロンプトが表示されない問題を緩和します。

### StopFailure エラータイプ応答

エージェント停止失敗時にエラータイプに応じて差別化された応答を提供します。長時間実行セッションでの安定性を保証します。

### PostCompact セッションメモ復元

コンテキスト圧縮後に重要なセッションメモ (進行状態、SPEC 参照) を自動的に復元します。コンテキスト圧縮はトークンを節約する代わりに情報を失う取引ですが、このフックがその損失から核心情報を守ります。

### SubagentStart コンテキスト注入

サブエージェント開始時に必要なコンテキスト (プロジェクトルール、MX タグ、進行状態) を自動注入します。

## マッチャー (Matchers)

マッチャーを使うと特定の条件でのみフックが実行されるようフィルタリングできます。すべてのイベントにフックを掛けるとその分だけ実行コストが増えるので、マッチャーで範囲を絞るのが基本です。

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": { "toolName": "Bash" },
      "hooks": [{
        "type": "command",
        "command": "echo 'Bash tool detected'",
        "timeout": 5
      }]
    }]
  }
}
```

### 利用可能なマッチャーフィールド

| マッチャーフィールド | 適用イベント | 説明 |
|----------|-----------|------|
| `toolName` | PreToolUse, PostToolUse, PostToolUseFailure, PermissionRequest, PermissionDenied | ツール名でフィルタ |
| `errorType` | StopFailure, PostToolUseFailure | エラー種別でフィルタ |
| `configSource` | ConfigChange | 設定ソースでフィルタ |

## CLAUDE_ENV_FILE

`CwdChanged` と `FileChanged` フックを通じて環境変数を継続的に管理できます。

```bash
# .claude/hooks/moai/handle-cwd-changed.sh
# CLAUDE_ENV_FILE を通じて環境変数を永続化
echo "MOAI_PROJECT_DIR=$(pwd)" >> "$CLAUDE_ENV_FILE"
```

これによりセッション間で環境変数を維持し、ディレクトリ変更時に自動的に環境を再設定できます。

## MoAI-ADK が使う主要なフック

| イベント | MoAI ハンドラー | 役割 |
|--------|-----------|------|
| `SessionStart` | `handle-session-start.sh` | Statusline 初期化、メトリクスセッション開始 |
| `PostToolUse` | `handle-post-tool.sh` | Task メトリクスロギング |
| `TeammateIdle` | `handle-teammate-idle.sh` | LSP 品質ゲート検証 |
| `TaskCompleted` | `handle-task-completed.sh` | SPEC ドキュメント存在確認 |
| `WorktreeCreate` | (なし — MoAI はデフォルトで非登録) | Claude Code のデフォルト worktree 動作を使用 (`isolation: worktree` agent 用)。登録時は active creator コントラクト (ディレクトリ生成 + path stdout echo) が義務。 |
| `WorktreeRemove` | (なし — MoAI はデフォルトで非登録) | Claude Code のデフォルト worktree クリーンアップ動作を使用。登録時は observer-only コントラクト (出力不要)。 |
| `UserPromptSubmit` | `handle-user-prompt.sh` | 品質ゲート自動実行 |

## 次のステップ

- [Hooks ガイド](/ja/advanced/hooks-guide) — フックの基本概念と設定方法
- [settings.json ガイド](/ja/advanced/settings-json) — settings.json 全リファレンス
- [CLI リファレンス](/ja/getting-started/cli) — `moai hook` コマンド詳細
