---
title: モデルポリシー
weight: 30
draft: false
---

## モデルポリシーとは?

モデルポリシーは MoAI-ADK トークノミクスの骨格です。「すべてのタスクに最高のモデル」
ではなく、エージェントごとに — 計画・監査のように推論が重い仕事と、ドキュメント化・Git のように
軽い仕事ごとに — 適切なモデルを宣言的に割り当てます。Claude Code のサブスクリプションプランに
合わせて品質を最大化しつつ、レート制限エラーを防ぎます。

MoAI-ADK v3.0 のエージェントカタログは **10 個** (MoAI カスタム 9 個 + Anthropic
内蔵 `Explore`) であり、以下の割り当て表はそのうちモデルポリシーが直接割り当てる中核 7
エージェントを扱います。

## 3 段階ポリシーの概要

| ポリシー | プラン | Opus | Sonnet | Haiku | 適した用途 |
|------|------|---------|-----------|----------|-----------|
| **High** | Max $200/月 | 5 | 1 | 1 | 最高品質、最大スループット |
| **Medium** | Max $100/月 | 2 | 3 | 2 | 品質とコストのバランス |
| **Low** | Plus $20/月 | 0 | 4 | 3 | 低予算、Opus 非対応 |

> **なぜ重要ですか?** Plus $20 プランは Opus にアクセスできません。`Low` ポリシーを設定すると、すべてのエージェントが Sonnet と Haiku のみを使用し、レート制限エラーを防ぎます。上位プランは中核エージェント (計画、監査) に Opus を割り当て、日常タスクには Sonnet/Haiku を使用します。

## エージェント別モデル割り当て表

### Manager Agents (4 個)

| エージェント | High | Medium | Low |
|---------|------|--------|-----|
| manager-spec | opus | opus | sonnet |
| manager-develop | opus | sonnet | sonnet |
| manager-docs | sonnet | haiku | haiku |
| manager-git | haiku | haiku | haiku |

### Evaluator & Builder Agents (3 個)

| エージェント | High | Medium | Low |
|---------|------|--------|-----|
| plan-auditor | opus | opus | sonnet |
| sync-auditor | opus | sonnet | sonnet |
| builder-harness | opus | sonnet | haiku |

> Anthropic 内蔵の `Explore` は読み取り専用の探索エージェントで、個別の割り当てなしに
> 動作します。Agent Teams 静的階層 (静的 role profile) は v3.0 で
> 引退し、並列作業は sub-agent 並列実行と動的ワークフローが
> 代替します。`moai cg` の teammate ランタイム (tmux pane) はそのまま維持されます。

## 割り当ての原則

- **常に Opus**: 計画監査 (plan-auditor)、SPEC 作成 (manager-spec) — 高い推論能力が必要
- **常に Haiku**: Git (manager-git) — 軽くて速いタスク
- **プランによって変動**: 実装 (manager-develop, cycle_type=tdd/ddd) — プランが上位ほど Opus

計画を作ったエージェントが監査しないよう、plan-auditor と sync-auditor は独立した
割り当てを維持します — コスト軸と品質軸 (バイアス防止) が一緒に設計された表です。

## v3.0 拡張: Tier×Phase 宣言軸

v3.0 では、エージェント単位の割り当ての上に **作業フェーズ (phase) と SPEC サイズ (Tier)**
の軸が追加されました。`internal/config/model_routing.go` が Tier×Phase →
{model, effort} マトリクスを宣言的に管理します:

- **model**: inherit / sonnet / opus / glm / fable
- **effort** (推論の深さ): low / medium / high / xhigh / max
- **tier** (SPEC サイズ): S / M / L
- **phase** (作業フェーズ): plan / run / sync / mx

同じワークフローでも API 従量課金とサブスクリプション料金では最適配分が異なるため、
料金プラン認識 (plan_type) プロファイルが料金プラン別のマトリクスを分離適用します。

## 設定方法

### プロジェクト初期化時

```bash
moai init my-project
# 対話型ウィザードにモデルポリシー選択を含む
```

### 既存プロジェクトの再設定

```bash
moai update
# 対話型プロンプト:
# - Reset model policy? (y/n) — モデルポリシーの再設定
# - Update GLM settings? (y/n) — GLM 環境変数の設定
```

> デフォルトポリシーは `High` です。GLM 設定は `settings.local.json` に分離され、Git にコミットされません。

## 次のステップ

- [CG モード](/ja/multi-llm/cg-mode) — Claude + GLM ハイブリッドでコスト削減
- [エージェントガイド](/ja/advanced/agent-guide) — エージェントのカスタマイズ
- [CLI リファレンス](/ja/getting-started/cli) — moai init、moai update の詳細
