---
title: Hooksイベントリファレンス
weight: 60
draft: false
---

MoAI-ADK v2.10.1現在、Claude Codeのフックシステムは **29個のイベントタイプ**、**5種類のフックタイプ**、**イベントごとのマッチャー**、**スマートな動作** をサポートしています。

> フックの基本的な概念とセットアップ手順については、[Hooksガイド](/ja/advanced/hooks-guide) を参照してください。このページは完全なイベントリファレンスです。

## フックタイプ

**5種類のフックタイプが利用可能です：**

| タイプ | 説明 | 例 |
|--------|------|-----|
| **command** | シェルスクリプト実行 | `".claude/hooks/moai/handle-session-start.sh"` |
| **prompt** | LLM評価 | プロンプトテキストをLLMが実行して結果を返す |
| **agent** | サブエージェント検証 | エージェントがタスクを検証して結果を返す |
| **http** | Webhookエンドポイント | HTTPリクエストで遠隔エンドポイントにイベントを配信 |
| **mcp_tool** | MCPツール呼び出し | MCPサーバーツールへのリモート呼び出し |

## 完全なイベントリファレンス (29個)

### ライフサイクルイベント

| イベント | 説明 | マッチャー |
|--------|------|--------|
| `SessionStart` | セッション開始 | — |
| `SessionEnd` | セッション終了 | — |
| `PostSession` | セッション終了後に実行 (self-hosted runner ライフサイクルイベント、CC 2.1.169+)。セッションが完全に破棄された後、`SessionEnd` よりも遅く発火します。MoAI-ADK は現在このフックをワイヤリングしません。セッション後のクリーンアップ/テレメトリが必要な self-hosted デプロイ向けの利用可能なオプションとして文書化されます。 | — |
| `Stop` | エージェント停止 | — |
| `SubagentStop` | サブエージェント停止 | — |
| `SubagentStart` | サブエージェント開始 | — |
| `StopFailure` | 停止失敗 | `errorType` |
| `Setup` | 初期設定 | — |

### ツールイベント

| イベント | 説明 | マッチャー |
|---------|------|-----------|
| `PreToolUse` | ツール実行前 | `toolName` |
| `PostToolUse` | ツール実行後 | `toolName` |
| `PostToolUseFailure` | ツール実行失敗 | `toolName`, `errorType` |
| `PostToolBatch` | 並列ツール呼び出しバッチ後 (v2.1.89+) | — |

### コンテキストイベント

| イベント | 説明 | マッチャー |
|--------|------|--------|
| `PreCompact` | コンテキスト圧縮前 | — |
| `PostCompact` | コンテキスト圧縮後 | — |
| `InstructionsLoaded` | インストラクションロード完了 | — |

### 入力イベント

| イベント | 説明 | マッチャー |
|---------|------|-----------|
| `UserPromptSubmit` | ユーザープロンプト送信 | — |
| `UserPromptExpansion` | スラッシュコマンドがプロンプトに展開 (v2.1.90+) | — |
| `Elicitation` | Elicitation開始 | — |
| `ElicitationResult` | Elicitation完了 | — |

### セキュリティイベント

| イベント | 説明 | マッチャー |
|--------|------|--------|
| `PermissionRequest` | 権限リクエスト | `toolName` |
| `PermissionDenied` | 権限拒否 | `toolName` |

### チームイベント

| イベント | 説明 | マッチャー |
|--------|------|--------|
| `TeammateIdle` | チームメンバーのアイドル状態への遷移 | — |
| `TaskCompleted` | タスク完了マーク | — |
| `TaskCreated` | タスク作成 | — |

### ワークツリーイベント

| イベント | 説明 | マッチャー |
|--------|------|--------|
| `WorktreeCreate` | ワークツリー作成 | — |
| `WorktreeRemove` | ワークツリー削除 | — |

### 環境イベント

| イベント | 説明 | マッチャー |
|--------|------|--------|
| `ConfigChange` | 設定変更 | `configSource` |
| `CwdChanged` | 作業ディレクトリ変更 | — |
| `FileChanged` | ファイル変更 | — |

### UIイベント

| イベント | 説明 | マッチャー |
|--------|------|--------|
| `Notification` | ユーザー通知 | — |

## スマート動作 (Smart Behaviors)

MoAI-ADKのフックは単純なイベント処理を超えて、インテリジェントな動作を実行します：

### PermissionDenied自動リトライ

読み取り専用ツール(Read, Grep, Glob)の権限が拒否されると、フックが自動的にリトライをトリガーします。これにより、バックグラウンドエージェントで権限プロンプトが表示されない問題を緩和します。

### StopFailureエラータイプ応答

エージェント停止失敗時にエラータイプに応じた差別化された応答を提供します。長時間実行セッションでの安定性を保証します。

### PostCompactセッションメモ復元

コンテキスト圧縮後、重要なセッションメモ(進行状態、SPEC参照)を自動的に復元します。これにより、コンテキスト圧縮時の重要な情報の消失を防ぎます。

### SubagentStartコンテキスト注入

サブエージェント開始時に必要なコンテキスト(プロジェクトルール、MXタグ、進行状態)を自動的に注入します。

## マッチャー (Matchers)

マッチャーを使用すると、特定の条件でのみフックが実行されるようフィルタリングできます：

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
| `errorType` | StopFailure, PostToolUseFailure | エラータイプでフィルタ |
| `configSource` | ConfigChange | 設定ソースでフィルタ |

## CLAUDE_ENV_FILE

`CwdChanged`と`FileChanged`フックを通じて、環境変数を継続的に管理できます：

```bash
# .claude/hooks/moai/handle-cwd-changed.sh
# CLAUDE_ENV_FILEを通じて環境変数を永続化
echo "MOAI_PROJECT_DIR=$(pwd)" >> "$CLAUDE_ENV_FILE"
```

これにより、セッション間で環境変数を維持し、ディレクトリ変更時に自動的に環境を再設定できます。

## MoAI-ADKが使用する主要フック

| イベント | MoAIハンドラー | 役割 |
|--------|-----------|------|
| `SessionStart` | `handle-session-start.sh` | Statusline初期化、メトリクスセッション開始 |
| `PostToolUse` | `handle-post-tool.sh` | Taskメトリクスロギング |
| `TeammateIdle` | `handle-teammate-idle.sh` | LSP品質ゲート検証 |
| `TaskCompleted` | `handle-task-completed.sh` | SPECドキュメント存在確認 |
| `WorktreeCreate` | (なし — MoAI 既定で未登録) | Claude Code 既定の worktree 動作を使用 (`isolation: worktree` エージェント向け)。登録時は active creator 規約 (ディレクトリ作成 + 絶対パスを stdout に echo) が必須。 |
| `WorktreeRemove` | (なし — MoAI 既定で未登録) | Claude Code 既定の worktree クリーンアップを使用。登録時は observer-only 規約 (stdout 不要)。 |
| `UserPromptSubmit` | `handle-user-prompt.sh` | 品質ゲート自動実行 |

## 次のステップ

- [Hooksガイド](/ja/advanced/hooks-guide) — フックの基本概念とセットアップ手順
- [settings.jsonガイド](/ja/advanced/settings-json) — settings.json完全リファレンス
- [CLIリファレンス](/ja/getting-started/cli) — `moai hook`コマンド詳細
