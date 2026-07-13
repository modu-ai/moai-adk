---
title: Hooks イベントリファレンス
weight: 60
draft: false
---

Claude Code のフックシステムは **29 のイベントタイプ**、**5 つのフックタイプ**、**イベント別マッチャー**、**スマート動作** をサポートします。フックはエージェンティック・ハーネスの中で唯一「必ず実行される」ことが保証される決定的(deterministic)な制御ポイントです — プロンプトは無視されうるが、フックは無視されません。

> フックの基本概念と設定方法は [Hooks ガイド](/ja/advanced/hooks-guide)を参照してください。このページはイベント全体のリファレンスです。

## フックタイプ

利用可能なフックタイプは 5 つです。

| タイプ | 説明 | 例 |
|------|------|------|
| **command** | シェルスクリプト実行 | `".claude/hooks/moai/handle-session-start.sh"` |
| **prompt** | LLM 評価 | プロンプトテキストを LLM が実行して結果を返す |
| **agent** | サブエージェント検証 | エージェントが作業を検証して結果を返す |
| **http** | Webhook エンドポイント | HTTP POST リクエストでイベントを送信 |
| **mcp_tool** | MCP ツール実行 | MCP サーバーツールをリモート呼び出し |

## イベント全体リファレンス (29 個)

### ライフサイクルイベント

| イベント | 説明 | マッチャー |
|--------|------|------|
| `SessionStart` | セッション開始 | — |
| `SessionEnd` | セッション終了 | — |
| `PostSession` | セッション終了後に実行 (self-hosted runner ライフサイクルイベント、CC 2.1.169+)。セッションが完全に解放された後、`SessionEnd` より遅く発火します。MoAI-ADK は現在このフックを wiring していません。セッション後のクリーンアップ/テレメトリが必要な self-hosted デプロイ向けの利用可能なオプションとして文書化されています。 | — |
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
| `InstructionsLoaded` | インストラクションのロード完了 | — |

### 入力イベント

| イベント | 説明 | マッチャー |
|--------|------|------|
| `UserPromptSubmit` | ユーザープロンプト送信 | — |
| `UserPromptExpansion` | スラッシュコマンドのプロンプト展開 (v2.1.90+) | — |
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
| `TeammateIdle` | チームメンバーのアイドル状態遷移 | — |
| `TaskCompleted` | タスク完了マーク | — |
| `TaskCreated` | タスク作成 | — |

### ワークツリーイベント

| イベント | 説明 | マッチャー |
|--------|------|------|
| `WorktreeCreate` | ワークツリー作成 | — |
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

## スマート動作 (Smart Behaviors)

MoAI-ADK のフックは単純なイベント処理を超えて、知的な動作を行います。

### PermissionDenied 自動リトライ

読み取り専用ツール(Read、Grep、Glob)の権限が拒否されると、フックが自動的にリトライをトリガーします。これはバックグラウンドエージェントで権限プロンプトが表示されない問題を緩和します。

### StopFailure エラータイプ応答

エージェント停止失敗時、エラータイプに応じて差別化された応答を提供します。長時間実行セッションでの安定性を保証します。

### PostCompact セッションメモ復元

コンテキスト圧縮後、重要なセッションメモ(進行状態、SPEC 参照)を自動的に復元します。コンテキスト圧縮はトークンを節約する代わりに情報を失う取引ですが、このフックがその損失から核心情報を守ります。

### SubagentStart コンテキスト注入

サブエージェント開始時、必要なコンテキスト(プロジェクトルール、MX タグ、進行状態)を自動注入します。

## マッチャー (Matchers)

マッチャーを使うと、特定の条件でのみフックが実行されるようフィルタリングできます。すべてのイベントにフックを掛ければその分だけ実行コストが増えるため、マッチャーで範囲を絞るのが基本です。

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

## MoAI-ADK が使用する主なフック

| イベント | MoAI ハンドラー | 役割 |
|--------|-----------|------|
| `SessionStart` | `handle-session-start.sh` | Statusline 初期化、メトリクスセッション開始 |
| `PostToolUse` | `handle-post-tool.sh` | Task メトリクスのロギング |
| `TeammateIdle` | `handle-teammate-idle.sh` | LSP 品質ゲート検証 |
| `TaskCompleted` | `handle-task-completed.sh` | SPEC ドキュメント存在確認 |
| `WorktreeCreate` | (なし — MoAI はデフォルト未登録) | Claude Code デフォルトの worktree 動作を使用 (`isolation: worktree` agent 用)。登録時は active creator コントラクト (ディレクトリ作成 + path stdout echo) が義務。 |
| `WorktreeRemove` | (なし — MoAI はデフォルト未登録) | Claude Code デフォルトの worktree クリーンアップ動作を使用。登録時は observer-only コントラクト (出力不要)。 |
| `UserPromptSubmit` | `handle-user-prompt.sh` | 品質ゲートの自動実行 |

## 次のステップ

- [Hooks ガイド](/ja/advanced/hooks-guide) — フックの基本概念と設定方法
- [settings.json ガイド](/ja/advanced/settings-json) — settings.json 全体リファレンス
- [CLI リファレンス](/ja/getting-started/cli) — `moai hook` コマンド詳細
