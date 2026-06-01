---
title: MCPサーバー活用ガイド
weight: 80
draft: false
---

Claude CodeのMCP (Model Context Protocol) サーバーの活用方法を詳しく解説します。

{{< callout type="info" >}}
**要約**: MCPはClaude Codeに**外部ツールを接続するUSBポート**です。Context7で最新ドキュメントを参照し、Adaptive Thinking（`--ultrathink`キーワード経由）で複雑な問題を分析します。
{{< /callout >}}

## MCPとは？

MCP (Model Context Protocol) はClaude Codeに**外部ツールとサービスを接続**する標準プロトコルです。

Claude Codeはデフォルトでファイル読み書き、ターミナルコマンドなどのツールを持っています。MCPを通じてこのツールセットを拡張し、ライブラリドキュメント参照、ナレッジグラフ保存、段階的推論などの機能を追加できます。

```mermaid
flowchart TD
    CC["Claude Code"] --> MCP_LAYER["MCPプロトコル層"]

    MCP_LAYER --> C7["Context7<br/>ライブラリドキュメント参照"]
    MCP_LAYER --> CHROME["Claude in Chrome<br/>ブラウザ自動化"]

    C7 --> C7_OUT["最新React、FastAPI<br/>公式ドキュメント参照"]
    CHROME --> CHROME_OUT["ウェブページ<br/>自動テスト"]
```

## MoAIで使用するMCPサーバー

### MCPサーバーリスト

| MCPサーバー | 用途 | ツール | 有効化 |
|----------|------|------|--------|
| **Context7** | ライブラリドキュメントリアルタイム参照 | `resolve-library-id`, `get-library-docs` | `.mcp.json` |
| **Claude in Chrome** | ブラウザ自動化 | `navigate`, `screenshot` 等 | `.mcp.json` |

## Context7活用法

Context7は**ライブラリ公式ドキュメントをリアルタイムで参照**するMCPサーバーです。

### 必要性

Claude Codeの学習データは特定時点までの情報のみを含みます。Context7を使用すると**最新バージョンの公式ドキュメント**をリアルタイムで参照して正確なコードを生成できます。

| 状況 | Context7なし | Context7使用時 |
|------|---------------|--------------|
| React 19新機能 | 学習データにない可能性あり | 最新公式ドキュメント参照 |
| Next.js 16設定 | 以前バージョンパターン使用可能性あり | 現行バージョンパターン適用 |
| FastAPI最新API | 古いバージョン構文使用可能性あり | 最新構文適用 |

### 使用方法

Context7は2段階で動作します。

**段階1: ライブラリID照会**

```bash
# Claude Codeが内部的に呼び出し
> Reactの最新ドキュメントを参照してコードを書いて

# Context7が実行する作業
# mcp__context7__resolve-library-id("react")
# → ライブラリID: /facebook/react
```

**段階2: ドキュメント検索**

```bash
# 特定トピックのドキュメント検索
# mcp__context7__get-library-docs("/facebook/react", "useEffect cleanup")
# → React公式ドキュメントでuseEffectクリーンアップ関数関連内容を返却
```

### 実戦活用シナリオ

```bash
# シナリオ: Next.js 16 App Router設定
> Next.js 16でプロジェクト設定をして

# Claude Code内部動作:
# 1. Context7でNext.js最新ドキュメント照会
# 2. App Router設定パターン確認
# 3. 最新設定ファイル作成
# 4. 公式推奨事項反映
```

### 対応ライブラリ例

| カテゴリ | ライブラリ |
|----------|-----------|
| フロントエンド | React, Next.js, Vue, Svelte, Angular |
| バックエンド | FastAPI, Django, Express, NestJS, Spring |
| データベース | PostgreSQL, MongoDB, Redis, Prisma |
| テスト | pytest, Jest, Vitest, Playwright |
| インフラ | Docker, Kubernetes, Terraform |
| その他 | TypeScript, Tailwind CSS, shadcn/ui |

## Adaptive Thinkingを用いた深い推論

`--ultrathink`キーワードを使用すると、Opus 4.7+（4.8を含む）およびSonnet 4.6の組み込み推論機能**Adaptive Thinking**が有効化されます。このモードは**タスク複雑度に基づいて推論トークンを動的に割り当て**ます。

従来のモデルが固定的な`budget_tokens`パラメータを使用していたのに対し、新しいモデルのAdaptive Thinkingは知的にスケーリングします。推論深度は固定予算ではなく、**effort**パラメータ（`xhigh`, `high`, `medium`, `low`）で制御します。

### `--ultrathink`いつ使うか

`--ultrathink`キーワードを使用すると、複雑な問題向けの強化された分析モードが有効になります。

```bash
# UltraThinkでアーキテクチャ分析
> 認証システムアーキテクチャを設計して --ultrathink

# Opus 4.7+/4.8またはSonnet 4.6で:
# 1. タスク複雑度に基づいて推論トークン動的割り当て
# 2. 複数の角度から問題分解を探索
# 3. トレードオフを体系的に評価
# 4. 検証済みの推論で最適解を導出
```

### 有効化される状況

Adaptive Thinkingは以下の状況で活用されます。

| 状況 | 例 |
|------|------|
| 複雑な問題分解 | "マイクロサービスアーキテクチャを設計して" |
| 3ファイル以上に影響 | "認証システム全体をリファクタリングして" |
| 技術選択比較 | "JWT vs セッション認証、どちらが良い？" |
| トレードオフ分析 | "パフォーマンスを上げつつ保守性も維持するには？" |
| 互換性破壊検討 | "このAPI変更が既存クライアントに与える影響は？" |

### モデル互換性

- **Opus 4.8, Opus 4.7, Sonnet 4.6**: Adaptive Thinking（動的割り当て推論）
- **Haiku 4.5**: 拡張推論未対応（`--ultrathink`キーワード有効化はno-op）
- **初期モデル**: 現在のClaudeモデルにアップグレードすると深い推論対応

## MCP設定方法

### .mcp.json設定

MCPサーバーはプロジェクトルートの`.mcp.json`ファイルで設定します。

```json
{
  "context7": {
    "command": "npx",
    "args": ["-y", "@anthropic/context7-mcp-server"]
  },
  "claude-in-chrome": {
    "command": "npx",
    "args": ["-y", "@anthropic/claude-in-chrome-mcp-server"]
  }
}
```

### settings.local.jsonで有効化

特定MCPサーバーを個人的に有効化するには`settings.local.json`に追加します。

```json
{
  "enabledMcpjsonServers": [
    "context7"
  ]
}
```

### settings.jsonで権限許可

MCPツールを使用するには`permissions.allow`に登録する必要があります。

```json
{
  "permissions": {
    "allow": [
      "mcp__context7__resolve-library-id",
      "mcp__context7__get-library-docs",
      "mcp__claude_in_chrome__*"
    ]
  }
}
```

## 実戦例

### ReactプロジェクトでContext7で最新ドキュメント参照

```bash
# 1. ユーザーがReact 19の新機能を使用したいとリクエスト
> React 19のuse()フックを使ってデータフェッチングを実装して

# 2. Claude Code内部動作
# a) Context7でReactライブラリID照会
#    → resolve-library-id("react") → "/facebook/react"
#
# b) React 19 use()関連ドキュメント検索
#    → get-library-docs("/facebook/react", "use hook data fetching")
#
# c) 最新公式ドキュメント基づきでコード生成
#    → use()フックの正しい使用法適用
#    → Suspenseバウンダリーと共に使用
#    → エラーバウンデリー処理包含

# 3. 結果: 最新パターンが反映された正確なコード生成
```

### 複雑なアーキテクチャ決定にUltraThink使用

```bash
# アーキテクチャ決定が必要な状況
> 自サービスの認証をJWTにするかセッションにするか分析して --ultrathink

# Adaptive Thinkingが動的に割り当てられた推論で:
# 1. 問題をサブ問題に分解
# 2. 各サブ問題を段階的に分析
# 3. 以前の結論を再検討し修正
# 4. 最適なソリューション導出
```

## 関連ドキュメント

- [settings.jsonガイド](/advanced/settings-json) - MCPサーバー権限設定
- [スキルガイド](/advanced/skill-guide) - スキルとMCPツールの関係
- [エージェントガイド](/advanced/agent-guide) - エージェントのMCPツール活用
- [CLAUDE.mdガイド](/advanced/claude-md-guide) - MCP関連設定参照
- [Google Stitchガイド](/advanced/stitch-guide) - AIベースUI/UXデザインツール詳細活用法

{{< callout type="info" >}}
**ヒント**: Context7は最新ライブラリドキュメントを参照する時に最も有用です。新フレームワーク導入時や最新バージョンへのアップグレード時にContext7を有効化すると正確なコードを得られます。
{{< /callout >}}
