---
title: "モデルポリシー"
weight: 30
draft: false
---

## モデルポリシーとは?

MoAI-ADKは8つの保持エージェント（7 MoAIカスタム + AnthropicビルトインExplore）それぞれに最適なAIモデルを割り当てます。Claude Codeサブスクリプションプランに合わせて品質を最大化しながらレート制限エラーを防ぎます。

## 3段階ポリシー概要

| ポリシー | プラン | 🟣 Opus | 🔵 Sonnet | 🟡 Haiku | 適した用途 |
|----------|--------|---------|-----------|----------|-----------|
| **High** | Max $200/月 | 5 | 1 | 1 | 最高品質、最大スループット |
| **Medium** | Max $100/月 | 2 | 3 | 2 | 品質とコストのバランス |
| **Low** | Plus $20/月 | 0 | 4 | 3 | 低予算、Opusなし |

> **なぜ重要?** Plus $20プランはOpusにアクセスできません。`Low`ポリシーを設定すると、すべてのエージェントがSonnetとHaikuのみを使用し、レート制限エラーを防ぎます。上位プランはコアエージェント（セキュリティ、戦略、アーキテクチャ）にOpusを割り当て、日常タスクにはSonnet/Haikuを使用します。

## エージェント別モデル割り当て表

### Managerエージェント (4個)

| エージェント | High | Medium | Low |
|--------------|------|--------|-----|
| manager-spec | 🟣 opus | 🟣 opus | 🔵 sonnet |
| manager-develop | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| manager-docs | 🔵 sonnet | 🟡 haiku | 🟡 haiku |
| manager-git | 🟡 haiku | 🟡 haiku | 🟡 haiku |

### 評価 & Builderエージェント (3個)

| エージェント | High | Medium | Low |
|--------------|------|--------|-----|
| plan-auditor | 🟣 opus | 🟣 opus | 🔵 sonnet |
| sync-auditor | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| builder-harness | 🟣 opus | 🔵 sonnet | 🟡 haiku |

> Teamモードの役割（researcher, analyst, architect, implementer, tester, designer, reviewer）は静的エージェントではなく、`workflow.yaml`のロールプロファイルから`Agent(general-purpose)`経由で動的に生成されます。

## 割り当て原則

- **常にOpus**: 計画監査（plan-auditor）、SPEC作成（manager-spec） — 高い推論能力が必要
- **常にHaiku**: ドキュメント（manager-docs）、Git（manager-git） — 軽量で高速なタスク
- **プランにより変動**: 実装（manager-develop, cycle_type=tdd/ddd） — 上位プランほどOpus

## 設定方法

### プロジェクト初期化時

```bash
moai init my-project
# 対話型ウィザードにモデルポリシー選択が含まれます
```

### 既存プロジェクトの再設定

```bash
moai update
# 対話型プロンプト:
# - Reset model policy? (y/n) — モデルポリシーをリセット
# - Update GLM settings? (y/n) — GLM環境変数を設定
```

> デフォルトポリシーは`High`です。GLM設定は`settings.local.json`に隔離され、Gitにコミットされません。

## 次のステップ

- [CGモード](/ja/multi-llm/cg-mode) — Claude + GLMハイブリッドでコスト削減
- [エージェントガイド](/ja/advanced/agent-guide) — エージェントのカスタマイズ
- [CLIリファレンス](/ja/getting-started/cli) — moai init, moai updateの詳細
