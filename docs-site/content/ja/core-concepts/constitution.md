---
title: Constitution システム
weight: 35
draft: false
---

MoAI-ADK の不変ルール (FROZEN) と進化可能なルール (Evolvable) を管理する、憲法的制約システムです。

## 概要

[ハーネスエンジニアリング](/ja/core-concepts/harness-engineering)で見たように、MoAI-ADK のハーネスはループが蓄積した観察をもとに自ら指針を進化させます。では、その進化を何が統制するのでしょうか? 答えが **Constitution (憲法)** システムです。

Constitution は、AI エージェントが勝手に変更できない不変制約 (FROZEN Zone) と、学習を通じて改善できる進化可能制約 (Evolvable Zone) を区別します。評価基準と安全ルールを進化ループの **外** に置くこと — これが自己進化ハーネスが暴走しない理由であり、ハーネスエンジニアリングの中核となる安全メカニズムです。

## FROZEN vs Evolvable

### FROZEN Zone (不変)

AI エージェントが絶対に修正できないルールです。人間の開発者だけが変更できます。

**代表項目**:

| 項目 | 説明 | ソース |
|------|------|------|
| TRUST 5 | 5つの品質基準 | moai-constitution.md |
| SPEC + EARS | 仕様書フォーマット | spec-workflow.md |
| AskUserQuestion 独占 | ユーザーへの質問チャネル | agent-common-protocol.md |
| 評価ディメンション 4つ | Functionality/Security/Craft/Consistency | harness/scorer.go |
| ルーブリックアンカー 4段階 | 0.25/0.50/0.75/1.00 | harness/rubric.go |
| 合格しきい値の下限 | 最低 0.60 (引き下げ不可) | design-constitution.md |
| デザインパイプラインの順序 | manager-spec が最初、sync-auditor が最後 | design-constitution.md |

### Evolvable Zone (進化可能)

学習(lessons)と研究(research)を通じて改善提案が可能なルールです。

**代表項目**:

| 項目 | 説明 |
|------|------|
| スキル本文の内容 | moai-domain-* スキルの詳細内容 |
| パイプラインの重み | design.yaml の phase_weights |
| 反復回数の上限 | design.yaml の iteration_limits |
| エージェント行動ルール | Surface Assumptions、Enforce Simplicity など |

## Zone Registry

すべての HARD 条項を列挙する **単一の信頼できる情報源**(Single Source of Truth)です。

### ID 割り当てルール

```
CONST-V3R2-NNN (3桁以上の zero-padding)

001-050: 既存の HARD 条項
051-099: design constitution ミラーエントリ
100-149: design overflow (自動拡張)
150+: 新規追加
```

### Canary Gate

FROZEN 条項は `canary_gate: true` を持ちます。変更前に canary 検証が必須です。

```yaml
# Zone Registry エントリの例
- id: CONST-V3R2-154
  zone: Frozen
  file: internal/harness/scorer.go
  anchor: "#dimension-enum"
  clause: "Dimension enum FROZEN at 4 values"
  canary_gate: true
```

## 安全アーキテクチャ (5層)

Constitution システムは5層の安全アーキテクチャで保護されます。ハーネスがどれだけ学習を積み重ねても、変更は以下の5つの関門を順に通過しなければなりません:

### Layer 1: Frozen Guard

書き込み操作の前に、対象ファイルが FROZEN zone でないことを確認します。違反時は書き込みブロック + ロギング + ユーザー通知。

### Layer 2: Canary Check

提案された変更をメモリ上で適用し、直近3プロジェクトを再評価します。スコアの低下が 0.10 を超える場合、変更を拒否します。

### Layer 3: Contradiction Detector

新しい学習が既存ルールと衝突する場合、両方をユーザーに提示します。自動上書きは絶対に発生しません。

### Layer 4: Rate Limiter

進化の速度を制限します:

| パラメータ | 既定値 | 説明 |
|-----------|--------|------|
| `max_evolution_rate_per_week` | 3 | 週あたりの最大進化回数 |
| `cooldown_hours` | 24 | 進化間の最小待機時間 |
| `max_active_learnings` | 50 | アクティブな学習項目の最大数 |

### Layer 5: Human Oversight

`require_approval: true` の場合、すべての進化提案にユーザー承認が必要です。

## CLI での活用

```bash
# registry 全体を照会
moai constitution list

# Frozen zone フィルタ
moai constitution list --zone frozen

# 特定ファイルの条項のみ照会
moai constitution list --file internal/harness/scorer.go

# JSON 形式で出力
moai constitution list --format json
```

## 関連ドキュメント

- [TRUST 5 品質](/ja/core-concepts/trust-5) — 5つの品質基準
- [ハーネスエンジニアリング](/ja/core-concepts/harness-engineering) — ハーネス概念の概要
- [SPEC ベース開発](/ja/core-concepts/spec-based-dev) — SPEC ワークフロー
