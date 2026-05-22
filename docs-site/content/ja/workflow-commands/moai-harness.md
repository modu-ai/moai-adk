---
title: /moai harness
weight: 55
draft: false
---

V3R4 Self-Evolving Harness 学習サブシステムを運用するコマンドです。4 段階の進化ラダー（observer → heuristic → rule → frozen-zone）と 5 層の安全パイプライン（frozen-guard → canary → contradiction → rate-limit → human oversight）を案内します。

{{< callout type="info" >}}
**スラッシュコマンド**: Claude Code で `/moai harness` を入力するとこのコマンドを直接実行できます。
{{< /callout >}}

## 概要

`/moai harness` は MoAI-ADK の自己進化学習サブシステムを安全に運用するための 4 つの verb（`status`、`apply`、`rollback`、`disable`）を提供します。PostToolUse フックがすべてのツール呼び出しを `.moai/harness/usage-log.jsonl` に append-only で記録し、提案は 4-tier 進化ラダーに沿って分類されます。Tier-4 のフローズンゾーン変更は必ず `AskUserQuestion` を通じてユーザー承認を得てから適用されます。

主要概念:

- **Observer**: PostToolUse フックがすべてのツール使用を `.moai/harness/usage-log.jsonl` に追記します。
- **4-Tier Evolution Ladder**: observation → heuristic → rule → frozen-zone 提案の 4 段階。
- **5-Layer Safety Pipeline**: すべての進化提案は 5 層の安全検証を通過する必要があります。
- **CLI Retirement**: V3R4 以降、すべての verb は workflow body のファイルシステム操作で実行され、Go バイナリのサブコマンドは呼び出されません。

## コマンド形式

```bash
/moai harness {status | apply | rollback <YYYY-MM-DD> | disable}
```

- 引数が空の場合はヘルプが表示されます。
- すべての verb は orchestrator のメインコンテキストで実行されます。

## verbs の詳細

### status

現在の harness 学習状態、保留中の Tier-4 提案、7 日間の rate-limit ウィンドウ使用量を表示します。

- **読み取り専用**: ファイルを変更しません。
- **出力内容**:
  - `.moai/config/sections/harness.yaml` の `learning.enabled` 値
  - `.moai/harness/proposals/` 内の保留 Tier-4 提案件数
  - `.moai/harness/learning-history/applied/` 配下、過去 7 日間の適用件数
  - 最近の tier 昇格イベント（`tier-promotions.jsonl`）
  - Frozen Guard 違反ログ（`frozen-guard-violations.jsonl`）

### apply

最も古い保留中の Tier-4 提案を 5-Layer Safety パイプラインへ送り、適用します。適用前には必ず orchestrator が `AskUserQuestion` ラウンドを実行し、ユーザーの明示的な承認を必要とします。

- **前提条件**:
  - 7 日間ウィンドウ内の適用件数が 1 件未満（REQ-HRN-FND-012 rate-limit floor）。
  - 提案ペイロードの整合性検証を通過。
- **ユーザー選択肢（推奨 / Modify / Defer / Reject）**: 最初の選択肢に `(推奨)` 表記。Apply を選ぶと事前スナップショットが `.moai/harness/learning-history/snapshots/<ISO-DATE>/` に保存されます。

### rollback `<YYYY-MM-DD>`

指定された日付のスナップショットを用いて直前の適用を取り消します。他の進化が累積している場合は競合レポートを出力し、再度ユーザー承認を求めます。

- **引数**: ISO-8601 日付（YYYY-MM-DD）。形式違反はエラー。
- **効果**: `.moai/harness/learning-history/applied/<DATE>.json` が `rolled-back/` に移され、対象ファイルがスナップショット状態に戻ります。

### disable

harness 学習を一時停止します（`learning.enabled: false`）。PostToolUse による観察は継続しますが、4-tier 分類器と提案生成器は無効になります。

- **使用場面**: 進化提案が疑わしい、または外部監査を実施するとき。
- **再有効化**: `.moai/config/sections/harness.yaml` で `learning.enabled: true` に戻します。

## 4-Tier Evolution Ladder

| Tier | 分類 | 自動適用 | 備考 |
|------|------|----------|------|
| Tier-1 | Observation | n/a（手動レビュー） | パッシブなログ蓄積のみ |
| Tier-2 | Heuristic | 提案のみ | orchestrator がユーザーへ推奨 |
| Tier-3 | Rule | 非 frozen 領域のみ自動適用可 | canary 通過必須 |
| Tier-4 | Frozen-zone | **ユーザー承認必須** | 5-Layer Safety を完走 |

Frozen ゾーンは `.claude/rules/moai/design/constitution.md` §2 と `.claude/rules/moai/core/zone-registry.md` で定義されます。

## 5-Layer Safety Pipeline

1. **L1 Frozen Guard**: Frozen ゾーンへの変更試行を遮断。
2. **L2 Canary**: 隔離サンドボックスで変更影響をシミュレーション。
3. **L3 Contradiction**: 他の有効規則との競合を検出。
4. **L4 Rate Limit**: 7 日間ウィンドウ内で最大 1 回の適用（REQ-HRN-FND-012）。
5. **L5 Human Oversight**: orchestrator 主導の `AskUserQuestion` 承認ラウンド。

5 層のいずれかが拒否すれば `apply` は中断され、提案は `pending` のまま保持されます。

## 使用例

```bash
# 1) 現在の状態を確認
/moai harness status

# 2) 保留中の Tier-4 提案を確認して適用
/moai harness apply

# 3) 直前の適用を昨日のスナップショットで取り消し
/moai harness rollback 2026-05-21

# 4) 学習を一時停止
/moai harness disable
```

## 関連資料

- [`.claude/skills/moai/workflows/harness.md`](https://github.com/modu-ai/moai-adk) — workflow body SSOT
- [`SPEC-V3R4-HARNESS-001`](https://github.com/modu-ai/moai-adk) — V3R4 foundation SPEC（3 つの V3R3 harness SPEC を統合）
- [`/moai plan`](/ja/workflow-commands/moai-plan) — SPEC ドキュメント作成
- [`/moai run`](/ja/workflow-commands/moai-run) — DDD/TDD 実装
- [`/moai sync`](/ja/workflow-commands/moai-sync) — ドキュメント同期 + PR
