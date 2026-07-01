---
title: 意思決定メモリシステム
weight: 50
draft: false
---

MoAIのユーザー選好学習と適応型推奨システムについてガイドします。

{{< callout type="info" >}}
**一行要約**: 意思決定メモリはユーザーの選択を記憶し、将来の同様の状況で個人化された推奨を提供します。
{{< /callout >}}

## システム概要

意思決定メモリ(Decision Memory)はMoAI-ADKの**長期学習レイヤー**です。AskUserQuestion ラウンドでユーザーの選択を観察し、将来の同じ意思決定ポイントで統計的多数選択に基づいた適応型推奨を提供します。

### 核心原則

| 原則 | 説明 |
|------|------|
| **観察ベース** | ユーザー選択の統計的多数を学習（ポリシーデフォルトではない） |
| **透明性** | 推奨根拠を常に明示（コールドスタート状態を含む） |
| **自律性** | ユーザーはいつでも推奨を拒否可能 |
| **適応型強度** | 熟練度に応じて推奨の強度を自動調整 |

## 5つの構成要素

### 1. 3-Tier Memory Layer（メモリレイヤー）

意思決定メモリは3つのレイヤーで構成されています。

#### L0: Immediate（即時メモリ）
- **範囲**: 現在のセッション内
- **用途**: ユーザーが選択したばかりのオプションを参照
- **永続性**: セッション終了時に消失

#### L1: Session Span（セッション範囲メモリ）
- **範囲**: 同じプロジェクトの直近3セッション
- **用途**: 最近の選好に基づく推奨
- **永続性**: `.claude/projects/{hash}/memory/` 自動メモリ

#### L2: Long-term（長期メモリ）
- **範囲**: 全セッション（無制限）
- **用途**: 統計的多数学習、長期トレンド
- **永続性**: MEMORY.md + topic ファイル（ユーザー管理）

### 2. Adaptive Recommendation Placement（適応型推奨配置）

推奨（最初のオプションの`（推奨）`ラベル）は**観察された統計的多数**に基づきます。

#### コールドスタート（初期状態）
- **観察 < N**: 十分な観察データなし
- **推奨配置**: 静的デフォルト（明示的に公開）
- **表示方式**: `based on static default, N observations needed for personalization`

#### ウォーム状態（学習中）
- **観察 = N~M**: 部分学習
- **推奨配置**: 観察された多数 + 信頼度シグナル
- **信頼度**: 観察数 × 選択一貫性

#### 成熟状態（安定化）
- **観察 > M**: 十分な学習
- **推奨配置**: 強い多数確信（統計的に有意）
- **信頼度**: 最高（≥95% 信頼度）

#### 熟練度ベースの適応型強度
- **専門家（セッション > 50）**: 弱い推奨強度（自律性優先、推測選好のみ公開）
- **初心者（セッション < 10）**: 強い推奨強度（`（推奨）`ラベル + 理由明示）
- **中級者（10 ≤ セッション ≤ 50）**: 中程度（状況に応じて調整）

### 3. PostToolUse Capture Hook（意思決定ポイント）

AskUserQuestion 応答が到着するとPostToolUse フックが自動的に意思決定をキャプチャします。

#### キャプチャされるデータ

```json
{
  "decision_id": "moai-ask-001",
  "timestamp": "2026-07-01T10:00:00Z",
  "question": "次のステップを選択してください",
  "user_choice": "オプション A（推奨）",
  "all_options": ["オプション A", "オプション B", "オプション C"],
  "context": {
    "spec_id": "SPEC-XXX-001",
    "phase": "run",
    "workflow": "/moai run"
  }
}
```

#### 保存場所

- **セッション中**: `.moai/state/decisions/`（一時 JSON）
- **セッション終了**: `~/.claude/projects/{hash}/memory/decisions.jsonl`（自動メモリ）

### 4. Decay Policy（減衰ポリシー）

古い意思決定の重みを段階的に減少させます。

#### 減衰関数

```
weight(t) = initial_weight × exp(-decay_rate × days_ago)
```

#### デフォルト値
- **初期重み**: 1.0
- **減衰率**: 0.1（7日ごとに約50% 減衰）
- **保持期間**: 90日（その後は自動アーカイブ）

#### 例示

```
昨日の選択: weight = 0.95
7日前の選択: weight = 0.50
30日前の選択: weight = 0.04
90日以上: アーカイブ（推奨反映除外）
```

### 5. Recovery Controls（復旧制御）

意思決定メモリのエラー復旧と再設定を管理します。

#### メモリ初期化

ユーザーが学習された選好をリセットできます:

```bash
/moai memory reset
```

#### 選好編集

特定の意思決定カテゴリの推奨を修正:

```bash
/moai memory set <category> <preferred-option>
```

#### 選好照会

現在学習された選好を確認:

```bash
/moai memory list
```

## 意思決定カテゴリ

メモリが追跡する主要な意思決定タイプ:

| カテゴリ | 例示 |
|----------|------|
| **Tier Selection** | Tier S/M/L 選択 |
| **Cycle Type** | DDD vs TDD モード |
| **Worktree Strategy** | Main vs Branch vs Worktree |
| **PR Routing** | Direct-to-main vs PR ベース |
| **Team Mode** | Solo vs Agent Teams |
| **Model Selection** | タスクごとのモデル選択 |
| **Effort Level** | Effort レベル（low/medium/high/xhigh） |

## 統計的多数学習の例示

### シナリオ 1: Tier Selection

ユーザーが10回のTier選択を実施した場合:

```
Tier S: 3回選択
Tier M: 6回選択  ← 統計的多数（60%）
Tier L: 1回選択

学習結果: Tier M が（推奨）と表示
信頼度: 中程度（6/10 = 60%, N=10）
推奨表示: "Tier M（推奨）— 最近の選択 60% ベース"
```

### シナリオ 2: Cycle Type

```
DDD: 4回
TDD: 5回選択  ← 統計的多数
その他: 1回

学習結果: TDD が（推奨）
信頼度: 中程度（5/10 = 50%, N=10）
推奨表示: "TDD（推奨）— 観察ベース"
```

## コールドスタート透明性

観察不足時に明示的に公開:

```
選択肢 1: Tier M（推奨）— based on static default, 5 observations needed for personalization
選択肢 2: Tier L
選択肢 3: Tier S
```

ユーザーは学習中の状態を明確に認識します。

## 熟練度ベースの強度調整の例

### 初心者ユーザー（セッション < 10）
```
Tier M（推奨）— 最近の選択ベースで提示
（強い推奨強度）
```

### 専門家ユーザー（セッション > 50）
```
選択肢:
- Tier M（最近の選択 60%）
- Tier L
- Tier S
（弱い推奨強度、推測選好のみ公開）
```

## 関連ドキュメント

- [AskUserQuestion プロトコル](/advanced/agent-guide) - 推奨配置ルール（HARD）
- [ワークフロー選択](/advanced/harness-v4-builder) - Tier 選択と意思決定
- [メモリシステム](/getting-started/memory) - ユーザー選好管理

{{< callout type="info" >}}
**ヒント**: 意思決定メモリは自動的に動作します。明示的な設定は不要です。ユーザーが意思決定するたびに自動的に学習します。
{{< /callout >}}
