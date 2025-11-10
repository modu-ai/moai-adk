# Hooksシステムリファレンス

Claude CodeのHookシステムを通じた自動ガードレールおよびコンテキスト管理です。

## 概要

**Hooks**は特定のイベント発生時に自動的に実行されるスクリプトです。MoAI-ADKは4つの主要Hookを提供します。

### Hook種類

| Hook                 | タイミング           | 用途               | タイムアウト |
| -------------------- | -------------------- | ------------------ | ------------ |
| **SessionStart**     | セッション開始       | プロジェクト状態確認 | 5秒      |
| **PreToolUse**       | ツール実行前         | 危険なコマンドブロック | 5秒      |
| **UserPromptSubmit** | ユーザー入力後       | 入力検証           | 5秒      |
| **PostToolUse**      | ツール実行後         | 結果分析           | 5秒      |

## Hook場所

```
.claude/
├── hooks/
│   ├── session_start.sh       # SessionStart Hook
│   ├── pre_tool_use.sh        # PreToolUse Hook
│   ├── post_tool_use.sh       # PostToolUse Hook
│   └── user_prompt_submit.sh  # UserPromptSubmit Hook
├── settings.json              # Hook設定
└── permissions.json           # 権限設定
```

## Hook実行フロー

```
┌─────────────────────────────────────┐
│  Claude Codeセッション開始            │
└────────────────┬────────────────────┘
                 │
            ┌────▼────────────┐
            │ SessionStart    │（プロジェクト状態確認）
            └────┬────────────┘
                 │
      ┌──────────▼──────────┐
      │ ユーザーコマンド入力  │
      └──────────┬──────────┘
                 │
            ┌────▼────────────┐
            │PreToolUse       │（実行前検証）
            └────┬────────────┘
                 │
          ┌──────▼──────────┐
          │ ツール実行      │
          └──────┬──────────┘
                 │
            ┌────▼────────────┐
            │PostToolUse      │（結果分析）
            └────┬────────────┘
                 │
      ┌──────────▼──────────┐
      │  ユーザーに結果配信  │
      └─────────────────────┘
```

## Hook設定

### .claude/settings.json

```json
{
  "hooks": {
    "enabled": true,
    "timeout": 5000,
    "session_start": ".claude/hooks/session_start.sh",
    "pre_tool_use": ".claude/hooks/pre_tool_use.sh",
    "post_tool_use": ".claude/hooks/post_tool_use.sh",
    "user_prompt_submit": ".claude/hooks/user_prompt_submit.sh"
  }
}
```

## 🆘 Hookエラー処理

### Hook実行失敗

```
:x: Hook失敗
│
├─ タイムアウト（5秒超過）
│  └─→ ツール実行ブロック
│
├─ スクリプトエラー
│  └─→ エラーログ保存
│
└─ 権限エラー
   └─→ 権限調整必要
```

### デバッグ

```bash
# Hookログ確認
cat ~/.claude/projects/*/hook-logs/*.log

# Hook手動実行
bash .claude/hooks/session_start.sh

# Hook無効化
# settings.jsonの"hooks.enabled" → false
```

## <span class="material-icons">library_books</span> 詳細ガイド

- **[SessionStart Hook](session.md)** - セッション開始時自動実行
- **[Tool Hooks](tool.md)** - ツール実行前/後処理

______________________________________________________________________

**次**: [SessionStart Hook](session.md)または[Tool Hooks](tool.md)



