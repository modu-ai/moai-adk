---
title: plan_type ティアプロファイル
weight: 4
draft: false
---

MoAI-ADKは同じワークフローでもAPI従量課金とサブスクリプションでは最適配分が異なることを認識しています。`plan_type`軸は課金モデルごとにTier × Phaseモデル/effortマトリクスを分離適用します。このページはSPEC-MODEL-TIER-PLANTYPE-001 (CLOSED)で実装された60セルプロファイルマトリクスを公式ドキュメント化します。

## plan_type軸

`plan_type`は2つの値を持ちます:

- `api` — 従量課金。ドルが唯一の制約です。課題あたりコスト最適化が目標です。
- `subscription` — サブスクリプションプラン。週次トークンクォータ + Opus加重差引が制約です。クォータあたり解決課題数の最大化が目標です。

サブスクリプションプランではOpus時間は別途加重差引されます(Max 5x: Sonnet 140-280h vs Opus 15-35h、約1/8)。したがってサブスクリプションではOpusを推論にのみ配し、実行は豊富なSonnet時間で回すopusplan構造が最適です。

## plan_type設定

```bash
moai init . --plan-type api           # 初期化時に設定
moai update --plan-type subscription  # 事後切替
```

`llm.yaml`の`llm.plan_type`フィールドで現在値を確認できます。

## 60セルプロファイルマトリクス

10エージェント × 3ティア × 2 plan_type = 60セル。下の表はSPEC-MODEL-TIER-PLANTYPE-001のApplyTierProfile実装です。

### Plan A — API従量課金 (rev2)

APIではドルが唯一の制約です。rev2修正: 単価はSonnetがOpusの半分ですが、課題あたりコストはOpus $13.22 < Sonnet $26.40に逆転します。したがってAPIでは実行もOpusを使用します。推論 = Fable(品質1位)、実行 = Opus(課題あたりコスト1位)、機械 = Sonnet low。

| エージェント (役割) | A-max (品質) | A-medium (推奨) | A-low (コスト) |
|---|---|---|---|
| manager-spec (推論) | fable / high | fable / high | opus / high |
| plan-auditor (推論) | fable / high | fable / high | opus / high |
| sync-auditor (推論) | fable / high | opus / high | opus / medium |
| manager-design (推論) | fable / high | fable / high | opus / high |
| super-advisor (最高推論) | fable / xhigh | fable / high | opus / high |
| manager-develop (実行) | fable / high | opus / high | opus / medium |
| builder-harness (実行) | opus / high | opus / medium | opus / medium |
| manager-docs (機械) | sonnet / medium | sonnet / low | sonnet / low |
| manager-git (機械) | sonnet / low | sonnet / low | sonnet / low |
| Explore (探索) | inherit / medium | inherit / low | inherit / low |

### Plan B — サブスクリプション (可用性最優先)

サブスクリプションの制約はドルではなく週次トークンクォータ + Opus加重差引です。目標は「クォータあたり解決課題数」の最大化 = 再試行ループの排除 + Opusは推論にのみ配分。Anthropic公式opusplanパターン(「計画はOpus、実行はSonnet」)の精密版です。

| エージェント (役割) | B-max (推奨) | B-medium | B-low (Pro) |
|---|---|---|---|
| manager-spec (推論) | opus / high | opus / high | opus / medium |
| plan-auditor (推論) | opus / high | opus / medium | sonnet / high |
| sync-auditor (推論) | opus / high | opus / medium | sonnet / high |
| manager-design (推論) | opus / high | opus / medium | sonnet / high |
| super-advisor (最高推論) | opus / xhigh | opus / high | opus / medium |
| manager-develop (実行) | sonnet / high | sonnet / high | sonnet / high |
| builder-harness (実行) | sonnet / high | sonnet / medium | sonnet / medium |
| manager-docs (機械) | sonnet / low | sonnet / low | sonnet / low |
| manager-git (機械) | sonnet / low | sonnet / low | sonnet / low |
| Explore (探索) | inherit / medium | inherit / low | inherit / low |

## ApplyTierProfileメカニズム

`ApplyTierProfile`はエージェントfrontmatterの`model`と`effort`を両方置換します(replace-both)。全エージェントに`effort:`フィールドがあるため「保存」モードは無効で、常にreplace-bothで動作します。

このメカニズムはSPEC-MODEL-TIER-PLANTYPE-001(実行フェーズ完了、CLOSED)で実装されました。上の表の全セルはライブ動作として検証済みです。

## GLMバックエンドeffortオーバーレイ

{{< icon warning warn >}} **正直性注記 (REQ-DA-060)**: GLMバックエンドeffortオーバーレイのwire有効性はライブGLMセッションのアウトバウンド観測が必要な実証課題です。

GLMバックエンド(`moai glm` / `moai cg` GLMパネル)ではClaudeの5段effort(max / xhigh / high / medium / low)をGLMの3段reasoning_effort(high / max)にcollapseして適用します。実装内容:

- `IsGLMBackend`検出でGLMセッションを識別
- 5段 → 3段collapseマッピング(max/xhigh → max, high → high, medium/low → GLM未サポート)
- coding作業時max override

**実装 + 配線完了、wire有効性実証予定** — z.aiのAnthropic-compat shimが`ANTHROPIC_REASONING_EFFORT`環境変数値を実際に消費するかは、ライブGLMセッションのアウトバウンド観測が必要な実行フェーズ実証課題です。このページには「動作保証」とは記載せず、「実装 + 配線完了、wire有効性実証予定」と記載します。

## モデルポリシーボード (moai web)

`moai web`の`/model-policy`ボードでplan_typeとティアプロファイルを視覚的に確認・設定できます。このボードはSPEC-WEB-CONSOLE-013の承認された例外としてplan_type書き込みを許可します。

## ロードマップ

{{< icon clock >}} **spawn-time 36セルルーティング** (SPEC-MODEL-TIER-ROUTING-PROFILES-001) — 現在ApplyTierProfileはエージェント単位でルーティングします。spawn-timeにphaseとSPEC Tierを組み合わせた36セル精密ルーティングはdescopedされた後続SPECです。現在はエージェントfrontmatterのmodel/effortがApplyTierProfileによりreplace-bothされる構造で運用されます。

## 次のステップ

- [3層エージェントアーキテクチャ](/ja/advanced/no-haiku-3tier/) — DeepSWEリーダーボード根拠と3層定義
- [トークノミクス概論](/ja/advanced/tokenomics-overview/) — 4層トークノミクス構造のLayer Bルーティング
