---
title: 意思決定メモリシステム
weight: 50
draft: false
---

エージェンティック・ループ・エンジニアリングの出発点は観測です — ループが回るたびに観測が蓄積され、蓄積された観測が学習の原料になります。意思決定メモリは、その観測対象をコードではなく **ユーザーの選択** に拡張した階層です。

{{< callout type="info" >}}
**ひと言要約**: 意思決定メモリはユーザーの選択を記憶し、今後の類似の状況でパーソナライズされた推奨を提供します。
{{< /callout >}}

## システム概要

意思決定メモリ(Decision Memory)は MoAI-ADK の **長期学習レイヤー** です。AskUserQuestion ラウンドでユーザーの選択を観測し、今後同じ意思決定ポイントで統計的多数の選択に基づいた適応型の推奨を提供します。

重要なのは方向性です。システムが押したいデフォルトを `(推奨)` で包装するのではなく、**ユーザーが実際に繰り返し選んできたもの** が推奨になります。

### コア原則

| 原則 | 説明 |
|------|------|
| **観測ベース** | ユーザー選択の統計的多数を学習 (ポリシーデフォルトではない) |
| **透明性** | 推奨の根拠を常に明示 (cold-start 状態を含む) |
| **自律性** | ユーザーは推奨をいつでも拒否可能 |
| **適応型強度** | 熟練度に応じて推奨の強度を自動調整 |

## 5 つの構成要素

### 1. 3-Tier Memory Layer (メモリ階層)

意思決定メモリは 3 つの階層で構成されます。下へ行くほど長く残ります。

#### L0: Immediate (即時メモリ)
- **範囲**: 現在のセッション内
- **用途**: たった今ユーザーが選択したオプションの参照
- **持続性**: セッション終了時に消失

#### L1: Session Span (セッション範囲メモリ)
- **範囲**: 同じプロジェクトの直近 3 セッション
- **用途**: 直近の好みに基づく推奨
- **持続性**: `.claude/projects/{hash}/memory/` の自動メモリ

#### L2: Long-term (長期メモリ)
- **範囲**: すべてのセッション (無制限)
- **用途**: 統計的多数の学習、長期トレンド
- **持続性**: MEMORY.md + topic ファイル (ユーザー管理)

### 2. Adaptive Recommendation Placement (適応型推奨配置)

推奨(先頭オプションの `(推奨)` ラベル)は **観測された統計的多数** に基づきます。観測量に応じて 3 つの状態を行き来します。

#### Cold-Start (初期状態)
- **観測 < N**: 十分な観測データの不在
- **推奨配置**: 静的デフォルト (明示的に公開)
- **表示方式**: `based on static default, N observations needed for personalization`

#### Warm State (学習中)
- **観測 = N〜M**: 部分学習
- **推奨配置**: 観測された多数 + 信頼度シグナル
- **信頼度**: 観測数 × 選択の一貫性

#### Mature State (安定化)
- **観測 > M**: 十分な学習
- **推奨配置**: 強い多数の確信 (統計的に有意)
- **信頼度**: 最高 (≥95% 信頼度)

#### 熟練度ベースの適応型強度

同じ推奨でも相手によって強度が変わります。専門家への強い推奨は自律性を侵害し、初心者への弱い推奨は決定疲労を増やすだけだからです。

- **専門家 (セッション > 50)**: 弱い推奨強度 (自律性優先、inferred preference の公開のみ)
- **初心者 (セッション < 10)**: 強い推奨強度 (`(推奨)` ラベル + 理由の明示)
- **中級者 (10 ≤ セッション ≤ 50)**: 中間強度 (状況に応じて調整)

### 3. PostToolUse Capture Hook (意思決定の捕捉)

AskUserQuestion の応答が届くと PostToolUse フックが自動的に意思決定を捕捉します。ユーザーが別途記録する必要はありません。

#### 捕捉されるデータ

```json
{
  "decision_id": "moai-ask-001",
  "timestamp": "2026-07-01T10:00:00Z",
  "question": "次のステップを選択してください",
  "user_choice": "Option A (推奨)",
  "all_options": ["Option A", "Option B", "Option C"],
  "context": {
    "spec_id": "SPEC-XXX-001",
    "phase": "run",
    "workflow": "/moai run"
  }
}
```

#### 保存場所

- **セッション中**: `.moai/state/decisions/` (一時 JSON)
- **セッション終了時**: `~/.claude/projects/{hash}/memory/decisions.jsonl` (自動メモリ)

### 4. Decay Policy (減衰ポリシー)

3 か月前の選択が今日の好みを代弁するとは限りません。古い意思決定の重みは徐々に減少します。

#### 減衰関数

```
weight(t) = initial_weight × exp(-decay_rate × days_ago)
```

#### デフォルト値
- **Initial weight**: 1.0
- **Decay rate**: 0.1 (7 日ごとに約 50% 減衰)
- **Retention period**: 90 日 (以降は自動アーカイブ)

#### 例

```
昨日の選択: weight = 0.95
7 日前の選択: weight = 0.50
30 日前の選択: weight = 0.04
90 日以上: アーカイブ (推奨への反映から除外)
```

### 5. Recovery Controls (復旧制御)

学習が誤った方向に固まったときのために、エラー復旧とリセットの手段が提供されます。

#### メモリのリセット

ユーザーは学習された好みをリセットできます。

```bash
/moai memory reset
```

#### 好みの編集

特定の意思決定カテゴリの推奨を修正します。

```bash
/moai memory set <category> <preferred-option>
```

#### 好みの照会

現在学習されている好みを確認します。

```bash
/moai memory list
```

## 意思決定カテゴリ

メモリが追跡する主な意思決定タイプです。

| カテゴリ | 例 |
|----------|------|
| **Tier Selection** | Tier S/M/L の選択 |
| **Cycle Type** | DDD vs TDD モード |
| **Worktree Strategy** | Main vs Branch vs Worktree |
| **PR Routing** | Direct-to-main vs PR-based |
| **Team Mode** | Solo vs Agent Teams |
| **Model Selection** | タスクごとのモデル選択 |
| **Effort Level** | Effort レベル (low/medium/high/xhigh) |

Model Selection と Effort Level がここに含まれる点は注目に値します — 意思決定メモリが学習した好みは、結局モデル・推論深度の割り当てにつながるため、このシステムはトークノミクスのパーソナライズ層でもあります。

## 統計的多数学習の例

### シナリオ 1: Tier Selection

ユーザーが 10 回の Tier 選択をした場合:

```
Tier S: 3 回選択
Tier M: 6 回選択  ← 統計的多数 (60%)
Tier L: 1 回選択

学習結果: Tier M が (推奨) として表示
信頼度: 中上 (6/10 = 60%, N=10)
推奨文言: "Tier M (推奨) — 直近の選択 60% に基づく"
```

### シナリオ 2: Cycle Type

```
DDD: 4 回
TDD: 5 回選択  ← 統計的多数
その他: 1 回

学習結果: TDD が (推奨)
信頼度: 中 (5/10 = 50%, N=10)
推奨文言: "TDD (推奨) — 観測ベース"
```

## Cold-Start の透明性

観測が不足しているときは、その事実を隠さず明示的に公開します。

```
選択肢 1: Tier M (推奨) — based on static default, 5 observations needed for personalization
選択肢 2: Tier L
選択肢 3: Tier S
```

ユーザーはまだ学習中の状態であることを明確に認識できます。

## 熟練度ベースの強度調整の例

### 初心者ユーザー (セッション < 10)
```
Tier M (推奨) — 直近の選択に基づいて提示
(強い推奨強度)
```

### 専門家ユーザー (セッション > 50)
```
選択肢:
- Tier M (直近の選択 60%)
- Tier L
- Tier S
(弱い推奨強度、inferred preference の公開のみ)
```

## 関連ドキュメント

- [エージェントガイド](/ja/advanced/agent-guide) - AskUserQuestion 推奨配置ルール (HARD)
- [Harness v4 Builder 深掘りガイド](/ja/advanced/harness-v4-builder) - Tier 選択と意思決定
- [メモリシステム](/ja/getting-started/memory) - ユーザー好みの管理

{{< callout type="info" >}}
**ヒント**: 意思決定メモリは自動的に動作します。明示的な設定は不要です — 意思決定を下すたびにシステムが静かに学習します。
{{< /callout >}}
